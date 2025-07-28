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
	pkg_utils "github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	userDataStoreInstance *UserDataStore
	userDataStoreOnce     sync.Once
)

type UserDataStoreInterface interface {
	GetUserByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.User, error)
	GetUserByUsername(ctx *appctx.AppContext, tenantID string, username string) (*entities.User, error)
	GetUsersByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.User], error)
	CreateUser(ctx *appctx.AppContext, tenantID string, user *entities.User) (*entities.User, error)
	UpdateUser(ctx *appctx.AppContext, tenantID string, user *entities.User) error
	UpdateUserPassword(ctx *appctx.AppContext, tenantID string, id string, password string) error
	BlockUser(ctx *appctx.AppContext, tenantID string, id string) error
	SetRefreshToken(ctx *appctx.AppContext, tenantID string, id string, refreshToken string) error
	DeleteUser(ctx *appctx.AppContext, tenantID string, id string) error
	GetRolesByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Role], error)
	GetClaimsByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Claim], error)
}

type UserDataStore struct {
	database.BaseDataStore
}

func GetUserDataStoreInstance() UserDataStoreInterface {
	return userDataStoreInstance
}

func InitializeUserDataStore() *diagnostics.Diagnostics {
	diag := diagnostics.New("initialize_user_data_store")
	cfg := config.GetInstance().Get()
	logging.Info("Initializing user store...")

	userDataStoreOnce.Do(func() {
		dbService := database.GetInstance()
		if dbService == nil {
			diag.AddError("database_service_not_initialized", "database service not initialized", "user_data_store", nil)
			return
		}

		store := &UserDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running user migrations")
			if migrateDiag := store.Migrate(); migrateDiag.HasErrors() {
				diag.Append(migrateDiag)
				return
			}
			logging.Info("User migrations completed")
		}

		userDataStoreInstance = store
	})

	logging.Info("User store initialized successfully")
	return diag
}

func (s *UserDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_user_data_store")
	if err := s.GetDB().AutoMigrate(&entities.UserRoles{}); err != nil {
		diag.AddError("failed_to_migrate_user_role_table", "failed to migrate user role table", "user_data_store", nil)
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.UserClaims{}); err != nil {
		diag.AddError("failed_to_migrate_user_claim_table", "failed to migrate user claim table", "user_data_store", nil)
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.Role{}); err != nil {
		diag.AddError("failed_to_migrate_role_table", "failed to migrate role table", "user_data_store", nil)
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.Claim{}); err != nil {
		diag.AddError("failed_to_migrate_claim_table", "failed to migrate claim table", "user_data_store", nil)
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.User{}); err != nil {
		diag.AddError("failed_to_migrate_user_table", "failed to migrate user table", "user_data_store", nil)
		return diag
	}

	return diag
}

func (s *UserDataStore) GetUsersByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.User], error) {
	return utils.PaginatedQueryWithPreload(s.GetDB(), tenantID, filterObj, entities.User{}, "Roles", "Claims")
}

// CreateUser creates a new user
func (s *UserDataStore) CreateUser(ctx *appctx.AppContext, tenantID string, user *entities.User) (*entities.User, error) {
	user.ID = uuid.New().String()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.TenantID = tenantID
	if user.Username != "" {
		user.Slug = pkg_utils.Slugify(user.Username)
	}

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
func (s *UserDataStore) GetUserByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.User, error) {
	var user entities.User
	result := s.GetDB().Preload("Roles").Preload("Claims").First(&user, "tenant_id = ? AND (id = ? OR slug = ?)", tenantID, id, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by id: %w", result.Error)
	}
	return &user, nil
}

// GetUserByUsername retrieves a user by username
func (s *UserDataStore) GetUserByUsername(ctx *appctx.AppContext, tenantID string, username string) (*entities.User, error) {
	var user entities.User
	result := s.GetDB().Preload("Roles").Preload("Claims").First(&user, "tenant_id = ? AND username = ?", tenantID, username)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get user by username: %w", result.Error)
	}
	return &user, nil
}

// UpdateUser updates an existing user
func (s *UserDataStore) UpdateUser(ctx *appctx.AppContext, tenantID string, user *entities.User) error {
	user.UpdatedAt = time.Now()
	user.TenantID = tenantID
	currentUser, err := s.GetUserByID(ctx, tenantID, user.ID)
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}
	if user.Password != "" {
		encryptionService := encryption.GetInstance()
		encryptedPassword, err := encryptionService.HashPassword(user.Password)
		if err != nil {
			return fmt.Errorf("failed to encrypt password: %w", err)
		}
		user.Password = encryptedPassword
	}
	if user.Username != "" {
		user.Slug = pkg_utils.Slugify(user.Username)
	}

	updates := utils.PartialUpdateMap(currentUser, user, "updated_at", "slug")
	if err := s.GetDB().Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

func (s *UserDataStore) UpdateUserPassword(ctx *appctx.AppContext, tenantID string, id string, password string) error {
	user, err := s.GetUserByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	encryptionService := encryption.GetInstance()
	encryptedPassword, err := encryptionService.HashPassword(password)
	if err != nil {
		return fmt.Errorf("failed to encrypt password: %w", err)
	}

	// Create a minimal user object with only the fields we want to update
	updatedUser := &entities.User{
		BaseModel: entities.BaseModel{
			ID:        user.ID,
			UpdatedAt: time.Now(),
		},
		Password: encryptedPassword,
	}

	// Use PartialUpdateMap to only update the password and updated_at fields
	updates := utils.PartialUpdateMap(user, updatedUser, "updated_at")
	if err := s.GetDB().Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

func (s *UserDataStore) BlockUser(ctx *appctx.AppContext, tenantID string, id string) error {
	user, err := s.GetUserByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Create a minimal user object with only the fields we want to update
	updatedUser := &entities.User{
		BaseModel: entities.BaseModel{
			ID:        user.ID,
			UpdatedAt: time.Now(),
		},
		Blocked: true,
	}

	// Use PartialUpdateMap to only update the blocked and updated_at fields
	updates := utils.PartialUpdateMap(user, updatedUser, "updated_at")
	if err := s.GetDB().Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

func (s *UserDataStore) SetRefreshToken(ctx *appctx.AppContext, tenantID string, id string, refreshToken string) error {
	user, err := s.GetUserByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}

	// Create a minimal user object with only the fields we want to update
	updatedUser := &entities.User{
		BaseModel: entities.BaseModel{
			ID:        user.ID,
			UpdatedAt: time.Now(),
		},
		RefreshToken:          refreshToken,
		RefreshTokenExpiresAt: time.Now().Add(24 * time.Hour),
	}

	// Use PartialUpdateMap to only update the refresh token fields and updated_at
	updates := utils.PartialUpdateMap(user, updatedUser, "updated_at")
	if err := s.GetDB().Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return err
	}
	return nil
}

// DeleteUser deletes a user
func (s *UserDataStore) DeleteUser(ctx *appctx.AppContext, tenantID string, id string) error {
	user, err := s.GetUserByID(ctx, tenantID, id)
	if err != nil {
		return fmt.Errorf("failed to get user: %w", err)
	}
	return s.GetDB().Delete(user).Error
}

func (s *UserDataStore) GetRolesByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Role], error) {
	return utils.PaginatedQueryWithPreload(s.GetDB(), tenantID, filterObj, entities.Role{}, "DefaultClaims")
}

func (s *UserDataStore) GetClaimsByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Claim], error) {
	return utils.PaginatedQueryWithPreload(s.GetDB(), tenantID, filterObj, entities.Claim{}, "Roles")
}
