package core

import (
	"github.com/Azure/k6ctl/internal/config"
	"github.com/Azure/k6ctl/internal/task"
)

// RegisterProviders registers the core config providers.
func RegisterProviders(
	registry config.ProviderRegistry,
	configProviders []task.ConfigProvider,
	userParameterInputs map[string]string,
) error {
	// "parameter" config provider
	resolvedParameters, err := resolveParameters(configProviders, userParameterInputs)
	if err != nil {
		return err
	}
	registry.Register(resolvedParameters.CreateProvider())

	return nil
}
