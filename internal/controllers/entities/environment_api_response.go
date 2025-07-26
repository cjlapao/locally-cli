package entities

type EnvironmentApiResponse struct {
	Valid     bool   `json:"valid" yaml:"valid"`
	Enabled   bool   `json:"enabled" yaml:"enabled"`
	Name      string `json:"name,omitempty" yaml:"name,omitempty"`
	Id        string `json:"id,omitempty" yaml:"id,omitempty"`
	Type      string `json:"type,omitempty" yaml:"type,omitempty"`
	Location  string `json:"location,omitempty" yaml:"location,omitempty"`
	Domain    string `json:"domain,omitempty" yaml:"domain,omitempty"`
	Subdomain string `json:"subdomain,omitempty" yaml:"subdomain,omitempty"`
}
