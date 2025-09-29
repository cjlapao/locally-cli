package stores

import (
	"fmt"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	db_utils "github.com/cjlapao/locally-cli/internal/database/utils"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_utils "github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	environmentDataStoreInstance *EnvironmentDataStore
	environmentDataStoreOnce     sync.Once
)

type EnvironmentDataStoreInterface interface {
	GetEnvironments(ctx *appctx.AppContext, tenantID string) ([]entities.Environment, *diagnostics.Diagnostics)
	GetEnvironmentsByQuery(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Environment], *diagnostics.Diagnostics)
	GetEnvironmentBy(ctx *appctx.AppContext, tenantID string, environmentID string) (*entities.Environment, *diagnostics.Diagnostics)
	CreateEnvironment(ctx *appctx.AppContext, tenantID string, environment *entities.Environment) (*entities.Environment, *diagnostics.Diagnostics)
	UpdateEnvironment(ctx *appctx.AppContext, tenantID string, environment *entities.Environment) *diagnostics.Diagnostics
	DeleteEnvironment(ctx *appctx.AppContext, tenantID string, environmentID string) *diagnostics.Diagnostics
	GetEnvironmentVaults(ctx *appctx.AppContext, tenantID string, environmentID string) ([]entities.EnvironmentVault, *diagnostics.Diagnostics)
	GetEnvironmentVaultBy(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string) (*entities.EnvironmentVault, *diagnostics.Diagnostics)
	GetEnvironmentVaultByQuery(ctx *appctx.AppContext, tenantID string, environmentID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.EnvironmentVault], *diagnostics.Diagnostics)
	CreateEnvironmentVault(ctx *appctx.AppContext, tenantID string, environmentID string, vault *entities.EnvironmentVault) (*entities.EnvironmentVault, *diagnostics.Diagnostics)
	UpdateEnvironmentVault(ctx *appctx.AppContext, tenantID string, environmentID string, vault *entities.EnvironmentVault) *diagnostics.Diagnostics
	DeleteEnvironmentVault(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string) *diagnostics.Diagnostics
	GetEnvironmentVaultItems(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string) ([]entities.EnvironmentVaultItem, *diagnostics.Diagnostics)
	GetEnvironmentVaultItemBy(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string, itemID string) (*entities.EnvironmentVaultItem, *diagnostics.Diagnostics)
	GetEnvironmentVaultItemByQuery(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.EnvironmentVaultItem], *diagnostics.Diagnostics)
	CreateEnvironmentVaultItem(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string, item *entities.EnvironmentVaultItem) (*entities.EnvironmentVaultItem, *diagnostics.Diagnostics)
	UpdateEnvironmentVaultItem(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string, item *entities.EnvironmentVaultItem) *diagnostics.Diagnostics
	DeleteEnvironmentVaultItem(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string, itemID string) *diagnostics.Diagnostics
}

type EnvironmentDataStore struct {
	database.BaseDataStore
}

func GetEnvironmentDataStoreInstance() EnvironmentDataStoreInterface {
	return environmentDataStoreInstance
}

func InitializeEnvironmentDataStore() (EnvironmentDataStoreInterface, *diagnostics.Diagnostics) {
	diag := diagnostics.New("initialize_environment_data_store")
	cfg := config.GetInstance().Get()
	logging.Info("Initializing environment store...")

	environmentDataStoreOnce.Do(func() {
		dbService := database.GetInstance()
		if dbService == nil {
			diag.AddError("database_service_not_initialized", "database service not initialized", "environment_data_store", nil)
			return
		}

		store := &EnvironmentDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running environment migrations")
			if migrateDiag := store.Migrate(); migrateDiag.HasErrors() {
				diag.Append(migrateDiag)
				return
			}
			logging.Info("Environment migrations completed")
		}

		environmentDataStoreInstance = store
	})
	return environmentDataStoreInstance, diag
}

