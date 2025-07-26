package stores

import (
	"context"
	"fmt"
	"sync"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
)

var (
	configurationDataStoreInstance *ConfigurationDataStore
	configurationDataStoreOnce     sync.Once
)

type ConfigurationDataStore struct {
	database.BaseDataStore
}

// GetConfigurationDataStoreInstance returns the singleton instance of the ConfigurationDataStore
func GetConfigurationDataStoreInstance() *ConfigurationDataStore {
	return configurationDataStoreInstance
}

// InitializeConfigurationDataStore initializes the ConfigurationDataStore singleton
func InitializeConfigurationDataStore() *diagnostics.Diagnostics {
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

	return diag
}

func (s *ConfigurationDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_configuration_data_store")
	if err := s.GetDB().AutoMigrate(&entities.Configuration{}); err != nil {
		diag.AddError("failed_to_migrate_configuration_table", fmt.Sprintf("failed to migrate configuration table, %v", err), "configuration_data_store", nil)
		return diag
	}
	return diag
}

func (s *ConfigurationDataStore) GetConfigurationValue(ctx context.Context, key string, value interface{}) (interface{}, error) {
	db := s.GetDB()

	db.Where("key = ?", key).First(&entities.Configuration{}).Scan(&value)

	return value, nil
}
