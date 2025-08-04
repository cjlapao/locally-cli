package service

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/role/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
)

func MapCreateRoleRequestToEntity(request *models.CreateRoleRequest) *entities.Role {
	result := &entities.Role{
		Name:          request.Name,
		Description:   request.Description,
		SecurityLevel: request.SecurityLevel,
	}

	result.Slug = utils.Slugify(request.Name)

	return result
}

func MapUpdateRoleRequestToEntity(request *models.UpdateRoleRequest) *entities.Role {
	result := &entities.Role{
		Name:          request.Name,
		Description:   request.Description,
		SecurityLevel: request.SecurityLevel,
	}

	result.ID = request.ID
	result.Slug = request.Slug

	return result
}
