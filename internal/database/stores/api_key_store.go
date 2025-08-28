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
	CreateApiKey(ctx *appctx.AppContext, tenantID string, apiKey *entities.ApiKey) (*entities.ApiKey, *diagnostics.Diagnostics)
	GetApiKeyByHash(ctx *appctx.AppContext, tenantID string, keyHash string) (*entities.ApiKey, *diagnostics.Diagnostics)
	GetApiKeyByDigest(ctx *appctx.AppContext, tenantID string, digest string) (*entities.ApiKey, *diagnostics.Diagnostics)
	GetApiKeyByPrefix(ctx *appctx.AppContext, tenantID string, keyPrefix string) (*entities.ApiKey, *diagnostics.Diagnostics)
	GetApiKeyByName(ctx *appctx.AppContext, tenantID string, name string) (*entities.ApiKey, *diagnostics.Diagnostics)
	GetApiKeys(ctx *appctx.AppContext, tenantID string) ([]entities.ApiKey, *diagnostics.Diagnostics)
	GetApiKeysByQuery(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.ApiKey], *diagnostics.Diagnostics)
	GetApiKeyByIDOrSlug(ctx *appctx.AppContext, tenantID string, id string) (*entities.ApiKey, *diagnostics.Diagnostics)
	RevokeApiKey(ctx *appctx.AppContext, tenantID string, id string, revokedBy string, reason string) *diagnostics.Diagnostics
	DeleteApiKey(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics
	AddClaimToApiKey(ctx *appctx.AppContext, tenantID string, id string, claimID string) *diagnostics.Diagnostics
	RemoveClaimFromApiKey(ctx *appctx.AppContext, tenantID string, id string, claimID string) *diagnostics.Diagnostics
	UpdateApiKeyLastUsed(ctx *appctx.AppContext, tenantID string, apiKeyID string) *diagnostics.Diagnostics
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
			if migrateDiag := store.Migrate(); migrateDiag.HasErrors() {
				diag.Append(migrateDiag)
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
func (s *ApiKeyDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_api_key_data_store")
	defer diag.Complete()

	if err := s.GetDB().AutoMigrate(&entities.ApiKey{}); err != nil {
		diag.AddError("failed_to_migrate_api_keys_table", fmt.Sprintf("failed to migrate api_keys table: %v", err), "api_key_data_store")
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.ApiKeyClaims{}); err != nil {
		diag.AddError("failed_to_migrate_api_key_claims_table", fmt.Sprintf("failed to migrate api_key_claims table: %v", err), "api_key_data_store")
		return diag
	}

	// add unique index to the api_key_claims table
	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_api_key_claims ON api_key_claims (api_key_id, claim_id);").Error; err != nil {
		diag.AddError("failed_to_create_unique_index_on_api_key_claims_table", fmt.Sprintf("failed to create unique index on api_key_claims table: %v", err), "api_key_data_store")
		return diag
	}

	return diag
}

// CreateApiKey creates a new API key for a user
func (s *ApiKeyDataStore) CreateApiKey(ctx *appctx.AppContext, tenantID string, apiKey *entities.ApiKey) (*entities.ApiKey, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_api_key")
	defer diag.Complete()

	if apiKey == nil {
		diag.AddError("api_key_is_nil", "api key is nil", "api_key_data_store")
		return nil, diag
	}

	if apiKey.KeyHash == "" {
		diag.AddError("api_key_hash_is_empty", "api key hash is empty", "api_key_data_store")
		return nil, diag
	}

	if apiKey.KeyPrefix == "" {
		diag.AddError("api_key_prefix_is_empty", "api key prefix is empty", "api_key_data_store")
		return nil, diag
	}

	if apiKey.Name == "" {
		diag.AddError("api_key_name_is_empty", "api key name is empty", "api_key_data_store")
		return nil, diag
	}

	if apiKey.Claims == nil {
		diag.AddError("api_key_claims_are_empty", "api key claims are empty", "api_key_data_store")
		return nil, diag
	}

	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		diag.AddError("api_key_expires_at_is_in_the_past", "api key expires at is in the past", "api_key_data_store")
		return nil, diag
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
		diag.AddError("failed_to_hash_api_key", fmt.Sprintf("failed to hash API key: %s", err.Error()), "api_key_data_store")
		return nil, diag
	}
	apiKey.KeyHash = keyHash

	// Create deterministic digest (SHA-256 of the full key)
	sha := sha256.Sum256([]byte(apiKey.KeyDigest))
	apiKey.KeyDigest = hex.EncodeToString(sha[:])

	result := s.GetDB().Create(apiKey)
	if result.Error != nil {
		diag.AddError("failed_to_create_api_key", fmt.Sprintf("failed to create API key: %s", result.Error.Error()), "api_key_data_store")
		return nil, diag
	}

	return apiKey, diag
}

