package entities

type Tenant struct {
	Source string
	ID     string `json:"id,omitempty" yaml:"id,omitempty"`
	Name   string `json:"name,omitempty" yaml:"name,omitempty"`
	URI    string `json:"uri,omitempty" yaml:"uri,omitempty"`
}
