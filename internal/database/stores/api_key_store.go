package stores

import (
	"crypto/sha256"
	"encoding/hex"
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
	pkg_utils "github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	authDataStoreInstance *ApiKeyDataStore
	authDataStoreOnce     sync.Once
)

type ApiKeyStoreInterface interface {
	CreateApiKey(ctx *appctx.AppContext, tenantID string, apiKey *entities.ApiKey) (*entities.ApiKey, error)
	GetApiKeyByHash(ctx *appctx.AppContext, tenantID string, keyHash string) (*entities.ApiKey, error)
	GetApiKeyByDigest(ctx *appctx.AppContext, tenantID string, digest string) (*entities.ApiKey, error)
	GetApiKeyByPrefix(ctx *appctx.AppContext, tenantID string, keyPrefix string) (*entities.ApiKey, error)
	GetApiKeyByName(ctx *appctx.AppContext, tenantID string, name string) (*entities.ApiKey, error)
	GetFilteredApiKeys(ctx *appctx.AppContext, tenantID string, filter *filters.Filter) (*filters.FilterResponse[entities.ApiKey], error)
	GetPaginatedApiKeys(ctx *appctx.AppContext, tenantID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.ApiKey], error)
	GetApiKeyByIDOrSlug(ctx *appctx.AppContext, tenantID string, id string) (*entities.ApiKey, error)
	RevokeApiKey(ctx *appctx.AppContext, tenantID string, id string, revokedBy string, reason string) error
	DeleteApiKey(ctx *appctx.AppContext, tenantID string, id string) error
	AddClaimToApiKey(ctx *appctx.AppContext, tenantID string, id string, claimID string) error
	RemoveClaimFromApiKey(ctx *appctx.AppContext, tenantID string, id string, claimID string) error
	UpdateApiKeyLastUsed(ctx *appctx.AppContext, tenantID string, apiKeyID string) error
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

// CreateApiKey creates a new API key for a user
func (s *ApiKeyDataStore) CreateApiKey(ctx *appctx.AppContext, tenantID string, apiKey *entities.ApiKey) (*entities.ApiKey, error) {
	if apiKey == nil {
		return nil, fmt.Errorf("api key is nil")
	}

	if apiKey.KeyHash == "" {
		return nil, fmt.Errorf("api key hash is empty")
	}

	if apiKey.KeyPrefix == "" {
		return nil, fmt.Errorf("api key prefix is empty")
	}

	if apiKey.Name == "" {
		return nil, fmt.Errorf("api key name is empty")
	}

	if apiKey.Claims == nil {
		return nil, fmt.Errorf("api key claims are empty")
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, fmt.Errorf("api key expires at is in the past")
	}

	if apiKey.ID == "" {
		apiKey.ID = uuid.New().String()
	}
	if apiKey.Slug == "" {
		apiKey.Slug = pkg_utils.Slugify(apiKey.Name)
	}

	apiKey.CreatedAt = time.Now()
	apiKey.UpdatedAt = time.Now()
	apiKey.BaseModelWithTenant = *utils.GetTenantBaseModelFromContext(ctx, &apiKey.BaseModelWithTenant)
	if apiKey.TenantID == "" {
		apiKey.TenantID = tenantID
	}

	// Hash the API key before storing
	encryptionService := encryption.GetInstance()
	keyHash, err := encryptionService.HashPassword(apiKey.KeyHash)
	if err != nil {
		return nil, fmt.Errorf("failed to hash API key: %w", err)
	}
	apiKey.KeyHash = keyHash

	// Create deterministic digest (SHA-256 of the full key)
	sha := sha256.Sum256([]byte(apiKey.KeyDigest))
	apiKey.KeyDigest = hex.EncodeToString(sha[:])

	result := s.GetDB().Create(apiKey)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create API key: %w", result.Error)
	}

	return apiKey, nil
}

// GetApiKeyByDigest retrieves an API key by its deterministic digest
func (s *ApiKeyDataStore) GetApiKeyByDigest(ctx *appctx.AppContext, tenantID string, digest string) (*entities.ApiKey, error) {
	var apiKey entities.ApiKey
	result := s.GetDB().Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "key_digest = ? AND tenant_id = ?", digest, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by digest: %w", result.Error)
	}
	return &apiKey, nil
}

// GetApiKeyByHash retrieves an API key by its hash
func (s *ApiKeyDataStore) GetApiKeyByHash(ctx *appctx.AppContext, tenantID string, keyHash string) (*entities.ApiKey, error) {
	var apiKey entities.ApiKey
	result := s.GetDB().Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "key_hash = ? AND tenant_id = ?", keyHash, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by hash: %w", result.Error)
	}
	return &apiKey, nil
}

// GetApiKeyByPrefix retrieves an API key by its prefix (for validation)
func (s *ApiKeyDataStore) GetApiKeyByPrefix(ctx *appctx.AppContext, tenantID string, keyPrefix string) (*entities.ApiKey, error) {
	var apiKey entities.ApiKey
	result := s.GetDB().Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "key_prefix = ? AND tenant_id = ?", keyPrefix, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by prefix: %w", result.Error)
	}
	return &apiKey, nil
}

// GetApiKeyByIDOrSlug retrieves an API key by ID
func (s *ApiKeyDataStore) GetApiKeyByIDOrSlug(ctx *appctx.AppContext, tenantID string, id string) (*entities.ApiKey, error) {
	var apiKey entities.ApiKey
	result := s.GetDB().Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "id = ? or slug = ? AND tenant_id = ?", id, id, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by ID: %w", result.Error)
	}
	return &apiKey, nil
}

