package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/k6ctl/internal/target"
)

func TestRegistry_Integration(t *testing.T) {
	r := NewRegistry()
	r.Register(
		Provide[string](
			"test-provider",
			func(
				ctx context.Context,
				target target.Target,
				userInput map[string]any,
			) (string, error) {
				return "test", nil
			},
			func(ctx context.Context, target target.Target, s string) (string, error) {
				return "test", nil
			},
		),
	)

	provider, ok := r.GetByName("test-provider")
	assert.True(t, ok)
	assert.NotNil(t, provider)

	provider, ok = r.GetByName("non-existent-provider")
	assert.False(t, ok)
	assert.Nil(t, provider)
}
