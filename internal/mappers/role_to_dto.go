package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapRoleToDto(role *entities.Role) *models.Role {
	if role == nil {
		return nil
	}
	result := &models.Role{
		ID:            role.ID,
		TenantID:      role.TenantID,
		Slug:          role.Slug,
		Name:          role.Name,
		Description:   role.Description,
		SecurityLevel: role.SecurityLevel,
		Claims:        make([]models.Claim, len(role.Claims)),
	}

	for i, claim := range role.Claims {
		result.Claims[i] = *MapClaimToDto(&claim)
	}

	return result
}

func MapRoleToDtoWithClaims(role *entities.Role) *models.Role {
	if role == nil {
		return nil
	}
	result := MapRoleToDto(role)
	for i, claim := range role.Claims {
		result.Claims[i] = *MapClaimToDto(&claim)
	}
	return result
}

func MapRolesToDto(roles []entities.Role) []models.Role {
	result := make([]models.Role, len(roles))
	for i, role := range roles {
		result[i] = *MapRoleToDto(&role)
	}
	return result
}
