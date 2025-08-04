package infrastructureworker

// import (
// 	"errors"
// 	"fmt"
// 	"os"
// 	"strings"

// 	"github.com/cjlapao/locally-cli/internal/common"
// 	"github.com/cjlapao/locally-cli/internal/context/pipeline_component"
// 	"github.com/cjlapao/locally-cli/internal/infrastructure"
// 	"github.com/cjlapao/locally-cli/internal/lanes/entities"
// 	"github.com/cjlapao/locally-cli/internal/lanes/interfaces"
// 	"github.com/cjlapao/locally-cli/internal/notifications"
// 	"github.com/cjlapao/locally-cli/internal/operations"

// 	"gopkg.in/yaml.v3"
// )

// var notify = notifications.Get()

// const (
// 	ErrorInvalidParameters = "400"
// 	ErrorInvalidCommand    = "401"
// )

// type InfrastructurePipelineWorker struct {
// 	name string
// }

// func (worker InfrastructurePipelineWorker) New() interfaces.PipelineWorker {
// 	return InfrastructurePipelineWorker{
// 		name: "infrastructure.worker",
// 	}
// }

// func (worker InfrastructurePipelineWorker) Name() string {
// 	return worker.name
// }

// func (worker InfrastructurePipelineWorker) Run(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
// 	result := entities.PipelineWorkerResult{}

// 	if task.Type != pipeline_component.InfrastructureTask {
// 		notify.Debug("[%s] %s: This is not a task for me, bye...", worker.name, task.Name)
// 		result.State = entities.StateIgnored
// 		return result
// 	}

// 	notify.Debug("[%s] picked up task %s to work on", worker.name, task.Name)

// 	validationResult := worker.Validate(task)
// 	if validationResult.State != entities.StateValid {
// 		return validationResult
// 	}

// 	inputs, err := worker.parseParameters(task)
// 	if err != nil {
// 		return entities.NewPipelineWorkerResultFromError(ErrorInvalidParameters, err)
// 	}

// 	inputs.Decode()

// 	oldOsArgs := os.Args
// 	if len(inputs.Arguments) > 0 {
// 		os.Args = append(os.Args, inputs.Arguments...)
// 	}

// 	notify.Debug("Run arguments: %s", strings.Join(os.Args, ","))
// 	notify.Reset()

// 	options := &infrastructure.TerraformServiceOptions{
// 		Name:              inputs.StackName,
// 		BuildDependencies: inputs.BuildDependencies,
// 		RootFolder:        inputs.WorkingDirectory,
// 	}

// 	notify.Debug("Running infrastructure with options: %s", fmt.Sprintf("%v", options))

// 	operations.InfrastructureOperations(inputs.Command, inputs.StackName, options)

// 	if len(inputs.Arguments) > 0 {
// 		os.Args = oldOsArgs
// 	}

// 	if notify.HasErrors() {
// 		return entities.NewPipelineWorkerResultFromError(ErrorInvalidCommand, errors.New("error running infrastructure"))
// 	}

// 	msg := fmt.Sprintf("Infrastructure executed successfully for task %s", task.Name)
// 	if common.IsDebug() {
// 		msg = fmt.Sprintf("[%s] %s", worker.name, msg)
// 	}

// 	notify.Success(msg)

// 	result.State = entities.StateExecuted

// 	return result
// }

// func (worker InfrastructurePipelineWorker) Validate(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
// 	result := entities.PipelineWorkerResult{}
// 	if task.Type != pipeline_component.InfrastructureTask {
// 		result.State = entities.StateIgnored
// 		return result
// 	}

// 	validationResult, err := worker.parseParameters(task)
// 	if err != nil {
// 		result.State = entities.StateErrored
// 		result.Error = err
// 	}

// 	if !validationResult.Validate() {
// 		result.State = entities.StateErrored
// 		result.Error = errors.New("failed validation")
// 	}

// 	result.State = entities.StateValid
// 	return result
// }

// func (worker InfrastructurePipelineWorker) parseParameters(task *pipeline_component.PipelineTask) (*InfrastructureParameters, error) {
// 	encoded, err := yaml.Marshal(task.Inputs)
// 	if err != nil {
// 		return nil, err
// 	}
// 	var inputs InfrastructureParameters
// 	err = yaml.Unmarshal(encoded, &inputs)
// 	if err != nil {
// 		return nil, err
// 	}

// 	return &inputs, nil
// }
