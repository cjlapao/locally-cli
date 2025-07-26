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
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	authDataStoreInstance *AuthDataStore
	authDataStoreOnce     sync.Once
)

// AuthDataStore handles auth-specific database operations
type AuthDataStore struct {
	database.BaseDataStore
}

// GetAuthDataStoreInstance returns the singleton instance of the auth store
func GetAuthDataStoreInstance() *AuthDataStore {
	return authDataStoreInstance
}

// InitializeAuthDataStore initializes the auth store singleton
func InitializeAuthDataStore() error {
	var initErr error
	cfg := config.GetInstance().Get()
	authDataStoreOnce.Do(func() {
		// Get the database service instance
		dbService := database.GetInstance()
		if dbService == nil {
			initErr = fmt.Errorf("database service not initialized")
			return
		}

		store := &AuthDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running auth migrations")
			if err := store.Migrate(); err != nil {
				initErr = fmt.Errorf("failed to run auth migrations: %w", err)
				return
			}
			logging.Info("Auth migrations completed")
		}

		authDataStoreInstance = store
	})

	return initErr
}

// Migrate implements the DataStore interface
func (s *AuthDataStore) Migrate() error {
	if err := s.GetDB().AutoMigrate(&entities.Role{}); err != nil {
		return fmt.Errorf("failed to migrate role table: %w", err)
	}

	if err := s.GetDB().AutoMigrate(&entities.Claim{}); err != nil {
		return fmt.Errorf("failed to migrate claim table: %w", err)
	}

	if err := s.GetDB().AutoMigrate(&entities.User{}); err != nil {
		return fmt.Errorf("failed to migrate user table: %w", err)
	}

	if err := s.GetDB().AutoMigrate(&entities.APIKey{}); err != nil {
		return fmt.Errorf("failed to migrate api_keys table: %w", err)
	}

	if err := s.GetDB().AutoMigrate(&entities.APIKeyUsage{}); err != nil {
		return fmt.Errorf("failed to migrate api_key_usage table: %w", err)
	}

	return nil
}

// CreateUser creates a new user
func (s *AuthDataStore) CreateUser(ctx *appctx.AppContext, user *entities.User) (*entities.User, error) {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	encryptionService := encryption.GetInstance()
	encryptedPassword, err := encryptionService.HashPassword(user.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt password: %w", err)
	}
	user.Password = encryptedPassword

	result := s.GetDB().Create(user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create user: %w", result.Error)
	}

	return user, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthDataStore) GetUserByID(ctx *appctx.AppContext, id string) (*entities.User, error) {
	var user entities.User
	result := s.GetDB().Preload("Roles").Preload("Claims").First(&user, "id = ?", id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by id: %w", result.Error)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (s *AuthDataStore) GetUserByUsername(ctx *appctx.AppContext, username string) (*entities.User, error) {
	var user entities.User
	result := s.GetDB().Preload("Roles").Preload("Claims").First(&user, "username = ?", username)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", result.Error)
	}
	return &user, nil
}

// UpdateUser updates an existing user
func (s *AuthDataStore) UpdateUser(ctx *appctx.AppContext, user *entities.User) error {
	user.UpdatedAt = time.Now()
	currentUser, err := s.GetUserByID(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	if user.Password != currentUser.Password {
		encryptionService := encryption.GetInstance()
		encryptedPassword, err := encryptionService.HashPassword(user.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		user.Password = encryptedPassword
	}
	if user.Name != currentUser.Name {
		user.Name = currentUser.Name
	}
	if user.Email != currentUser.Email {
		user.Email = currentUser.Email
	}
	if len(user.Roles) != len(currentUser.Roles) {
		user.Roles = currentUser.Roles
	}
	if user.Status != currentUser.Status {
		user.Status = currentUser.Status
	}
	if user.Blocked != currentUser.Blocked {
		user.Blocked = currentUser.Blocked
	}
	user.UpdatedAt = time.Now()
	return s.GetDB().Save(currentUser).Error
}

func (s *AuthDataStore) UpdateUserPassword(ctx *appctx.AppContext, id string, password string) error {
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	encryptionService := encryption.GetInstance()
	encryptedPassword, err := encryptionService.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}
	user.Password = encryptedPassword
	user.UpdatedAt = time.Now()
	return s.GetDB().Save(user).Error
}

func (s *AuthDataStore) BlockUser(ctx *appctx.AppContext, id string) error {
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	user.Blocked = true
	user.UpdatedAt = time.Now()
	return s.GetDB().Save(user).Error
}

func (s *AuthDataStore) SetRefreshToken(ctx *appctx.AppContext, id string, refreshToken string) error {
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	user.RefreshToken = refreshToken
	user.RefreshTokenExpiresAt = time.Now().Add(24 * time.Hour)
	user.UpdatedAt = time.Now()
	return s.GetDB().Save(user).Error
}

// DeleteUser deletes a user
func (s *AuthDataStore) DeleteUser(ctx *appctx.AppContext, id string) error {
	user, err := s.GetUserByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	return s.GetDB().Delete(user).Error
}

// API Key Management Methods

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
	return utils.PaginatedQuery(s.GetDB(), filterObj, entities.APIKey{})
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
