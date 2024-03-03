package target

// StaticTarget provides static target settings.
type StaticTarget struct {
	Kubeconfig string
}

var _ Target = (*StaticTarget)(nil)

func (t *StaticTarget) GetKubeconfig() (string, bool) {
	return t.Kubeconfig, t.Kubeconfig != ""
}
