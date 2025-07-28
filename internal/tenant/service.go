package tenant

import (
	"errors"
	"sync"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"gorm.io/gorm"
)

var (
	globalTenantService *TenantService
	tenantServiceOnce   sync.Once
	tenantServiceMutex  sync.Mutex
)

type TenantServiceInterface interface {
	GetName() string
	GetTenantsByFilter(ctx *appctx.AppContext, filter *filters.Filter) (*api.PaginatedResponse[models.Tenant], *diagnostics.Diagnostics)
	GetTenantByID(ctx *appctx.AppContext, idOrSlug string) (*models.Tenant, *diagnostics.Diagnostics)
	CreateTenant(ctx *appctx.AppContext, tenant *models.Tenant) (*models.Tenant, *diagnostics.Diagnostics)
	UpdateTenant(ctx *appctx.AppContext, tenantRequest *TenantUpdateRequest) (*models.Tenant, *diagnostics.Diagnostics)
	DeleteTenant(ctx *appctx.AppContext, idOrSlug string) *diagnostics.Diagnostics
}

type TenantService struct {
	tenantStore stores.TenantDataStoreInterface
}

func Initialize(tenantStore stores.TenantDataStoreInterface) TenantServiceInterface {
	tenantServiceMutex.Lock()
	defer tenantServiceMutex.Unlock()

	tenantServiceOnce.Do(func() {
		globalTenantService = new(tenantStore)
	})
	return globalTenantService
}

func GetInstance() TenantServiceInterface {
	if globalTenantService == nil {
		panic("tenant service not initialized")
	}
	return globalTenantService
}

// Reset resets the singleton for testing purposes
func Reset() {
	tenantServiceMutex.Lock()
	defer tenantServiceMutex.Unlock()
	globalTenantService = nil
	tenantServiceOnce = sync.Once{}
}

func new(tenantStore stores.TenantDataStoreInterface) *TenantService {
	return &TenantService{
		tenantStore: tenantStore,
	}
}

func (s *TenantService) GetName() string {
	return "tenant"
}

func (s *TenantService) GetTenantsByFilter(ctx *appctx.AppContext, filter *filters.Filter) (*api.PaginatedResponse[models.Tenant], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_tenants")
	defer diag.Complete()

	dbTenants, err := s.tenantStore.GetTenantsByFilter(ctx, filter)
	if err != nil {
		diag.AddError("failed_to_get_tenants", "failed to get tenants", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	tenants := mappers.MapTenantsToDto(dbTenants.Items)
	pagination := api.Pagination{
		Page:       dbTenants.Page,
		PageSize:   dbTenants.PageSize,
		TotalPages: dbTenants.TotalPages,
	}

	response := api.PaginatedResponse[models.Tenant]{
		Data:       tenants,
		TotalCount: int(dbTenants.Total),
		Pagination: pagination,
	}

	return &response, diag
}

func (s *TenantService) GetTenantByID(ctx *appctx.AppContext, idOrSlug string) (*models.Tenant, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_tenant_by_id")
	defer diag.Complete()

	dbTenant, err := s.tenantStore.GetTenantByIdOrSlug(ctx, idOrSlug)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			ctx.Log().WithField("id_or_slug", idOrSlug).Infof("Tenant with id or slug %v not found", idOrSlug)
			return nil, diag
		}
		diag.AddError("failed_to_get_tenant_by_id_or_slug", "failed to get tenant by id or slug", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	tenant := mappers.MapTenantToDto(dbTenant)

	return tenant, diag
}

func (s *TenantService) CreateTenant(ctx *appctx.AppContext, tenant *models.Tenant) (*models.Tenant, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_tenant")
	defer diag.Complete()

	dbTenant := mappers.MapTenantToEntity(tenant)

	createdTenant, err := s.tenantStore.CreateTenant(ctx, dbTenant)
	if err != nil {
		diag.AddError("failed_to_create_tenant", "failed to create tenant", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	result := mappers.MapTenantToDto(createdTenant)

	return result, diag
}

func (s *TenantService) UpdateTenant(ctx *appctx.AppContext, tenantRequest *TenantUpdateRequest) (*models.Tenant, *diagnostics.Diagnostics) {
	diag := diagnostics.New("update_tenant")
	defer diag.Complete()

	tenant := models.Tenant{
		ID:           tenantRequest.ID,
		Description:  tenantRequest.Description,
		Name:         tenantRequest.Name,
		Domain:       tenantRequest.Domain,
		OwnerID:      tenantRequest.OwnerID,
		ContactEmail: tenantRequest.ContactEmail,
	}

	dbTenant := mappers.MapTenantToEntity(&tenant)

	err := s.tenantStore.UpdateTenant(ctx, dbTenant)
	if err != nil {
		diag.AddError("failed_to_update_tenant", "failed to update tenant", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	return &tenant, diag
}

func (s *TenantService) DeleteTenant(ctx *appctx.AppContext, idOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_tenant")
	defer diag.Complete()

	dbTenant, err := s.tenantStore.GetTenantByIdOrSlug(ctx, idOrSlug)
	if err != nil {
		diag.AddError("failed_to_delete_tenant", "failed to delete tenant", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	err = s.tenantStore.DeleteTenant(ctx, dbTenant)
	if err != nil {
		diag.AddError("failed_to_delete_tenant", "failed to delete tenant", "tenant", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}
