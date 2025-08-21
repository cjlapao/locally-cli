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
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_utils "github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	roleDataStoreInstance *RoleDataStore
	roleDataStoreOnce     sync.Once
)

type RoleDataStoreInterface interface {
	GetRoles(ctx *appctx.AppContext, tenantID string) ([]entities.Role, *diagnostics.Diagnostics)
	GetRolesByQuery(ctx *appctx.AppContext, tenantID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *diagnostics.Diagnostics)
	GetRoleBySlugOrID(ctx *appctx.AppContext, tenantID string, slugOrID string) (*entities.Role, *diagnostics.Diagnostics)
	GetRoleUsers(ctx *appctx.AppContext, tenantID string, roleID string) ([]entities.User, *diagnostics.Diagnostics)
	GetRoleUsersByQuery(ctx *appctx.AppContext, tenantID string, roleID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *diagnostics.Diagnostics)
	CreateRole(ctx *appctx.AppContext, tenantID string, role *entities.Role) (*entities.Role, *diagnostics.Diagnostics)
	UpdateRole(ctx *appctx.AppContext, tenantID string, role *entities.Role) *diagnostics.Diagnostics
	DeleteRole(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics
	GetRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string) ([]entities.Claim, *diagnostics.Diagnostics)
	GetRoleClaimsByQuery(ctx *appctx.AppContext, tenantID string, roleID string, queryObj *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *diagnostics.Diagnostics)
	GetUserRoles(ctx *appctx.AppContext, tenantID string, userID string) ([]entities.Role, *diagnostics.Diagnostics)
	AddUserToRole(ctx *appctx.AppContext, tenantID string, userID string, roleIdOrSlug string) *diagnostics.Diagnostics
	RemoveUserFromRole(ctx *appctx.AppContext, tenantID string, userID string, roleIdOrSlug string) *diagnostics.Diagnostics
	AddClaimToRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) *diagnostics.Diagnostics
	RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) *diagnostics.Diagnostics
}

type RoleDataStore struct {
	database.BaseDataStore
}

func GetRoleDataStoreInstance() RoleDataStoreInterface {
	return roleDataStoreInstance
}

func InitializeRoleDataStore() (RoleDataStoreInterface, *diagnostics.Diagnostics) {
	diag := diagnostics.New("initialize_role_data_store")
	cfg := config.GetInstance().Get()
	logging.Info("Initializing role store...")

	roleDataStoreOnce.Do(func() {
		dbService := database.GetInstance()
		if dbService == nil {
			diag.AddError("database_service_not_initialized", "database service not initialized", "role_data_store", nil)
			return
		}

		store := &RoleDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running role migrations")
			if migrateDiag := store.Migrate(); migrateDiag.HasErrors() {
				diag.Append(migrateDiag)
				return
			}
			logging.Info("Role migrations completed")
		}

		roleDataStoreInstance = store
	})

	logging.Info("Role store initialized successfully")
	return roleDataStoreInstance, diag
}

func (s *RoleDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_role_data_store")

	if err := s.GetDB().AutoMigrate(&entities.RoleClaims{}); err != nil {
		diag.AddError("failed_to_migrate_role_claim_table", fmt.Sprintf("failed to migrate role claim table: %v", err), "role_data_store", nil)
		return diag
	}

	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_role_claims_unique ON role_claims(role_id, claim_id);").Error; err != nil {
		diag.AddError("failed_to_create_unique_index_on_role_claims", fmt.Sprintf("failed to create unique index on role claims: %v", err), "role_data_store", nil)
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.Role{}); err != nil {
		diag.AddError("failed_to_migrate_role_table", fmt.Sprintf("failed to migrate role table: %v", err), "role_data_store", nil)
		return diag
	}

	return diag
}

func (s *RoleDataStore) GetRoles(ctx *appctx.AppContext, tenantID string) ([]entities.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_roles")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "role_data_store")
		return nil, diag
	}

	var roles []entities.Role
	result := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Where("tenant_id = ?", tenantID).
		Find(&roles)
	if result.Error != nil {
		diag.AddError("failed_to_get_roles", fmt.Sprintf("failed to get roles: %v", result.Error), "role_data_store", map[string]interface{}{
			"tenant_id": tenantID,
		})
		return nil, diag
	}
	return roles, diag
}

