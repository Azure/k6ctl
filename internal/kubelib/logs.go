package kubelib

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"sync"

	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// FollowLogsParams is the parameters for following logs.
type FollowLogsParams struct {
	// Namespace is the namespace of the pods.
	Namespace string
	// Selector is the selector of the pods.
	Selector labels.Selector
	// Container is the container name of the pods.
	Container string
	// MaxConcurrency is the maximum number of concurrent log streams to follow.
	MaxConcurrency int
	// AddPrefix specifies whether to add the prefix of the pod name to the logs.
	AddPrefix bool

	// Output is the writer to write the logs to.
	Output io.Writer
}

func (p *FollowLogsParams) defaults() error {
	if p.Namespace == "" {
		return fmt.Errorf(".Namespace is required")
	}
	if p.Selector == nil {
		return fmt.Errorf(".Selector is required")
	}
	if p.Container == "" {
		return fmt.Errorf(".Container is required")
	}
	if p.MaxConcurrency == 0 {
		p.MaxConcurrency = 5
	}
	if p.Output == nil {
		p.Output = os.Stderr
	}

	return nil
}

type podLog struct {
	target  k8scorev1.ObjectReference
	request rest.ResponseWrapper
}

type logsFollower struct {
	client         kubernetes.Interface
	namespace      string
	selector       labels.Selector
	container      string
	maxConcurrency int
	addPrefix      bool
	out            io.Writer

	wg *sync.WaitGroup

	podLogsChan      chan podLog
	handledTargets   map[k8scorev1.ObjectReference]struct{}
	handledTargetsMu *sync.Mutex
}

func (f *logsFollower) Wait() {
	f.wg.Wait()
}

func (f *logsFollower) Start(ctx context.Context) {
	f.wg.Add(1)
	go f.discoverPods(ctx)

	f.wg.Add(f.maxConcurrency)
	for i := 0; i < f.maxConcurrency; i++ {
		go f.followPodLogs(ctx)
	}
}

func (f *logsFollower) isNewTarget(target k8scorev1.ObjectReference) bool {
	f.handledTargetsMu.Lock()
	defer f.handledTargetsMu.Unlock()

	_, exists := f.handledTargets[target]
	if exists {
		// already handled
		return false
	}

	f.handledTargets[target] = struct{}{}
	return true
}

var logsAblePodPhase = map[k8scorev1.PodPhase]struct{}{
	k8scorev1.PodRunning:   {},
	k8scorev1.PodFailed:    {},
	k8scorev1.PodSucceeded: {},
}

func (f *logsFollower) discoverPods(ctx context.Context) {
	defer f.wg.Done()

	shouldHandlePod := func(pod *k8scorev1.Pod) bool {
		matchContainer := false
		for _, container := range pod.Spec.Containers {
			if container.Name == f.container {
				matchContainer = true
				break
			}
		}
		if !matchContainer {
			return false
		}

		if _, ok := logsAblePodPhase[pod.Status.Phase]; !ok {
			return false
		}

		return true
	}

	err := func() error {
		podsClient := f.client.CoreV1().Pods(f.namespace)

		watch, err := podsClient.Watch(ctx, k8smetav1.ListOptions{
			LabelSelector: f.selector.String(),
		})
		if err != nil {
			return err
		}
		defer watch.Stop()

		for {
			select {
			case <-ctx.Done():
				return nil
			case event, ok := <-watch.ResultChan():
				if !ok {
					return nil
				}

				pod, ok := event.Object.(*k8scorev1.Pod)
				if !ok {
					continue
				}
				if !shouldHandlePod(pod) {
					continue
				}

				objectRef := k8scorev1.ObjectReference{
					Kind:      "Pod",
					Namespace: pod.Namespace,
					Name:      pod.Name,
				}
				if !f.isNewTarget(objectRef) {
					continue
				}
				f.podLogsChan <- podLog{
					target: objectRef,
					request: podsClient.GetLogs(pod.Name, &k8scorev1.PodLogOptions{
						Container: f.container,
						Follow:    true,
					}),
				}
			}

		}
	}()

	if err != nil {
		// TODO: log error
	}
}

func (f *logsFollower) followPodLogs(ctx context.Context) {
	defer f.wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case p := <-f.podLogsChan:
			if err := f.followPodLog(ctx, p); err != nil {
				// TODO: log error
			}
		}
	}
}

type prefixWriter struct {
	prefix []byte
	w      io.Writer
}

func (w *prefixWriter) Write(p []byte) (n int, err error) {
	return w.w.Write(append(w.prefix, p...))
}

func (f *logsFollower) followPodLog(ctx context.Context, podLog podLog) error {
	if _, err := fmt.Fprintf(f.out, "Following logs of %s/%s...\n", podLog.target.Namespace, podLog.target.Name); err != nil {
		return err
	}

	logsStream, err := podLog.request.Stream(ctx)
	if err != nil {
		return err
	}
	defer func() {
		_ = logsStream.Close()
	}()

	r := bufio.NewReader(logsStream)

	out := f.out
	if f.addPrefix {
		out = &prefixWriter{
			prefix: []byte(fmt.Sprintf("[%s/%s] ", podLog.target.Namespace, podLog.target.Name)),
			w:      out,
		}
	}

	for {
		line, err := r.ReadBytes('\n')
		if _, writeErr := out.Write(line); writeErr != nil {
			return writeErr
		}
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
	}
}

// FollowLogs follows the logs of the pods that match the selector.
func FollowLogs(
	ctx context.Context,
	client kubernetes.Interface,
	params *FollowLogsParams,
) error {
	if err := params.defaults(); err != nil {
		return err
	}

	pr, pw := io.Pipe()
	follower := &logsFollower{
		client:         client,
		namespace:      params.Namespace,
		selector:       params.Selector,
		container:      params.Container,
		maxConcurrency: params.MaxConcurrency,
		addPrefix:      params.AddPrefix,
		out:            pw,

		wg: new(sync.WaitGroup),

		podLogsChan:      make(chan podLog),
		handledTargets:   map[k8scorev1.ObjectReference]struct{}{},
		handledTargetsMu: new(sync.Mutex),
	}

	go func() {
		follower.Start(ctx)
		follower.Wait()
		_ = pw.Close()
	}()

	_, err := io.Copy(params.Output, pr)
	return err
}
