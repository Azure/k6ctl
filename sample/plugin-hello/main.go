package main

import (
	"context"
	"fmt"

	"github.com/Azure/k6ctl"
)

type params struct {
	Message string `mapstructure:"message"`
}

func (p params) Defaulting(
	ctx context.Context,
	target k6ctl.Target,
) (params, error) {
	if p.Message == "" {
		return params{
			Message: "world",
		}, nil
	}
	return p, nil
}

func main() {
	reg := k6ctl.NewConfigProviderRegistry()
	reg.Register(
		k6ctl.ProvideConfig(
			"message",
			k6ctl.LoadConfigForStruct[params],
			func(ctx context.Context, target k6ctl.Target, p params) (string, error) {
				return fmt.Sprintf("hello %s", p.Message), nil
			},
		),
	)

	k6ctl.ServeConfigRegistryPlugin(reg)
}