func (s *RoleDataStore) GetRolesByQuery(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_roles_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "role_data_store")
		return nil, diag
	}

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	db := s.GetDB()
	db = db.Preload("Claims", func(db *gorm.DB) *gorm.DB {
		return db.Order("claims.created_at DESC")
	})

	result, err := utils.QueryDatabase[entities.Role](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_roles", fmt.Sprintf("failed to get roles: %v", err), "role_data_store", map[string]interface{}{
			"tenant_id": tenantID,
		})
		return nil, diag
	}
	return result, diag
}

func (s *RoleDataStore) GetRoleBySlugOrID(ctx *appctx.AppContext, tenantID string, slugOrID string) (*entities.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_role_by_slug_or_id")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "role_data_store")
		return nil, diag
	}

	var role entities.Role
	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Where("tenant_id = ?", tenantID).
		First(&role, "(slug = ? OR id = ?)", slugOrID, slugOrID)
	if db.Error != nil {
		if errors.Is(db.Error, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %v", db.Error), "role_data_store", map[string]interface{}{
			"role_id": slugOrID,
		})
		return nil, diag
	}

	return &role, diag
}

func (s *RoleDataStore) GetRoleUsersByQuery(ctx *appctx.AppContext, tenantID string, roleID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_role_users_by_query")
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, roleID)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return nil, diag
	}
	if role == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}

	db := s.GetDB().
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", role.ID)
	result, err := utils.QueryDatabase[entities.User](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_role_users", fmt.Sprintf("failed to get role users: %v", err), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}

	return result, diag
}

func (s *RoleDataStore) GetRoleUsers(ctx *appctx.AppContext, tenantID string, roleID string) ([]entities.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_role_users")
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, roleID)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return nil, diag
	}
	if role == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}

	db := s.GetDB().
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", role.ID)

	if tenantID != "" {
		db = db.Where("users.tenant_id = ?", tenantID)
	}

	var users []entities.User
	result := db.Find(&users)
	if result.Error != nil {
		diag.AddError("failed_to_get_role_users", fmt.Sprintf("failed to get role users: %v", result.Error), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}

	return users, diag
}

func (s *RoleDataStore) GetRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string) ([]entities.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_role_claims")
	var role entities.Role
	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Where("tenant_id = ?", tenantID).
		Where("id = ?", roleID).
		Find(&role)
	if db.Error != nil {
		diag.AddError("failed_to_get_role_claims", fmt.Sprintf("failed to get role claims: %v", db.Error), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}
	return role.Claims, diag
}

func (s *RoleDataStore) GetRoleClaimsByQuery(ctx *appctx.AppContext, tenantID string, roleID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_role_claims_by_query")
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, roleID)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return nil, diag
	}
	if role == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}
	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN role_claims ON claims.id = role_claims.claim_id").
		Where("role_claims.role_id = ?", role.ID)
	result, err := utils.QueryDatabase[entities.Claim](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_role_claims", fmt.Sprintf("failed to get role claims: %v", err), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return nil, diag
	}
	return result, diag
}

func (s *RoleDataStore) CreateRole(ctx *appctx.AppContext, tenantID string, role *entities.Role) (*entities.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_create_role")
	role.ID = uuid.New().String()
	role.CreatedAt = time.Now()
	role.UpdatedAt = time.Now()
	role.TenantID = tenantID
	if role.Slug != "" {
		role.Slug = pkg_utils.Slugify(role.Slug)
	}

	// Store the claims to associate after creating the role
	claimsToAssociate := role.Claims
	role.Claims = nil // Clear claims to avoid GORM trying to create them

	result := s.GetDB().Create(role)
	if result.Error != nil {
		diag.AddError("failed_to_create_role", fmt.Sprintf("failed to create role: %v", result.Error), "role_data_store", map[string]interface{}{
			"role": role,
		})
		return nil, diag
	}

	// Associate claims if any were provided
	if len(claimsToAssociate) > 0 {
		// Get the full claim entities from the database
		var dbClaims []entities.Claim
		for _, claim := range claimsToAssociate {
			var dbClaim entities.Claim
			if result := s.GetDB().Where("id = ?", claim.ID).First(&dbClaim); result.Error != nil {
				diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim with id %s: %v", claim.ID, result.Error), "role_data_store", map[string]interface{}{
					"claim_id": claim.ID,
				})
				return nil, diag
			}
			dbClaims = append(dbClaims, dbClaim)
		}

		// Associate the claims with the role using Replace to avoid duplicates
		// First, clear any existing associations
		if err := s.GetDB().Model(role).Association("Claims").Clear(); err != nil {
			diag.AddError("failed_to_clear_existing_claims_associations", fmt.Sprintf("failed to clear existing claims associations: %v", err), "role_data_store", map[string]interface{}{
				"role_id": role.ID,
			})
			return nil, diag
		}

		// Then add the new associations
		if err := s.GetDB().Model(role).Association("Claims").Append(dbClaims); err != nil {
			diag.AddError("failed_to_associate_claims_with_role", fmt.Sprintf("failed to associate claims with role: %v", err), "role_data_store", map[string]interface{}{
				"role_id": role.ID,
			})
			return nil, diag
		}
	}

	return role, diag
}

