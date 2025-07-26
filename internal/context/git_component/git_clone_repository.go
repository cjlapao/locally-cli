package git_component

type GitCloneRepository struct {
	Enabled     bool            `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Url         string          `json:"url,omitempty" yaml:"url,omitempty"`
	Destination string          `json:"destination,omitempty" yaml:"destination,omitempty"`
	Path        string          `json:"-" yaml:"-"`
	Credentials *GitCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}