// ListAPIKeysByUserIDWithFilter retrieves API keys for a user with pagination and filtering
func (s *ApiKeyDataStore) ListAPIKeysByUserIDWithFilter(ctx *appctx.AppContext, tenantID string, userID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.ApiKey], error) {
	// Add the user_id filter to the existing filter
	filterObj.WithField("user_id", filters.FilterOperatorEqual, userID, filters.FilterJoinerAnd)

	// Use the generic pagination helper
	return utils.PaginatedFilteredQueryWithPreload(s.GetDB(), tenantID, filterObj, entities.ApiKey{}, "Claims")
}

// RevokeApiKey revokes an API key
func (s *ApiKeyDataStore) RevokeApiKey(ctx *appctx.AppContext, tenantID string, id string, revokedBy string, reason string) error {
	now := time.Now()
	result := s.GetDB().Model(&entities.ApiKey{}).Where("id = ? AND tenant_id = ?", id, tenantID).Updates(map[string]interface{}{
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

// DeleteApiKey permanently deletes an API key
func (s *ApiKeyDataStore) DeleteApiKey(ctx *appctx.AppContext, tenantID string, id string) error {
	// deleting the api key relationships first
	result := s.GetDB().Where("api_key_id = ?", id).Delete(&entities.ApiKeyClaims{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete API key claims: %w", result.Error)
	}

	result = s.GetDB().Where("id = ?", id).Delete(&entities.ApiKey{})
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

// GetFilteredApiKeys retrieves API keys based on a filter
func (s *ApiKeyDataStore) GetFilteredApiKeys(ctx *appctx.AppContext, tenantID string, filter *filters.Filter) (*filters.FilterResponse[entities.ApiKey], error) {
	return utils.PaginatedFilteredQueryWithPreload(s.GetDB(), tenantID, filter, entities.ApiKey{}, "Claims")
}

// GetPaginatedApiKeys retrieves paginated API keys based on a pagination and filter
func (s *ApiKeyDataStore) GetPaginatedApiKeys(ctx *appctx.AppContext, tenantID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.ApiKey], error) {
	return utils.PaginatedQuery(s.GetDB(), tenantID, pagination, entities.ApiKey{}, "Claims")
}

func (s *ApiKeyDataStore) AddClaimToApiKey(ctx *appctx.AppContext, tenantID string, id string, claimID string) error {
	var dbClaim entities.Claim
	result := s.GetDB().First(&dbClaim, "id = ? AND tenant_id = ?", claimID, tenantID)
	if result.Error != nil {
		return fmt.Errorf("failed to get claim: %w", result.Error)
	}
	if dbClaim.ID == "" {
		return fmt.Errorf("claim not found")
	}

	existingApiKey, err := s.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}
	if existingApiKey == nil {
		return fmt.Errorf("API key not found")
	}
	// checking if the claim is already assigned to the api key
	var apiKeyClaims entities.ApiKeyClaims
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ?", id, claimID).First(&apiKeyClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get user claim: %w", result.Error)
		}
	}
	if apiKeyClaims.ClaimID != "" {
		return fmt.Errorf("claim already assigned to API key")
	}

	// creating the api key claim
	apiKeyClaims.ApiKeyID = id
	apiKeyClaims.ClaimID = dbClaim.ID
	result = s.GetDB().Create(&apiKeyClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to create API key claim: %w", result.Error)
	}
	return nil
}

func (s *ApiKeyDataStore) RemoveClaimFromApiKey(ctx *appctx.AppContext, tenantID string, id string, claimID string) error {
	var dbClaim entities.Claim
	result := s.GetDB().First(&dbClaim, "id = ? AND tenant_id = ?", claimID, tenantID)
	if result.Error != nil {
		return fmt.Errorf("failed to get claim: %w", result.Error)
	}
	if dbClaim.ID == "" {
		return fmt.Errorf("claim not found")
	}

	existingApiKey, err := s.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to get API key: %w", err)
	}
	if existingApiKey == nil {
		return fmt.Errorf("API key not found")
	}

	// checking if the claim is assigned to the api key
	var apiKeyClaims entities.ApiKeyClaims
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ? AND tenant_id = ?", id, claimID, tenantID).First(&apiKeyClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to get API key claim: %w", result.Error)
	}
	if apiKeyClaims.ClaimID == "" {
		return fmt.Errorf("claim not assigned to API key")
	}

	// deleting the api key claim
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ? AND tenant_id = ?", id, claimID, tenantID).Delete(&apiKeyClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to delete API key claim: %w", result.Error)
	}

	return nil
}

func (s *ApiKeyDataStore) GetApiKeyByName(ctx *appctx.AppContext, tenantID string, name string) (*entities.ApiKey, error) {
	var apiKey entities.ApiKey
	result := s.GetDB().Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "name = ? AND tenant_id = ?", name, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get API key by name: %w", result.Error)
	}

	return &apiKey, nil
}

func (s *ApiKeyDataStore) UpdateApiKeyLastUsed(ctx *appctx.AppContext, tenantID string, apiKeyID string) error {
	now := time.Now()
	result := s.GetDB().
		Model(&entities.ApiKey{}).
		Where("id = ? AND tenant_id = ?", apiKeyID, tenantID).
		Update("last_used_at", now)
	if result.Error != nil {
		return fmt.Errorf("failed to update API key last used: %w", result.Error)
	}
	return nil
}