func (s *RoleDataStore) UpdateRole(ctx *appctx.AppContext, tenantID string, role *entities.Role) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_update_role")
	role.UpdatedAt = time.Now()
	if role.Slug != "" {
		role.Slug = pkg_utils.Slugify(role.Slug)
	}
	// check if the role exists in the database
	existingRole, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, role.Slug)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return diag
	}
	if existingRole == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": role.ID,
		})
		return diag
	}

	// using the partial update map to update the role
	updates := utils.PartialUpdateMap(existingRole, role, "updated_at", "slug")
	if err := s.GetDB().Model(&entities.Role{}).Where("id = ?", role.ID).Updates(updates).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		diag.AddError("failed_to_update_role", fmt.Sprintf("failed to update role: %v", err), "role_data_store", map[string]interface{}{
			"role_id": role.ID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) DeleteRole(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_delete_role")
	result := s.GetDB().Where("tenant_id = ?", tenantID).Delete(&entities.Role{}, "id = ?", id)
	if result.Error != nil {
		diag.AddError("failed_to_delete_role", fmt.Sprintf("failed to delete role: %v", result.Error), "role_data_store", map[string]interface{}{
			"role_id": id,
		})
		return diag
	}
	return diag
}

func (s *RoleDataStore) GetUserRoles(ctx *appctx.AppContext, tenantID string, userID string) ([]entities.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_user_roles")
	var user entities.User
	result := s.GetDB().
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Preload("Claims", func(db *gorm.DB) *gorm.DB {
				return db.Order("claims.created_at DESC")
			}).Order("roles.created_at DESC")
		}).
		Where("users.id = ?", userID).
		Find(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user_roles", fmt.Sprintf("failed to get user roles: %v", result.Error), "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return nil, diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return nil, diag
	}

	return user.Roles, diag
}

