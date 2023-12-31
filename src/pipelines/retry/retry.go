package retry

import (
	"errors"
	"fmt"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/notifications"
	"github.com/cjlapao/locally-cli/pipelines/entities"
	"time"
)

func RetryRun(task *configuration.PipelineTask, funcToExecute func(task *configuration.PipelineTask) entities.PipelineWorkerResult, retryCount, waitFor int) entities.PipelineWorkerResult {
	var result entities.PipelineWorkerResult
	notify := notifications.Get()

	waiting := time.Second * 0

	// if retryCount == 0 {
	// 	retryCount = common.DEFAULT_RETRY_COUNT
	// }
	if waitFor > 0 {
		waiting = time.Second * time.Duration(waitFor)
	}

	if funcToExecute == nil {
		result.State = entities.StateIgnored
		result.ErrorCode = "100"
		result.Error = errors.New("no task to execute")

		return result
	}

	for {
		result = funcToExecute(task)
		if result.Error == nil {
			return result
		}

		retryCount -= 1

		if retryCount <= 0 {
			notify.Error("Exceeded maximum number of retries, returning bad result")
			break
		}

		notify.Info("Will retry for %s more time(s)", fmt.Sprintf("%v", retryCount))

		if waiting.Seconds() > 0 {
			notify.Info("Waiting for %s before next retry", waiting.String())
			time.Sleep(waiting)
		}
	}

	return result
}
