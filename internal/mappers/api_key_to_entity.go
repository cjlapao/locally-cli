package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapApiKeyToEntity(apiKey *models.ApiKey) *entities.ApiKey {
	result := &entities.ApiKey{
		Name:             apiKey.Name,
		KeyHash:          apiKey.KeyHash,
		KeyPrefix:        apiKey.KeyPrefix,
		ExpiresAt:        apiKey.ExpiresAt,
		IsActive:         apiKey.IsActive,
		RevokedAt:        apiKey.RevokedAt,
		RevokedBy:        apiKey.RevokedBy,
		RevocationReason: apiKey.RevocationReason,
	}

	if apiKey.Claims != nil {
		result.Claims = make([]entities.Claim, len(apiKey.Claims))
		for i, claim := range apiKey.Claims {
			result.Claims[i] = *MapClaimToEntity(&claim)
		}
	}

	result.BaseModelWithTenant = *MapBaseModelWithTenantToEntity(&apiKey.BaseModelWithTenant)

	return result
}

func MapApiKeysToEntity(apiKeys []models.ApiKey) []entities.ApiKey {
	result := make([]entities.ApiKey, len(apiKeys))
	for i, apiKey := range apiKeys {
		result[i] = *MapApiKeyToEntity(&apiKey)
	}
	return result
}
