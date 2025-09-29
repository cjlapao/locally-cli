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
	db_utils "github.com/cjlapao/locally-cli/internal/database/utils"
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
	GetUserByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.User, *diagnostics.Diagnostics)
	GetUserByUsername(ctx *appctx.AppContext, tenantID string, username string) (*entities.User, *diagnostics.Diagnostics)
	GetUsers(ctx *appctx.AppContext, tenantID string) ([]entities.User, *diagnostics.Diagnostics)
	GetUsersByQuery(ctx *appctx.AppContext, tenantID string, filterObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *diagnostics.Diagnostics)
	CreateUser(ctx *appctx.AppContext, tenantID string, user *entities.User) (*entities.User, *diagnostics.Diagnostics)
	UpdateUser(ctx *appctx.AppContext, tenantID string, user *entities.User) *diagnostics.Diagnostics
	UpdateUserPassword(ctx *appctx.AppContext, tenantID string, id string, password string) *diagnostics.Diagnostics
	BlockUser(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics
	SetRefreshToken(ctx *appctx.AppContext, tenantID string, id string, refreshToken string) *diagnostics.Diagnostics
	DeleteUser(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics
	GetUserClaims(ctx *appctx.AppContext, tenantID string, userID string) ([]entities.Claim, *diagnostics.Diagnostics)
	GetUserClaimsByQuery(ctx *appctx.AppContext, tenantID string, userID string, filterObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *diagnostics.Diagnostics)
	AddClaimToUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics
	RemoveClaimFromUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics
	GetUserRoles(ctx *appctx.AppContext, tenantID string, userID string) ([]entities.Role, *diagnostics.Diagnostics)
	GetUserRolesByQuery(ctx *appctx.AppContext, tenantID string, userID string, filterObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *diagnostics.Diagnostics)
	AddUserToRole(ctx *appctx.AppContext, tenantID string, userID string, roleId string) *diagnostics.Diagnostics
	RemoveUserFromRole(ctx *appctx.AppContext, tenantID string, userID string, roleId string) *diagnostics.Diagnostics
}

type UserDataStore struct {
	database.BaseDataStore
}

func GetUserDataStoreInstance() UserDataStoreInterface {
	return userDataStoreInstance
}

func InitializeUserDataStore() (UserDataStoreInterface, *diagnostics.Diagnostics) {
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
	return userDataStoreInstance, diag
}

func (s *UserDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_user_data_store")
	if err := s.GetDB().AutoMigrate(&entities.UserRoles{}); err != nil {
		diag.AddError("failed_to_migrate_user_role_table", fmt.Sprintf("failed to migrate user role table: %s", err.Error()), "user_data_store", nil)
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.UserClaims{}); err != nil {
		diag.AddError("failed_to_migrate_user_claim_table", fmt.Sprintf("failed to migrate user claim table: %s", err.Error()), "user_data_store", nil)
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.User{}); err != nil {
		diag.AddError("failed_to_migrate_user_table", fmt.Sprintf("failed to migrate user table: %s", err.Error()), "user_data_store", nil)
		return diag
	}

	// Add unique constraints to prevent duplicates
	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_roles_unique ON user_roles(user_id, role_id);").Error; err != nil {
		diag.AddError("failed_to_create_unique_index_on_user_roles", fmt.Sprintf("failed to create unique index on user roles: %s", err.Error()), "user_data_store", nil)
		return diag
	}

	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_user_claims_unique ON user_claims(user_id, claim_id);").Error; err != nil {
		diag.AddError("failed_to_create_unique_index_on_user_claims", fmt.Sprintf("failed to create unique index on user claims: %s", err.Error()), "user_data_store", nil)
		return diag
	}

	return diag
}

