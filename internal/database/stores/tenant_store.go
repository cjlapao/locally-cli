package stores

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/utils"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_utils "github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	tenantDataStoreInstance *TenantDataStore
	tenantDataStoreOnce     sync.Once
)

type TenantDataStoreInterface interface {
	GetTenantByIdOrSlug(ctx *appctx.AppContext, idOrSlug string) (*entities.Tenant, *diagnostics.Diagnostics)
	GetTenants(ctx *appctx.AppContext) ([]entities.Tenant, *diagnostics.Diagnostics)
	GetTenantsByQuery(ctx *appctx.AppContext, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Tenant], *diagnostics.Diagnostics)
	CreateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) (*entities.Tenant, *diagnostics.Diagnostics)
	UpdateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) *diagnostics.Diagnostics
	DeleteTenant(ctx *appctx.AppContext, id string) *diagnostics.Diagnostics
	Migrate() *diagnostics.Diagnostics
}

type TenantDataStore struct {
	database.BaseDataStore
}

func GetTenantDataStoreInstance() TenantDataStoreInterface {
	return tenantDataStoreInstance
}

func InitializeTenantDataStore() (TenantDataStoreInterface, *diagnostics.Diagnostics) {
	diag := diagnostics.New("initialize_tenant_data_store")
	cfg := config.GetInstance().Get()
	logging.Info("Initializing tenant store...")

	tenantDataStoreOnce.Do(func() {
		dbService := database.GetInstance()
		if dbService == nil {
			diag.AddError("database_service_not_initialized", "database service not initialized", "tenant_data_store", nil)
			return
		}

		store := &TenantDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running tenant migrations")
			if migrateDiag := store.Migrate(); migrateDiag.HasErrors() {
				diag.Append(migrateDiag)
				return
			}
			logging.Info("Tenant migrations completed")
		}

		tenantDataStoreInstance = store
	})

	logging.Info("Tenant store initialized successfully")
	return tenantDataStoreInstance, diag
}

func (s *TenantDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_tenant_data_store")
	if err := s.GetDB().AutoMigrate(&entities.Tenant{}); err != nil {
		diag.AddError("failed_to_migrate_tenant_table", "failed to migrate tenant table", "tenant_data_store", nil)
		return diag
	}

	return diag
}

func (s *TenantDataStore) GetTenantByIdOrSlug(ctx *appctx.AppContext, idOrSlug string) (*entities.Tenant, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_tenant_by_id_or_slug")
	var tenant entities.Tenant
	if err := s.GetDB().Where("id = ? OR slug = ?", idOrSlug, idOrSlug).First(&tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_tenant", fmt.Sprintf("failed to get tenant: %v", err), "tenant_data_store", nil)
		return nil, diag
	}
	return &tenant, diag
}

func (s *TenantDataStore) GetTenants(ctx *appctx.AppContext) ([]entities.Tenant, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_tenants")
	var tenants []entities.Tenant
	if err := s.GetDB().Find(&tenants).Error; err != nil {
		diag.AddError("failed_to_get_tenants", fmt.Sprintf("failed to get tenants: %v", err), "tenant_data_store", nil)
		return nil, diag
	}
	return tenants, diag
}

func (s *TenantDataStore) GetTenantsByQuery(ctx *appctx.AppContext, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Tenant], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_tenants_by_query")
	result, err := utils.QueryDatabase[entities.Tenant](s.GetDB(), "", queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_tenants", fmt.Sprintf("failed to get tenants: %v", err), "tenant_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *TenantDataStore) CreateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) (*entities.Tenant, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_create_tenant")
	if tenant.ID == "" {
		tenant.ID = uuid.New().String()
	}
	tenant.Slug = pkg_utils.Slugify(tenant.Name)
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	if err := s.GetDB().Create(tenant).Error; err != nil {
		diag.AddError("failed_to_create_tenant", fmt.Sprintf("failed to create tenant: %v", err), "tenant_data_store", nil)
		return nil, diag
	}
	return tenant, diag
}

func (s *TenantDataStore) UpdateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_update_tenant")
	// Get the original tenant from the database
	originalTenant, err := s.GetTenantByIdOrSlug(ctx, tenant.ID)
	if err != nil {
		diag.Append(err)
		return diag
	}
	if tenant.Name != "" {
		tenant.Slug = pkg_utils.Slugify(tenant.Name)
	}

	// Generate partial update map by comparing original with updated
	updates := utils.PartialUpdateMap(originalTenant, tenant, "updated_at", "slug")
	if err := s.GetDB().Model(&entities.Tenant{}).Where("id = ?", tenant.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_tenant", fmt.Sprintf("failed to update tenant: %v", err), "tenant_data_store", nil)
		return diag
	}
	return diag
}

func (s *TenantDataStore) DeleteTenant(ctx *appctx.AppContext, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_delete_tenant")
	// deleting all claims for the tenant
	if err := s.GetDB().Delete(&entities.Claim{}, "tenant_id = ?", id).Error; err != nil {
		diag.AddError("failed_to_delete_claims", fmt.Sprintf("failed to delete claims: %v", err), "tenant_data_store", nil)
		return diag
	}
	// deleting all roles for the tenant
	if err := s.GetDB().Delete(&entities.Role{}, "tenant_id = ?", id).Error; err != nil {
		diag.AddError("failed_to_delete_roles", fmt.Sprintf("failed to delete roles: %v", err), "tenant_data_store", nil)
		return diag
	}
	// deleting all users for the tenant
	if err := s.GetDB().Delete(&entities.User{}, "tenant_id = ?", id).Error; err != nil {
		diag.AddError("failed_to_delete_users", fmt.Sprintf("failed to delete users: %v", err), "tenant_data_store", nil)
		return diag
	}

	if err := s.GetDB().Delete(&entities.Tenant{}, "id = ?", id).Error; err != nil {
		diag.AddError("failed_to_delete_tenant", fmt.Sprintf("failed to delete tenant: %v", err), "tenant_data_store", nil)
		return diag
	}
	return diag
}
