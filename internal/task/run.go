package task

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/sourcegraph/conc/iter"
	k8sbatchv1 "k8s.io/api/batch/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/Azure/k6ctl/internal/config"
	"github.com/Azure/k6ctl/internal/kubelib"
	"github.com/Azure/k6ctl/internal/target"
)

const (
	labelKeyTaskName     = "k6ctl/task"
	scriptsVolumeName    = "k6-scripts"
	containerScriptsPath = "/scripts"
)

type runTaskOption struct {
	// KubeClientFactory provides the kubernetes client to use for the task.
	// If not provided, createKubeClientFromKubeConfig is used.
	// Unit test can provide a mock implementation.
	KubeClientFactory kubelib.KubeClientFactory
}

func defaultRunTaskOption() *runTaskOption {
	return &runTaskOption{
		KubeClientFactory: kubelib.CreateKubeClientFromKubeConfig,
	}
}

type RunTaskOption interface {
	apply(option *runTaskOption) error
}

type applyRunTaskOptionFunc func(option *runTaskOption) error

func (f applyRunTaskOptionFunc) apply(option *runTaskOption) error {
	return f(option)
}

func RunTask(
	ctx context.Context,
	target target.Target,
	getConfigProviderByName config.GetConfigProviderByName,
	taskConfig *Schema,
	sourceBaseDir string,
	script string,
	options ...RunTaskOption,
) error {
	opt := defaultRunTaskOption()
	for _, o := range options {
		if err := o.apply(opt); err != nil {
			return err
		}
	}

	if s, err := filepath.Abs(filepath.Clean(sourceBaseDir)); err != nil {
		return fmt.Errorf("invalid source base dir %q: %w", sourceBaseDir, err)
	} else {
		sourceBaseDir = s
	}

	kubeconfig, ok := target.GetKubeconfig()
	if !ok {
		return fmt.Errorf("target does not have kubeconfig")
	}
	kubeClient, err := opt.KubeClientFactory(kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	tr := &taskRunner{
		target:                  target,
		kubeClient:              kubeClient,
		getConfigProviderByName: getConfigProviderByName,
		taskConfig:              taskConfig,
		sourceBaseDir:           sourceBaseDir,
		script:                  script,
	}

	return tr.Run(ctx)
}

type taskRunner struct {
	target     target.Target
	kubeClient kubernetes.Interface

	getConfigProviderByName config.GetConfigProviderByName
	taskConfig              *Schema
	sourceBaseDir           string
	script                  string
}

type createOrUpdateClient[T any] interface {
	Create(ctx context.Context, obj T, opts k8smetav1.CreateOptions) (T, error)
	Update(ctx context.Context, obj T, opts k8smetav1.UpdateOptions) (T, error)
}

func createOrUpdateObject[T any](
	ctx context.Context,
	client createOrUpdateClient[T],
	obj T,
) (T, error) {
	objCreated, err := client.Create(ctx, obj, k8smetav1.CreateOptions{})
	switch {
	case err == nil:
		return objCreated, nil
	case k8serrors.IsAlreadyExists(err):
	// update
	default:
		var empty T
		return empty, err
	}

	return client.Update(ctx, obj, k8smetav1.UpdateOptions{})
}

func (tr *taskRunner) Run(ctx context.Context) error {
	var (
		secretsToCreate    []*k8scorev1.Secret
		configMapsToCreate []*k8scorev1.ConfigMap
		jobsToCreate       []*k8sbatchv1.Job
	)

	if len(tr.taskConfig.Configs) > 0 {
		configs, err := tr.resolveConfigs(ctx, tr.taskConfig.Configs)
		if err != nil {
			return fmt.Errorf("failed to resolve configs: %w", err)
		}
		configsSecretObject, err := tr.buildConfigSecretObject(ctx, configs)
		if err != nil {
			return fmt.Errorf("failed to build config secret object: %w", err)
		}
		secretsToCreate = append(secretsToCreate, configsSecretObject)
	}

	if len(tr.taskConfig.Files) > 0 {
		scriptsConfigMapObject, err := tr.buildScriptsConfigMapObject(ctx)
		if err != nil {
			return fmt.Errorf("failed to build scripts config map object: %w", err)
		}
		configMapsToCreate = append(configMapsToCreate, scriptsConfigMapObject)
	}

	jobObject, err := tr.buildJobObject()
	if err != nil {
		return err
	}
	jobsToCreate = append(jobsToCreate, jobObject)

	secretsClient := tr.kubeClient.CoreV1().Secrets(tr.objectNamespace())
	configMapsClient := tr.kubeClient.CoreV1().ConfigMaps(tr.objectNamespace())
	jobsClient := tr.kubeClient.BatchV1().Jobs(tr.objectNamespace())

	for _, secret := range secretsToCreate {
		_, err := createOrUpdateObject(ctx, secretsClient, secret)
		if err != nil {
			return fmt.Errorf("failed to create secret %q: %w", secret.Name, err)
		}
	}
	for _, configMap := range configMapsToCreate {
		_, err := createOrUpdateObject(ctx, configMapsClient, configMap)
		if err != nil {
			return fmt.Errorf("failed to create config map %q: %w", configMap.Name, err)
		}
	}
	for _, job := range jobsToCreate {
		_, err := createOrUpdateObject(ctx, jobsClient, job)
		if err != nil {
			return fmt.Errorf("failed to create job %q: %w", job.Name, err)
		}
	}

	return nil
}

func (tr *taskRunner) taskJobName() string {
	return fmt.Sprintf("k6ctl-job-%s", tr.taskConfig.Name)
}

func (tr *taskRunner) configsSecretName() string {
	return fmt.Sprintf("k6ctl-configs-secret-%s", tr.taskConfig.Name)
}

func (tr *taskRunner) scriptsConfigMapName() string {
	return fmt.Sprintf("k6ctl-scripts-config-%s", tr.taskConfig.Name)
}

func (tr *taskRunner) objectNamespace() string {
	return tr.taskConfig.K6.Namespace
}

func isChildPath(basePath string, childPath string) bool { // TODO: revisit this implementation
	relPath, err := filepath.Rel(basePath, childPath)
	if err != nil {
		return false
	}
	return relPath != ".." && !filepath.IsAbs(relPath)
}

type resolvedConfig struct {
	Value string
	Env   string
}

func (tr *taskRunner) resolveConfigs(
	ctx context.Context,
	configProviders []ConfigProvider,
) ([]resolvedConfig, error) {
	return iter.MapErr(configProviders, func(cp *ConfigProvider) (resolvedConfig, error) {
		return tr.resolveConfig(ctx, *cp)
	})
}

func (tr *taskRunner) resolveConfig(
	ctx context.Context,
	configProvider ConfigProvider,
) (resolvedConfig, error) {
	p, ok := tr.getConfigProviderByName(configProvider.Provider.Name)
	if !ok {
		return resolvedConfig{}, fmt.Errorf("no config provider %q", configProvider.Provider.Name)
	}

	value, err := p.Resolve(ctx, tr.target, configProvider.Provider.Params)
	if err != nil {
		return resolvedConfig{}, fmt.Errorf("%s: failed to resolve config: %w", p.Name(), err)
	}

	rv := resolvedConfig{
		Value: value,
		Env:   configProvider.Env,
	}

	return rv, nil
}
