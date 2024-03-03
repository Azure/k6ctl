package config

import (
	"context"

	"github.com/Azure/k6ctl/internal/target"
)

// LoadAndValidateParams restores and validates the parameters from the given user input map.
type LoadAndValidateParams[T any] func(ctx context.Context, target target.Target, userInput map[string]any) (T, error)

// ResolveConfig resolves the configuration for the given target and parameters.
type ResolveConfig[T any] func(ctx context.Context, target target.Target, params T) (string, error)

// Provider provides a configuration.
type Provider interface {
	// Name - name of the provider.
	Name() string
	// Resolve - resolves the configuration for the given target and user input map.
	Resolve(ctx context.Context, target target.Target, userInput map[string]any) (string, error)

	configInternal
}

// ProviderRegistry is a registry of configuration providers.
type ProviderRegistry interface {
	// Register - registers a config provider.
	Register(provider Provider) ProviderRegistry
	// GetByName - gets a provider by name.
	GetByName(name string) (Provider, bool)
	// GetNames - gets the names of all registered providers.
	GetNames() []string
}

// GetConfigProviderByName gets a provider by name.
type GetConfigProviderByName func(name string) (Provider, bool)

type configInternal interface {
	// require to provider are being created via k6ctl package
	configInternal()
}

type configInternalImpl struct{}

func (configInternalImpl) configInternal() {}
