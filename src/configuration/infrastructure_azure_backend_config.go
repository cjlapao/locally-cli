package configuration

type InfrastructureAzureBackendConfig struct {
	Location           string `json:"location,omitempty" yaml:"location,omitempty"`
	SubscriptionId     string `json:"subscriptionId,omitempty" yaml:"subscriptionId,omitempty"`
	ResourceGroupName  string `json:"resourceGroupName,omitempty" yaml:"resourceGroupName,omitempty"`
	StorageAccountName string `json:"storageAccountName,omitempty" yaml:"storageAccountName,omitempty"`
	ContainerName      string `json:"containerName,omitempty" yaml:"containerName,omitempty"`
	AccessKey          string `json:"accessKey,omitempty" yaml:"accessKey,omitempty"`
}
