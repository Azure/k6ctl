package config

import (
	"context"

	"github.com/Azure/k6ctl/internal/target"
)

type configProvider[T any] struct {
	name     string
	loader   LoadAndValidateParams[T]
	resolver ResolveConfig[T]

	configInternalImpl
}

// Provide creates a config provider using the loader and resolver functions.
func Provide[T any](
	name string,
	loader LoadAndValidateParams[T],
	resolver ResolveConfig[T],
) Provider {
	return &configProvider[T]{
		name:     name,
		loader:   loader,
		resolver: resolver,
	}
}

var _ Provider = (*configProvider[any])(nil)

func (c *configProvider[T]) Name() string {
	return c.name
}

func (c *configProvider[T]) Resolve(
	ctx context.Context,
	target target.Target,
	userInput map[string]any,
) (string, error) {
	validatedParams, err := c.loader(ctx, target, userInput)
	if err != nil {
		return "", err
	}

	return c.resolver(ctx, target, validatedParams)
}
