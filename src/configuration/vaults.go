package configuration

type EnvironmentVariables struct {
	Global    map[string]interface{} `json:"global,omitempty" yaml:"global,omitempty"`
	KeyVault  map[string]interface{} `json:"keyvault,omitempty" yaml:"keyvault,omitempty"`
	Terraform map[string]interface{} `json:"terraform,omitempty" yaml:"terraform,omitempty"`
}
