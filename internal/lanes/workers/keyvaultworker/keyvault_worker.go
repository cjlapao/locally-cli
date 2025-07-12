package keyvaultworker

import (
	"errors"
	"fmt"

	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/context/pipeline_component"
	"github.com/cjlapao/locally-cli/internal/lanes/entities"
	"github.com/cjlapao/locally-cli/internal/lanes/interfaces"
	"github.com/cjlapao/locally-cli/internal/notifications"
	"github.com/cjlapao/locally-cli/internal/vaults/azure_keyvault"

	"gopkg.in/yaml.v3"
)

var notify = notifications.Get()

const (
	ErrorExecuting         = "500"
	ErrorInvalidParameters = "400"
)

type KeyvaultPipelineWorker struct {
	name string
}

func (worker KeyvaultPipelineWorker) New() interfaces.PipelineWorker {
	return KeyvaultPipelineWorker{
		name: "keyvault.worker",
	}
}

func (worker KeyvaultPipelineWorker) Name() string {
	return worker.name
}

func (worker KeyvaultPipelineWorker) Run(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}

	if task.Type != pipeline_component.KeyvaultSyncTask {
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

	notify.Debug("Starting to sync the keyvault on %s", inputs.KeyvaultUrl)
	kvSync := azure_keyvault.New(inputs.Name, &azure_keyvault.AzureKeyVaultOptions{
		KeyVaultUri:  inputs.KeyvaultUrl,
		DecodeBase64: inputs.Base64Decode,
	})

	if _, err := kvSync.Sync(); err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorExecuting, err)
	}

	msg := fmt.Sprintf("Azure KeyVault sync executed successfully for task %s", task.Name)
	if common.IsDebug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)
	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker KeyvaultPipelineWorker) Validate(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != pipeline_component.KeyvaultSyncTask {
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

func (worker KeyvaultPipelineWorker) parseParameters(task *pipeline_component.PipelineTask) (*KeyvaultParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}
	var inputs KeyvaultParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
