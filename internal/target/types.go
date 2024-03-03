package target

type Target interface {
	// GetKubeconfig returns the path to the kubeconfig.
	// If the kubeconfig is not available, the second return value is false.
	GetKubeconfig() (string, bool)
}
