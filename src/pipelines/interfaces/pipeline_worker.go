package interfaces

import (
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/pipelines/entities"
)

type PipelineWorker interface {
	Name() string
	New() PipelineWorker
	Run(task *configuration.PipelineTask) entities.PipelineWorkerResult
	Validate(task *configuration.PipelineTask) entities.PipelineWorkerResult
}
