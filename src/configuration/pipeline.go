package configuration

import (
	"strings"
)

type Pipeline struct {
	source     string
	Disabled   bool           `json:"disabled" yaml:"disabled"`
	Name       string         `json:"name,omitempty" yaml:"name,omitempty"`
	Jobs       []*PipelineJob `json:"jobs,omitempty" yaml:"jobs,omitempty"`
	DependsOn  []string       `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	RequiredBy []string       `json:"-" yaml:"-"`
}

func (pipeline *Pipeline) GetName() string {
	return pipeline.Name
}

func (pipeline *Pipeline) GetDependencies() []string {
	return pipeline.DependsOn
}

func (pipeline *Pipeline) GetSource() string {
	return pipeline.source
}

func (pipeline *Pipeline) AddDependency(value string) {
	if pipeline.DependsOn == nil {
		pipeline.DependsOn = make([]string, 0)
	}

	found := false
	for _, d := range pipeline.DependsOn {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		pipeline.DependsOn = append(pipeline.DependsOn, value)
	}
}

func (pipeline *Pipeline) AddRequiredBy(value string) {
	if pipeline.RequiredBy == nil {
		pipeline.RequiredBy = make([]string, 0)
	}

	found := false
	for _, d := range pipeline.RequiredBy {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		pipeline.RequiredBy = append(pipeline.RequiredBy, value)
	}
}

func (pipeline *Pipeline) BuildDependency() error {
	err := BuildDependencyGraph(pipeline.Jobs, false)
	return err
}

func (pipeline *Pipeline) SaveFragment() error {
	return nil
}
