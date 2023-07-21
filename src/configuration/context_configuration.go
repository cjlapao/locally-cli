package configuration

type ContextConfiguration struct {
	SchemaVersion        string                `json:"schemaVersion,omitempty" yaml:"schemaVersion,omitempty"`
	ConfigFolder         string                `json:"folder,omitempty" yaml:"folder,omitempty"`
	RootURI              string                `json:"rootUri,omitempty" yaml:"rootUri,omitempty"`
	OutputPath           string                `json:"outputPath,omitempty" yaml:"outputPath,omitempty"`
	LocallyConfigService *LocallyConfigService `json:"locallyConfigService,omitempty" yaml:"locallyConfigService,omitempty"`
}

type LocallyConfigService struct {
	Url             string `json:"url" yaml:"url"`
	ReverseProxyUrl string `json:"reverseProxyUrl"`
}
