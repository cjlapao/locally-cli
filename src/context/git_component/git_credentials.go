package git_component

type GitCredentials struct {
	Username       string `json:"username,omitempty" yaml:"username,omitempty"`
	Password       string `json:"password,omitempty" yaml:"password,omitempty"`
	AccessToken    string `json:"accessToken,omitempty" yaml:"accessToken,omitempty"`
	PublicKeyPath  string `json:"publicKeyPath,omitempty" yaml:"publicKeyPath,omitempty"`
	PrivateKeyPath string `json:"privateKeyPath,omitempty" yaml:"privateKeyPath,omitempty"`
}
