package infrastructureworker

// import (
// 	"strings"

// 	"github.com/cjlapao/locally-cli/internal/environment"
// )

// type InfrastructureParameters struct {
// 	Command           string   `json:"command,omitempty" yaml:"command,omitempty"`
// 	StackName         string   `json:"stackName,omitempty" yaml:"stackName,omitempty"`
// 	Arguments         []string `json:"arguments,omitempty" yaml:"arguments,omitempty"`
// 	BuildDependencies bool     `json:"buildDependencies,omitempty" yaml:"buildDependencies,omitempty"`
// 	WorkingDirectory  string   `json:"workingDir,omitempty" yaml:"workingDir,omitempty"`
// }

// func (c *InfrastructureParameters) Validate() bool {
// 	return c.Command != "" && c.StackName != ""
// }

// func (c *InfrastructureParameters) Decode() {
// 	env := environment.GetInstance()

// 	parts := strings.Split(c.Command, " ")
// 	if len(parts) > 1 {
// 		c.Command = parts[0]
// 		reorderArgs := parts[1:]
// 		reorderArgs = append(reorderArgs, c.Arguments...)
// 		c.Arguments = reorderArgs
// 	}

// 	c.Command = env.Replace(c.Command)
// 	c.WorkingDirectory = env.Replace(c.WorkingDirectory)
// 	c.StackName = env.Replace(c.StackName)
// 	for key, value := range c.Arguments {
// 		c.Arguments[key] = env.Replace(value)
// 	}
// }
