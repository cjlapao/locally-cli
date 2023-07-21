package configuration

type PipelineStep struct {
	Tasks []*PipelineTask `json:"tasks,omitempty" yaml:"tasks,omitempty"`
}
