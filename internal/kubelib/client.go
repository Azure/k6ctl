package kubelib

import (
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

// KubeClientFactory is a function that creates a kubernetes client from the kubeconfig file at the given path.
type KubeClientFactory func(kubeConfigPath string) (kubernetes.Interface, error)

// CreateKubeClientFromKubeConfig creates a kubernetes client from the kubeconfig file at the given path.
func CreateKubeClientFromKubeConfig(kubeConfigPath string) (kubernetes.Interface, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeConfigPath)
	if err != nil {
		return nil, err
	}

	return kubernetes.NewForConfig(config)
}
