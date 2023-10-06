package efmigrationsworker

import (
	"errors"
	"fmt"

	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/context/pipeline_component"
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

type EFMigrationsPipelineWorker struct {
	name string
}

func (worker EFMigrationsPipelineWorker) New() interfaces.PipelineWorker {
	return EFMigrationsPipelineWorker{
		name: "ef.migrations.worker",
	}
}

func (worker EFMigrationsPipelineWorker) Name() string {
	return worker.name
}

func (worker EFMigrationsPipelineWorker) Run(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}

	if task.Type != pipeline_component.EFMigrationTask {
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

	err = dockerSvc.RunEfMigrations(inputs.Context, inputs.BaseImage, fmt.Sprintf("%s-migration", common.EncodeName(task.Name)),
		inputs.RepoUrl, inputs.ProjectPath, inputs.StartupProjectPath, inputs.DbConnectionString, inputs.Arguments, inputs.EnvironmentVariables)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorRunningImage, err)
	}

	msg := fmt.Sprintf("Ef Migrations executed successfully for task %s", task.Name)
	if common.IsDebug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)
	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker EFMigrationsPipelineWorker) Validate(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != pipeline_component.EFMigrationTask {
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

func (worker EFMigrationsPipelineWorker) parseParameters(task *pipeline_component.PipelineTask) (*EfMigrationsPipelineWorkerParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}
	var inputs EfMigrationsPipelineWorkerParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
