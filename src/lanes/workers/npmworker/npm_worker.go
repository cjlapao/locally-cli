package npmworker

import (
	"errors"
	"fmt"
	"os"

	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/help"
	"github.com/cjlapao/locally-cli/lanes/entities"
	"github.com/cjlapao/locally-cli/lanes/interfaces"
	"github.com/cjlapao/locally-cli/notifications"
	"github.com/cjlapao/locally-cli/npm"

	"gopkg.in/yaml.v3"
)

var notify = notifications.Get()

const (
	ErrorExecuting         = "500"
	ErrorInvalidParameters = "400"
)

type NpmPipelineWorker struct {
	name string
}

func (worker NpmPipelineWorker) New() interfaces.PipelineWorker {
	return NpmPipelineWorker{
		name: "npm.worker",
	}
}

func (worker NpmPipelineWorker) Name() string {
	return worker.name
}

func (worker NpmPipelineWorker) Run(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	config := configuration.Get()
	result := entities.PipelineWorkerResult{}

	if task.Type != configuration.NpmTask {
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

	npmSvc := npm.Get()
	npmSvc.CheckForNpm(false)

	switch inputs.Command {
	case "ci":
		err := npmSvc.CI(inputs.WorkingDir, inputs.MinVersion)
		if err != nil {
			return entities.NewPipelineWorkerResultFromError(ErrorExecuting, err)
		}
	case "install":
		err := npmSvc.Install(inputs.WorkingDir, inputs.MinVersion)
		if err != nil {
			return entities.NewPipelineWorkerResultFromError(ErrorExecuting, err)
		}
	case "publish":
		err := npmSvc.Publish(inputs.WorkingDir, inputs.MinVersion)
		if err != nil {
			return entities.NewPipelineWorkerResultFromError(ErrorExecuting, err)
		}
	case "custom":
		err := npmSvc.Custom(inputs.CustomCommand, inputs.WorkingDir, inputs.MinVersion)
		if err != nil {
			return entities.NewPipelineWorkerResultFromError(ErrorExecuting, err)
		}
	default:
		help.ShowHelpForInfrastructureCommand()
		os.Exit(0)
	}

	msg := fmt.Sprintf("Npm %s executed successfully for task %s", inputs.Command, task.Name)
	if config.Debug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)
	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker NpmPipelineWorker) Validate(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != configuration.NpmTask {
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

func (worker NpmPipelineWorker) parseParameters(task *configuration.PipelineTask) (*NpmPipelineWorkerParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}

	var inputs NpmPipelineWorkerParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
