package models

import (
	"github.com/cjlapao/locally-cli/pkg/types"
)

type EnvironmentVaultItem struct {
	BaseModelWithTenant
	EnvironmentID   string                         `json:"environment_id" yaml:"environment_id"`
	EnvironmentName string                         `json:"environment_name" yaml:"environment_name"`
	VaultID         string                         `json:"vault_id" yaml:"vault_id"`
	VaultName       string                         `json:"vault_name" yaml:"vault_name"`
	Key             string                         `json:"key" yaml:"key"`
	Value           string                         `json:"value" yaml:"value"`
	ValueType       types.EnvironmentVaultItemType `json:"value_type" yaml:"value_type"`
	Encrypted       bool                           `json:"encrypted" yaml:"is_encrypted"`
	Secret          bool                           `json:"secret" yaml:"is_secret"`
}

func (e EnvironmentVaultItem) GetKey() string {
	return e.Key
}

func (e EnvironmentVaultItem) GetValue() string {
	return e.Value
}

func (e EnvironmentVaultItem) IsEncrypted() bool {
	return e.Encrypted
}

func (e EnvironmentVaultItem) IsSecret() bool {
	return e.Secret
}