func (s *EnvironmentDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_environment_data_store")
	defer diag.Complete()

	if err := s.GetDB().AutoMigrate(&entities.Environment{}); err != nil {
		diag.AddError("failed_to_migrate_environment_table", "failed to migrate environment table", "environment_data_store", nil)
		return diag
	}
	if err := s.GetDB().AutoMigrate(&entities.EnvironmentVault{}); err != nil {
		diag.AddError("failed_to_migrate_environment_vault_table", "failed to migrate environment vault table", "environment_data_store", nil)
		return diag
	}
	if err := s.GetDB().AutoMigrate(&entities.EnvironmentVaultItem{}); err != nil {
		diag.AddError("failed_to_migrate_environment_vault_item_table", "failed to migrate environment vault item table", "environment_data_store", nil)
		return diag
	}

	// creating an index for the environment for the tenant, project and name to be unique
	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_environments_tenant_project_name ON environments(tenant_id, project_id, name);").Error; err != nil {
		diag.AddError("failed_to_create_unique_index_on_environments", fmt.Sprintf("failed to create unique index on environments: %s", err.Error()), "environment_data_store", nil)
		return diag
	}

	// creating an index for the environment vault for the tenant, project and name to be unique
	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_environment_vaults_tenant_name ON environment_vaults(tenant_id, environment_id, name);").Error; err != nil {
		diag.AddError("failed_to_create_unique_index_on_environment_vaults", fmt.Sprintf("failed to create unique index on environment vaults: %s", err.Error()), "environment_data_store", nil)
		return diag
	}

	// creating an index for the environment vault item for the tenant, project and key to be unique
	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_environment_vault_items_tenant_vault_key ON environment_vault_items(tenant_id,  environment_vault_id, key);").Error; err != nil {
		diag.AddError("failed_to_create_unique_index_on_environment_vault_items", fmt.Sprintf("failed to create unique index on environment vault items: %s", err.Error()), "environment_data_store", nil)
		return diag
	}

	return diag
}

func (s *EnvironmentDataStore) GetEnvironments(ctx *appctx.AppContext, tenantID string) ([]entities.Environment, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_environments")
	var environments []entities.Environment
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}

	if err := s.GetDB().
		Preload("Vaults", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Items")
		}).
		Where("tenant_id = ?", tenantID).
		Find(&environments).Error; err != nil {
		diag.AddError("failed_to_get_environments", "failed to get environments", "environment_data_store", nil)
		return nil, diag
	}

	return environments, diag
}

func (s *EnvironmentDataStore) GetEnvironmentsByQuery(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Environment], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_environments_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}

	db := s.GetDB()
	db = db.Preload("Vaults", func(db *gorm.DB) *gorm.DB {
		return db.Preload("Items")
	})

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := db_utils.QueryDatabase[entities.Environment](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_environments_by_query", "failed to get environments by query", "environment_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *EnvironmentDataStore) GetEnvironmentBy(ctx *appctx.AppContext, tenantID string, environmentID string) (*entities.Environment, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_environment_by")
	var environment entities.Environment
	if err := s.GetDB().
		Preload("Vaults", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Items")
		}).
		Where("tenant_id = ? AND id = ?", tenantID, environmentID).
		First(&environment).Error; err != nil {
		diag.AddError("failed_to_get_environment_by", "failed to get environment by", "environment_data_store", nil)
		return nil, diag
	}
	return &environment, diag
}

func (s *EnvironmentDataStore) CreateEnvironment(ctx *appctx.AppContext, tenantID string, environment *entities.Environment) (*entities.Environment, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_create_environment")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}

	if environment.ID == "" {
		environment.ID = uuid.New().String()
	}
	if environment.Slug == "" {
		environment.Slug = pkg_utils.Slugify(environment.Name)
	}
	environment.TenantID = tenantID
	environment.CreatedAt = time.Now()
	environment.UpdatedAt = time.Now()

	if err := s.GetDB().Create(environment).Error; err != nil {
		diag.AddError("failed_to_create_environment", "failed to create environment", "environment_data_store", nil)
		return nil, diag
	}
	return environment, diag
}

