package dockerworker

import (
	"fmt"
	"strings"

	"github.com/cjlapao/locally-cli/internal/context/docker_component"
	"github.com/cjlapao/locally-cli/internal/environment"
)

type DockerParameters struct {
	Command              string                          `json:"command,omitempty" yaml:"command,omitempty"`
	Registry             string                          `json:"registry,omitempty" yaml:"registry,omitempty"`
	Username             string                          `json:"username,omitempty" yaml:"username,omitempty"`
	Password             string                          `json:"password,omitempty" yaml:"password,omitempty"`
	SubscriptionId       string                          `json:"subscriptionId,omitempty" yaml:"subscriptionId,omitempty"`
	TenantId             string                          `json:"tenantId,omitempty" yaml:"tenantId,omitempty"`
	BasePath             string                          `json:"basePath,omitempty" yaml:"basePath,omitempty"`
	ImagePath            string                          `json:"imagePath,omitempty" yaml:"imagePath,omitempty"`
	ImageTag             string                          `json:"imageTag,omitempty" yaml:"imageTag,omitempty"`
	FullImagePath        string                          `json:"-" yaml:"-"`
	ConfigName           string                          `json:"configName,omitempty" yaml:"configName,omitempty"`
	ComponentName        string                          `json:"componentName,omitempty" yaml:"componentName,omitempty"`
	Arguments            map[string]string               `json:"arguments,omitempty" yaml:"arguments,omitempty"`
	EnvironmentVariables map[string]string               `json:"environmentVars,omitempty" yaml:"environmentVars,omitempty"`
	DockerCompose        *docker_component.DockerCompose `json:"dockerCompose,omitempty" yaml:"dockerCompose,omitempty"`
}

func (c *DockerParameters) Validate() bool {
	if c.Registry == "" {
		return false
	}
	if c.ImagePath == "" {
		return false
	}

	return true
}

func (c *DockerParameters) Decode() {
	env := environment.Get()

	c.Registry = env.Replace(c.Registry)
	c.ImagePath = env.Replace(c.ImagePath)
	if c.Username != "" {
		c.Username = env.Replace(c.Username)
	}
	if c.Password != "" {
		c.Password = env.Replace(c.Password)
	}
	if c.ImageTag != "" {
		c.ImageTag = env.Replace(c.ImageTag)
	}
	if c.BasePath != "" {
		c.BasePath = env.Replace(c.BasePath)
	}
	if c.SubscriptionId != "" {
		c.SubscriptionId = env.Replace(c.SubscriptionId)
	}
	if c.TenantId != "" {
		c.TenantId = env.Replace(c.TenantId)
	}

	if c.BasePath != "" {
		c.BasePath = strings.Trim(c.BasePath, "/")
		path := c.BasePath
		if c.ImagePath != "" {
			c.ImagePath = strings.Trim(c.ImagePath, "/")
			path = fmt.Sprintf("%s/%s", c.BasePath, c.ImagePath)
		}

		c.FullImagePath = path
	}

	if c.FullImagePath == "" {
		c.FullImagePath = c.ImagePath
	}

	if c.ConfigName != "" {
		c.ConfigName = env.Replace(c.ConfigName)
	}

	for key, value := range c.Arguments {
		c.Arguments[key] = env.Replace(value)
	}

	for key, value := range c.EnvironmentVariables {
		c.EnvironmentVariables[key] = env.Replace(value)
	}
}
