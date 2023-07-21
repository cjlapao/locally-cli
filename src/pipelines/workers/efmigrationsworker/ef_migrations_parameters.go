package efmigrationsworker

import (
	"github.com/cjlapao/locally-cli/environment"
	"github.com/cjlapao/locally-cli/git"
	"net/url"

	"github.com/cjlapao/locally-cli/configuration"
)

type EfMigrationsPipelineWorkerParameters struct {
	Context              string            `json:"context,omitempty" yaml:"context,omitempty"`
	BaseImage            string            `json:"baseImage,omitempty" yaml:"baseImage,omitempty"`
	RepoAccessToken      string            `json:"repoAccessToken,omitempty" yaml:"repoAccessToken,omitempty"`
	RepoUrl              string            `json:"repoUrl,omitempty" yaml:"repoUrl,omitempty"`
	ProjectPath          string            `json:"projectPath,omitempty" yaml:"projectPath,omitempty"`
	StartupProjectPath   string            `json:"startupProjectPath,omitempty" yaml:"startupProjectPath,omitempty"`
	DbConnectionString   string            `json:"dbConnectionString,omitempty" yaml:"dbConnectionString,omitempty"`
	Arguments            map[string]string `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	EnvironmentVariables map[string]string `json:"environmentVars,omitempty" yaml:"environmentVars,omitempty"`
}

func (c *EfMigrationsPipelineWorkerParameters) Validate() bool {
	if c.RepoUrl == "" {
		return false
	}
	if c.ProjectPath == "" {
		return false
	}

	if c.Context == "" {
		c.Context = "."
	}

	return true
}

func (c *EfMigrationsPipelineWorkerParameters) Decode() {
	env := environment.Get()

	c.RepoUrl = env.Replace(c.RepoUrl)
	if c.RepoAccessToken != "" {
		cred := configuration.GitCredentials{
			AccessToken: env.Replace(c.RepoAccessToken),
		}
		parsedUrl, err := url.Parse(c.RepoUrl)
		if err == nil {
			c.RepoUrl, _, _ = git.InsertCredentials(parsedUrl, &cred)
		}
	}
	c.ProjectPath = env.Replace(c.ProjectPath)

	c.StartupProjectPath = env.Replace(c.StartupProjectPath)

	c.DbConnectionString = env.Replace(c.DbConnectionString)

	for key, value := range c.Arguments {
		c.Arguments[key] = env.Replace(value)
	}

	for key, value := range c.EnvironmentVariables {
		c.EnvironmentVariables[key] = env.Replace(value)
	}
}
