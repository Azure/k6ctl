package task

import (
	"context"
	"fmt"
	"os"

	k8sbatchv1 "k8s.io/api/batch/v1"
	k8scorev1 "k8s.io/api/core/v1"
	k8smetav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/Azure/k6ctl/internal/kubelib"
	"github.com/Azure/k6ctl/internal/stdlib"
)

func (tr *taskRunner) buildJobObject() (*k8sbatchv1.Job, error) {
	taskName := tr.taskConfig.Name
	k6RunnerImage := tr.taskConfig.K6.PodImage
	scriptToRun := tr.script

	var envFrom []k8scorev1.EnvFromSource
	if len(tr.taskConfig.Configs) > 0 {
		envFrom = append(envFrom, k8scorev1.EnvFromSource{
			SecretRef: &k8scorev1.SecretEnvSource{
				LocalObjectReference: k8scorev1.LocalObjectReference{
					Name: tr.configsSecretName(),
				},
			},
		})
	}

	var volumes []k8scorev1.Volume
	if len(tr.taskConfig.Files) > 0 {
		volumes = append(volumes, k8scorev1.Volume{
			Name: scriptsVolumeName,
			VolumeSource: k8scorev1.VolumeSource{
				ConfigMap: &k8scorev1.ConfigMapVolumeSource{
					LocalObjectReference: k8scorev1.LocalObjectReference{
						Name: tr.scriptsConfigMapName(),
					},
				},
			},
		})
	}

	rv := &k8sbatchv1.Job{
		ObjectMeta: k8smetav1.ObjectMeta{
			Name:      tr.taskJobName(),
			Namespace: tr.objectNamespace(),
			Labels: map[string]string{
				labelKeyTaskName: taskName,
			},
		},
		Spec: k8sbatchv1.JobSpec{
			TTLSecondsAfterFinished: stdlib.Ptr[int32](100),
			Parallelism:             stdlib.Ptr[int32](1), // TODO: configurable
			Completions:             stdlib.Ptr[int32](1), // TODO: configurable
			ManualSelector:          stdlib.Ptr(true),
			Selector: &k8smetav1.LabelSelector{
				MatchLabels: map[string]string{
					labelKeyTaskName: taskName,
				},
			},
			Template: k8scorev1.PodTemplateSpec{
				ObjectMeta: k8smetav1.ObjectMeta{
					Labels: map[string]string{
						labelKeyTaskName: taskName,
					},
					Annotations: map[string]string{
						"linkerd.io/inject": "enabled",
					},
				},
				Spec: k8scorev1.PodSpec{
					Containers: []k8scorev1.Container{
						{
							Name:  containerNameRunner,
							Image: k6RunnerImage,
							Args:  []string{"run", scriptToRun},
							VolumeMounts: []k8scorev1.VolumeMount{
								{
									Name:      scriptsVolumeName,
									MountPath: containerScriptsPath,
								},
							},
							EnvFrom:    envFrom,
							WorkingDir: containerScriptsPath,
						},
					},
					RestartPolicy: "Never",
					Volumes:       volumes,
					Affinity: &k8scorev1.Affinity{
						PodAntiAffinity: &k8scorev1.PodAntiAffinity{
							RequiredDuringSchedulingIgnoredDuringExecution: []k8scorev1.PodAffinityTerm{
								{
									LabelSelector: &k8smetav1.LabelSelector{
										MatchExpressions: []k8smetav1.LabelSelectorRequirement{
											{
												Key:      labelKeyTaskName,
												Operator: "In",
												Values:   []string{taskName},
											},
										},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							},
						},
					},
				},
			},
		},
	}

	return rv, nil
}

func (tr *taskRunner) followJobLogs(
	ctx context.Context,
	job *k8sbatchv1.Job,
) error {
	selector, err := k8smetav1.LabelSelectorAsSelector(job.Spec.Selector)
	if err != nil {
		return fmt.Errorf("invalid job selector: %w", err)
	}

	addPrefix := false
	if stdlib.ValOrZero(job.Spec.Parallelism) > 1 {
		// there might be multiple pods, add prefix to distinguish them
		addPrefix = true
	}

	return kubelib.FollowLogs(
		ctx,
		tr.kubeClient,
		&kubelib.FollowLogsParams{
			Namespace: job.Namespace,
			Selector:  selector,
			Container: containerNameRunner,
			AddPrefix: addPrefix,
			Output:    os.Stderr,
		},
	)
}