// CreateUser creates a new user
func (s *UserDataStore) CreateUser(ctx *appctx.AppContext, tenantID string, user *entities.User) (*entities.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_user")
	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()
	user.TenantID = tenantID
	if user.Username != "" {
		user.Slug = pkg_utils.Slugify(user.Username)
	}

	encryptionService := encryption.GetInstance()
	encryptedPassword, err := encryptionService.HashPassword(user.Password)
	if err != nil {
		diag.AddError("failed_to_encrypt_password", fmt.Sprintf("failed to encrypt password: %s", err.Error()), "user_data_store", nil)
		return nil, diag
	}
	user.Password = encryptedPassword

	// Store the roles and claims to associate after creating the user
	rolesToAssociate := user.Roles
	claimsToAssociate := user.Claims
	user.Roles = nil  // Clear roles to avoid GORM trying to create them
	user.Claims = nil // Clear claims to avoid GORM trying to create them

	result := s.GetDB().Create(user)
	if result.Error != nil {
		diag.AddError("failed_to_create_user", fmt.Sprintf("failed to create user: %s", result.Error.Error()), "user_data_store", nil)
		return nil, diag
	}

	// Associate roles if any were provided
	if len(rolesToAssociate) > 0 {
		// Get the full role entities from the database
		var dbRoles []entities.Role
		for _, role := range rolesToAssociate {
			var dbRole entities.Role
			if result := s.GetDB().Where("id = ?", role.ID).First(&dbRole); result.Error != nil {
				diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role with id %s: %s", role.ID, result.Error.Error()), "user_data_store", nil)
				return nil, diag
			}
			dbRoles = append(dbRoles, dbRole)
		}

		// First, clear any existing role associations
		if err := s.GetDB().Model(user).Association("Roles").Clear(); err != nil {
			diag.AddError("failed_to_clear_existing_role_associations", fmt.Sprintf("failed to clear existing role associations: %s", err.Error()), "user_data_store", nil)
			return nil, diag
		}

		// Then add the new role associations
		if err := s.GetDB().Model(user).Association("Roles").Append(dbRoles); err != nil {
			diag.AddError("failed_to_associate_roles_with_user", fmt.Sprintf("failed to associate roles with user: %s", err.Error()), "user_data_store", nil)
			return nil, diag
		}
	}

	// Associate claims if any were provided
	if len(claimsToAssociate) > 0 {
		// Get the full claim entities from the database
		var dbClaims []entities.Claim
		for _, claim := range claimsToAssociate {
			var dbClaim entities.Claim
			if result := s.GetDB().Where("id = ?", claim.ID).First(&dbClaim); result.Error != nil {
				diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim with id %s: %s", claim.ID, result.Error.Error()), "user_data_store", nil)
				return nil, diag
			}
			dbClaims = append(dbClaims, dbClaim)
		}

		// First, clear any existing claim associations
		if err := s.GetDB().Model(user).Association("Claims").Clear(); err != nil {
			diag.AddError("failed_to_clear_existing_claim_associations", fmt.Sprintf("failed to clear existing claim associations: %s", err.Error()), "user_data_store", nil)
			return nil, diag
		}

		// Then add the new claim associations
		if err := s.GetDB().Model(user).Association("Claims").Append(dbClaims); err != nil {
			diag.AddError("failed_to_associate_claims_with_user", fmt.Sprintf("failed to associate claims with user: %s", err.Error()), "user_data_store", nil)
			return nil, diag
		}
	}

	return user, diag
}

// GetUserByID retrieves a user by ID
func (s *UserDataStore) GetUserByID(ctx *appctx.AppContext, tenantID string, id string) (*entities.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_by_id")
	var user entities.User
	result := s.GetDB().Preload("Roles").Preload("Claims").First(&user, "tenant_id = ? AND (id = ? OR slug = ?)", tenantID, id, id)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_user_by_id", fmt.Sprintf("failed to get user by id: %s", result.Error.Error()), "user_data_store", nil)
		return nil, diag
	}
	return &user, diag
}

// GetUserByUsername retrieves a user by username
func (s *UserDataStore) GetUserByUsername(ctx *appctx.AppContext, tenantID string, username string) (*entities.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_by_username")
	var user entities.User
	result := s.GetDB().
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		First(&user, "tenant_id = ? AND username = ?", tenantID, username)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_user_by_username", fmt.Sprintf("failed to get user by username: %s", result.Error.Error()), "user_data_store", nil)
		return nil, diag
	}
	return &user, diag
}

