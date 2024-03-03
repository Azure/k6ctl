package k6ctl

import (
	"context"

	"github.com/Azure/k6ctl/internal/config"
	configplugin "github.com/Azure/k6ctl/internal/config/plugin"
)

// ConfigProvider - config provider.
type ConfigProvider = config.Provider

// ConfigProviderRegistry - registry of config providers.
type ConfigProviderRegistry = config.ProviderRegistry

// NewConfigProviderRegistry creates a new config provider registry instance.
var NewConfigProviderRegistry = config.NewRegistry

// LoadConfigForStruct loads the given user input map into the given struct and validates it.
func LoadConfigForStruct[T any](
	ctx context.Context,
	target Target,
	userInput map[string]any,
) (T, error) {
	return config.LoadForStruct[T](ctx, target, userInput)
}

// ProvideConfig creates a config provider using the loader and resolver functions.
func ProvideConfig[T any](
	name string,
	// FIXME: generic alias
	loader func(ctx context.Context, target Target, userInput map[string]any) (T, error),
	resolver func(ctx context.Context, target Target, params T) (string, error),
) ConfigProvider {
	return config.Provide[T](
		name,
		loader,
		resolver,
	)
}

// ServeConfigRegistryPlugin serves the given config provider registry as a plugin.
var ServeConfigRegistryPlugin = configplugin.ServeRegistry
