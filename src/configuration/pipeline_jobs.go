package configuration

import (
	"strings"
)

type PipelineJob struct {
	source     string
	Disabled   bool            `json:"disabled" yaml:"disabled"`
	Name       string          `json:"name,omitempty" yaml:"name,omitempty"`
	Steps      []*PipelineTask `json:"steps,omitempty" yaml:"steps,omitempty"`
	DependsOn  []string        `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	RequiredBy []string        `json:"-" yaml:"-"`
}

func (job *PipelineJob) GetName() string {
	return job.Name
}

func (job *PipelineJob) GetDependencies() []string {
	return job.DependsOn
}

func (job *PipelineJob) GetSource() string {
	return job.source
}

func (job *PipelineJob) AddDependency(value string) {
	if job.DependsOn == nil {
		job.DependsOn = make([]string, 0)
	}

	found := false
	for _, d := range job.DependsOn {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		job.DependsOn = append(job.DependsOn, value)
	}
}

func (job *PipelineJob) AddRequiredBy(value string) {
	if job.RequiredBy == nil {
		job.RequiredBy = make([]string, 0)
	}

	found := false
	for _, d := range job.RequiredBy {
		if strings.EqualFold(d, value) {
			found = true
			break
		}
	}

	if !found {
		job.RequiredBy = append(job.RequiredBy, value)
	}
}

func (job *PipelineJob) BuildDependency() error {
	err := BuildDependencyGraph(job.Steps, false)
	return err
}

func (job *PipelineJob) SaveFragment() error {
	return nil
}
