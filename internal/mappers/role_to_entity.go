package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapRoleToEntity(role *models.Role) *entities.Role {
	result := &entities.Role{
		Name:        role.Name,
		Description: role.Description,
		IsAdmin:     role.IsAdmin,
		IsSuperUser: role.IsSuperUser,
	}

	result.Slug = role.Slug

	return result
}

func MapRolesToEntities(roles []models.Role) []entities.Role {
	result := make([]entities.Role, len(roles))
	for i, role := range roles {
		result[i] = *MapRoleToEntity(&role)
	}
	return result
}
