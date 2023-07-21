package gitworker

import (
	"github.com/cjlapao/locally-cli/environment"
	"github.com/cjlapao/locally-cli/mappers"

	"github.com/cjlapao/locally-cli/configuration"
)

type GitParameters struct {
	RepoUrl     string                        `json:"repoUrl,omitempty" yaml:"repoUrl,omitempty"`
	Destination string                        `json:"destination,omitempty" yaml:"destination,omitempty"`
	Clean       bool                          `json:"clean,omitempty" yaml:"clean,omitempty"`
	Credentials *configuration.GitCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

func (c *GitParameters) Validate() bool {
	return c.RepoUrl != ""
}

func (c *GitParameters) Decode() {
	env := environment.Get()

	c.RepoUrl = env.Replace(c.RepoUrl)
	c.Destination = env.Replace(c.Destination)
	c.Credentials = mappers.DecodeGitCredentials(c.Credentials)
}
