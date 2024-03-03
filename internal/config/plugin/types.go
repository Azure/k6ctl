package plugin

import (
	"context"
	"net/rpc"
	"time"

	"github.com/hashicorp/go-plugin"

	"github.com/Azure/k6ctl/internal/target"
)

const pluginName = "k6ctl"

var handshakeConfig = plugin.HandshakeConfig{
	ProtocolVersion:  1,
	MagicCookieKey:   "k6ctl_plugin",
	MagicCookieValue: "ltck6",
}

type ResolveRequest struct {
	Name                      string
	ContextDeadlineInUnixNano int64
	UserInput                 map[string]any
	TargetKubeconfig          string
}

func (rr ResolveRequest) Context() (context.Context, context.CancelFunc) {
	if rr.ContextDeadlineInUnixNano <= 0 {
		return context.WithCancel(context.Background())
	}

	deadline := time.Unix(0, rr.ContextDeadlineInUnixNano)
	return context.WithDeadline(context.Background(), deadline)
}

func (rr ResolveRequest) Target() target.Target {
	return &target.StaticTarget{
		Kubeconfig: rr.TargetKubeconfig,
	}
}

// Interface defines the config plugin interface.
type Interface interface {
	// GetNames returns the names of the available config providers
	GetNames() ([]string, error)

	// Resolve resolves a config from a config provider.
	Resolve(req ResolveRequest) (string, error)
}

type Plugin struct {
	Impl Interface
}

func (p *Plugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &rpcServer{Impl: p.Impl}, nil
}

func (p *Plugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &rpcClient{client: c}, nil
}
