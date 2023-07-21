package whatsnewworker

import (
	"github.com/cjlapao/locally-cli/environment"
)

type WhatsNewPipelineWorkerParameters struct {
	Register bool   `json:"register,omitempty" yaml:"register,omitempty"`
	Path     string `json:"path,omitempty" yaml:"path,omitempty"`
}

func (c *WhatsNewPipelineWorkerParameters) Validate() bool {
	if c == nil {
		return false
	}

	if c.Path == "" {
		return false
	}

	if !c.Register {
		return false
	}

	return true
}

func (c *WhatsNewPipelineWorkerParameters) Decode() {
	env := environment.Get()

	c.Path = env.Replace(c.Path)
}
