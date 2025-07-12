package dotnetworker

import (
	"errors"
	"fmt"

	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/context/pipeline_component"
	"github.com/cjlapao/locally-cli/internal/docker"
	"github.com/cjlapao/locally-cli/internal/lanes/entities"
	"github.com/cjlapao/locally-cli/internal/lanes/interfaces"
	"github.com/cjlapao/locally-cli/internal/notifications"

	"gopkg.in/yaml.v3"
)

var notify = notifications.Get()

const (
	ErrorInvalidParameters = "500"
	ErrorRunningImage      = "400"
	ErrorDeletingImage     = "401"
	ErrorDeletingFile      = "402"
)

type DotnetPipelineWorker struct {
	name string
}

func (worker DotnetPipelineWorker) New() interfaces.PipelineWorker {
	return DotnetPipelineWorker{
		name: "dotnet.worker",
	}
}

func (worker DotnetPipelineWorker) Name() string {
	return worker.name
}

func (worker DotnetPipelineWorker) Run(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}

	if task.Type != pipeline_component.DotnetTask {
		notify.Debug("[%s] %s: This is not a task for me, bye...", worker.name, task.Name)
		result.State = entities.StateIgnored
		return result
	}

	notify.Debug("[%s] picked up task %s to work on", worker.name, task.Name)

	validationResult := worker.Validate(task)

	if validationResult.State != entities.StateValid {
		return validationResult
	}

	inputs, err := worker.parseParameters(task)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorInvalidParameters, err)
	}

	inputs.Decode()

	dockerSvc := docker.Get()

	err = dockerSvc.RunDotnetContainer(inputs.Command, inputs.Context, inputs.BaseImage, fmt.Sprintf("%s-migration", common.EncodeName(task.Name)), inputs.RepoUrl, inputs.ProjectPath, inputs.Arguments, inputs.EnvironmentVariables, inputs.BuildArguments)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorRunningImage, err)
	}

	msg := fmt.Sprintf("Dotnet %s executed successfully for task %s", inputs.Command, task.Name)
	if common.IsDebug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)
	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker DotnetPipelineWorker) Validate(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != pipeline_component.DotnetTask {
		result.State = entities.StateIgnored
		return result
	}

	validationResult, err := worker.parseParameters(task)
	if err != nil {
		result.State = entities.StateErrored
		result.Error = err
	}

	if !validationResult.Validate() {
		result.State = entities.StateErrored
		result.Error = errors.New("failed validation")
	}

	result.State = entities.StateValid
	return result
}

func (worker DotnetPipelineWorker) parseParameters(task *pipeline_component.PipelineTask) (*DotnetPipelineWorkerParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}

	var inputs DotnetPipelineWorkerParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
