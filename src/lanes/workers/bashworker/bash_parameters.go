package bashworker

import (
	"github.com/cjlapao/locally-cli/environment"
	"strings"
)

type BashParameters struct {
	Command          string   `json:"command,omitempty" yaml:"command,omitempty"`
	Arguments        []string `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	WorkingDirectory string   `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
	RetryCount       int      `json:"retryCount,omitempty" yaml:"retryCount,omitempty"`
	WaitForInSeconds int      `json:"waitFor,omitempty" yaml:"waitFor,omitempty"`
}

func (c *BashParameters) Validate() bool {
	return c.Command != ""
}

func (c *BashParameters) Decode() {
	env := environment.Get()

	parts := strings.Split(c.Command, " ")
	if len(parts) > 1 {
		c.Command = parts[0]
		reorderArgs := parts[1:]
		reorderArgs = append(reorderArgs, c.Arguments...)
		c.Arguments = reorderArgs
	}

	c.Command = env.Replace(c.Command)
	c.WorkingDirectory = env.Replace(c.WorkingDirectory)
	for key, value := range c.Arguments {
		c.Arguments[key] = env.Replace(value)
	}
}
