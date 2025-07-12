package interfaces

import (
	"github.com/cjlapao/locally-cli/internal/context/pipeline_component"
	"github.com/cjlapao/locally-cli/internal/lanes/entities"
)

type PipelineWorker interface {
	Name() string
	New() PipelineWorker
	Run(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult
	Validate(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult
}
