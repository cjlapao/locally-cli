package interfaces

import (
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/lanes/entities"
)

type PipelineWorker interface {
	Name() string
	New() PipelineWorker
	Run(task *configuration.PipelineTask) entities.PipelineWorkerResult
	Validate(task *configuration.PipelineTask) entities.PipelineWorkerResult
}
