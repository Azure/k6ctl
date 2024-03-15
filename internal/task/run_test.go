package task

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/Azure/k6ctl/internal/config"
	"github.com/Azure/k6ctl/internal/target"
)

func TestRunTask_Integration(t *testing.T) {
	configReg := config.NewRegistry()
	configReg.Register(
		config.Provide[string](
			"echo",
			func(
				ctx context.Context,
				target target.Target,
				params map[string]any,
			) (string, error) {
				v, ok := params["message"]
				if ok {
					return fmt.Sprint(v), nil
				}
				return "", fmt.Errorf("message not found")
			},
			func(ctx context.Context, target target.Target, s string) (string, error) {
				return s, nil
			},
		),
	)

	taskConfig := &Schema{
		Name: "test",
		Configs: []ConfigProvider{
			{
				Provider: ConfigProviderProviderSpec{
					Name: "echo",
					Params: map[string]any{
						"message": "hello world!",
					},
				},
				Env: "TEST_MESSAGE",
			},
		},
		Files: []FileMount{
			{
				Source: "test.js",
				Dest:   "test.js",
			},
		},
		K6: K6{
			Namespace: "test",
		},
	}

	fakeTarget := &target.StaticTarget{
		Kubeconfig: "/tmp/fake-kubeconfig",
	}

	ctx := context.Background()

	err := RunTask(
		ctx,
		fakeTarget,
		configReg.GetByName,
		taskConfig,
		"./testdata/integration",
		"test.js",
		applyRunTaskOptionFunc(func(option *runTaskOption) error {
			option.KubeClientFactory = func(kubeconfig string) (kubernetes.Interface, error) {
				return fake.NewSimpleClientset(), nil
			}

			return nil
		}),
		WithFollowLogs(false),
	)
	assert.NoError(t, err)
}
