package service

import (
	"time"

	tenant_models "github.com/cjlapao/locally-cli/internal/tenant/models"
	"github.com/cjlapao/locally-cli/pkg/models"
)

func MapTenantCreateRequestToTenant(request *tenant_models.TenantCreateRequest) *models.Tenant {
	now := time.Now()
	return &models.Tenant{
		ID:            request.ID,
		Name:          request.Name,
		Description:   request.Description,
		Domain:        request.Domain,
		ContactEmail:  request.ContactEmail,
		Status:        "active",
		ActivatedAt:   &now,
		DeactivatedAt: nil,
		Metadata:      request.Metadata,
		LogoURL:       "",
		Require2FA:    false,
	}
}
