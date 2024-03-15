package task

import (
	"context"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/Azure/k6ctl/internal/config"
	"github.com/Azure/k6ctl/internal/stdlib"
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

	const namespace = "test"
	const instances = 2

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
			Namespace: namespace,
		},
	}

	fakeTarget := &target.StaticTarget{
		Kubeconfig: "/tmp/fake-kubeconfig",
	}

	ctx := context.Background()

	kubeClient := fake.NewSimpleClientset(
		&k8scorev1.Namespace{
			ObjectMeta: k8smetav1.ObjectMeta{
				Name: namespace,
			},
		},
	)

	err := RunTask(
		ctx,
		fakeTarget,
		configReg.GetByName,
		taskConfig,
		"./testdata/integration",
		"test.js",
		applyRunTaskOptionFunc(func(option *runTaskOption) error {
			option.KubeClientFactory = func(kubeconfig string) (kubernetes.Interface, error) {
				return kubeClient, nil
			}

			return nil
		}),
		WithFollowLogs(false),
		WithInstances(instances),
	)
	assert.NoError(t, err)

	jobsList, err := kubeClient.BatchV1().Jobs(namespace).List(ctx, k8smetav1.ListOptions{})
	assert.NoError(t, err)
	assert.Len(t, jobsList.Items, 1)

	job := jobsList.Items[0]
	assert.Equal(t, stdlib.ValOrZero(job.Spec.Parallelism), int32(instances))
	assert.Equal(t, stdlib.ValOrZero(job.Spec.Parallelism), int32(instances))
}
