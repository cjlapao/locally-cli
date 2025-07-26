package docker

type Services []ServiceElement

type ServiceElement struct {
	Name        string `json:"Name"`
	Status      string `json:"Status"`
	ConfigFiles string `json:"ConfigFiles"`
}
