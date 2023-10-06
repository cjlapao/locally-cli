package curlworker

import (
	"errors"
	"fmt"
	"io"

	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/context/pipeline_component"
	"github.com/cjlapao/locally-cli/environment"
	"github.com/cjlapao/locally-cli/lanes/entities"
	"github.com/cjlapao/locally-cli/lanes/interfaces"
	"github.com/cjlapao/locally-cli/lanes/retry"
	"github.com/cjlapao/locally-cli/notifications"

	"net/http"
	"net/url"
	"strings"

	_ "github.com/microsoft/go-mssqldb"
	"gopkg.in/yaml.v3"
)

var notify = notifications.Get()

const (
	ErrorInvalidParameters = "500"
	ErrorInvalidConnection = "501"
)

type CurlPipelineWorker struct {
	name string
}

func (worker CurlPipelineWorker) New() interfaces.PipelineWorker {
	return CurlPipelineWorker{
		name: "curl.worker",
	}
}

func (worker CurlPipelineWorker) Name() string {
	return worker.name
}

func (worker CurlPipelineWorker) Run(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}

	if task.Type != pipeline_component.CurlTask {
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

	if result.Error != nil {
		return result
	}

	msg := fmt.Sprintf("Curl executed successfully for task %s, response status code %s", task.Name, result.StatusCode)
	if common.IsDebug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)

	}

	notify.Success(msg)

	result.State = entities.StateExecuted
	return result
}

func (worker CurlPipelineWorker) Validate(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != pipeline_component.CurlTask {
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

func (worker CurlPipelineWorker) runTask(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	env := environment.Get()

	inputs, err := worker.parseParameters(task)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorInvalidParameters, err)
	}

	inputs.Decode()

	var request *http.Request
	inputs.Host = env.Replace(inputs.Host)
	if inputs.Content != nil {
		if inputs.Content.Json != "" {
			if inputs.Content.ContentType == "" {
				inputs.Content.ContentType = "application/json"
			}
			request, err = http.NewRequest(inputs.Verb, inputs.Host, strings.NewReader(inputs.Content.Json))
			if err != nil {
				return entities.NewPipelineWorkerResultFromError(ErrorInvalidConnection, err)
			}
			request.Header.Add("Content-Type", inputs.Content.ContentType)
		} else if inputs.Content.UrlEncoded != nil {
			if inputs.Content.ContentType == "" {
				inputs.Content.ContentType = "application/x-www-form-urlencoded"
			}

			data := url.Values{}
			for key, value := range inputs.Content.UrlEncoded {
				data.Add(key, value)
			}

			notify.Debug("Data: %s", data.Encode())
			request, err = http.NewRequest(inputs.Verb, inputs.Host, strings.NewReader(data.Encode()))
			if err != nil {
				return entities.NewPipelineWorkerResultFromError(ErrorInvalidConnection, err)
			}
			request.Header.Add("Content-Type", inputs.Content.ContentType)
		}
	} else {
		request, err = http.NewRequest(inputs.Verb, inputs.Host, nil)
		if err != nil {
			return entities.NewPipelineWorkerResultFromError(ErrorInvalidConnection, err)
		}
	}

	if request == nil {
		return entities.NewPipelineWorkerResultFromError(ErrorInvalidConnection, err)
	}

	// Adding the extra headers into the project
	if inputs.Headers != nil {
		for key, value := range inputs.Headers {
			request.Header.Add(key, value)
		}
	}

	for k, v := range request.Header {
		notify.Debug("Header %s: %s", k, v)
	}
	client := &http.Client{}

	response, err := client.Do(request)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorInvalidConnection, err)
	}

	result.StatusCode = fmt.Sprintf("%d", response.StatusCode)

	defer response.Body.Close()

	b, err := io.ReadAll(response.Body)
	// b, err := ioutil.ReadAll(resp.Body)  Go.1.15 and earlier
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorInvalidConnection, err)
	}

	notify.Debug("%s: got %s\n %s", inputs.Host, fmt.Sprintf("%v", response.StatusCode), string(b))

	if response.StatusCode <= 199 || response.StatusCode >= 400 {
		result.State = entities.StateErrored
		result.Error = fmt.Errorf("invalid success status code, %v", response.StatusCode)
		result.ErrorCode = fmt.Sprintf("%v", response.StatusCode)
		return result
	}

	return result
}

func (worker CurlPipelineWorker) parseParameters(task *pipeline_component.PipelineTask) (*CurlParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}
	var inputs CurlParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
