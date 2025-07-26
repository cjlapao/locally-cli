package entities

type LocallyConfigService struct {
	Url             string `json:"url" yaml:"url"`
	ReverseProxyUrl string `json:"reverseProxyUrl"`
}
