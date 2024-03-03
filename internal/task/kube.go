package task

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func createKubeClientFromKubeConfig(kubeConfigPath string) (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