// GetApiKeyByDigest retrieves an API key by its deterministic digest
func (s *ApiKeyDataStore) GetApiKeyByDigest(ctx *appctx.AppContext, tenantID string, digest string) (*entities.ApiKey, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_api_key_by_digest")
	defer diag.Complete()

	var apiKey entities.ApiKey
	result := s.GetDB().Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "key_digest = ? AND tenant_id = ?", digest, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_api_key_by_digest", fmt.Sprintf("failed to get API key by digest: %s", result.Error.Error()), "api_key_data_store")
		return nil, diag
	}
	return &apiKey, diag
}

// GetApiKeyByHash retrieves an API key by its hash
func (s *ApiKeyDataStore) GetApiKeyByHash(ctx *appctx.AppContext, tenantID string, keyHash string) (*entities.ApiKey, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_api_key_by_hash")
	defer diag.Complete()

	var apiKey entities.ApiKey
	result := s.GetDB().Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "key_hash = ? AND tenant_id = ?", keyHash, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_api_key_by_hash", fmt.Sprintf("failed to get API key by hash: %s", result.Error.Error()), "api_key_data_store")
		return nil, diag
	}
	return &apiKey, diag
}

// GetApiKeyByPrefix retrieves an API key by its prefix (for validation)
func (s *ApiKeyDataStore) GetApiKeyByPrefix(ctx *appctx.AppContext, tenantID string, keyPrefix string) (*entities.ApiKey, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_api_key_by_prefix")
	defer diag.Complete()

	var apiKey entities.ApiKey
	result := s.GetDB().Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "key_prefix = ? AND tenant_id = ?", keyPrefix, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_api_key_by_prefix", fmt.Sprintf("failed to get API key by prefix: %s", result.Error.Error()), "api_key_data_store")
		return nil, diag
	}
	return &apiKey, diag
}

// GetApiKeyByIDOrSlug retrieves an API key by ID
func (s *ApiKeyDataStore) GetApiKeyByIDOrSlug(ctx *appctx.AppContext, tenantID string, id string) (*entities.ApiKey, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_api_key_by_id_or_slug")
	defer diag.Complete()

	var apiKey entities.ApiKey
	result := s.GetDB().Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "id = ? or slug = ? AND tenant_id = ?", id, id, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_api_key_by_id_or_slug", fmt.Sprintf("failed to get API key by ID: %s", result.Error.Error()), "api_key_data_store")
		return nil, diag
	}
	return &apiKey, diag
}

// RevokeApiKey revokes an API key
func (s *ApiKeyDataStore) RevokeApiKey(ctx *appctx.AppContext, tenantID string, id string, revokedBy string, reason string) *diagnostics.Diagnostics {
	diag := diagnostics.New("revoke_api_key")
	defer diag.Complete()

	now := time.Now()
	result := s.GetDB().Model(&entities.ApiKey{}).Where("id = ? AND tenant_id = ?", id, tenantID).Updates(map[string]interface{}{
		"is_active":         false,
		"revoked_at":        now,
		"revoked_by":        revokedBy,
		"revocation_reason": reason,
		"updated_at":        now,
	})
	if result.Error != nil {
		diag.AddError("failed_to_revoke_api_key", fmt.Sprintf("failed to revoke API key: %s", result.Error.Error()), "api_key_data_store")
		return diag
	}
	return diag
}

// DeleteApiKey permanently deletes an API key
func (s *ApiKeyDataStore) DeleteApiKey(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_api_key")
	defer diag.Complete()

	// deleting the api key relationships first
	result := s.GetDB().Where("api_key_id = ?", id).Delete(&entities.ApiKeyClaims{})
	if result.Error != nil {
		diag.AddError("failed_to_delete_api_key_claims", fmt.Sprintf("failed to delete API key claims: %s", result.Error.Error()), "api_key_data_store")
		return diag
	}

	result = s.GetDB().Where("id = ?", id).Delete(&entities.ApiKey{})
	if result.Error != nil {
		diag.AddError("failed_to_delete_api_key", fmt.Sprintf("failed to delete API key: %s", result.Error.Error()), "api_key_data_store")
		return diag
	}
	return diag
}

// CleanupExpiredAPIKeys removes expired API keys
func (s *ApiKeyDataStore) CleanupExpiredAPIKeys(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("cleanup_expired_api_keys")
	defer diag.Complete()

	now := time.Now()
	result := s.GetDB().Where("expires_at IS NOT NULL AND expires_at < ?", now).Delete(&entities.ApiKey{})
	if result.Error != nil {
		diag.AddError("failed_to_cleanup_expired_api_keys", fmt.Sprintf("failed to cleanup expired API keys: %s", result.Error.Error()), "api_key_data_store")
		return diag
	}
	return diag
}

func (s *ApiKeyDataStore) GetApiKeys(ctx *appctx.AppContext, tenantID string) ([]entities.ApiKey, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_api_keys")
	defer diag.Complete()

	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "api_key_data_store")
		return nil, diag
	}

	var apiKeys []entities.ApiKey
	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		})

	if tenantID != "" {
		db = db.Where("tenant_id = ?", tenantID)
	}

	result := db.Find(&apiKeys)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_keys", fmt.Sprintf("failed to get API keys: %s", result.Error.Error()), "api_key_data_store")
		return nil, diag
	}

	return apiKeys, diag
}

