package stores

import (
	"errors"
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
	GetTenantBySlug(ctx *appctx.AppContext, slug string) (*entities.Tenant, error)
	GetTenantByID(ctx *appctx.AppContext, id string) (*entities.Tenant, error)
	GetTenantByIdOrSlug(ctx *appctx.AppContext, idOrSlug string) (*entities.Tenant, error)
	GetTenants(ctx *appctx.AppContext) ([]entities.Tenant, error)
	GetTenantsByFilter(ctx *appctx.AppContext, filterObj *filters.Filter) (*filters.FilterResponse[entities.Tenant], error)
	CreateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) (*entities.Tenant, error)
	UpdateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) error
	DeleteTenant(ctx *appctx.AppContext, tenant *entities.Tenant) error
	Migrate() *diagnostics.Diagnostics
}

type TenantDataStore struct {
	database.BaseDataStore
}

func GetTenantDataStoreInstance() TenantDataStoreInterface {
	return tenantDataStoreInstance
}

func InitializeTenantDataStore() *diagnostics.Diagnostics {
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
	return diag
}

func (s *TenantDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_tenant_data_store")
	if err := s.GetDB().AutoMigrate(&entities.Tenant{}); err != nil {
		diag.AddError("failed_to_migrate_tenant_table", "failed to migrate tenant table", "tenant_data_store", nil)
		return diag
	}

	return diag
}

func (s *TenantDataStore) GetTenantBySlug(ctx *appctx.AppContext, slug string) (*entities.Tenant, error) {
	var tenant entities.Tenant
	if err := s.GetDB().Where("slug = ?", slug).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *TenantDataStore) GetTenantByID(ctx *appctx.AppContext, id string) (*entities.Tenant, error) {
	var tenant entities.Tenant
	if err := s.GetDB().Where("id = ?", id).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *TenantDataStore) GetTenantByIdOrSlug(ctx *appctx.AppContext, idOrSlug string) (*entities.Tenant, error) {
	var tenant entities.Tenant
	if err := s.GetDB().Where("id = ? OR slug = ?", idOrSlug, idOrSlug).First(&tenant).Error; err != nil {
		return nil, err
	}
	return &tenant, nil
}

func (s *TenantDataStore) GetTenants(ctx *appctx.AppContext) ([]entities.Tenant, error) {
	var tenants []entities.Tenant
	if err := s.GetDB().Find(&tenants).Error; err != nil {
		return nil, err
	}
	return tenants, nil
}

func (s *TenantDataStore) GetTenantsByFilter(ctx *appctx.AppContext, filterObj *filters.Filter) (*filters.FilterResponse[entities.Tenant], error) {
	return utils.PaginatedQuery(s.GetDB(), filterObj, entities.Tenant{})
}

func (s *TenantDataStore) CreateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) (*entities.Tenant, error) {
	tenant.ID = uuid.New().String()
	tenant.Slug = pkg_utils.Slugify(tenant.Name)
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()
	if err := s.GetDB().Create(tenant).Error; err != nil {
		return nil, err
	}
	return tenant, nil
}

func (s *TenantDataStore) UpdateTenant(ctx *appctx.AppContext, tenant *entities.Tenant) error {
	// Get the original tenant from the database
	originalTenant, err := s.GetTenantByID(ctx, tenant.ID)
	if err != nil {
		return err
	}
	if tenant.Name != "" {
		tenant.Slug = pkg_utils.Slugify(tenant.Name)
	}

	// Generate partial update map by comparing original with updated
	updates := utils.PartialUpdateMap(originalTenant, tenant, "updated_at", "slug")
	if err := s.GetDB().Model(&entities.Tenant{}).Where("id = ?", tenant.ID).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

func (s *TenantDataStore) DeleteTenant(ctx *appctx.AppContext, tenant *entities.Tenant) error {
	if err := s.GetDB().Delete(tenant).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil
		}
		return err
	}
	return nil
}
