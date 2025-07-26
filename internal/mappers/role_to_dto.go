package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapRoleToDto(role *entities.Role) *models.Role {
	result := &models.Role{
		ID:          role.ID,
		Slug:        role.Slug,
		Name:        role.Name,
		Description: role.Description,
		IsAdmin:     role.IsAdmin,
		IsSuperUser: role.IsSuperUser,
	}

	return result
}
