package dotnetworker

import (
	"net/url"

	"github.com/cjlapao/locally-cli/internal/context/git_component"
	"github.com/cjlapao/locally-cli/internal/environment"
	"github.com/cjlapao/locally-cli/internal/git"
)

type DotnetPipelineWorkerParameters struct {
	Context              string            `json:"context,omitempty" yaml:"context,omitempty"`
	BaseImage            string            `json:"baseImage,omitempty" yaml:"baseImage,omitempty"`
	RepoAccessToken      string            `json:"repoAccessToken,omitempty" yaml:"repoAccessToken,omitempty"`
	RepoUrl              string            `json:"repoUrl,omitempty" yaml:"repoUrl,omitempty"`
	ProjectPath          string            `json:"projectPath,omitempty" yaml:"projectPath,omitempty"`
	Command              string            `json:"command,omitempty" yaml:"command,omitempty"`
	Arguments            map[string]string `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	BuildArguments       []string          `json:"buildArguments,omitempty" yaml:"buildArguments,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVars,omitempty" yaml:"environmentVars,omitempty"`
}

func (c *DotnetPipelineWorkerParameters) Validate() bool {
	if c == nil {
		return false
	}

	if c.RepoUrl == "" {
		return false
	}
	if c.ProjectPath == "" {
		return false
	}

	if c.Context == "" {
		c.Context = "."
	}

	if c.Command == "" {
		c.Command = "build"
	}

	return true
}

func (c *DotnetPipelineWorkerParameters) Decode() {
	env := environment.Get()

	c.RepoUrl = env.Replace(c.RepoUrl)
	if c.RepoAccessToken != "" {
		cred := git_component.GitCredentials{
			AccessToken: env.Replace(c.RepoAccessToken),
		}
		parsedUrl, err := url.Parse(c.RepoUrl)
		if err == nil {
			c.RepoUrl, _, _ = git.InsertCredentials(parsedUrl, &cred)
		}
	}
	c.ProjectPath = env.Replace(c.ProjectPath)

	for key, value := range c.Arguments {
		c.Arguments[key] = env.Replace(value)
	}

	for key, value := range c.BuildArguments {
		c.BuildArguments[key] = env.Replace(value)
	}

	for key, value := range c.EnvironmentVariables {
		c.EnvironmentVariables[key] = env.Replace(value)
	}
}
