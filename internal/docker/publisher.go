package docker

type Publisher struct {
	URL           string `json:"URL"`
	TargetPort    int64  `json:"TargetPort"`
	PublishedPort int64  `json:"PublishedPort"`
	Protocol      string `json:"Protocol"`
}
