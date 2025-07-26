package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapClaimToDto(claim *entities.Claim) *models.Claim {
	result := &models.Claim{
		ID:        claim.ID,
		Slug:      claim.Slug,
		Service:   claim.Service,
		Module:    claim.Module,
		Action:    claim.Action,
		CreatedAt: claim.CreatedAt,
		UpdatedAt: claim.UpdatedAt,
	}

	return result
}
