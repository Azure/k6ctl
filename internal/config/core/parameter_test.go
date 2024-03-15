package core

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Azure/k6ctl/internal/config"
	"github.com/Azure/k6ctl/internal/target"
	"github.com/Azure/k6ctl/internal/task"
)

func TestParameterSettings(t *testing.T) {
	t.Run("defaulting", func(t *testing.T) {
		cases := []struct {
			name  string
			input parameterSettings

			expectErr bool
			expected  parameterSettings
		}{
			{
				name:  "empty",
				input: parameterSettings{},
				expected: parameterSettings{
					OnMissing: parameterOnMissingPrompt,
				},
			},
			{
				name: "with onMissing",
				input: parameterSettings{
					OnMissing: parameterOnMissingEmpty,
				},
				expected: parameterSettings{
					OnMissing: parameterOnMissingEmpty,
				},
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				actual, err := tc.input.defaulting()
				if tc.expectErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
				assert.Equal(t, tc.expected, actual)
			})
		}
	})

	t.Run("Validate", func(t *testing.T) {
		cases := []struct {
			name      string
			input     parameterSettings
			expectErr bool
		}{
			{
				name: "valid",
				input: parameterSettings{
					OnMissing: parameterOnMissingEmpty,
				},
				expectErr: false,
			},
			{
				name: "invalid onMissing",
				input: parameterSettings{
					OnMissing: "foobar",
				},
				expectErr: true,
			},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				err := tc.input.Validate()
				if tc.expectErr {
					assert.Error(t, err)
					return
				}
				assert.NoError(t, err)
			})

		}
	})
}

func TestResolveParameters(t *testing.T) {
	parameter := func(
		envName string,
		parameterName string,
		mu ...func(provider *task.ConfigProvider),
	) task.ConfigProvider {
		rv := task.ConfigProvider{
			Provider: task.ConfigProviderProviderSpec{
				Name: "parameter",
				Params: map[string]any{
					"name": parameterName,
				},
			},
			Env: envName,
		}

		for _, m := range mu {
			m(&rv)
		}

		return rv
	}

	cases := []struct {
		name                 string
		configProviders      []task.ConfigProvider
		inputsFromUserInputs map[string]string

		expectErr bool
		expected  resolvedParameters
	}{
		{
			name:                 "empty config providers",
			configProviders:      []task.ConfigProvider{},
			inputsFromUserInputs: map[string]string{},
			expected:             resolvedParameters{},
		},
		{
			name: "fill from user input",
			configProviders: []task.ConfigProvider{
				parameter("FOO", "foo"),
				parameter("FOO2", "foo", func(provider *task.ConfigProvider) {
					provider.Provider.Params["onMissing"] = parameterOnMissingError
				}),
			},
			inputsFromUserInputs: map[string]string{
				"foo": "bar",
			},
			expected: resolvedParameters{
				"foo": "bar",
			},
		},
		{
			name: "fill missing with empty",
			configProviders: []task.ConfigProvider{
				parameter("FOO", "foo", func(provider *task.ConfigProvider) {
					provider.Provider.Params["onMissing"] = parameterOnMissingEmpty
				}),
			},
			inputsFromUserInputs: map[string]string{},
			expected: resolvedParameters{
				"foo": "",
			},
		},
		{
			name: "error out on missing value",
			configProviders: []task.ConfigProvider{
				parameter("FOO", "foo", func(provider *task.ConfigProvider) {
					provider.Provider.Params["onMissing"] = parameterOnMissingError
				}),
			},
			inputsFromUserInputs: map[string]string{},
			expectErr:            true,
		},
		// TODO: testing prompt logic
		//{
		//	name: "prompt for missing value",
		//	configProviders: []task.ConfigProvider{
		//		parameter("FOO", "foo"),
		//	},
		//	inputsFromUserInputs: map[string]string{},
		//	expectErr:            true,
		//},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := resolveParameters(tc.configProviders, tc.inputsFromUserInputs)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestResolvedParameters_ConfigProvider(t *testing.T) {
	resolve := func(t *testing.T, cp config.Provider, userInput map[string]any) (string, error) {
		t.Helper()

		ctx := context.Background()
		var fakeTarget target.Target

		return cp.Resolve(ctx, fakeTarget, userInput)
	}

	cases := []struct {
		name      string
		v         resolvedParameters
		userInput map[string]any

		expectErr bool
		expected  string
	}{
		{
			name: "missing parameter, error out",
			v:    resolvedParameters{},
			userInput: map[string]any{
				"name": "bar",
			},
			expectErr: true,
		},
		{
			name: "missing parameter, error out (onMissing: error)",
			v:    resolvedParameters{},
			userInput: map[string]any{
				"name":      "bar",
				"onMissing": parameterOnMissingError,
			},
			expectErr: true,
		},
		{
			name: "missing parameter, error out (onMissing: prompt)",
			v:    resolvedParameters{},
			userInput: map[string]any{
				"name":      "bar",
				"onMissing": parameterOnMissingPrompt,
			},
			expectErr: true,
		},
		{
			name: "missing parameter, error out (onMissing: empty)",
			v:    resolvedParameters{},
			userInput: map[string]any{
				"name":      "bar",
				"onMissing": parameterOnMissingEmpty,
			},
			expectErr: false,
			expected:  "",
		},
		{
			name: "loaded parameter",
			v: resolvedParameters{
				"foo": "bar",
			},
			userInput: map[string]any{
				"name": "foo",
			},
			expected: "bar",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cp := tc.v.CreateProvider()
			actual, err := resolve(t, cp, tc.userInput)
			if tc.expectErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expected, actual)
		})
	}
}
