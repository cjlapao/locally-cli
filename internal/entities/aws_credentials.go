package entities

type AwsCredentials struct {
	KeyId     string `json:"keyId,omitempty" yaml:"keyId,omitempty"`
	KeySecret string `json:"keySecret,omitempty" yaml:"keySecret,omitempty"`
	Region    string `json:"region,omitempty" yaml:"region,omitempty"`
}
