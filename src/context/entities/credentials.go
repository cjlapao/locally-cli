package entities

import (
	"github.com/cjlapao/locally-cli/entities"
)

type Credentials struct {
	Azure *entities.AzureCredentials `json:"azure,omitempty" yaml:"azure,omitempty"`
}
