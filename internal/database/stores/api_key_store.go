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
	"github.com/cjlapao/locally-cli/internal/encryption"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	authDataStoreInstance *ApiKeyDataStore
	authDataStoreOnce     sync.Once
)

type ApiKeyStoreInterface interface {
	CreateAPIKey(ctx *appctx.AppContext, apiKey *entities.ApiKey) (*entities.ApiKey, error)
	GetAPIKeyByHash(ctx *appctx.AppContext, keyHash string) (*entities.ApiKey, error)
	GetAPIKeyByPrefix(ctx *appctx.AppContext, keyPrefix string) (*entities.ApiKey, error)
	GetFilteredAPIKeys(ctx *appctx.AppContext, filter *filters.Filter) (*filters.FilterResponse[entities.ApiKey], error)
	GetPaginatedAPIKeys(ctx *appctx.AppContext, pagination *filters.Pagination) (*filters.PaginationResponse[entities.ApiKey], error)
	GetAPIKeyByID(ctx *appctx.AppContext, id string) (*entities.ApiKey, error)
	RevokeAPIKey(ctx *appctx.AppContext, id string, revokedBy string, reason string) error
	DeleteAPIKey(ctx *appctx.AppContext, id string) error
	GetDB() *gorm.DB
}

// ApiKeyDataStore handles auth-specific database operations
type ApiKeyDataStore struct {
	database.BaseDataStore
}

// GetApiKeyDataStoreInstance returns the singleton instance of the auth store
func GetApiKeyDataStoreInstance() ApiKeyStoreInterface {
	return authDataStoreInstance
}

// InitializeApiKeyDataStore initializes the api key store singleton
func InitializeApiKeyDataStore() (ApiKeyStoreInterface, *diagnostics.Diagnostics) {
	diag := diagnostics.New("initialize_api_key_data_store")
	cfg := config.GetInstance().Get()
	authDataStoreOnce.Do(func() {
		// Get the database service instance
		dbService := database.GetInstance()
		if dbService == nil {
			diag.AddError("database_service_not_initialized", "database service not initialized", "auth_data_store", nil)
			return
		}

		store := &ApiKeyDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running api key migrations")
			if err := store.Migrate(); err != nil {
				diag.AddError("failed_to_run_api_key_migrations", "failed to run api key migrations", "api_key_data_store", nil)
				return
			}
			logging.Info("Api key migrations completed")
		}

		authDataStoreInstance = store
	})

	logging.Info("Api key store initialized successfully")
	return authDataStoreInstance, diag
}

// Migrate implements the DataStore interface
func (s *ApiKeyDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.ApiKey{}); err != nil {
		return fmt.Errorf("failed to migrate api_keys table: %w", err)
	}

	if err := s.GetDB().AutoMigrate(&entities.ApiKeyClaims{}); err != nil {
		return fmt.Errorf("failed to migrate api_key_claims table: %w", err)
	}

	// add unique index to the api_key_claims table
	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_api_key_claims ON api_key_claims (api_key_id, claim_id);").Error; err != nil {
		return fmt.Errorf("failed to create unique index on api_key_claims table: %w", err)
	}

	return nil
}

// CreateAPIKey creates a new API key for a user
func (s *ApiKeyDataStore) CreateAPIKey(ctx *appctx.AppContext, apiKey *entities.ApiKey) (*entities.ApiKey, error) {
	apiKey.ID = uuid.New().String()
	apiKey.CreatedAt = time.Now()
	apiKey.UpdatedAt = time.Now()
	apiKey.BaseModelWithTenant = *utils.GetTenantBaseModelFromContext(ctx)

	// Hash the API key before storing
	encryptionService := encryption.GetInstance()
	keyHash, err := encryptionService.HashPassword(apiKey.KeyHash)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key: %w", err)
	}
	apiKey.KeyHash = keyHash

	result := s.GetDB().Create(apiKey)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create API key: %w", result.Error)
	}

	return apiKey, nil
}

