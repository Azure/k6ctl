package config

type registry struct {
	providers map[string]Provider
}

var _ ProviderRegistry = (*registry)(nil)

func NewRegistry() ProviderRegistry {
	return &registry{
		providers: make(map[string]Provider),
	}
}

func (r *registry) Register(provider Provider) ProviderRegistry {
	r.providers[provider.Name()] = provider

	return r
}

func (r *registry) GetByName(name string) (Provider, bool) {
	provider, ok := r.providers[name]
	return provider, ok
}

func (r *registry) GetNames() []string {
	var names []string
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}
