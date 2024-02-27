package docker_component

import (
	"strings"

	"github.com/cjlapao/locally-cli/context/entities"
	"github.com/cjlapao/locally-cli/context/git_component"
)

type DockerContainer struct {
	Source               string                            `json:"-" yaml:"-"`
	Name                 string                            `json:"name,omitempty" yaml:"name,omitempty"`
	Location             *entities.Location                `json:"location,omitempty" yaml:"location,omitempty"`
	Repository           *git_component.GitCloneRepository `json:"repository,omitempty" yaml:"repository,omitempty"`
	Environment          map[string]interface{}            `json:"environment,omitempty" yaml:"environment,omitempty"`
	EnvironmentVariables map[string]string                 `json:"environmentVariables,omitempty" yaml:"environmentVariables,omitempty"`
	ManifestPath         string                            `json:"manifestPath,omitempty" yaml:"manifestPath,omitempty"`
	ManifestTag          string                            `json:"tag,omitempty" yaml:"tag,omitempty"`
	BuildArguments       []string                          `json:"buildArguments,omitempty" yaml:"buildArguments,omitempty"`
	Tags                 []string                          `json:"tags,omitempty" yaml:"tags,omitempty"`
	Components           []*DockerContainer                `json:"components" yaml:"components"`
	DockerRegistry       *DockerRegistry                   `json:"dockerRegistry,omitempty" yaml:"dockerRegistry,omitempty"`
	DockerCompose        *DockerCompose                    `json:"dockerCompose,omitempty" yaml:"dockerCompose,omitempty"`
	DependsOn            []string                          `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	RequiredBy           []string                          `json:"-" yaml:"-"`
}

func (svc *DockerContainer) GetName() string {
	return svc.Name
}

func (svc *DockerContainer) GetDependencies() []string {
	return svc.DependsOn
}

func (svc *DockerContainer) GetSource() string {
	return svc.Source
}

func (stack *DockerContainer) AddDependency(value string) {
	if stack.DependsOn == nil {
		stack.DependsOn = make([]string, 0)
	}

	found := false
	for _, d := range stack.DependsOn {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		stack.DependsOn = append(stack.DependsOn, value)
	}
}

func (stack *DockerContainer) AddRequiredBy(value string) {
	if stack.RequiredBy == nil {
		stack.RequiredBy = make([]string, 0)
	}

	found := false
	for _, d := range stack.RequiredBy {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		stack.RequiredBy = append(stack.RequiredBy, value)
	}
}

func (task *DockerContainer) SaveFragment() error {
	return nil
}

func (task *DockerContainer) GetPath() error {
	return nil
}
