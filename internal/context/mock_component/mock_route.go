package mock_component

type MockRoute struct {
	Name     string              `json:"name,omitempty" yaml:"name,omitempty"`
	Regex    string              `json:"regex,omitempty" yaml:"regex,omitempty"`
	Headers  []map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
	Responds MockRouteResponse   `json:"responds,omitempty" yaml:"responds,omitempty"`
}