func (s *EnvironmentDataStore) UpdateEnvironment(ctx *appctx.AppContext, tenantID string, environment *entities.Environment) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_update_environment")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return diag
	}
	environment.UpdatedAt = time.Now()
	if environment.Slug == "" {
		environment.Slug = pkg_utils.Slugify(environment.Name)
	}
	environment.TenantID = tenantID

	// check if the environment exists in the database
	existingEnvironment, getEnvironmentDiag := s.GetEnvironmentBy(ctx, tenantID, environment.ID)
	if getEnvironmentDiag.HasErrors() {
		diag.Append(getEnvironmentDiag)
		return diag
	}
	if existingEnvironment == nil {
		diag.AddError("environment_not_found", "environment not found", "environment_data_store", nil)
		return diag
	}

	// using the partial update map to update the environment
	updates := db_utils.PartialUpdateMap(existingEnvironment, environment, "updated_at", "slug")
	if err := s.GetDB().
		Model(&entities.Environment{}).
		Where("id = ?", environment.ID).
		Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_environment", "failed to update environment", "environment_data_store", nil)
		return diag
	}

	return diag
}

func (s *EnvironmentDataStore) DeleteEnvironment(ctx *appctx.AppContext, tenantID string, environmentID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_delete_environment")
	if err := s.GetDB().
		Where("tenant_id = ?", tenantID).
		Where("id = ?", environmentID).
		Delete(&entities.Environment{}).Error; err != nil {
		diag.AddError("failed_to_delete_environment", "failed to delete environment", "environment_data_store", nil)
		return diag
	}
	return diag
}

