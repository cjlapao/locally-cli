package types

import (
	"encoding/json"
	"strings"
)

type ContainerState string

const (
	ContainerStateCreated    ContainerState = "created"
	ContainerStateStarting   ContainerState = "starting"
	ContainerStateRestarting ContainerState = "restarting"
	ContainerStateRunning    ContainerState = "running"
	ContainerStateRemoving   ContainerState = "removing"
	ContainerStatePaused     ContainerState = "paused"
	ContainerStateExited     ContainerState = "exited"
	ContainerStateDestroy    ContainerState = "destroy"
	ContainerStateDead       ContainerState = "dead"
	ContainerStateError      ContainerState = "error"
	ContainerStateStopping   ContainerState = "stopping"
	ContainerStatePausing    ContainerState = "stopped"
	ContainerStateAborting   ContainerState = "aborting"
	ContainerStateFreezing   ContainerState = "freezing"
	ContainerStateFrozen     ContainerState = "frozen"
	ContainerStateThawed     ContainerState = "thawed"
)

func (c ContainerState) String() string {
	return strings.ToLower(string(c))
}

func (c ContainerState) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}

func (c ContainerState) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}

	return nil
}

func (c ContainerState) FromString(s string) ContainerState {
	s = strings.ToLower(s)
	return ContainerState(s)
}

func (c ContainerState) IsRunning() bool {
	return c == ContainerStateRunning || c == ContainerStateStarting || c == ContainerStateRestarting
}

func (c ContainerState) IsStopped() bool {
	return c == ContainerStateExited || c == ContainerStateDestroy || c == ContainerStateDead
}

func (c ContainerState) IsPaused() bool {
	return c == ContainerStatePaused || c == ContainerStatePausing
}

func (c ContainerState) IsDestroyed() bool {
	return c == ContainerStateDestroy || c == ContainerStateDead
}

func (c ContainerState) IsError() bool {
	return c == ContainerStateError
}

func (c ContainerState) IsStarting() bool {
	return c == ContainerStateStarting
}

func (c ContainerState) IsStopping() bool {
	return c == ContainerStateStopping
}

func (c ContainerState) IsPausing() bool {
	return c == ContainerStatePausing
}

func (c ContainerState) IsAborting() bool {
	return c == ContainerStateAborting
}

type ContainerDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	State string `json:"state"`
}

type ContainerRemoveRequest struct {
	Force         bool `json:"force"`
	RemoveVolumes bool `json:"remove_volumes"`
	RemoveLinks   bool `json:"remove_links"`
}

type ContainerOperationResponse struct {
	ContainerID string `json:"container_id"`
	Success     bool   `json:"success"`
	Error       string `json:"error,omitempty"`
}

type ContainerExposePort struct {
	HostPort      string `json:"host_port"`
	ContainerPort string `json:"container_port"`
}

type ContainerCreateRequest struct {
	Name        string                `json:"name"`
	Image       string                `json:"image"`
	ExposePorts []ContainerExposePort `json:"expose_ports"`
	Cmd         []string              `json:"cmd"`
	RunOnCreate bool                  `json:"run_on_create"`
}
