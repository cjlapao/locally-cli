package entities

import (
	"github.com/cjlapao/locally-cli/internal/entities"
)

type Credentials struct {
	Azure *entities.AzureCredentials `json:"azure,omitempty" yaml:"azure,omitempty"`
}
