package entities

type RouteReplace struct {
	Type string `json:"type,omitempty" yaml:"type,omitempty"`
	Old  string `json:"old,omitempty" yaml:"old,omitempty"`
	New  string `json:"new,omitempty" yaml:"new,omitempty"`
}
