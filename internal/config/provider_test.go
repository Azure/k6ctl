package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/k6ctl/internal/target"
)

func TestProvide_Integration(t *testing.T) {
	t.Run("static value", func(t *testing.T) {
		p := Provide(
			"static-name",
			func(
				ctx context.Context,
				target target.Target,
				userInput map[string]any,
			) (string, error) {
				return "static-value", nil
			},
			func(ctx context.Context, target target.Target, params string) (string, error) {
				return params, nil
			},
		)

		assert.Equal(t, "static-name", p.Name())
		ctx := context.Background()
		fakeTarget := &target.StaticTarget{}
		params := map[string]any{}
		resolvedValue, err := p.Resolve(ctx, fakeTarget, params)
		assert.NoError(t, err)
		assert.Equal(t, "static-value", resolvedValue)
	})

	t.Run("validate params error", func(t *testing.T) {
		p := Provide(
			"static-name",
			func(
				ctx context.Context,
				target target.Target,
				userInput map[string]any,
			) (string, error) {
				return "static-value", assert.AnError
			},
			func(ctx context.Context, target target.Target, s string) (string, error) {
				panic("should not be called")
			},
		)

		ctx := context.Background()
		fakeTarget := &target.StaticTarget{}
		params := map[string]any{}
		_, err := p.Resolve(ctx, fakeTarget, params)
		assert.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})

	t.Run("resolve error", func(t *testing.T) {
		p := Provide(
			"static-name",
			func(
				ctx context.Context,
				target target.Target,
				userInput map[string]any,
			) (string, error) {
				return "static-value", nil
			},
			func(ctx context.Context, target target.Target, s string) (string, error) {
				return "", assert.AnError
			},
		)

		ctx := context.Background()
		fakeTarget := &target.StaticTarget{}
		params := map[string]any{}
		_, err := p.Resolve(ctx, fakeTarget, params)
		assert.Error(t, err)
		assert.ErrorIs(t, err, assert.AnError)
	})
}
