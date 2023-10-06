package service_component

import (
	"strings"

	"github.com/cjlapao/locally-cli/context/docker_component"
	"github.com/cjlapao/locally-cli/context/entities"
	"github.com/cjlapao/locally-cli/context/git_component"
	"github.com/cjlapao/locally-cli/context/mock_component"
)

type SpaService struct {
	Source               string
	Default              bool                              `json:"default,omitempty" yaml:"default,omitempty"`
	Name                 string                            `json:"name,omitempty" yaml:"name,omitempty"`
	Path                 string                            `json:"path,omitempty" yaml:"path,omitempty"`
	ReverseProxyURI      string                            `json:"reverseProxyUri,omitempty" yaml:"reverseProxyUri,omitempty"`
	UseReverseProxy      bool                              `json:"useReverseProxy,omitempty" yaml:"useReverseProxy,omitempty"`
	DockerRegistry       *docker_component.DockerRegistry  `json:"dockerRegistry,omitempty" yaml:"dockerRegistry,omitempty"`
	DockerCompose        *docker_component.DockerCompose   `json:"dockerCompose,omitempty" yaml:"dockerCompose,omitempty"`
	Location             *entities.Location                `json:"location,omitempty" yaml:"location,omitempty"`
	Repository           *git_component.GitCloneRepository `json:"repository,omitempty" yaml:"repository,omitempty"`
	URI                  string                            `json:"uri,omitempty" yaml:"uri,omitempty"`
	AllowedOrigins       []string                          `json:"allowedOrigins,omitempty" yaml:"allowedOrigins,omitempty"`
	Environment          map[string]interface{}            `json:"environment,omitempty" yaml:"environment,omitempty"`
	EnvironmentVariables map[string]string                 `json:"environmentVariables,omitempty" yaml:"environmentVariables,omitempty"`
	BuildArguments       []string                          `json:"buildArguments,omitempty" yaml:"buildArguments,omitempty"`
	MockRoutes           []*mock_component.MockRoute       `json:"mockRoutes,omitempty" yaml:"mockRoutes,omitempty"`
	RouteReplace         *entities.RouteReplace            `json:"routeReplace,omitempty" yaml:"routeReplace,omitempty"`
	Tags                 []string                          `json:"tags,omitempty" yaml:"tags,omitempty"`
	DependsOn            []string                          `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	RequiredBy           []string                          `json:"-" yaml:"-"`
}

func (spaService *SpaService) GetName() string {
	return spaService.Name
}

func (svc *SpaService) GetDependencies() []string {
	return svc.DependsOn
}

func (spaService *SpaService) GetSource() string {
	return spaService.Source
}

func (spaService *SpaService) AddDependency(value string) {
	if spaService.DependsOn == nil {
		spaService.DependsOn = make([]string, 0)
	}

	found := false
	for _, d := range spaService.DependsOn {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		spaService.DependsOn = append(spaService.DependsOn, value)
	}
}

func (spaService *SpaService) AddRequiredBy(value string) {
	if spaService.RequiredBy == nil {
		spaService.RequiredBy = make([]string, 0)
	}

	found := false
	for _, d := range spaService.RequiredBy {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		spaService.RequiredBy = append(spaService.RequiredBy, value)
	}
}

func (spaService *SpaService) BuildDependency() error {
	return nil
}

func (spaService *SpaService) SaveFragment() error {
	return nil
}

func (spaService *SpaService) HasPath() bool {
	if spaService.Location != nil && spaService.Location.RootFolder != "" {
		// Local debugging setup - Repo already cloned. Try to find docker compose in it
		return true
	}

	if spaService.Repository != nil && spaService.Repository.Enabled {
		// Local debugging setup - Clone the repo and try to find docker compose in it
		return true
	}

	if spaService.DockerCompose != nil &&
		spaService.DockerRegistry != nil &&
		spaService.DockerRegistry.Enabled &&
		spaService.DockerRegistry.ManifestPath != "" &&
		spaService.DockerRegistry.Registry != "" {

		// Not doing a local build. Using image from ACR. We will be building a docker compose file
		return true
	}

	return false
}
