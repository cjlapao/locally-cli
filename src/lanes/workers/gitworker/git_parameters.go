package gitworker

import (
	"github.com/cjlapao/locally-cli/context/git_component"
	"github.com/cjlapao/locally-cli/environment"
	"github.com/cjlapao/locally-cli/mappers"
)

type GitParameters struct {
	RepoUrl     string                        `json:"repoUrl,omitempty" yaml:"repoUrl,omitempty"`
	Destination string                        `json:"destination,omitempty" yaml:"destination,omitempty"`
	Clean       bool                          `json:"clean,omitempty" yaml:"clean,omitempty"`
	Credentials *git_component.GitCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
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
