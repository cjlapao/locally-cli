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
	authDataStoreInstance *AuthDataStore
	authDataStoreOnce     sync.Once
)

type AuthDataStoreInterface interface {
	CreateAPIKey(ctx *appctx.AppContext, apiKey *entities.APIKey) (*entities.APIKey, error)
	GetAPIKeyByHash(ctx *appctx.AppContext, keyHash string) (*entities.APIKey, error)
	GetAPIKeyByPrefix(ctx *appctx.AppContext, keyPrefix string) (*entities.APIKey, error)
	GetAPIKeyByID(ctx *appctx.AppContext, id string) (*entities.APIKey, error)
	UpdateAPIKeyLastUsed(ctx *appctx.AppContext, id string) error
	ListAPIKeysByUserID(ctx *appctx.AppContext, userID string) ([]entities.APIKey, error)
	RevokeAPIKey(ctx *appctx.AppContext, id string, revokedBy string, reason string) error
	DeleteAPIKey(ctx *appctx.AppContext, id string) error
}

// AuthDataStore handles auth-specific database operations
type AuthDataStore struct {
	database.BaseDataStore
}

// GetAuthDataStoreInstance returns the singleton instance of the auth store
func GetAuthDataStoreInstance() AuthDataStoreInterface {
	return authDataStoreInstance
}

// InitializeAuthDataStore initializes the auth store singleton
func InitializeAuthDataStore() (AuthDataStoreInterface, *diagnostics.Diagnostics) {
	diag := diagnostics.New("initialize_auth_data_store")
	cfg := config.GetInstance().Get()
	authDataStoreOnce.Do(func() {
		// Get the database service instance
		dbService := database.GetInstance()
		if dbService == nil {
			diag.AddError("database_service_not_initialized", "database service not initialized", "auth_data_store", nil)
			return
		}

		store := &AuthDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running auth migrations")
			if err := store.Migrate(); err != nil {
				diag.AddError("failed_to_run_auth_migrations", "failed to run auth migrations", "auth_data_store", nil)
				return
			}
			logging.Info("Auth migrations completed")
		}

		authDataStoreInstance = store
	})

	logging.Info("Auth store initialized successfully")
	return authDataStoreInstance, diag
}

// Migrate implements the DataStore interface
func (s *AuthDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.APIKey{}); err != nil {
		return fmt.Errorf("failed to migrate api_keys table: %w", err)
	}

	if err := s.GetDB().AutoMigrate(&entities.APIKeyUsage{}); err != nil {
		return fmt.Errorf("failed to migrate api_key_usage table: %w", err)
	}

	return nil
}

// CreateAPIKey creates a new API key for a user
func (s *AuthDataStore) CreateAPIKey(ctx *appctx.AppContext, apiKey *entities.APIKey) (*entities.APIKey, error) {
	apiKey.ID = uuid.New().String()
	apiKey.CreatedAt = time.Now()
	apiKey.UpdatedAt = time.Now()

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
func (s *AuthDataStore) GetAPIKeyByHash(ctx *appctx.AppContext, keyHash string) (*entities.APIKey, error) {
	var apiKey entities.APIKey
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
func (s *AuthDataStore) GetAPIKeyByPrefix(ctx *appctx.AppContext, keyPrefix string) (*entities.APIKey, error) {
	var apiKey entities.APIKey
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
func (s *AuthDataStore) GetAPIKeyByID(ctx *appctx.AppContext, id string) (*entities.APIKey, error) {
	var apiKey entities.APIKey
	result := s.GetDB().First(&apiKey, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by ID: %w", result.Error)
	}
	return &apiKey, nil
}

// ListAPIKeysByUserID retrieves all API keys for a user
func (s *AuthDataStore) ListAPIKeysByUserID(ctx *appctx.AppContext, userID string) ([]entities.APIKey, error) {
	var apiKeys []entities.APIKey
	result := s.GetDB().Where("user_id = ?", userID).Order("created_at DESC").Find(&apiKeys)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", result.Error)
	}
	return apiKeys, nil
}

// ListAPIKeysByUserIDWithFilter retrieves API keys for a user with pagination and filtering
func (s *AuthDataStore) ListAPIKeysByUserIDWithFilter(ctx *appctx.AppContext, userID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.APIKey], error) {
	// Add the user_id filter to the existing filter
	filterObj.WithField("user_id", filters.FilterOperatorEqual, userID, filters.FilterJoinerAnd)

	// Use the generic pagination helper
	return utils.PaginatedFilteredQuery(s.GetDB(), "", filterObj, entities.APIKey{})
}

// UpdateAPIKeyLastUsed updates the last used timestamp for an API key
func (s *AuthDataStore) UpdateAPIKeyLastUsed(ctx *appctx.AppContext, id string) error {
	now := time.Now()
	result := s.GetDB().Model(&entities.APIKey{}).Where("id = ?", id).Update("last_used_at", now)
	if result.Error != nil {
		return fmt.Errorf("failed to update API key last used: %w", result.Error)
	}
	return nil
}

// RevokeAPIKey revokes an API key
func (s *AuthDataStore) RevokeAPIKey(ctx *appctx.AppContext, id string, revokedBy string, reason string) error {
	now := time.Now()
	result := s.GetDB().Model(&entities.APIKey{}).Where("id = ?", id).Updates(map[string]interface{}{
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
func (s *AuthDataStore) DeleteAPIKey(ctx *appctx.AppContext, id string) error {
	result := s.GetDB().Delete(&entities.APIKey{}, "id = ?", id)
	if result.Error != nil {
		return fmt.Errorf("failed to delete API key: %w", result.Error)
	}
	return nil
}

// CreateAPIKeyUsage creates a usage record for an API key
func (s *AuthDataStore) CreateAPIKeyUsage(ctx *appctx.AppContext, usage *entities.APIKeyUsage) error {
	usage.ID = uuid.New().String()
	usage.CreatedAt = time.Now()
	usage.UpdatedAt = time.Now()

	result := s.GetDB().Create(usage)
	if result.Error != nil {
		return fmt.Errorf("failed to create API key usage: %w", result.Error)
	}
	return nil
}

// GetAPIKeyUsageStats retrieves usage statistics for an API key
func (s *AuthDataStore) GetAPIKeyUsageStats(ctx *appctx.AppContext, apiKeyID string, days int) ([]entities.APIKeyUsage, error) {
	var usage []entities.APIKeyUsage
	startDate := time.Now().AddDate(0, 0, -days)

	result := s.GetDB().Where("api_key_id = ? AND created_at >= ?", apiKeyID, startDate).
		Order("created_at DESC").Find(&usage)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get API key usage stats: %w", result.Error)
	}
	return usage, nil
}

// CleanupExpiredAPIKeys removes expired API keys
func (s *AuthDataStore) CleanupExpiredAPIKeys(ctx *appctx.AppContext) error {
	now := time.Now()
	result := s.GetDB().Where("expires_at IS NOT NULL AND expires_at < ?", now).Delete(&entities.APIKey{})
	if result.Error != nil {
		return fmt.Errorf("failed to cleanup expired API keys: %w", result.Error)
	}
	return nil
}
