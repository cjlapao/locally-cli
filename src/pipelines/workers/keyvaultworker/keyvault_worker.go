package keyvaultworker

import (
	"errors"
	"fmt"
	"github.com/cjlapao/locally-cli/azure_keyvault"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/notifications"
	"github.com/cjlapao/locally-cli/pipelines/entities"
	"github.com/cjlapao/locally-cli/pipelines/interfaces"

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

func (worker KeyvaultPipelineWorker) Run(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	config := configuration.Get()
	result := entities.PipelineWorkerResult{}

	if task.Type != configuration.KeyvaultSyncTask {
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
	if config.Debug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)
	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker KeyvaultPipelineWorker) Validate(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != configuration.KeyvaultSyncTask {
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

func (worker KeyvaultPipelineWorker) parseParameters(task *configuration.PipelineTask) (*KeyvaultParameters, error) {
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
