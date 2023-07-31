package dotnetworker

import (
	"errors"
	"fmt"

	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/docker"
	"github.com/cjlapao/locally-cli/lanes/entities"
	"github.com/cjlapao/locally-cli/lanes/interfaces"
	"github.com/cjlapao/locally-cli/notifications"

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

func (worker DotnetPipelineWorker) Run(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	config := configuration.Get()
	result := entities.PipelineWorkerResult{}

	if task.Type != configuration.DotnetTask {
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

	err = dockerSvc.RunDotnetContainer(inputs.Command, inputs.Context, inputs.BaseImage, fmt.Sprintf("%s-migration", configuration.EncodeName(task.Name)), inputs.RepoUrl, inputs.ProjectPath, inputs.Arguments, inputs.EnvironmentVariables, inputs.BuildArguments)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorRunningImage, err)
	}

	msg := fmt.Sprintf("Dotnet %s executed successfully for task %s", inputs.Command, task.Name)
	if config.Debug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)
	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker DotnetPipelineWorker) Validate(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != configuration.DotnetTask {
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

func (worker DotnetPipelineWorker) parseParameters(task *configuration.PipelineTask) (*DotnetPipelineWorkerParameters, error) {
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
