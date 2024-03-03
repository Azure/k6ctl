package task

import (
	"context"
	"fmt"
	"os/exec"

	"github.com/Azure/k6ctl/internal/config"
	configplugin "github.com/Azure/k6ctl/internal/config/plugin"
)

// resolvePluginBinaryPathFromName attempts to resolve the binary path from a plugin name.
// It looks for k6ctl-<baseName> from $PATH. The second return value is false if no binary is found.
func resolvePluginBinaryPathFromName(baseName string) (string, bool) {
	const prefix = "k6ctl-"
	binaryPath, err := exec.LookPath(prefix + baseName)
	if err != nil {
		return "", false
	}
	return binaryPath, true
}

func LoadConfigPlugins(
	ctx context.Context,
	reg config.ProviderRegistry,
	k6 K6,
) (func(), error) {
	if len(k6.ConfigPlugins) < 1 {
		return func() {}, nil
	}

	var settingsList []configplugin.ClientBinarySettings
	for _, plugin := range k6.ConfigPlugins {
		// TODO: validation settings
		binaryPath := plugin.BinaryPath
		if binaryPath == "" {
			if p, ok := resolvePluginBinaryPathFromName(plugin.Namespace); ok {
				binaryPath = p
			} else {
				return func() {}, fmt.Errorf("plugin binary not found: %s", plugin.Namespace)
			}
		}

		settings := configplugin.ClientBinarySettings{
			Namespace: plugin.Namespace,
			Path:      binaryPath,
			Args:      plugin.Args[:],
		}

		settingsList = append(settingsList, settings)
	}

	return configplugin.RegisterFromClientBinaries(ctx, reg, settingsList)
}
