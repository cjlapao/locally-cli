package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapClaimToDto(claim *entities.Claim) *models.Claim {
	result := &models.Claim{
		ID:            claim.ID,
		TenantID:      claim.TenantID,
		Slug:          claim.Slug,
		Service:       claim.Service,
		Module:        claim.Module,
		Action:        claim.Action,
		SecurityLevel: claim.SecurityLevel,
		CreatedAt:     claim.CreatedAt,
		UpdatedAt:     claim.UpdatedAt,
	}

	return result
}

func MapClaimsToDto(claims []entities.Claim) []models.Claim {
	result := make([]models.Claim, len(claims))
	for i, claim := range claims {
		result[i] = *MapClaimToDto(&claim)
	}
	return result
}
