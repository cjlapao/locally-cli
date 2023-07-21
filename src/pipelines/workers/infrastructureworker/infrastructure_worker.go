package infrastructureworker

import (
	"errors"
	"fmt"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/infrastructure"
	"github.com/cjlapao/locally-cli/notifications"
	"github.com/cjlapao/locally-cli/operations"
	"github.com/cjlapao/locally-cli/pipelines/entities"
	"github.com/cjlapao/locally-cli/pipelines/interfaces"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

var notify = notifications.Get()

const (
	ErrorInvalidParameters = "400"
	ErrorInvalidCommand    = "401"
)

type InfrastructurePipelineWorker struct {
	name string
}

func (worker InfrastructurePipelineWorker) New() interfaces.PipelineWorker {
	return InfrastructurePipelineWorker{
		name: "infrastructure.worker",
	}
}

func (worker InfrastructurePipelineWorker) Name() string {
	return worker.name
}

func (worker InfrastructurePipelineWorker) Run(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	config := configuration.Get()
	result := entities.PipelineWorkerResult{}

	if task.Type != configuration.InfrastructureTask {
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

	oldOsArgs := os.Args
	if len(inputs.Arguments) > 0 {
		os.Args = append(os.Args, inputs.Arguments...)
	}

	notify.Debug("Run arguments: %s", strings.Join(os.Args, ","))
	notify.Reset()

	options := &infrastructure.TerraformServiceOptions{
		Name:              inputs.StackName,
		BuildDependencies: inputs.BuildDependencies,
		RootFolder:        inputs.WorkingDirectory,
	}

	notify.Debug("Running infrastructure with options: %s", fmt.Sprintf("%v", options))

	operations.InfrastructureOperations(inputs.Command, inputs.StackName, options)

	if len(inputs.Arguments) > 0 {
		os.Args = oldOsArgs
	}

	if notify.HasErrors() {
		return entities.NewPipelineWorkerResultFromError(ErrorInvalidCommand, errors.New("error running infrastructure"))
	}

	msg := fmt.Sprintf("Infrastructure executed successfully for task %s", task.Name)
	if config.Debug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)
	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker InfrastructurePipelineWorker) Validate(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != configuration.InfrastructureTask {
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

func (worker InfrastructurePipelineWorker) parseParameters(task *configuration.PipelineTask) (*InfrastructureParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}
	var inputs InfrastructureParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
