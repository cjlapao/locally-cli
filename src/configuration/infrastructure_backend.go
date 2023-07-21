package configuration

type InfrastructureBackend struct {
	ResourceGroupName  string `json:"resourceGroupName,omitempty" yaml:"resourceGroupName,omitempty"`
	StorageAccountName string `json:"storageAccountName,omitempty" yaml:"storageAccountName,omitempty"`
	ContainerName      string `json:"containerName,omitempty" yaml:"containerName,omitempty"`
	StateFileName      string `json:"stateFileName,omitempty" yaml:"stateFileName,omitempty"`
	AccessKey          string `json:"accessKey,omitempty" yaml:"accessKey,omitempty"`
}
