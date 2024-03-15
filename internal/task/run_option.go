package task

import "github.com/Azure/k6ctl/internal/kubelib"

type runTaskOption struct {
	// Instances specifies the number of instances to run.
	// Defaults to 1.
	Instances int32
	// FollowLogs specifies whether to follow the logs of the task.
	FollowLogs bool
	// KubeClientFactory provides the kubernetes client to use for the task.
	// If not provided, createKubeClientFromKubeConfig is used.
	// Unit test can provide a mock implementation.
	KubeClientFactory kubelib.KubeClientFactory
}

func defaultRunTaskOption() *runTaskOption {
	return &runTaskOption{
		Instances:         1,
		FollowLogs:        true,
		KubeClientFactory: kubelib.CreateKubeClientFromKubeConfig,
	}
}

// RunTaskOption configures the behavior of RunTask.
type RunTaskOption interface {
	apply(option *runTaskOption) error
}

type applyRunTaskOptionFunc func(option *runTaskOption) error

func (f applyRunTaskOptionFunc) apply(option *runTaskOption) error {
	return f(option)
}

// WithFollowLogs specifies whether to follow the logs of the task.
func WithFollowLogs(followLogs bool) RunTaskOption {
	return applyRunTaskOptionFunc(func(option *runTaskOption) error {
		option.FollowLogs = followLogs
		return nil
	})
}

// WithInstances specifies the number of instances to run.
func WithInstances(replicas int32) RunTaskOption {
	return applyRunTaskOptionFunc(func(option *runTaskOption) error {
		option.Instances = replicas
		return nil
	})
}
