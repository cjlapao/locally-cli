package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapTenantToEntity(tenant *models.Tenant) *entities.Tenant {
	result := &entities.Tenant{
		Name:          tenant.Name,
		Description:   tenant.Description,
		Domain:        tenant.Domain,
		OwnerID:       tenant.OwnerID,
		ContactEmail:  tenant.ContactEmail,
		Status:        tenant.Status,
		LogoURL:       tenant.LogoURL,
		Require2FA:    tenant.Require2FA,
		ActivatedAt:   tenant.ActivatedAt,
		DeactivatedAt: tenant.DeactivatedAt,
	}

	if tenant.Metadata != nil {
		obj := types.JSONObject[map[string]interface{}]{}
		obj.Set(tenant.Metadata)
		result.Metadata = obj
	}

	result.ID = tenant.ID
	result.Slug = tenant.Slug
	result.CreatedAt = tenant.CreatedAt
	result.UpdatedAt = tenant.UpdatedAt

	return result
}