func (s *UserDataStore) GetUsers(ctx *appctx.AppContext, tenantID string) ([]entities.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_users")
	var users []entities.User
	result := s.GetDB().Where("tenant_id = ?", tenantID).Find(&users)
	if result.Error != nil {
		diag.AddError("failed_to_get_users", fmt.Sprintf("failed to get users: %s", result.Error.Error()), "user_data_store", nil)
		return nil, diag
	}
	return users, diag
}

func (s *UserDataStore) GetUsersByQuery(ctx *appctx.AppContext, tenantID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_users_by_query")
	db := s.GetDB()
	db = db.Preload("Roles", func(db *gorm.DB) *gorm.DB {
		return db.Order("roles.created_at DESC")
	}).Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	}).Where("tenant_id = ?", tenantID)

	result, err := db_utils.QueryDatabase[entities.User](db, tenantID, queryObj)
	if err != nil {
		diag.AddError("failed_to_get_users_by_query", fmt.Sprintf("failed to get users by query: %s", err.Error()), "user_data_store", nil)
		return nil, diag
	}
	return result, diag
}

// UpdateUser updates an existing user
func (s *UserDataStore) UpdateUser(ctx *appctx.AppContext, tenantID string, user *entities.User) *diagnostics.Diagnostics {
	diag := diagnostics.New("update_user")
	user.UpdatedAt = time.Now()
	user.TenantID = tenantID
	currentUser, getUserDiag := s.GetUserByID(ctx, tenantID, user.ID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user.Password != "" {
		encryptionService := encryption.GetInstance()
		encryptedPassword, err := encryptionService.HashPassword(user.Password)
		if err != nil {
			diag.AddError("failed_to_encrypt_password", fmt.Sprintf("failed to encrypt password: %s", err.Error()), "user_data_store", nil)
			return diag
		}
		user.Password = encryptedPassword
	}
	if user.Username != "" {
		user.Slug = pkg_utils.Slugify(user.Username)
	}

	updates := db_utils.PartialUpdateMap(currentUser, user, "updated_at", "slug")
	if err := s.GetDB().Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_user", fmt.Sprintf("failed to update user: %s", err.Error()), "user_data_store", nil)
		return diag
	}
	return diag
}

func (s *UserDataStore) UpdateUserPassword(ctx *appctx.AppContext, tenantID string, id string, password string) *diagnostics.Diagnostics {
	diag := diagnostics.New("update_user_password")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, id)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	encryptionService := encryption.GetInstance()
	encryptedPassword, err := encryptionService.HashPassword(password)
	if err != nil {
		diag.AddError("failed_to_encrypt_password", fmt.Sprintf("failed to encrypt password: %s", err.Error()), "user_data_store", nil)
		return diag
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
	updates := db_utils.PartialUpdateMap(user, updatedUser, "updated_at")
	if err := s.GetDB().Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_update_user_password", fmt.Sprintf("failed to update user password: %s", err.Error()), "user_data_store", nil)
		return diag
	}
	return diag
}

func (s *UserDataStore) BlockUser(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("block_user")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, id)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
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
	updates := db_utils.PartialUpdateMap(user, updatedUser, "updated_at")
	if err := s.GetDB().Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_block_user", fmt.Sprintf("failed to block user: %s", err.Error()), "user_data_store", nil)
		return diag
	}
	return diag
}

func (s *UserDataStore) SetRefreshToken(ctx *appctx.AppContext, tenantID string, id string, refreshToken string) *diagnostics.Diagnostics {
	diag := diagnostics.New("set_refresh_token")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, id)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
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
	updates := db_utils.PartialUpdateMap(user, updatedUser, "updated_at")
	if err := s.GetDB().Model(&entities.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		diag.AddError("failed_to_set_refresh_token", fmt.Sprintf("failed to set refresh token: %s", err.Error()), "user_data_store", nil)
		return diag
	}
	return diag
}

