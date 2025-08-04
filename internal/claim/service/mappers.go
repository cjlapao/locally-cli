package service

import (
	"github.com/cjlapao/locally-cli/internal/claim/models"
	"github.com/cjlapao/locally-cli/internal/database/entities"
)

func MapCreateClaimRequestToEntity(request *models.CreateClaimRequest) *entities.Claim {
	result := &entities.Claim{
		Module:        request.Module,
		Service:       request.Service,
		Action:        request.Action,
		SecurityLevel: request.SecurityLevel,
	}

	result.Slug = request.GetSlug()

	return result
}

func MapUpdateClaimRequestToEntity(request *models.UpdateClaimRequest) *entities.Claim {
	result := &entities.Claim{
		SecurityLevel: request.SecurityLevel,
	}

	result.ID = request.ID

	return result
}
