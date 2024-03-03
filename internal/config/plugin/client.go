package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"

	"github.com/hashicorp/go-plugin"

	"github.com/Azure/k6ctl/internal/config"
	"github.com/Azure/k6ctl/internal/target"
)

func remoteNamespacedConfigProvider(
	namespace string,
	name string,
	impl Interface,
) config.Provider {
	return config.Provide[ResolveRequest](
		fmt.Sprintf("%s/%s", namespace, name),
		func(
			ctx context.Context,
			target target.Target,
			userInput map[string]any,
		) (ResolveRequest, error) {
			kubeconfig, _ := target.GetKubeconfig()
			rv := ResolveRequest{
				Name:             name,
				TargetKubeconfig: kubeconfig,
				UserInput:        userInput,
			}

			deadline, hasDeadline := ctx.Deadline()
			if hasDeadline {
				rv.ContextDeadlineInUnixNano = deadline.UnixNano()
			}

			return rv, nil
		},
		func(ctx context.Context, target target.Target, params ResolveRequest) (string, error) {
			return impl.Resolve(params)
		},
	)
}

// ClientBinarySettings defines a namespaced plugin binary.
// The binary is expected to be used as a remote plugin.
type ClientBinarySettings struct {
	// Namespace - the namespace of the plugin
	Namespace string
	// Path - the absolute path to the plugin binary
	Path string
	// Args - optional arguments to pass to the plugin binary
	Args []string
}

func (s ClientBinarySettings) validate() error {
	if s.Namespace == "" {
		return fmt.Errorf("namespace is required")
	}
	if s.Path == "" {
		return fmt.Errorf("path is required")
	}
	return nil
}

func registerFromClientBinary(
	ctx context.Context,
	reg config.ProviderRegistry,
	settings ClientBinarySettings,
) (func(), error) {
	var stop func()

	errOut := func(err error) (func(), error) {
		if stop != nil {
			stop()
		}
		return func() {}, err
	}

	if err := settings.validate(); err != nil {
		return errOut(err)
	}

	pluginClient := plugin.NewClient(
		&plugin.ClientConfig{
			Stderr:          os.Stderr,
			HandshakeConfig: handshakeConfig,
			Plugins: map[string]plugin.Plugin{
				pluginName: &Plugin{},
			},
			Cmd:              exec.CommandContext(ctx, settings.Path),
			AllowedProtocols: []plugin.Protocol{plugin.ProtocolNetRPC},
		},
	)

	cc, err := pluginClient.Client()
	if err != nil {
		return errOut(err)
	}

	stop = func() {
		pluginClient.Kill()
	}

	raw, err := cc.Dispense(pluginName)
	if err != nil {
		return errOut(err)
	}

	p := raw.(Interface)
	supportedProviderNames, err := p.GetNames()
	if err != nil {
		return errOut(err)
	}

	for _, name := range supportedProviderNames {
		reg.Register(remoteNamespacedConfigProvider(settings.Namespace, name, p))
	}

	return stop, nil
}

// RegisterFromClientBinaries registers the given client binaries as remote plugins.
func RegisterFromClientBinaries(
	ctx context.Context,
	reg config.ProviderRegistry,
	settingsList []ClientBinarySettings,
) (func(), error) {
	var stopFuncs []func()

	for _, settings := range settingsList {
		stop, err := registerFromClientBinary(ctx, reg, settings)
		if err != nil {
			for _, stop := range stopFuncs {
				stop()
			}
			return func() {}, err
		}
		stopFuncs = append(stopFuncs, stop)
	}

	stop := func() {
		for _, stop := range stopFuncs {
			stop()
		}
	}

	return stop, nil
}
