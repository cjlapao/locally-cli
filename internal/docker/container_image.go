package docker

type ContainerImage struct {
	Container  string `json:"Container"`
	Repository string `json:"Repository"`
	Tag        string `json:"Tag"`
	ImageId    string `json:"ImageId"`
	Size       string `json:"Size"`
}
