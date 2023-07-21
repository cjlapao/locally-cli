package configuration

import (
	"strings"
)

type Infrastructure struct {
	Source        string                       `json:"-" yaml:"-"`
	ConfigFile    string                       `json:"configFile,omitempty" yaml:"configFile,omitempty"`
	Terraform     TerraformTool                `json:"terraform,omitempty" yaml:"terraform,omitempty"`
	Authorization *InfrastructureAuthorization `json:"-" yaml:"-"`
	Stacks        []*InfrastructureStack       `json:"stacks,omitempty" yaml:"stacks,omitempty"`
}

func (infra *Infrastructure) GetStackByBackend(backendKey string) *InfrastructureStack {
	if infra.Stacks == nil {
		return nil
	}

	for _, s := range infra.Stacks {
		if s.Backend != nil {
			stateBackend := s.Backend.StateFileName
			if strings.HasSuffix(s.Backend.StateFileName, ".tfstate") {
				stateBackend = strings.ReplaceAll(stateBackend, ".tfstate", "")
			}
			notify.Debug("Stack %s -> logging for %s and stateFile is %s ", s.Name, backendKey, stateBackend)
			if strings.EqualFold(stateBackend, backendKey) {
				return s
			}
		}
	}

	return nil
}

func (infra *Infrastructure) GetStackByName(name string) *InfrastructureStack {
	if infra.Stacks == nil {
		return nil
	}

	for _, s := range infra.Stacks {
		if strings.EqualFold(s.Name, name) {
			return s
		}
	}

	return nil
}

func (infra *Infrastructure) AddStack(stack *InfrastructureStack) {
	if infra.Stacks == nil {
		infra.Stacks = make([]*InfrastructureStack, 0)
	}

	found := false
	for _, s := range infra.Stacks {
		if strings.EqualFold(s.Name, stack.Name) {
			found = true
			break
		}
	}

	if !found {
		infra.Stacks = append(infra.Stacks, stack)
	}
}

func (infra *Infrastructure) BuildDependency() error {
	err := BuildDependencyGraph(infra.Stacks, false)
	return err
}

func (infra *Infrastructure) SaveFragment() error {
	return nil
}
