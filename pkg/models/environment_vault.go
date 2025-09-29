package models

import "time"

type EnvironmentVault struct {
	BaseModelWithTenant
	EnvironmentID   string                 `json:"environment_id" yaml:"environment_id"`
	EnvironmentName string                 `json:"environment_name" yaml:"environment_name"`
	Name            string                 `json:"name" yaml:"name"`
	VaultType       string                 `json:"vault_type" yaml:"vault_type"`
	Description     string                 `json:"description" yaml:"description"`
	CacheResults    bool                   `json:"cache_results" yaml:"cache_results"`
	CacheTTL        int                    `json:"cache_ttl" yaml:"cache_ttl"`
	LastSyncedAt    *time.Time             `json:"last_synced_at" yaml:"last_synced_at"`
	Metadata        map[string]interface{} `json:"metadata" yaml:"metadata"`
	Enabled         bool                   `json:"enabled" yaml:"enabled"`
	Items           []EnvironmentVaultItem `json:"items" yaml:"items"`
}

func (e *EnvironmentVault) GetAvailableVaultItems() []EnvironmentVaultItem {
	items := make([]EnvironmentVaultItem, len(e.Items))
	for i, item := range e.Items {
		items[i] = item
		items[i].EnvironmentID = e.EnvironmentID
		items[i].EnvironmentName = e.EnvironmentName
		items[i].VaultID = e.ID
		items[i].VaultName = e.Name
		items[i].Key = item.Key
		items[i].Value = item.Value
		items[i].ValueType = item.ValueType
		items[i].Encrypted = item.Encrypted
		items[i].Secret = item.Secret
	}
	return items
}
