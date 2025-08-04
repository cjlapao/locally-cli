package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/user/models"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

func MapClaimToEntity(claim *pkg_models.Claim) *entities.Claim {
	result := &entities.Claim{
		BaseModelWithTenant: entities.BaseModelWithTenant{
			ID:       claim.ID,
			TenantID: claim.TenantID,
			Slug:     claim.GetSlug(),
		},
		Service:       claim.Service,
		Module:        claim.Module,
		Action:        claim.Action,
		SecurityLevel: claim.SecurityLevel,
	}

	return result
}

func MapClaimsToEntities(claims []pkg_models.Claim) []entities.Claim {
	result := make([]entities.Claim, len(claims))
	for i, claim := range claims {
		result[i] = *MapClaimToEntity(&claim)
	}
	return result
}

func MapCreateClaimRequestToEntity(claim *models.CreateClaimRequest) *entities.Claim {
	result := &entities.Claim{
		Service:       claim.Service,
		Module:        claim.Module,
		Action:        pkg_models.AccessLevel(claim.Action),
		SecurityLevel: claim.SecurityLevel,
	}

	result.Slug = claim.GetSlug()

	return result
}

func MapUpdateClaimRequestToEntity(claim *models.UpdateClaimRequest) *entities.Claim {
	result := &entities.Claim{
		SecurityLevel: claim.SecurityLevel,
	}

	result.ID = claim.ID

	return result
}
