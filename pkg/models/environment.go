package models

import (
	"github.com/cjlapao/locally-cli/pkg/types"
)

type Environment struct {
	BaseModelWithTenant
	Name      string                `json:"name" yaml:"name"`
	Type      types.EnvironmentType `json:"type" yaml:"type"`
	ProjectID string                `json:"project_id,omitempty" yaml:"project_id,omitempty"`
	Enabled   bool                  `json:"enabled" yaml:"enabled"`
	Vaults    []EnvironmentVault    `json:"vaults,omitempty" yaml:"vaults,omitempty"`
}

func (e *Environment) GetAvailableVaults() []EnvironmentVault {
	vaults := make([]EnvironmentVault, len(e.Vaults))
	for i, vault := range e.Vaults {
		vaults[i] = vault
		vaults[i].EnvironmentID = e.ID
		vaults[i].EnvironmentName = e.Name
	}
	return vaults
}

func (e *Environment) GetVault(name string) (*EnvironmentVault, bool) {
	for _, vault := range e.Vaults {
		if vault.Name == name {
			return &vault, true
		}
	}
	return nil, false
}
