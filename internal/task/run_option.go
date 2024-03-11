package task

import "github.com/Azure/k6ctl/internal/kubelib"

type runTaskOption struct {
	// FollowLogs specifies whether to follow the logs of the task.
	FollowLogs bool
	// KubeClientFactory provides the kubernetes client to use for the task.
	// If not provided, createKubeClientFromKubeConfig is used.
	// Unit test can provide a mock implementation.
	KubeClientFactory kubelib.KubeClientFactory
}

func defaultRunTaskOption() *runTaskOption {
	return &runTaskOption{
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
