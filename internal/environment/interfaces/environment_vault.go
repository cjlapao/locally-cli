package interfaces

import "github.com/cjlapao/locally-cli/pkg/diagnostics"

type EnvironmentVault interface {
	Name() string
	Sync() ([]EnvironmentVaultItem, *diagnostics.Diagnostics)
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}) *diagnostics.Diagnostics
	Remove(key string) *diagnostics.Diagnostics
	GetItems() []EnvironmentVaultItem
	GetMetadata() map[string]interface{}
	IsEnabled() bool
	IsSynced() bool
}
