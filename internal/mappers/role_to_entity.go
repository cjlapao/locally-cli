package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/user/models"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
)

func MapRoleToEntity(role *pkg_models.Role) *entities.Role {
	result := &entities.Role{
		BaseModelWithTenant: entities.BaseModelWithTenant{
			ID:       role.ID,
			TenantID: role.TenantID,
			Slug:     role.Slug,
		},
		Name:          role.Name,
		Description:   role.Description,
		SecurityLevel: role.SecurityLevel,
		Claims:        make([]entities.Claim, len(role.Claims)),
	}

	for i, claim := range role.Claims {
		result.Claims[i] = *MapClaimToEntity(&claim)
	}

	return result
}

func MapRolesToEntities(roles []pkg_models.Role) []entities.Role {
	result := make([]entities.Role, len(roles))
	for i, role := range roles {
		result[i] = *MapRoleToEntity(&role)
	}
	return result
}

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
