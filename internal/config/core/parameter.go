package core

import (
	"context"
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/mitchellh/mapstructure"

	"github.com/Azure/k6ctl/internal/config"
	"github.com/Azure/k6ctl/internal/target"
	"github.com/Azure/k6ctl/internal/task"
)

const configProviderNameParameter = "parameter"

const (
	parameterOnMissingError  = "error"
	parameterOnMissingPrompt = "prompt"
	parameterOnMissingEmpty  = "empty"
)

type parameterSettings struct {
	Name      string `mapstructure:"name" validate:"required"`
	OnMissing string `mapstructure:"onMissing"`
}

func loadParameterSettings(cp task.ConfigProvider) (parameterSettings, error) {
	var params parameterSettings
	if err := mapstructure.Decode(cp.Provider.Params, &params); err != nil {
		return parameterSettings{}, fmt.Errorf("invalid params for resolving %q: %w", cp.Env, err)
	}
	if err := config.DefaultValidator.Struct(params); err != nil {
		return parameterSettings{}, fmt.Errorf("invalid params for resolving %q: %w", cp.Env, err)

	}
	if defaultedParams, err := params.defaulting(); err == nil {
		params = defaultedParams
	} else {
		return parameterSettings{}, fmt.Errorf("invalid params for resolving %q: %w", cp.Env, err)
	}
	if err := params.Validate(); err != nil {
		return parameterSettings{}, fmt.Errorf("invalid params for resolving %q: %w", cp.Env, err)
	}

	return params, nil
}

func (p parameterSettings) defaulting() (parameterSettings, error) {
	rv := p

	if rv.OnMissing == "" {
		rv.OnMissing = parameterOnMissingPrompt
	}

	return rv, nil
}

func (p parameterSettings) Defaulting(_ context.Context, _ target.Target) (parameterSettings, error) {
	return p.defaulting()
}

func (p parameterSettings) Validate() error {
	switch p.OnMissing {
	case parameterOnMissingError, parameterOnMissingPrompt, parameterOnMissingEmpty:
		// valid values
	default:
		return fmt.Errorf("invalid onMissing value %q", p.OnMissing)
	}

	return nil
}

// TODO: move to ui package
func promptForParameter(
	title string,
) (string, error) {
	var v string
	err := huh.NewInput().Title(title).
		Prompt("? ").
		Value(&v).
		Validate(func(s string) error {
			if s == "" {
				return fmt.Errorf("value is required")
			}

			return nil
		}).
		Run()
	if err != nil {
		return "", err
	}

	return v, nil
}

func resolveParameters(
	configProviders []task.ConfigProvider,
	inputsFromUserInputs map[string]string,
) (ResolvedParameters, error) {
	rv := make(ResolvedParameters)
	for k, v := range inputsFromUserInputs {
		rv[k] = v
	}

	var paramsList []parameterSettings
	for _, cp := range configProviders {
		if cp.Provider.Name != configProviderNameParameter {
			continue
		}
		params, err := loadParameterSettings(cp)
		if err != nil {
			return nil, err
		}

		paramsList = append(paramsList, params)
	}
	if len(paramsList) == 0 {
		// no parameter to resolve
		return rv, nil
	}

	// pass 1: fill with user inputs or prompt
	for _, params := range paramsList {
		if _, ok := rv[params.Name]; ok {
			// already resolved
			continue
		}

		if params.OnMissing == parameterOnMissingPrompt {
			value, err := promptForParameter(fmt.Sprintf("Please input value for parameter %q", params.Name))
			if err != nil {
				return nil, fmt.Errorf("failed to prompt for parameter %q: %w", params.Name, err)
			}
			rv[params.Name] = value
		}
	}

	// pass 2: backfill default or error
	for _, params := range paramsList {
		if _, ok := rv[params.Name]; ok {
			// already resolved
			continue
		}

		switch params.OnMissing {
		case parameterOnMissingError:
			return nil, fmt.Errorf("missing required parameter %q", params.Name)
		case parameterOnMissingEmpty:
			rv[params.Name] = ""
		}
	}

	return rv, nil
}

type ResolvedParameters map[string]string

func (p ResolvedParameters) CreateProvider() config.Provider {
	return config.Provide(
		configProviderNameParameter,
		config.LoadForStruct[parameterSettings],
		func(ctx context.Context, _ target.Target, params parameterSettings) (string, error) {
			v, ok := p[params.Name]
			if ok {
				return v, nil
			}

			if params.OnMissing == parameterOnMissingEmpty {
				return "", nil
			}
			// NOTE: prompt is done in previous stage
			return "", fmt.Errorf("missing required parameter %q", params.Name)
		},
	)
}
