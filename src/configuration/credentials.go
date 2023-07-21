package configuration

import (
	"github.com/cjlapao/locally-cli/entities"
)

type Credentials struct {
	Azure *entities.AzureCredentials `json:"azure,omitempty" yaml:"azure,omitempty"`
}

func AddAzureCredential(cred *entities.AzureCredentials) {
	config := Get()
	context := config.GetCurrentContext()

	if context.Credentials == nil {
		context.Credentials = &Credentials{}
	}

	context.Credentials.Azure = cred
}
