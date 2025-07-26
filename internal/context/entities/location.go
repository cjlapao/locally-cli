package entities

type Location struct {
	RootFolder string `json:"rootFolder,omitempty" yaml:"rootFolder,omitempty"`
	Path       string `json:"path,omitempty" yaml:"path,omitempty"`
	DistPath   string `json:"distPath,omitempty" yaml:"distPath,omitempty"`
}
