package docker

type DockerVersion struct {
	Client Client `json:"Client"`
	Server Server `json:"Server"`
}
