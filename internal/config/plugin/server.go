package plugin

import (
	"fmt"

	"github.com/hashicorp/go-plugin"

	"github.com/Azure/k6ctl/internal/config"
)

type registryServer struct {
	registry config.ProviderRegistry
}

var _ Interface = (*registryServer)(nil)

func (c *registryServer) GetNames() ([]string, error) {
	return c.registry.GetNames(), nil
}

func (c *registryServer) Resolve(req ResolveRequest) (string, error) {
	provider, ok := c.registry.GetByName(req.Name)
	if !ok {
		return "", fmt.Errorf("config provider %q not found", req.Name)
	}

	ctx, cancel := req.Context()
	defer cancel()

	return provider.Resolve(ctx, req.Target(), req.UserInput)
}

// ServeRegistry serves the given registry as a plugin.
func ServeRegistry(registry config.ProviderRegistry) {
	plugin.Serve(&plugin.ServeConfig{
		HandshakeConfig: handshakeConfig,
		Plugins: map[string]plugin.Plugin{
			pluginName: &Plugin{
				Impl: &registryServer{registry: registry},
			},
		},
	})
}
