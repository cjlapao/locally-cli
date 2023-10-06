package service_component

import (
	"strings"

	"github.com/cjlapao/locally-cli/context/entities"
	"github.com/cjlapao/locally-cli/context/mock_component"
)

type BackendComponent struct {
	Source               string                      `json:"-" yaml:"-"`
	Name                 string                      `json:"name,omitempty" yaml:"name,omitempty"`
	EnvironmentVariables map[string]string           `json:"environmentVariables,omitempty" yaml:"environmentVariables,omitempty"`
	BuildArguments       []string                    `json:"buildArguments,omitempty" yaml:"buildArguments,omitempty"`
	ReverseProxyURI      string                      `json:"reverseProxyUri,omitempty" yaml:"reverseProxyUri,omitempty"`
	ManifestPath         string                      `json:"manifestPath,omitempty" yaml:"manifestPath,omitempty"`
	ManifestTag          string                      `json:"tag,omitempty" yaml:"tag,omitempty"`
	Routes               []entities.Route            `json:"routes,omitempty" yaml:"routes,omitempty"`
	MockRoutes           []*mock_component.MockRoute `json:"mockRoutes,omitempty" yaml:"mockRoutes,omitempty"`
	DependsOn            []string                    `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	RequiredBy           []string                    `json:"-" yaml:"-"`
}

func (backendComponent *BackendComponent) GetName() string {
	return backendComponent.Name
}

func (backendComponent *BackendComponent) GetDependencies() []string {
	return backendComponent.DependsOn
}

func (backendComponent *BackendComponent) GetSource() string {
	return backendComponent.Source
}

func (backendComponent *BackendComponent) AddDependency(value string) {
	if backendComponent.DependsOn == nil {
		backendComponent.DependsOn = make([]string, 0)
	}

	found := false
	for _, d := range backendComponent.DependsOn {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		backendComponent.DependsOn = append(backendComponent.DependsOn, value)
	}
}

func (backendComponent *BackendComponent) AddRequiredBy(value string) {
	if backendComponent.RequiredBy == nil {
		backendComponent.RequiredBy = make([]string, 0)
	}

	found := false
	for _, d := range backendComponent.RequiredBy {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		backendComponent.RequiredBy = append(backendComponent.RequiredBy, value)
	}
}

func (backendComponent *BackendComponent) BuildDependency() error {
	return nil
}

func (backendComponent *BackendComponent) SaveFragment() error {
	return nil
}
