package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapEnvironmentToDto(environment *entities.Environment) *models.Environment {
	result := &models.Environment{
		BaseModelWithTenant: *MapBaseModelWithTenantToDto(&environment.BaseModelWithTenant),
		Name:                environment.Name,
		Type:                environment.Type,
		ProjectID:           environment.ProjectID,
		Enabled:             environment.Enabled,
		Vaults:              make([]models.EnvironmentVault, len(environment.Vaults)),
	}

	for i, vault := range environment.Vaults {
		result.Vaults[i] = *MapEnvironmentVaultToDto(&vault)
		result.Vaults[i].EnvironmentID = environment.ID
		result.Vaults[i].EnvironmentName = environment.Name
	}

	return result
}

func MapEnvironmentsToDto(environments []entities.Environment) []models.Environment {
	result := make([]models.Environment, len(environments))
	for i, environment := range environments {
		result[i] = *MapEnvironmentToDto(&environment)
	}
	return result
}

func MapEnvironmentVaultToDto(environmentVault *entities.EnvironmentVault) *models.EnvironmentVault {
	result := &models.EnvironmentVault{
		BaseModelWithTenant: *MapBaseModelWithTenantToDto(&environmentVault.BaseModelWithTenant),
		EnvironmentID:       environmentVault.EnvironmentID,
		Name:                environmentVault.Name,
		VaultType:           environmentVault.VaultType,
		Description:         environmentVault.Description,
		CacheResults:        environmentVault.CacheResults,
		CacheTTL:            environmentVault.CacheTTL,
		LastSyncedAt:        environmentVault.LastSyncedAt,
		Metadata:            environmentVault.Metadata.Get(),
		Enabled:             environmentVault.Enabled,
	}

	for i, item := range environmentVault.Items {
		result.Items[i] = *MapEnvironmentVaultItemToDto(&item)
		result.Items[i].EnvironmentID = result.EnvironmentID
		result.Items[i].EnvironmentName = result.EnvironmentName
		result.Items[i].VaultID = result.ID
		result.Items[i].VaultName = result.Name
	}

	return result
}

func MapEnvironmentVaultsToDto(environmentVaults []entities.EnvironmentVault) []models.EnvironmentVault {
	result := make([]models.EnvironmentVault, len(environmentVaults))
	for i, environmentVault := range environmentVaults {
		result[i] = *MapEnvironmentVaultToDto(&environmentVault)
	}
	return result
}

func MapEnvironmentVaultItemToDto(environmentVaultItem *entities.EnvironmentVaultItem) *models.EnvironmentVaultItem {
	result := &models.EnvironmentVaultItem{
		BaseModelWithTenant: *MapBaseModelWithTenantToDto(&environmentVaultItem.BaseModelWithTenant),
		Key:                 environmentVaultItem.Key,
		Value:               environmentVaultItem.Value,
		ValueType:           environmentVaultItem.ValueType,
		Encrypted:           environmentVaultItem.IsEncrypted,
		Secret:              environmentVaultItem.IsSecret,
	}

	return result
}
