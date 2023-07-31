package whatsnewworker

import (
	"errors"
	"fmt"

	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/environment"
	"github.com/cjlapao/locally-cli/lanes/common"
	"github.com/cjlapao/locally-cli/lanes/entities"
	"github.com/cjlapao/locally-cli/lanes/interfaces"
	"github.com/cjlapao/locally-cli/notifications"

	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

var notify = notifications.Get()

const (
	ErrorInvalidParameters   = "500"
	ErrorReadingFile         = "501"
	ErrorRegisteringManifest = "502"
	ErrorRetrievingOpsToken  = "503"
)

type WhatsNewPipelineWorker struct {
	name string
}

func (worker WhatsNewPipelineWorker) New() interfaces.PipelineWorker {
	return WhatsNewPipelineWorker{
		name: "whatsnew.worker",
	}
}

func (worker WhatsNewPipelineWorker) Name() string {
	return worker.name
}

func (worker WhatsNewPipelineWorker) Run(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	env := environment.Get()
	result := entities.PipelineWorkerResult{}

	if task.Type != configuration.WhatsNewTask {
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

	content, err := helper.ReadFromFile(inputs.Path)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorReadingFile, err)
	}

	// Get Ops Token
	access_token, err := common.RequestOpsToken(env.Replace("${{ config.context.baseUrl }}"))
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorRetrievingOpsToken, err)
	}

	// Making the WhatsNew API request
	url := env.Replace("${{ config.context.baseUrl }}") + "/api/user/ops/whatsnew"
	status_code, err := common.SendPostRequest(access_token, content, url)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorRegisteringManifest, err)
	}

	// Return status of the request
	msg := ""
	if status_code == "204" || status_code == "200" {
		msg = fmt.Sprintf("Whats New Registration executed successfully for task %s, with status code %s", task.Name, status_code)
	} else {
		err := errors.New(fmt.Sprintf("Whats New Registration executed but returned with status code %s", status_code))
		return entities.NewPipelineWorkerResultFromError(ErrorRegisteringManifest, err)
	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker WhatsNewPipelineWorker) Validate(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}

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

func (worker WhatsNewPipelineWorker) parseParameters(task *configuration.PipelineTask) (*WhatsNewPipelineWorkerParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}

	var inputs WhatsNewPipelineWorkerParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
