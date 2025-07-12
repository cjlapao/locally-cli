package docker

type RunningResponse struct {
	Image string `json:"image"`
	State string `json:"state"`
}
