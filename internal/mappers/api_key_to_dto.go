package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapApiKeyToDto(apiKey *entities.ApiKey) *models.ApiKey {
	result := &models.ApiKey{
		Name:             apiKey.Name,
		KeyPrefix:        apiKey.KeyPrefix,
		KeyHash:          apiKey.KeyHash,
		ExpiresAt:        apiKey.ExpiresAt,
		IsActive:         apiKey.IsActive,
		RevokedAt:        apiKey.RevokedAt,
		RevokedBy:        apiKey.RevokedBy,
		RevocationReason: apiKey.RevocationReason,
		LastUsedAt:       apiKey.LastUsedAt,
	}

	if apiKey.Claims != nil {
		result.Claims = make([]models.Claim, len(apiKey.Claims))
		for i, claim := range apiKey.Claims {
			result.Claims[i] = *MapClaimToDto(&claim)
		}
	}

	result.BaseModelWithTenant = *MapBaseModelWithTenantToDto(&apiKey.BaseModelWithTenant)

	return result
}

func MapApiKeysToDto(apiKeys []entities.ApiKey) []models.ApiKey {
	result := make([]models.ApiKey, len(apiKeys))
	for i, apiKey := range apiKeys {
		result[i] = *MapApiKeyToDto(&apiKey)
	}
	return result
}
