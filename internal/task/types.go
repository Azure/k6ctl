package task

type Schema struct {
	Version string           `json:"version"`
	Name    string           `json:"name"`
	Files   []FileMount      `json:"files"`
	Configs []ConfigProvider `json:"configs"`
	K6      K6               `json:"k6"`
}

type FileMount struct {
	Source string `json:"source"`
	Dest   string `json:"dest"`
}

type ConfigProviderProviderSpec struct {
	Name   string         `json:"name"`
	Params map[string]any `json:"params"`
}

type ConfigProvider struct {
	Provider ConfigProviderProviderSpec `json:"provider"`
	Env      string                     `json:"env"`
}

type K6 struct {
	Namespace      string           `json:"namespace"`
	PodImage       string           `json:"image"`
	ControllerKind string           `json:"controllerKind"`
	JobSpec        struct{}         `json:"jobSpec"`
	PodSpec        struct{}         `json:"podSpec"`
	ConfigPlugins  []K6ConfigPlugin `json:"configPlugins"`
}

type K6ConfigPlugin struct {
	Namespace  string   `json:"namespace"`
	BinaryPath string   `json:"binaryPath"`
	Args       []string `json:"args"`
}
