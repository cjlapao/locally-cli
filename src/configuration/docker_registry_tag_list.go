package configuration

type DockerRegistryTagList struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}