// GetAPIKeyByHash retrieves an API key by its hash
func (s *ApiKeyDataStore) GetAPIKeyByHash(ctx *appctx.AppContext, keyHash string) (*entities.ApiKey, error) {
	var apiKey entities.ApiKey
	result := s.GetDB().First(&apiKey, "key_hash = ?", keyHash)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by hash: %w", result.Error)
	}
	return &apiKey, nil
}

// GetAPIKeyByPrefix retrieves an API key by its prefix (for validation)
func (s *ApiKeyDataStore) GetAPIKeyByPrefix(ctx *appctx.AppContext, keyPrefix string) (*entities.ApiKey, error) {
	var apiKey entities.ApiKey
	result := s.GetDB().First(&apiKey, "key_prefix = ?", keyPrefix)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by prefix: %w", result.Error)
	}
	return &apiKey, nil
}

// GetAPIKeyByID retrieves an API key by ID
func (s *ApiKeyDataStore) GetAPIKeyByID(ctx *appctx.AppContext, id string) (*entities.ApiKey, error) {
	var apiKey entities.ApiKey
	result := s.GetDB().First(&apiKey, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by ID: %w", result.Error)
	}
	return &apiKey, nil
}

// ListAPIKeysByUserIDWithFilter retrieves API keys for a user with pagination and filtering
func (s *ApiKeyDataStore) ListAPIKeysByUserIDWithFilter(ctx *appctx.AppContext, userID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.ApiKey], error) {
	// Add the user_id filter to the existing filter
	filterObj.WithField("user_id", filters.FilterOperatorEqual, userID, filters.FilterJoinerAnd)

	// Use the generic pagination helper
	return utils.PaginatedFilteredQuery(s.GetDB(), "", filterObj, entities.ApiKey{})
}

// RevokeAPIKey revokes an API key
func (s *ApiKeyDataStore) RevokeAPIKey(ctx *appctx.AppContext, id string, revokedBy string, reason string) error {
	now := time.Now()
	result := s.GetDB().Model(&entities.ApiKey{}).Where("id = ?", id).Updates(map[string]interface{}{
		"is_active":         false,
		"revoked_at":        now,
		"revoked_by":        revokedBy,
		"revocation_reason": reason,
		"updated_at":        now,
	})
	if result.Error != nil {
		return fmt.Errorf("failed to revoke API key: %w", result.Error)
	}
	return nil
}

// DeleteAPIKey permanently deletes an API key
func (s *ApiKeyDataStore) DeleteAPIKey(ctx *appctx.AppContext, id string) error {
	result := s.GetDB().Delete(&entities.ApiKey{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete API key: %w", result.Error)
	}
	return nil
}

// CleanupExpiredAPIKeys removes expired API keys
func (s *ApiKeyDataStore) CleanupExpiredAPIKeys(ctx *appctx.AppContext) error {
	now := time.Now()
	result := s.GetDB().Where("expires_at IS NOT NULL AND expires_at < ?", now).Delete(&entities.ApiKey{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired API keys: %w", result.Error)
	}
	return nil
}

// GetFilteredAPIKeys retrieves API keys based on a filter
func (s *ApiKeyDataStore) GetFilteredAPIKeys(ctx *appctx.AppContext, filter *filters.Filter) (*filters.FilterResponse[entities.ApiKey], error) {
	return utils.PaginatedFilteredQueryWithPreload(s.GetDB(), "", filter, entities.ApiKey{}, "Claims")
}

// GetPaginatedAPIKeys retrieves paginated API keys based on a pagination and filter
func (s *ApiKeyDataStore) GetPaginatedAPIKeys(ctx *appctx.AppContext, pagination *filters.Pagination) (*filters.PaginationResponse[entities.ApiKey], error) {
	return utils.PaginatedQuery(s.GetDB(), "", pagination, entities.ApiKey{}, "Claims")
}
