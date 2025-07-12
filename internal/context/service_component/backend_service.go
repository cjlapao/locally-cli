package service_component

import (
	"strings"

	"github.com/cjlapao/locally-cli/internal/context/docker_component"
	"github.com/cjlapao/locally-cli/internal/context/entities"
	"github.com/cjlapao/locally-cli/internal/context/git_component"
)

type BackendService struct {
	Source         string                            `json:"-" yaml:"-"`
	Name           string                            `json:"name,omitempty" yaml:"name,omitempty"`
	Location       *entities.Location                `json:"location,omitempty" yaml:"location,omitempty"`
	DockerRegistry *docker_component.DockerRegistry  `json:"dockerRegistry,omitempty" yaml:"dockerRegistry,omitempty"`
	DockerCompose  *docker_component.DockerCompose   `json:"dockerCompose,omitempty" yaml:"dockerCompose,omitempty"`
	Repository     *git_component.GitCloneRepository `json:"repository,omitempty" yaml:"repository,omitempty"`
	URI            string                            `json:"uri,omitempty" yaml:"uri,omitempty"`
	AllowedOrigins []string                          `json:"allowedOrigins,omitempty" yaml:"allowedOrigins,omitempty"`
	Components     []*BackendComponent               `json:"components,omitempty" yaml:"components,omitempty"`
	Tags           []string                          `json:"tags,omitempty" yaml:"tags,omitempty"`
	DependsOn      []string                          `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	RequiredBy     []string                          `json:"-" yaml:"-"`
}

func (svc *BackendService) GetName() string {
	return svc.Name
}

func (svc *BackendService) GetDependencies() []string {
	return svc.DependsOn
}

func (svc *BackendService) GetSource() string {
	return svc.Source
}

func (stack *BackendService) AddDependency(value string) {
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

func (stack *BackendService) AddRequiredBy(value string) {
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

func (task *BackendService) SaveFragment() error {
	return nil
}

func (svc *BackendService) AllComponentsHaveManifestPath() bool {
	for _, component := range svc.Components {
		if component.ManifestPath == "" {
			return false
		}
	}

	return true
}

func (svc *BackendService) HasPath() bool {
	if svc.Location != nil && svc.Location.RootFolder != "" {
		// Local debugging setup - Repo already cloned. Try to find docker compose in it
		return true
	}

	if svc.Repository != nil && svc.Repository.Enabled {
		// Local debugging setup - Clone the repo and try to find docker compose in it
		return true
	}

	if svc.DockerCompose != nil &&
		svc.DockerRegistry != nil &&
		svc.DockerRegistry.Enabled &&
		svc.AllComponentsHaveManifestPath() &&
		svc.DockerRegistry.Registry != "" {

		// Not doing a local build. Using image from ACR. We will be building a docker compose file
		return true
	}

	return false
}
