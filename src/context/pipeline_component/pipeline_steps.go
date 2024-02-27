package pipeline_component

type PipelineStep struct {
	Tasks []*PipelineTask `json:"tasks,omitempty" yaml:"tasks,omitempty"`
}
