package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapClaimToEntity(claim *models.Claim) *entities.Claim {
	result := &entities.Claim{
		Service: claim.Service,
		Module:  claim.Module,
		Action:  claim.Action,
	}

	result.Slug = claim.GetSlug()

	return result
}

func MapClaimsToEntities(claims []models.Claim) []entities.Claim {
	result := make([]entities.Claim, len(claims))
	for i, claim := range claims {
		result[i] = *MapClaimToEntity(&claim)
	}
	return result
}
