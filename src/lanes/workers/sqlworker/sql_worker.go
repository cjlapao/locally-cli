package sqlworker

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/context/pipeline_component"
	"github.com/cjlapao/locally-cli/lanes/entities"
	"github.com/cjlapao/locally-cli/lanes/interfaces"
	"github.com/cjlapao/locally-cli/notifications"

	_ "github.com/microsoft/go-mssqldb"
	"gopkg.in/yaml.v3"
)

var notify = notifications.Get()

const (
	ErrorInvalidParameters = "400"
	ErrorInvalidConnection = "401"
	ErrorCannotConnect     = "402"
	ErrorFailedExecution   = "403"
)

type SqlPipelineWorker struct {
	name string
}

func (worker SqlPipelineWorker) New() interfaces.PipelineWorker {
	return SqlPipelineWorker{
		name: "sql.worker",
	}
}

func (worker SqlPipelineWorker) Name() string {
	return worker.name
}

func (worker SqlPipelineWorker) Run(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}

	if task.Type != pipeline_component.SqlTask {
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

	notify.Debug("SQL Query: %s", inputs.Query)

	// Create connection pool
	db, err := sql.Open("sqlserver", inputs.ConnectionString)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorInvalidConnection, err)
	}

	ctx := context.Background()
	err = db.PingContext(ctx)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorCannotConnect, err)
	}

	r, err := db.ExecContext(ctx, inputs.Query)
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorFailedExecution, err)
	}

	affected, err := r.RowsAffected()
	if err != nil {
		return entities.NewPipelineWorkerResultFromError(ErrorFailedExecution, err)
	}

	msg := fmt.Sprintf("Sql executed successfully for task %s, affected %s rows", task.Name, fmt.Sprintf("%d", affected))
	if common.IsDebug() {
		msg = fmt.Sprintf("[%s] %s", worker.name, msg)

	}

	notify.Success(msg)

	result.State = entities.StateExecuted

	return result
}

func (worker SqlPipelineWorker) Validate(task *pipeline_component.PipelineTask) entities.PipelineWorkerResult {
	result := entities.PipelineWorkerResult{}
	if task.Type != pipeline_component.SqlTask {
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

func (worker SqlPipelineWorker) parseParameters(task *pipeline_component.PipelineTask) (*SqlParameters, error) {
	encoded, err := yaml.Marshal(task.Inputs)
	if err != nil {
		return nil, err
	}
	var inputs SqlParameters
	err = yaml.Unmarshal(encoded, &inputs)
	if err != nil {
		return nil, err
	}

	return &inputs, nil
}
