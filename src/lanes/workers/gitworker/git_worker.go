package gitworker

import (
	"errors"
	"fmt"
	"strings"

	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/git"
	"github.com/cjlapao/locally-cli/lanes/entities"
	"github.com/cjlapao/locally-cli/lanes/interfaces"
	"github.com/cjlapao/locally-cli/mappers"
	"github.com/cjlapao/locally-cli/notifications"

	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

var notify = notifications.Get()

const (
	ErrorExecuting         = "500"
	ErrorInvalidParameters = "400"
)

type GitPipelineWorker struct {
	name string
}

func (worker GitPipelineWorker) New() interfaces.PipelineWorker {
	return &GitPipelineWorker{
		name: "git.worker",
	}
}

func (worker GitPipelineWorker) Name() string {
	return worker.name
}

func (worker GitPipelineWorker) Run(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	config := configuration.Get()
	git := git.Get()

	result := entities.PipelineWorkerResult{}

	if task.Type != configuration.GitTask {
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

	if inputs.Destination == "" {
		sources := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.SOURCES_PATH)
		repoFolder := worker.extractRepoName(inputs.RepoUrl)
		inputs.Destination = helper.JoinPath(sources, repoFolder)
	}
	notify.Debug("Cloning repo %s to %s", inputs.RepoUrl, inputs.Destination)
	if inputs.Credentials != nil {
		mappers.DecodeGitCredentials(inputs.Credentials)
		if err := git.CloneWithCredentials(inputs.RepoUrl, inputs.Destination, inputs.Credentials, inputs.Clean); err != nil {
			notify.Error(err.Error())
			return entities.NewPipelineWorkerResultFromError(ErrorExecuting, err)
		}
	} else {
		if err := git.Clone(inputs.RepoUrl, inputs.Destination, inputs.Clean); err != nil {
			notify.Error(err.Error())
			return entities.NewPipelineWorkerResultFromError(ErrorExecuting, err)
		}
	}

	msg := fmt.Sprintf("Git executed successfully for task %s", task.Name)
	if config.Debug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)
	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker GitPipelineWorker) Validate(task *configuration.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != configuration.GitTask {
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

func (worker GitPipelineWorker) extractRepoName(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) == 1 {
		return url
	}

	name := parts[len(parts)-1]
	name = strings.ReplaceAll(name, ".git", "")
	return name
}

func (worker GitPipelineWorker) parseParameters(task *configuration.PipelineTask) (*GitParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}
	var inputs GitParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
