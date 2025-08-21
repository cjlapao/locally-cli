package stores

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"gorm.io/gorm"
)

var (
	configurationDataStoreInstance *ConfigurationDataStore
	configurationDataStoreOnce     sync.Once
)

type ConfigurationDataStoreInterface interface {
	GetConfigurationValue(ctx context.Context, tenantID string, key string, value interface{}) (interface{}, *diagnostics.Diagnostics)
}

type ConfigurationDataStore struct {
	database.BaseDataStore
}

// GetConfigurationDataStoreInstance returns the singleton instance of the ConfigurationDataStore
func GetConfigurationDataStoreInstance() ConfigurationDataStoreInterface {
	return configurationDataStoreInstance
}

// InitializeConfigurationDataStore initializes the ConfigurationDataStore singleton
func InitializeConfigurationDataStore() (ConfigurationDataStoreInterface, *diagnostics.Diagnostics) {
	diag := diagnostics.New("initialize_configuration_data_store")
	cfg := config.GetInstance().Get()
	configurationDataStoreOnce.Do(func() {
		dbService := database.GetInstance()
		if dbService == nil {
			diag.AddError("database_service_not_initialized", "database service not initialized", "configuration_data_store", nil)
			return
		}

		store := &ConfigurationDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running configuration migrations")
			if migrateDiag := store.Migrate(); migrateDiag.HasErrors() {
				diag.Append(migrateDiag)
				return
			}
			logging.Info("Configuration migrations completed")
		}

		configurationDataStoreInstance = store
	})

	logging.Info("Configuration store initialized successfully")
	return configurationDataStoreInstance, diag
}

func (s *ConfigurationDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_configuration_data_store")
	if err := s.GetDB().AutoMigrate(&entities.Configuration{}); err != nil {
		diag.AddError("failed_to_migrate_configuration_table", fmt.Sprintf("failed to migrate configuration table: %v", err), "configuration_data_store", nil)
		return diag
	}
	return diag
}

func (s *ConfigurationDataStore) GetConfigurationValue(ctx context.Context, tenantID string, key string, value interface{}) (interface{}, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_configuration_value")
	db := s.GetDB()
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "configuration_data_store")
		return nil, diag
	}
	if key == "" {
		diag.AddError("key_cannot_be_empty", "key cannot be empty", "configuration_data_store")
		return nil, diag
	}

	err := db.Where("key = ?", key).
		Where("tenant_id = ?", tenantID).
		First(&entities.Configuration{}).
		Scan(&value).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_configuration_value", fmt.Sprintf("failed to get configuration value: %v", err), "configuration_data_store", nil)
		return nil, diag
	}

	return value, diag
}