func (s *RoleDataStore) AddUserToRole(ctx *appctx.AppContext, tenantID string, userID string, roleIdOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_add_user_to_role")
	var user entities.User
	result := s.GetDB().
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		First(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user", fmt.Sprintf("failed to get user: %v", result.Error), "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}

	// check if the roles exist
	existingRole, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, roleIdOrSlug)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return diag
	}
	if existingRole == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleIdOrSlug,
		})
		return diag
	}

	// check if the user is already in the role
	var userRole entities.UserRoles
	result = s.GetDB().Where("user_id = ? AND role_id = ?", user.ID, existingRole.ID).First(&userRole)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user_role", fmt.Sprintf("failed to get user role: %v", result.Error), "role_data_store", map[string]interface{}{
				"user_id": userID,
				"role_id": existingRole.ID,
			})
			return diag
		}
	}
	if userRole.RoleID != "" {
		diag.AddError("user_already_in_role", "user already in role", "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": existingRole.ID,
		})
		return diag
	}

	// add the role to the user
	userRole.UserID = user.ID
	userRole.RoleID = existingRole.ID
	result = s.GetDB().Create(&userRole)
	if result.Error != nil {
		diag.AddError("failed_to_add_user_to_role", fmt.Sprintf("failed to add user to role: %v", result.Error), "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": existingRole.ID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) RemoveUserFromRole(ctx *appctx.AppContext, tenantID string, userID string, roleIdOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_remove_user_from_role")
	var user entities.User
	result := s.GetDB().
		Preload("Roles").
		Where("users.id = ?", userID).
		Find(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user", fmt.Sprintf("failed to get user: %v", result.Error), "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "role_data_store", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}

	// check if the role exists
	role, getRoleDiag := s.GetRoleBySlugOrID(ctx, tenantID, roleIdOrSlug)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
		return diag
	}
	if role == nil {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleIdOrSlug,
		})
		return diag
	}

	// remove the role from the user
	var userRole entities.UserRoles
	result = s.GetDB().Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRole)
	if result.Error != nil {
		diag.AddError("failed_to_get_user_role", fmt.Sprintf("failed to get user role: %v", result.Error), "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": role.ID,
		})
		return diag
	}

	if userRole.RoleID == "" {
		diag.AddError("user_role_not_found", "user role not found", "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": role.ID,
		})
		return diag
	}

	result = s.GetDB().Where("user_id = ? AND role_id = ?", user.ID, role.ID).Delete(&userRole)
	if result.Error != nil {
		diag.AddError("failed_to_remove_user_from_role", fmt.Sprintf("failed to remove user from role: %v", result.Error), "role_data_store", map[string]interface{}{
			"user_id": userID,
			"role_id": role.ID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) AddClaimToRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_add_claim_to_role")
	var role entities.Role
	result := s.GetDB().
		Where("id = ? AND tenant_id = ?", roleID, tenantID).
		First(&role)
	if result.Error != nil {
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %v", result.Error), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return diag
	}
	if role.ID == "" {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return diag
	}

	// check if the claim exists
	var claim entities.Claim
	result = s.GetDB().Where("id = ?", claimID).First(&claim)
	if result.Error != nil {
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %v", result.Error), "role_data_store", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}
	if claim.ID == "" {
		diag.AddError("claim_not_found", "claim not found", "role_data_store", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}

	// check if the claim is already in the role
	var roleClaim entities.RoleClaims
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, claim.ID).First(&roleClaim)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_role_claim", fmt.Sprintf("failed to get role claim: %v", result.Error), "role_data_store", map[string]interface{}{
				"role_id":  roleID,
				"claim_id": claimID,
			})
			return diag
		}
	}
	if roleClaim.RoleID != "" {
		diag.AddError("claim_already_in_role", "claim already in role", "role_data_store", map[string]interface{}{
			"role_id":  roleID,
			"claim_id": claimID,
		})
		return diag
	}

	// add the claim to the role
	roleClaim.RoleID = role.ID
	roleClaim.ClaimID = claim.ID
	result = s.GetDB().Create(&roleClaim)
	if result.Error != nil {
		diag.AddError("failed_to_add_claim_to_role", fmt.Sprintf("failed to add claim to role: %v", result.Error), "role_data_store", map[string]interface{}{
			"role_id":  roleID,
			"claim_id": claimID,
		})
		return diag
	}

	return diag
}

func (s *RoleDataStore) RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_remove_claim_from_role")
	var role entities.Role
	result := s.GetDB().
		Where("id = ? AND tenant_id = ?", roleID, tenantID).
		First(&role)
	if result.Error != nil {
		diag.AddError("failed_to_get_role", fmt.Sprintf("failed to get role: %v", result.Error), "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return diag
	}
	if role.ID == "" {
		diag.AddError("role_not_found", "role not found", "role_data_store", map[string]interface{}{
			"role_id": roleID,
		})
		return diag
	}

	// check if the claim exists
	var claim entities.Claim
	result = s.GetDB().Where("id = ?", claimID).First(&claim)
	if result.Error != nil {
		diag.AddError("failed_to_get_claim", fmt.Sprintf("failed to get claim: %v", result.Error), "role_data_store", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}
	if claim.ID == "" {
		diag.AddError("claim_not_found", "claim not found", "role_data_store", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}

	// check if the claim is in the role
	var roleClaim entities.RoleClaims
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, claim.ID).First(&roleClaim)
	if result.Error != nil {
		diag.AddError("failed_to_get_role_claim", fmt.Sprintf("failed to get role claim: %v", result.Error), "role_data_store", map[string]interface{}{
			"role_id":  roleID,
			"claim_id": claimID,
		})
		return diag
	}

	// remove the claim from the role
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, claim.ID).Delete(&roleClaim)
	if result.Error != nil {
		diag.AddError("failed_to_remove_claim_from_role", fmt.Sprintf("failed to remove claim from role: %v", result.Error), "role_data_store", map[string]interface{}{
			"role_id":  roleID,
			"claim_id": claimID,
		})
		return diag
	}

	return diag
}
