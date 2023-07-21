package docker

type Component struct {
	Name    string  `json:"Name"`
	Version string  `json:"Version"`
	Details Details `json:"Details"`
}
