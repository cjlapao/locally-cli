package entities

import "fmt"

type PipelineWorkerResult struct {
	State      PipelineWorkerResultState
	Output     string
	ErrorCode  string
	StatusCode string
	Error      error
}

func (a PipelineWorkerResult) String() string {
	switch a.State {
	case StateValid:
		return "Task is valid"
	case StateExecuted:
		return "Task was executed"
	case StateIgnored:
		return "Task was ignored"
	case StateErrored:
		if a.ErrorCode != "" {
			return fmt.Sprintf("Task errored with code %s, err. %s", a.ErrorCode, a.Error.Error())
		} else {
			return fmt.Sprintf("Task errored, err. %s", a.Error.Error())
		}
	default:
		return "Task has unknown state"
	}
}

func NewPipelineWorkerResultFromError(code string, err error) PipelineWorkerResult {
	return PipelineWorkerResult{
		State:     StateErrored,
		ErrorCode: code,
		Error:     err,
	}
}

type PipelineWorkerResultState uint

const (
	StateIgnored PipelineWorkerResultState = iota
	StateErrored
	StateExecuted
	StateValid
)

func (s PipelineWorkerResultState) String() string {
	return toPipelineWorkerResultStateString[s]
}

func (s *PipelineWorkerResultState) FromString(value string) {
	*s = toPipelineWorkerResultStateType[value]
}

var toPipelineWorkerResultStateString = map[PipelineWorkerResultState]string{
	StateErrored:  "errored",
	StateIgnored:  "ignored",
	StateExecuted: "executed",
	StateValid:    "valid",
}

var toPipelineWorkerResultStateType = map[string]PipelineWorkerResultState{
	"errored":  StateErrored,
	"ignored":  StateIgnored,
	"executed": StateExecuted,
	"valid":    StateValid,
}
