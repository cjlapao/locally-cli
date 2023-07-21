package configuration

import (
	"strings"
)

type PipelineTask struct {
	source              string
	Disabled            bool                   `json:"disabled" yaml:"disabled"`
	Name                string                 `json:"name,omitempty" yaml:"name,omitempty"`
	Type                PipelineTaskType       `json:"type,omitempty" yaml:"type,omitempty"`
	RetryCountOnFailure int                    `json:"retryCountOnFailure,omitempty" yaml:"retryCountOnFailure,omitempty"`
	WorkingDirectory    string                 `json:"workingDirectory,omitempty" yaml:"workingDirectory,omitempty"`
	Inputs              map[string]interface{} `json:"inputs,omitempty" yaml:"inputs,omitempty"`
	Body                string                 `json:"body,omitempty" yaml:"body,omitempty"`
	DependsOn           []string               `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	RequiredBy          []string               `json:"-" yaml:"-"`
}

func (task *PipelineTask) GetName() string {
	return task.Name
}

func (task *PipelineTask) GetDependencies() []string {
	return task.DependsOn
}

func (task *PipelineTask) GetSource() string {
	return task.source
}

func (task *PipelineTask) AddDependency(value string) {
	if task.DependsOn == nil {
		task.DependsOn = make([]string, 0)
	}

	found := false
	for _, d := range task.DependsOn {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		task.DependsOn = append(task.DependsOn, value)
	}
}

func (task *PipelineTask) AddRequiredBy(value string) {
	if task.RequiredBy == nil {
		task.RequiredBy = make([]string, 0)
	}

	found := false
	for _, d := range task.RequiredBy {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		task.RequiredBy = append(task.RequiredBy, value)
	}
}

func (task *PipelineTask) BuildDependency() error {
	return nil
}

func (task *PipelineTask) SaveFragment() error {
	return nil
}
