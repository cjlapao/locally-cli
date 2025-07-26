package docker

type Components []ComponentElement

type ComponentElement struct {
	ID         string      `json:"ID"`
	Name       string      `json:"Name"`
	Command    string      `json:"Command"`
	Project    string      `json:"Project"`
	Service    string      `json:"Service"`
	State      string      `json:"State"`
	Health     string      `json:"Health"`
	ExitCode   int64       `json:"ExitCode"`
	Publishers []Publisher `json:"Publishers"`
}
