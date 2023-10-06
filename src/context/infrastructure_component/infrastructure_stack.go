package infrastructure_component

import (
	"strings"
	"time"

	"github.com/cjlapao/locally-cli/context/entities"
	"github.com/cjlapao/locally-cli/context/git_component"
)

type InfrastructureStack struct {
	Source          string                            `json:"-" yaml:"-"`
	CurrentLocation string                            `json:"-" yaml:"-"`
	LastApplied     *time.Time                        `json:"lastApplied,omitempty" yaml:"lastApplied,omitempty"`
	Name            string                            `json:"name,omitempty" yaml:"name,omitempty"`
	DependsOn       []string                          `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	RequiredBy      []string                          `json:"requiredBy,omitempty" yaml:"requiredBy,omitempty"`
	RequiredStates  []string                          `json:"requiredStates,omitempty" yaml:"requiredStates,omitempty"`
	Location        *entities.Location                `json:"location,omitempty" yaml:"location,omitempty"`
	Repository      *git_component.GitCloneRepository `json:"repository,omitempty" yaml:"repository,omitempty"`
	VariableFile    string                            `json:"-" yaml:"-"`
	Variables       map[string]interface{}            `json:"variables,omitempty" yaml:"variables,omitempty"`
	Backend         *InfrastructureAzureBackend       `json:"backend,omitempty" yaml:"backend,omitempty"`
	Tags            []string                          `json:"tags,omitempty" yaml:"tags,omitempty"`
}

func (stack *InfrastructureStack) GetName() string {
	return stack.Name
}

func (stack *InfrastructureStack) GetDependencies() []string {
	return stack.DependsOn
}

func (stack *InfrastructureStack) GetSource() string {
	return stack.Source
}

func (stack *InfrastructureStack) AddDependency(value string) {
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

func (stack *InfrastructureStack) AddRequiredBy(value string) {
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

func (stack *InfrastructureStack) AddRequiredState(value string) {
	if stack.RequiredStates == nil {
		stack.RequiredStates = make([]string, 0)
	}

	found := false
	for _, d := range stack.RequiredStates {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		stack.RequiredStates = append(stack.RequiredStates, value)
	}
}

func (infra *InfrastructureStack) BuildDependency() error {
	return nil
}

func (infra *InfrastructureStack) SaveFragment() error {
	return nil
}
