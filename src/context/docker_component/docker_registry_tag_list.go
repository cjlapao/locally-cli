package docker_component

type DockerRegistryTagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
