package interfaces

import (
	"github.com/cjlapao/locally-cli/context/pipeline_component"
	"github.com/cjlapao/locally-cli/lanes/entities"
)

type PipelineWorker interface {
	Name() string
	New() PipelineWorker
	Run(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult
	Validate(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult
}
