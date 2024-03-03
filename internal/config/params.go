package config

import (
	"context"

	"github.com/mitchellh/mapstructure"

	"github.com/Azure/k6ctl/internal/target"
)

type paramsWithValidation interface {
	Validate() error
}

type paramsWithDefaulting[T any] interface {
	Defaulting(ctx context.Context, target target.Target) (T, error)
}

func applyDefaultingIfNeeded[T any](ctx context.Context, target target.Target, v T) (T, error) {
	if withDefaulting, ok := (any(v)).(paramsWithDefaulting[T]); ok {
		return withDefaulting.Defaulting(ctx, target)
	}

	return v, nil
}

func applyValidationIfNeeded(v any) error {
	if withValidation, ok := v.(paramsWithValidation); ok {
		return withValidation.Validate()
	}

	return nil
}

func withErr[T any](err error) (T, error) {
	var empty T
	return empty, err
}

// LoadForStruct loads the given user input map into the given struct and validates it.
func LoadForStruct[T any](
	ctx context.Context,
	target target.Target,
	userInput map[string]any,
) (T, error) {
	var (
		rv  T
		err error
	)

	err = mapstructure.Decode(userInput, &rv)
	if err != nil {
		return withErr[T](err)
	}
	err = defaultValidator.Struct(rv)
	if err != nil {
		return withErr[T](err)
	}

	rv, err = applyDefaultingIfNeeded(ctx, target, rv)
	if err != nil {
		return withErr[T](err)
	}

	err = applyValidationIfNeeded(rv)
	if err != nil {
		return withErr[T](err)
	}

	return rv, nil
}
