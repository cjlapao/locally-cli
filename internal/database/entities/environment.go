package entities

import (
	"time"

	database_types "github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/pkg/types"
)

type Environment struct {
	BaseModelWithTenant
	Type      types.EnvironmentType `json:"type" yaml:"type" gorm:"column:type;type:varchar(255);not null"`
	Name      string                `json:"name" yaml:"name" gorm:"column:name;type:varchar(255);not null"`
	ProjectID string                `json:"project_id" yaml:"project_id" gorm:"column:project_id;type:varchar(255);not null"`
	Enabled   bool                  `json:"enabled" yaml:"enabled" gorm:"column:enabled;type:boolean;not null;default:true"`
	Vaults    []EnvironmentVault    `json:"vaults" yaml:"vaults" gorm:"foreignKey:EnvironmentID;references:ID;constraint:OnDelete:CASCADE"`
}

func (e *Environment) TableName() string {
	return "environments"
}

type EnvironmentVault struct {
	BaseModelWithTenant
	EnvironmentID string                                            `json:"environment_id" yaml:"environment_id" gorm:"column:environment_id;type:varchar(255);not null"`
	Name          string                                            `json:"name" yaml:"name" gorm:"column:name;type:varchar(255);not null"`
	VaultType     string                                            `json:"vault_type" yaml:"vault_type" gorm:"column:vault_type;type:varchar(255);not null"`
	Description   string                                            `json:"description" yaml:"description" gorm:"column:description;type:text"`
	Enabled       bool                                              `json:"enabled" yaml:"enabled" gorm:"column:enabled;type:boolean;not null;default:true"`
	CacheResults  bool                                              `json:"cache_results" yaml:"cache_results" gorm:"column:cache_results;type:boolean;not null;default:false"`
	CacheTTL      int                                               `json:"cache_ttl" yaml:"cache_ttl" gorm:"column:cache_ttl;type:int;not null;default:0"`
	LastSyncedAt  *time.Time                                        `json:"last_synced_at" yaml:"last_synced_at" gorm:"column:last_synced_at;type:timestamp;"`
	Metadata      database_types.JSONObject[map[string]interface{}] `json:"metadata" yaml:"metadata" gorm:"column:metadata;type:jsonb;not null"`
	Items         []EnvironmentVaultItem                            `json:"items" yaml:"items" gorm:"foreignKey:EnvironmentVaultID;references:ID;constraint:OnDelete:CASCADE"`
}

func (e *EnvironmentVault) TableName() string {
	return "environment_vaults"
}

type EnvironmentVaultItem struct {
	BaseModelWithTenant
	EnvironmentVaultID string                         `json:"environment_vault_id" yaml:"environment_vault_id" gorm:"column:environment_vault_id;type:varchar(255);not null"`
	Key                string                         `json:"key" yaml:"key" gorm:"column:key;type:varchar(255);not null"`
	Value              string                         `json:"value" yaml:"value" gorm:"column:value;type:text;not null"`
	IsEncrypted        bool                           `json:"is_encrypted" yaml:"is_encrypted" gorm:"column:is_encrypted;type:boolean;not null;default:false"`
	IsSecret           bool                           `json:"is_secret" yaml:"is_secret" gorm:"column:is_secret;type:boolean;not null;default:false"`
	ValueType          types.EnvironmentVaultItemType `json:"value_type" yaml:"value_type" gorm:"column:value_type;type:varchar(255);not null"`
}

func (e *EnvironmentVaultItem) TableName() string {
	return "environment_vault_items"
}
