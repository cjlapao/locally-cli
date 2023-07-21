package configuration

type GitCloneRepository struct {
	Enabled     bool            `json:"enabled,omitempty" yaml:"enabled,omitempty"`
	Url         string          `json:"url,omitempty" yaml:"url,omitempty"`
	Destination string          `json:"destination,omitempty" yaml:"destination,omitempty"`
	Path        string          `json:"-" yaml:"-"`
	Credentials *GitCredentials `json:"credentials,omitempty" yaml:"credentials,omitempty"`
}

type GitCredentials struct {
	Username       string `json:"username,omitempty" yaml:"username,omitempty"`
	Password       string `json:"password,omitempty" yaml:"password,omitempty"`
	AccessToken    string `json:"accessToken,omitempty" yaml:"accessToken,omitempty"`
	PublicKeyPath  string `json:"publicKeyPath,omitempty" yaml:"publicKeyPath,omitempty"`
	PrivateKeyPath string `json:"privateKeyPath,omitempty" yaml:"privateKeyPath,omitempty"`
}
