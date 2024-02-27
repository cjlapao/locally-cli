package bashworker

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/context/pipeline_component"
	"github.com/cjlapao/locally-cli/executer"
	"github.com/cjlapao/locally-cli/lanes/entities"
	"github.com/cjlapao/locally-cli/lanes/interfaces"
	"github.com/cjlapao/locally-cli/lanes/retry"
	"github.com/cjlapao/locally-cli/notifications"

	"gopkg.in/yaml.v3"
)

var notify = notifications.Get()

const (
	ErrorExecuting         = "500"
	ErrorInvalidParameters = "400"
)

type BashPipelineWorker struct {
	name string
}

func (worker BashPipelineWorker) New() interfaces.PipelineWorker {
	return BashPipelineWorker{
		name: "bash.worker",
	}
}

func (worker BashPipelineWorker) Name() string {
	return worker.name
}

func (worker BashPipelineWorker) Run(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}

	if task.Type != pipeline_component.BashTask {
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

	result = retry.RetryRun(task, worker.runTask, inputs.RetryCount, inputs.WaitForInSeconds)

	msg := fmt.Sprintf("Command executed successfully for task %s", task.Name)

	if inputs.WorkingDirectory != "" {
		msg += fmt.Sprintf(" in folder %s", inputs.WorkingDirectory)
	}

	if common.IsDebug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)
	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker BashPipelineWorker) runTask(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}

	inputs, err := worker.parseParameters(task)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorInvalidParameters, err)
	}

	inputs.Decode()

	var changeDirErr error
	currentFolder := ""

	if inputs.WorkingDirectory != "" {
		currentFolder, changeDirErr = os.Getwd()
		if changeDirErr != nil {
			return entities.NewPipelineWorkerResultFromError(ErrorInvalidParameters, changeDirErr)
		}
		changeDirErr = os.Chdir(inputs.WorkingDirectory)
		if changeDirErr != nil {
			return entities.NewPipelineWorkerResultFromError(ErrorInvalidParameters, changeDirErr)
		}
	}

	notify.Debug("Run arguments: %s", strings.Join(inputs.Arguments, ","))
	output, err := executer.ExecuteAndWatch(inputs.Command, inputs.Arguments...)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorExecuting, err)
	}

	if inputs.WorkingDirectory != "" {
		changeDirErr = os.Chdir(currentFolder)
		if changeDirErr != nil {
			return entities.NewPipelineWorkerResultFromError(ErrorInvalidParameters, changeDirErr)
		}
	}

	result.Output = output.GetAllOutput()
	return result
}

func (worker BashPipelineWorker) Validate(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != pipeline_component.BashTask {
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

func (worker BashPipelineWorker) parseParameters(task *pipeline_component.PipelineTask) (*BashParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}
	var inputs BashParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
