package config

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/k6ctl/internal/target"
)

type loadForStructTestWithDefaultingError struct{}

func (l loadForStructTestWithDefaultingError) Defaulting(
	ctx context.Context,
	target target.Target,
) (loadForStructTestWithDefaultingError, error) {
	return l, assert.AnError
}

var _ paramsWithDefaulting[loadForStructTestWithDefaultingError] = loadForStructTestWithDefaultingError{}

type loadForStructTestWithValidationError struct{}

func (l loadForStructTestWithValidationError) Validate() error {
	return assert.AnError
}

var _ paramsWithValidation = loadForStructTestWithValidationError{}

type loadForStructIntegration struct {
	RequiredField string `mapstructure:"RequiredField" validate:"required"`
	OptionalField bool   `mapstructure:"OptionalField"`
}

var _ paramsWithValidation = loadForStructIntegration{}
var _ paramsWithDefaulting[loadForStructIntegration] = loadForStructIntegration{}

func (l loadForStructIntegration) Defaulting(
	ctx context.Context,
	target target.Target,
) (loadForStructIntegration, error) {
	l.OptionalField = true
	return l, nil
}

func (l loadForStructIntegration) Validate() error {
	return nil
}

func TestLoadForStruct(t *testing.T) {
	fakeContext := func() context.Context {
		return context.Background()
	}
	fakeTarget := func() target.Target {
		return nil
	}

	t.Run("basic", func(t *testing.T) {
		type T struct {
			A string `validate:"required"`
		}

		t.Run("validate failure", func(t *testing.T) {
			_, err := LoadForStruct[T](
				fakeContext(), fakeTarget(),
				map[string]any{},
			)
			assert.Error(t, err)
		})

		t.Run("no error", func(t *testing.T) {
			v, err := LoadForStruct[T](
				fakeContext(), fakeTarget(),
				map[string]any{"A": "a"},
			)
			assert.NoError(t, err)
			assert.Equal(t, "a", v.A)
		})
	})

	t.Run("defaulting with error", func(t *testing.T) {
		_, err := LoadForStruct[loadForStructTestWithDefaultingError](
			fakeContext(), fakeTarget(),
			map[string]any{},
		)
		assert.Error(t, err)
	})

	t.Run("validating with error", func(t *testing.T) {
		_, err := LoadForStruct[loadForStructTestWithValidationError](
			fakeContext(), fakeTarget(),
			map[string]any{},
		)
		assert.Error(t, err)
	})

	t.Run("integration", func(t *testing.T) {
		t.Run("missing required value", func(t *testing.T) {
			_, err := LoadForStruct[loadForStructIntegration](
				fakeContext(), fakeTarget(),
				map[string]any{},
			)
			assert.Error(t, err)
		})

		t.Run("no error", func(t *testing.T) {
			v, err := LoadForStruct[loadForStructIntegration](
				fakeContext(), fakeTarget(),
				map[string]any{
					"RequiredField": "a",
				},
			)
			assert.NoError(t, err)
			assert.Equal(t, "a", v.RequiredField)
			assert.True(t, v.OptionalField)
		})
	})
}