// DeleteUser deletes a user
func (s *UserDataStore) DeleteUser(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_user")
	user, getUserDiag := s.GetUserByID(ctx, tenantID, id)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if err := s.GetDB().Delete(user).Error; err != nil {
		diag.AddError("failed_to_delete_user", fmt.Sprintf("failed to delete user: %s", err.Error()), "user_data_store", nil)
		return diag
	}
	return diag
}

func (s *UserDataStore) GetUserClaims(ctx *appctx.AppContext, tenantID string, userID string) ([]entities.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_claims")
	var user entities.User
	result := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Where("users.id = ?", userID).
		Find(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user_claims", fmt.Sprintf("failed to get user claims: %s", result.Error.Error()), "user_data_store", nil)
		return nil, diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return nil, diag
	}

	return user.Claims, diag
}

func (s *UserDataStore) GetUserClaimsByQuery(ctx *appctx.AppContext, tenantID string, userID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_claims_by_query")
	db := s.GetDB()
	db = db.Table("user_claims").
		Joins("JOIN claims ON claims.id = user_claims.claim_id").
		Where("user_claims.user_id = ?", userID).
		Where("claims.tenant_id = ?", tenantID)

	result, err := db_utils.QueryDatabase[entities.Claim](db, tenantID, queryObj)
	if err != nil {
		diag.AddError("failed_to_get_user_claims_by_query", fmt.Sprintf("failed to get user claims by query: %s", err.Error()), "user_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *UserDataStore) GetUserRoles(ctx *appctx.AppContext, tenantID string, userID string) ([]entities.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_roles")
	var user entities.User
	result := s.GetDB().
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Where("users.id = ?", userID).
		Find(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user_roles", fmt.Sprintf("failed to get user roles: %s", result.Error.Error()), "user_data_store", nil)
		return nil, diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return nil, diag
	}

	return user.Roles, diag
}

func (s *UserDataStore) GetUserRolesByQuery(ctx *appctx.AppContext, tenantID string, userID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_roles_by_query")
	db := s.GetDB()

	// query the user_roles table and join the roles table and filter by the user_id
	// and apply the query object to the query
	db = db.Table("user_roles").
		Joins("JOIN roles ON roles.id = user_roles.role_id").
		Where("user_roles.user_id = ?", userID).
		Where("roles.tenant_id = ?", tenantID)

	result, err := db_utils.QueryDatabase[entities.Role](db, tenantID, queryObj)
	if err != nil {
		diag.AddError("failed_to_get_user_roles_by_query", fmt.Sprintf("failed to get user roles by query: %s", err.Error()), "user_data_store", nil)
		return nil, diag
	}
	return result, diag
}

func (s *UserDataStore) AddUserToRole(ctx *appctx.AppContext, tenantID string, userID string, roleId string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_user_to_role")
	_, getUserDiag := s.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}

	// checking if the dbRole exists in the database
	var dbRole entities.Role
	roleDbResult := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, roleId).First(&dbRole)
	if roleDbResult.Error != nil {
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %s", roleDbResult.Error.Error()), "user_data_store", nil)
		return diag
	}
	if dbRole.ID == "" {
		diag.AddError("role_not_found", "role not found", "user_data_store", nil)
		return diag
	}

	// role exists in the relationship
	var userRole entities.UserRoles
	userRoleDbResult := s.GetDB().Where("user_id = ? AND role_id = ?", userID, roleId).First(&userRole)
	if userRoleDbResult.Error != nil {
		diag.AddError("failed_to_get_user_role", fmt.Sprintf("failed to get user role: %s", userRoleDbResult.Error.Error()), "user_data_store", nil)
		return diag
	}
	if userRole.RoleID != "" {
		diag.AddError("role_already_associated_with_user", "role already associated with user", "user_data_store", nil)
		return diag
	}

	// create the user role association
	userRole = entities.UserRoles{
		UserID: userID,
		RoleID: roleId,
	}

	createUserRoleDbResult := s.GetDB().Create(&userRole)
	if createUserRoleDbResult.Error != nil {
		diag.AddError("failed_to_create_user_role", fmt.Sprintf("failed to create user role: %s", createUserRoleDbResult.Error.Error()), "user_data_store", nil)
		return diag
	}

	return diag
}

