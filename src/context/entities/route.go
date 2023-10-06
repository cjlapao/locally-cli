package entities

type Route struct {
	Name    string              `json:"name,omitempty" yaml:"name,omitempty"`
	Regex   string              `json:"regex,omitempty" yaml:"regex,omitempty"`
	Replace RouteReplace        `json:"replace,omitempty" yaml:"replace,omitempty"`
	Headers []map[string]string `json:"headers,omitempty" yaml:"headers,omitempty"`
}
