package dockerworker

import (
	"errors"
	"fmt"

	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/context/docker_component"
	"github.com/cjlapao/locally-cli/context/pipeline_component"
	"github.com/cjlapao/locally-cli/docker"
	"github.com/cjlapao/locally-cli/lanes/entities"
	"github.com/cjlapao/locally-cli/lanes/interfaces"
	"github.com/cjlapao/locally-cli/notifications"
	"github.com/cjlapao/locally-cli/operations"

	_ "github.com/microsoft/go-mssqldb"
	"gopkg.in/yaml.v3"
)

var notify = notifications.Get()

const (
	ErrorInvalidParameters = "400"
	ErrorInvalidLogin      = "401"
	ErrorExecutingWrapper  = "4012"
)

type DockerPipelineWorker struct {
	name string
}

func (worker DockerPipelineWorker) New() interfaces.PipelineWorker {
	return DockerPipelineWorker{
		name: "docker.worker",
	}
}

func (worker DockerPipelineWorker) Name() string {
	return worker.name
}

func (worker DockerPipelineWorker) Run(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}

	if task.Type != pipeline_component.DockerTask {
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

	options := docker.DockerServiceOptions{
		Name:          inputs.ConfigName,
		ComponentName: inputs.ComponentName,
		DockerRegistry: &docker_component.DockerRegistry{
			Registry:     inputs.Registry,
			BasePath:     inputs.BasePath,
			ManifestPath: inputs.ImagePath,
			Enabled:      true,
			Tag:          inputs.ImageTag,
			Credentials: &docker_component.DockerRegistryCredentials{
				Username:       inputs.Username,
				Password:       inputs.Password,
				SubscriptionId: inputs.SubscriptionId,
				TenantId:       inputs.TenantId,
			},
		},
		DockerCompose: inputs.DockerCompose,
	}

	notify.Debug("Command: %s", inputs.Command)
	notify.Debug("Command: %s", fmt.Sprintf("%v", inputs))
	notify.Debug("Options: %s", fmt.Sprintf("%v", options))
	notify.Debug("Docker Registry: %s", fmt.Sprintf("%v", options.DockerRegistry))
	notify.Debug("Docker Compose: %s", fmt.Sprintf("%v", options.DockerCompose))
	notify.Reset()
	operations.DockerOperations(inputs.Command, &options)
	if notify.HasErrors() {
		return entities.NewPipelineWorkerResultFromError(ErrorInvalidParameters, err)
	}

	msg := fmt.Sprintf("Docker executed successfully for task %s", task.Name)
	if common.IsDebug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)

	}
	notify.Success(msg)

	result.State = entities.StateExecuted
	return result
}

func (worker DockerPipelineWorker) Validate(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != pipeline_component.DockerTask {
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

func (worker DockerPipelineWorker) parseParameters(task *pipeline_component.PipelineTask) (*DockerParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}
	var inputs DockerParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