func (s *UserDataStore) RemoveUserFromRole(ctx *appctx.AppContext, tenantID string, userID string, roleId string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_user_from_role")
	_, getUserDiag := s.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}

	// checking if the user role exists in the database
	var userRole entities.UserRoles
	userRoleDbResult := s.GetDB().Where("user_id = ? AND role_id = ?", userID, roleId).First(&userRole)
	if userRoleDbResult.Error != nil {
		diag.AddError("failed_to_get_user_role", fmt.Sprintf("failed to get user role: %s", userRoleDbResult.Error.Error()), "user_data_store", nil)
		return diag
	}
	if userRole.RoleID == "" {
		diag.AddError("user_role_not_found", "user role not found", "user_data_store", nil)
		return diag
	}

	// delete the user role association
	userRoleDbResult = s.GetDB().Delete(&userRole)
	if userRoleDbResult.Error != nil {
		diag.AddError("failed_to_delete_user_role", fmt.Sprintf("failed to delete user role: %s", userRoleDbResult.Error.Error()), "user_data_store", nil)
		return diag
	}

	return diag
}

func (s *UserDataStore) AddClaimToUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_claim_to_user")
	_, getUserDiag := s.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	var user entities.User
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, userID).First(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user", fmt.Sprintf("failed to get user: %s", result.Error.Error()), "user_data_store", nil)
		return diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag
	}

	var claim entities.Claim
	result = s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, claimIdOrSlug).First(&claim)
	if result.Error != nil {
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", result.Error.Error()), "user_data_store", nil)
		return diag
	}
	if claim.ID == "" {
		diag.AddError("claim_not_found", "claim not found", "user_data_store", nil)
		return diag
	}

	// Check if the claim is already associated with the user
	var userClaims entities.UserClaims
	result = s.GetDB().Where("user_id = ? AND claim_id = ?", user.ID, claim.ID).First(&userClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user_claim", fmt.Sprintf("failed to get user claim: %s", result.Error.Error()), "user_data_store", nil)
			return diag
		}
	}
	if userClaims.ClaimID != "" {
		diag.AddError("claim_already_associated_with_user", "claim already associated with user", "user_data_store", nil)
		return diag
	}

	// Create the user claim association
	userClaim := entities.UserClaims{
		UserID:  user.ID,
		ClaimID: claim.ID,
	}

	result = s.GetDB().Create(&userClaim)
	if result.Error != nil {
		diag.AddError("failed_to_create_user_claim", fmt.Sprintf("failed to create user claim: %s", result.Error.Error()), "user_data_store", nil)
		return diag
	}

	return diag
}

func (s *UserDataStore) RemoveClaimFromUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_claim_from_user")
	_, getUserDiag := s.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}

	var user entities.User
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, userID).First(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user", fmt.Sprintf("failed to get user: %s", result.Error.Error()), "user_data_store", nil)
		return diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "user_data_store", nil)
		return diag
	}

	var claim entities.Claim
	result = s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, claimIdOrSlug).First(&claim)
	if result.Error != nil {
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %s", result.Error.Error()), "user_data_store", nil)
		return diag
	}
	if claim.ID == "" {
		diag.AddError("claim_not_found", "claim not found", "user_data_store", nil)
		return diag
	}

	// Check if the claim is associated with the user
	var userClaims entities.UserClaims
	result = s.GetDB().Where("user_id = ? AND claim_id = ?", user.ID, claim.ID).First(&userClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user_claim", fmt.Sprintf("failed to get user claim: %s", result.Error.Error()), "user_data_store", nil)
			return diag
		}
	}
	if userClaims.ClaimID == "" {
		diag.AddError("claim_not_associated_with_user", "claim not associated with user", "user_data_store", nil)
		return diag
	}

	// Delete the user claim association
	result = s.GetDB().Delete(&userClaims)
	if result.Error != nil {
		diag.AddError("failed_to_delete_user_claim", fmt.Sprintf("failed to delete user claim: %s", result.Error.Error()), "user_data_store", nil)
		return diag
	}

	return diag
}
