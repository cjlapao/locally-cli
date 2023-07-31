package webclientmanifestworker

import (
	"github.com/cjlapao/locally-cli/environment"
)

type WebClientManifestPipelineWorkerParameters struct {
	Register bool   `json:"register,omitempty" yaml:"register,omitempty"`
	Path     string `json:"path,omitempty" yaml:"path,omitempty"`
	Endpoint string `json:"endpoint,omitempty" yaml:"endpoint,omitempty"`
}

func (c *WebClientManifestPipelineWorkerParameters) Validate() bool {
	if c == nil {
		return false
	}

	if c.Path == "" {
		return false
	}

	if !c.Register {
		return false
	}

	if c.Endpoint == "" {
		return false
	}

	return true
}

func (c *WebClientManifestPipelineWorkerParameters) Decode() {
	env := environment.Get()

	c.Path = env.Replace(c.Path)
	c.Endpoint = env.Replace(c.Endpoint)
}
