package npmworker

import (
	"github.com/cjlapao/locally-cli/internal/environment"
)

type NpmPipelineWorkerParameters struct {
	Command       string `json:"command,omitempty" yaml:"command,omitempty"`
	CustomCommand string `json:"customCommand,omitempty" yaml:"customCommand,omitempty"`
	WorkingDir    string `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	MinVersion    string `json:"minVersion,omitempty" yaml:"minVersion,omitempty"`
}

func (c *NpmPipelineWorkerParameters) Validate() bool {
	if c == nil {
		return false
	}

	if c.Command == "" {
		return false
	}

	if c.Command == "custom" && c.CustomCommand == "" {
		return false
	}

	if !(c.Command == "ci" || c.Command == "install" || c.Command == "publish" || c.Command == "custom") {
		return false
	}

	if c.WorkingDir == "" {
		return false
	}

	return true
}

func (c *NpmPipelineWorkerParameters) Decode() {
	env := environment.Get()

	c.Command = env.Replace(c.Command)
	c.CustomCommand = env.Replace(c.CustomCommand)
	c.WorkingDir = env.Replace(c.WorkingDir)
	c.MinVersion = env.Replace(c.MinVersion)
}
