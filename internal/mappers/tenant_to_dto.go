package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapTenantToDto(tenant *entities.Tenant) *models.Tenant {
	result := &models.Tenant{
		ID:            tenant.ID,
		Name:          tenant.Name,
		Slug:          tenant.Slug,
		Description:   tenant.Description,
		Domain:        tenant.Domain,
		OwnerID:       tenant.OwnerID,
		ContactEmail:  tenant.ContactEmail,
		Status:        tenant.Status,
		ActivatedAt:   tenant.ActivatedAt,
		DeactivatedAt: tenant.DeactivatedAt,
		LogoURL:       tenant.LogoURL,
		Require2FA:    tenant.Require2FA,
		CreatedAt:     tenant.CreatedAt,
		UpdatedAt:     tenant.UpdatedAt,
	}

	if tenant.Metadata.Get() != nil {
		result.Metadata = tenant.Metadata.Get()
	}

	return result
}

func MapTenantsToDto(tenants []entities.Tenant) []models.Tenant {
	result := make([]models.Tenant, len(tenants))
	for i, tenant := range tenants {
		result[i] = *MapTenantToDto(&tenant)
	}
	return result
}