func (s *EnvironmentDataStore) GetEnvironmentVaults(ctx *appctx.AppContext, tenantID string, environmentID string) ([]entities.EnvironmentVault, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_environment_vaults")
	var environmentVaults []entities.EnvironmentVault
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if environmentID == "" {
		diag.AddError("environment_id_cannot_be_empty", "environment ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if err := s.GetDB().
		Preload("Items").
		Where("tenant_id = ?", tenantID).
		Where("environment_id = ?", environmentID).
		Find(&environmentVaults).Error; err != nil {
		diag.AddError("failed_to_get_environment_vaults", "failed to get environment vaults", "environment_data_store", nil)
		return nil, diag
	}
	return environmentVaults, diag
}

func (s *EnvironmentDataStore) GetEnvironmentVaultBy(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string) (*entities.EnvironmentVault, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_environment_vault_by")
	var environmentVault entities.EnvironmentVault
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if environmentID == "" {
		diag.AddError("environment_id_cannot_be_empty", "environment ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if vaultID == "" {
		diag.AddError("vault_id_cannot_be_empty", "vault ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if err := s.GetDB().
		Preload("Items").
		Where("tenant_id = ?", tenantID).
		Where("environment_id = ?", environmentID).
		Where("id = ?", vaultID).
		First(&environmentVault).Error; err != nil {
		diag.AddError("failed_to_get_environment_vault_by", "failed to get environment vault by", "environment_data_store", nil)
		return nil, diag
	}
	return &environmentVault, diag
}

func (s *EnvironmentDataStore) GetEnvironmentVaultByQuery(ctx *appctx.AppContext, tenantID string, environmentID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.EnvironmentVault], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_environment_vault_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if environmentID == "" {
		diag.AddError("environment_id_cannot_be_empty", "environment ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}
	db := s.GetDB()
	db = db.Preload("Items")
	db = db.Where("tenant_id = ?", tenantID)
	db = db.Where("environment_id = ?", environmentID)
	result, err := db_utils.QueryDatabase[entities.EnvironmentVault](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_environment_vault_by_query", "failed to get environment vault by query", "environment_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *EnvironmentDataStore) CreateEnvironmentVault(ctx *appctx.AppContext, tenantID string, environmentID string, vault *entities.EnvironmentVault) (*entities.EnvironmentVault, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_create_environment_vault")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if environmentID == "" {
		diag.AddError("environment_id_cannot_be_empty", "environment ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if vault.ID == "" {
		vault.ID = uuid.New().String()
	}
	vault.EnvironmentID = environmentID
	vault.TenantID = tenantID
	vault.CreatedAt = time.Now()
	vault.UpdatedAt = time.Now()

	// check if the environment vault exists in the database
	existingEnvironmentVault, getEnvironmentVaultDiag := s.GetEnvironmentVaultBy(ctx, tenantID, environmentID, vault.ID)
	if getEnvironmentVaultDiag.HasErrors() {
		diag.Append(getEnvironmentVaultDiag)
		return nil, diag
	}
	if existingEnvironmentVault != nil {
		diag.AddError("environment_vault_already_exists", "environment vault already exists", "environment_data_store", nil)
		return nil, diag
	}

	if err := s.GetDB().
		Create(vault).Error; err != nil {
		diag.AddError("failed_to_create_environment_vault", "failed to create environment vault", "environment_data_store", nil)
		return nil, diag
	}
	return vault, diag
}

func (s *EnvironmentDataStore) UpdateEnvironmentVault(ctx *appctx.AppContext, tenantID string, environmentID string, vault *entities.EnvironmentVault) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_update_environment_vault")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return diag
	}
	if environmentID == "" {
		diag.AddError("environment_id_cannot_be_empty", "environment ID cannot be empty", "environment_data_store")
		return diag
	}
	if vault.ID == "" {
		diag.AddError("vault_id_cannot_be_empty", "vault ID cannot be empty", "environment_data_store")
		return diag
	}
	vault.EnvironmentID = environmentID
	vault.TenantID = tenantID
	vault.UpdatedAt = time.Now()

	// check if the environment vault exists in the database
	existingEnvironmentVault, getEnvironmentVaultDiag := s.GetEnvironmentVaultBy(ctx, tenantID, environmentID, vault.ID)
	if getEnvironmentVaultDiag.HasErrors() {
		diag.Append(getEnvironmentVaultDiag)
		return diag
	}
	if existingEnvironmentVault == nil {
		diag.AddError("environment_vault_not_found", "environment vault not found", "environment_data_store", nil)
		return diag
	}

	// using the partial update map to update the environment vault
	updates := db_utils.PartialUpdateMap(existingEnvironmentVault, vault, "updated_at")
	if err := s.GetDB().
		Model(&entities.EnvironmentVault{}).
		Where("id = ?", vault.ID).
		Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_environment_vault", "failed to update environment vault", "environment_data_store", nil)
		return diag
	}
	return diag
}

func (s *EnvironmentDataStore) DeleteEnvironmentVault(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_delete_environment_vault")
	if err := s.GetDB().
		Where("tenant_id = ?", tenantID).
		Where("environment_id = ?", environmentID).
		Where("id = ?", vaultID).
		Delete(&entities.EnvironmentVault{}).Error; err != nil {
		diag.AddError("failed_to_delete_environment_vault", "failed to delete environment vault", "environment_data_store", nil)
		return diag
	}
	return diag
}

func (s *EnvironmentDataStore) GetEnvironmentVaultItems(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string) ([]entities.EnvironmentVaultItem, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_environment_vault_items")
	var environmentVaultItems []entities.EnvironmentVaultItem
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if environmentID == "" {
		diag.AddError("environment_id_cannot_be_empty", "environment ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if vaultID == "" {
		diag.AddError("vault_id_cannot_be_empty", "vault ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if err := s.GetDB().
		Where("tenant_id = ?", tenantID).
		Where("environment_id = ?", environmentID).
		Where("vault_id = ?", vaultID).
		Find(&environmentVaultItems).Error; err != nil {
		diag.AddError("failed_to_get_environment_vault_items", "failed to get environment vault items", "environment_data_store", nil)
		return nil, diag
	}
	return environmentVaultItems, diag
}

func (s *EnvironmentDataStore) GetEnvironmentVaultItemBy(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string, itemID string) (*entities.EnvironmentVaultItem, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_environment_vault_item_by")
	var environmentVaultItem entities.EnvironmentVaultItem
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if environmentID == "" {
		diag.AddError("environment_id_cannot_be_empty", "environment ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if vaultID == "" {
		diag.AddError("vault_id_cannot_be_empty", "vault ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if itemID == "" {
		diag.AddError("item_id_cannot_be_empty", "item ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if err := s.GetDB().
		Where("tenant_id = ?", tenantID).
		Where("environment_id = ?", environmentID).
		Where("vault_id = ?", vaultID).
		Where("id = ?", itemID).
		First(&environmentVaultItem).Error; err != nil {
		diag.AddError("failed_to_get_environment_vault_item_by", "failed to get environment vault item by", "environment_data_store", nil)
		return nil, diag
	}
	return &environmentVaultItem, diag
}

func (s *EnvironmentDataStore) GetEnvironmentVaultItemByQuery(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.EnvironmentVaultItem], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_environment_vault_item_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if environmentID == "" {
		diag.AddError("environment_id_cannot_be_empty", "environment ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if vaultID == "" {
		diag.AddError("vault_id_cannot_be_empty", "vault ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}
	db := s.GetDB()
	db = db.Where("tenant_id = ?", tenantID)
	db = db.Where("environment_id = ?", environmentID)
	db = db.Where("vault_id = ?", vaultID)
	result, err := db_utils.QueryDatabase[entities.EnvironmentVaultItem](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_environment_vault_item_by_query", "failed to get environment vault item by query", "environment_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *EnvironmentDataStore) CreateEnvironmentVaultItem(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string, item *entities.EnvironmentVaultItem) (*entities.EnvironmentVaultItem, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_create_environment_vault_item")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if environmentID == "" {
		diag.AddError("environment_id_cannot_be_empty", "environment ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if vaultID == "" {
		diag.AddError("vault_id_cannot_be_empty", "vault ID cannot be empty", "environment_data_store")
		return nil, diag
	}
	if item.ID == "" {
		item.ID = uuid.New().String()
	}
	item.EnvironmentVaultID = vaultID
	item.TenantID = tenantID
	item.CreatedAt = time.Now()
	item.UpdatedAt = time.Now()

	// check if the environment vault item exists in the database
	existingEnvironmentVaultItem, getEnvironmentVaultItemDiag := s.GetEnvironmentVaultItemBy(ctx, tenantID, environmentID, vaultID, item.ID)
	if getEnvironmentVaultItemDiag.HasErrors() {
		diag.Append(getEnvironmentVaultItemDiag)
		return nil, diag
	}
	if existingEnvironmentVaultItem != nil {
		diag.AddError("environment_vault_item_already_exists", "environment vault item already exists", "environment_data_store", nil)
		return nil, diag
	}
	if err := s.GetDB().Create(item).Error; err != nil {
		diag.AddError("failed_to_create_environment_vault_item", "failed to create environment vault item", "environment_data_store", nil)
		return nil, diag
	}
	return item, diag
}

func (s *EnvironmentDataStore) UpdateEnvironmentVaultItem(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string, item *entities.EnvironmentVaultItem) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_update_environment_vault_item")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "environment_data_store")
		return diag
	}
	if environmentID == "" {
		diag.AddError("environment_id_cannot_be_empty", "environment ID cannot be empty", "environment_data_store")
		return diag
	}
	if vaultID == "" {
		diag.AddError("vault_id_cannot_be_empty", "vault ID cannot be empty", "environment_data_store")
		return diag
	}
	if item.ID == "" {
		diag.AddError("item_id_cannot_be_empty", "item ID cannot be empty", "environment_data_store")
		return diag
	}
	item.EnvironmentVaultID = vaultID
	item.TenantID = tenantID
	item.UpdatedAt = time.Now()

	// check if the environment vault item exists in the database
	existingEnvironmentVaultItem, getEnvironmentVaultItemDiag := s.GetEnvironmentVaultItemBy(ctx, tenantID, environmentID, vaultID, item.ID)
	if getEnvironmentVaultItemDiag.HasErrors() {
		diag.Append(getEnvironmentVaultItemDiag)
		return diag
	}
	if existingEnvironmentVaultItem == nil {
		diag.AddError("environment_vault_item_not_found", "environment vault item not found", "environment_data_store", nil)
		return diag
	}

	// using the partial update map to update the environment vault item
	updates := db_utils.PartialUpdateMap(existingEnvironmentVaultItem, item, "updated_at")
	if err := s.GetDB().
		Model(&entities.EnvironmentVaultItem{}).
		Where("id = ?", item.ID).
		Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_environment_vault_item", "failed to update environment vault item", "environment_data_store", nil)
		return diag
	}
	return diag
}

func (s *EnvironmentDataStore) DeleteEnvironmentVaultItem(ctx *appctx.AppContext, tenantID string, environmentID string, vaultID string, itemID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_delete_environment_vault_item")
	if err := s.GetDB().
		Where("tenant_id = ?", tenantID).
		Where("environment_id = ?", environmentID).
		Where("vault_id = ?", vaultID).
		Where("id = ?", itemID).
		Delete(&entities.EnvironmentVaultItem{}).Error; err != nil {
		diag.AddError("failed_to_delete_environment_vault_item", "failed to delete environment vault item", "environment_data_store", nil)
		return diag
	}
	return diag
}