// GetApiKeysByQuery retrieves paginated API keys based on a pagination and filter
func (s *ApiKeyDataStore) GetApiKeysByQuery(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.ApiKey], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_paginated_api_keys")
	defer diag.Complete()

	db := s.GetDB()
	db = db.Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	})

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := utils.QueryDatabase[entities.ApiKey](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_api_keys", fmt.Sprintf("failed to get API keys: %s", err.Error()), "api_key_data_store")
		return nil, diag
	}

	return result, diag
}

func (s *ApiKeyDataStore) AddClaimToApiKey(ctx *appctx.AppContext, tenantID string, id string, claimID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_claim_to_api_key")
	defer diag.Complete()

	var dbClaim entities.Claim
	result := s.GetDB().First(&dbClaim, "id = ? AND tenant_id = ?", claimID, tenantID)
	if result.Error != nil {
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", result.Error.Error()), "api_key_data_store")
		return diag
	}
	if dbClaim.ID == "" {
		diag.AddError("claim_not_found", "claim not found", "api_key_data_store")
		return diag
	}

	existingApiKey, getApiKeyDiag := s.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return diag
	}
	if existingApiKey == nil {
		diag.AddError("api_key_not_found", "API key not found", "api_key_data_store")
		return diag
	}
	// checking if the claim is already assigned to the api key
	var apiKeyClaims entities.ApiKeyClaims
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ?", id, claimID).First(&apiKeyClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user_claim", fmt.Sprintf("failed to get user claim: %s", result.Error.Error()), "api_key_data_store")
			return diag
		}
	}
	if apiKeyClaims.ClaimID != "" {
		diag.AddError("claim_already_assigned_to_api_key", "claim already assigned to API key", "api_key_data_store")
		return diag
	}

	// creating the api key claim
	apiKeyClaims.ApiKeyID = id
	apiKeyClaims.ClaimID = dbClaim.ID
	result = s.GetDB().Create(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_create_api_key_claim", fmt.Sprintf("failed to create API key claim: %s", result.Error.Error()), "api_key_data_store")
		return diag
	}
	return diag
}

func (s *ApiKeyDataStore) RemoveClaimFromApiKey(ctx *appctx.AppContext, tenantID string, id string, claimID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_claim_from_api_key")
	defer diag.Complete()

	var dbClaim entities.Claim
	result := s.GetDB().First(&dbClaim, "id = ? AND tenant_id = ?", claimID, tenantID)
	if result.Error != nil {
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", result.Error.Error()), "api_key_data_store")
		return diag
	}
	if dbClaim.ID == "" {
		diag.AddError("claim_not_found", "claim not found", "api_key_data_store")
		return diag
	}

	existingApiKey, getApiKeyDiag := s.GetApiKeyByIDOrSlug(ctx, tenantID, id)
	if getApiKeyDiag.HasErrors() {
		diag.Append(getApiKeyDiag)
		return diag
	}
	if existingApiKey == nil {
		diag.AddError("api_key_not_found", "API key not found", "api_key_data_store")
		return diag
	}

	// checking if the claim is assigned to the api key
	var apiKeyClaims entities.ApiKeyClaims
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ? AND tenant_id = ?", id, claimID, tenantID).First(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_key_claim", fmt.Sprintf("failed to get API key claim: %s", result.Error.Error()), "api_key_data_store")
		return diag
	}
	if apiKeyClaims.ClaimID == "" {
		diag.AddError("claim_not_assigned_to_api_key", "claim not assigned to API key", "api_key_data_store")
		return diag
	}

	// deleting the api key claim
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ? AND tenant_id = ?", id, claimID, tenantID).Delete(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_delete_api_key_claim", fmt.Sprintf("failed to delete API key claim: %s", result.Error.Error()), "api_key_data_store")
		return diag
	}

	return diag
}

func (s *ApiKeyDataStore) GetApiKeyByName(ctx *appctx.AppContext, tenantID string, name string) (*entities.ApiKey, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_api_key_by_name")
	defer diag.Complete()

	var apiKey entities.ApiKey
	result := s.GetDB().Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).First(&apiKey, "name = ? AND tenant_id = ?", name, tenantID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_api_key_by_name", fmt.Sprintf("failed to get API key by name: %s", result.Error.Error()), "api_key_data_store")
		return nil, diag
	}

	return &apiKey, diag
}

func (s *ApiKeyDataStore) UpdateApiKeyLastUsed(ctx *appctx.AppContext, tenantID string, apiKeyID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("update_api_key_last_used")
	defer diag.Complete()

	now := time.Now()
	result := s.GetDB().
		Model(&entities.ApiKey{}).
		Where("id = ? AND tenant_id = ?", apiKeyID, tenantID).
		Update("last_used_at", now)
	if result.Error != nil {
		diag.AddError("failed_to_update_api_key_last_used", fmt.Sprintf("failed to update API key last used: %s", result.Error.Error()), "api_key_data_store")
		return diag
	}
	return diag
}
