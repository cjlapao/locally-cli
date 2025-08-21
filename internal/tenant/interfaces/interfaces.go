// Package interfaces provides the tenant service interface.
package interfaces

import (
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	tenant_models "github.com/cjlapao/locally-cli/internal/tenant/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
)

type TenantServiceInterface interface {
	GetName() string
	GetTenants(ctx *appctx.AppContext, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[models.Tenant], *diagnostics.Diagnostics)
	GetTenantByIDOrSlug(ctx *appctx.AppContext, idOrSlug string) (*models.Tenant, *diagnostics.Diagnostics)
	CreateTenant(ctx *appctx.AppContext, request *tenant_models.TenantCreateRequest) (*models.Tenant, *diagnostics.Diagnostics)
	UpdateTenant(ctx *appctx.AppContext, tenantRequest *tenant_models.TenantUpdateRequest) (*models.Tenant, *diagnostics.Diagnostics)
	DeleteTenant(ctx *appctx.AppContext, idOrSlug string) *diagnostics.Diagnostics
}
