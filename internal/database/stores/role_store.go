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
	GetRoles(ctx *appctx.AppContext, tenantID string) ([]entities.Role, error)
	GetRolesByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Role], error)
	GetRoleBySlugOrID(ctx *appctx.AppContext, tenantID string, slugOrID string) (*entities.Role, error)
	GetRoleUsers(ctx *appctx.AppContext, tenantID string, roleID string) ([]entities.User, error)
	GetRoleUsersByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.User], error)
	GetPaginatedRoleUsers(ctx *appctx.AppContext, tenantID string, roleID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.User], error)
	CreateRole(ctx *appctx.AppContext, tenantID string, role *entities.Role) (*entities.Role, error)
	UpdateRole(ctx *appctx.AppContext, tenantID string, role *entities.Role) error
	DeleteRole(ctx *appctx.AppContext, tenantID string, id string) error
	GetRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string) ([]entities.Claim, error)
	GetPaginatedRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.Claim], error)
	GetUserRoles(ctx *appctx.AppContext, tenantID string, userID string) ([]entities.Role, error)
	AddUserToRole(ctx *appctx.AppContext, tenantID string, userID string, roleIdOrSlug string) error
	RemoveUserFromRole(ctx *appctx.AppContext, tenantID string, userID string, roleIdOrSlug string) error
	AddClaimToRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) error
	RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) error
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
		diag.AddError("failed_to_migrate_role_claim_table", "failed to migrate role claim table", "role_data_store", nil)
		return diag
	}

	if err := s.GetDB().Exec("CREATE UNIQUE INDEX IF NOT EXISTS idx_role_claims_unique ON role_claims(role_id, claim_id);").Error; err != nil {
		diag.AddError("failed_to_create_unique_index_on_role_claims", "failed to create unique index on role claims", "role_data_store", nil)
		return diag
	}

	if err := s.GetDB().AutoMigrate(&entities.Role{}); err != nil {
		diag.AddError("failed_to_migrate_role_table", "failed to migrate role table", "role_data_store", nil)
		return diag
	}

	return diag
}

func (s *RoleDataStore) GetRoles(ctx *appctx.AppContext, tenantID string) ([]entities.Role, error) {
	var roles []entities.Role
	result := s.GetDB().
		Preload("Claims").
		Where("tenant_id = ?", tenantID).
		Find(&roles)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get roles: %w", result.Error)
	}
	return roles, nil
}

func (s *RoleDataStore) GetRolesByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Role], error) {
	return utils.PaginatedFilteredQueryWithPreload(s.GetDB(), tenantID, filterObj, entities.Role{}, "Claims")
}

func (s *RoleDataStore) GetRoleBySlugOrID(ctx *appctx.AppContext, tenantID string, slugOrID string) (*entities.Role, error) {
	var role entities.Role
	result := s.GetDB().
		Preload("Claims").
		Where("tenant_id = ?", tenantID).
		First(&role, "(slug = ? OR id = ?)", slugOrID, slugOrID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get role: %w", result.Error)
	}
	return &role, nil
}

func (s *RoleDataStore) GetPaginatedRoleUsers(ctx *appctx.AppContext, tenantID string, roleID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.User], error) {
	role, err := s.GetRoleBySlugOrID(ctx, tenantID, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	query := s.GetDB().
		Preload("Roles").
		Preload("Claims").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", role.ID)
	result, err := utils.PaginatedQuery(query, tenantID, pagination, entities.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated role users: %w", err)
	}

	return result, nil
}

func (s *RoleDataStore) GetRoleUsers(ctx *appctx.AppContext, tenantID string, roleID string) ([]entities.User, error) {
	var users []entities.User
	role, err := s.GetRoleBySlugOrID(ctx, tenantID, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	query := s.GetDB().
		Preload("Roles").
		Preload("Claims").
		Joins("JOIN user_roles ON users.id = user_roles.user_id").
		Where("user_roles.role_id = ?", role.ID)
	if tenantID != "" {
		query = query.Where("users.tenant_id = ?", tenantID)
	}
	result := query.Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get role users: %w", result.Error)
	}
	return users, nil
}

func (s *RoleDataStore) GetRoleUsersByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.User], error) {
	return utils.PaginatedFilteredQueryWithPreload(s.GetDB(), tenantID, filterObj, entities.User{}, "Roles", "Claims")
}

func (s *RoleDataStore) GetRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string) ([]entities.Claim, error) {
	var role entities.Role
	result := s.GetDB().
		Preload("Claims").
		Where("tenant_id = ?", tenantID).
		Where("id = ?", roleID).
		Find(&role)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get role claims: %w", result.Error)
	}
	return role.Claims, nil
}

func (s *RoleDataStore) GetPaginatedRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.Claim], error) {
	role, err := s.GetRoleBySlugOrID(ctx, tenantID, roleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return nil, fmt.Errorf("role not found")
	}
	query := s.GetDB().
		Preload("Claims").
		Joins("JOIN role_claims ON claims.id = role_claims.claim_id").
		Where("role_claims.role_id = ?", role.ID)
	result, err := utils.PaginatedQuery(query, tenantID, pagination, entities.Claim{})
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated role claims: %w", err)
	}
	return result, nil
}

func (s *RoleDataStore) CreateRole(ctx *appctx.AppContext, tenantID string, role *entities.Role) (*entities.Role, error) {
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
		return nil, fmt.Errorf("failed to create role: %w", result.Error)
	}

	// Associate claims if any were provided
	if len(claimsToAssociate) > 0 {
		// Get the full claim entities from the database
		var dbClaims []entities.Claim
		for _, claim := range claimsToAssociate {
			var dbClaim entities.Claim
			if result := s.GetDB().Where("id = ?", claim.ID).First(&dbClaim); result.Error != nil {
				return nil, fmt.Errorf("failed to get claim with id %s: %w", claim.ID, result.Error)
			}
			dbClaims = append(dbClaims, dbClaim)
		}

		// Associate the claims with the role using Replace to avoid duplicates
		// First, clear any existing associations
		if err := s.GetDB().Model(role).Association("Claims").Clear(); err != nil {
			return nil, fmt.Errorf("failed to clear existing claims associations: %w", err)
		}

		// Then add the new associations
		if err := s.GetDB().Model(role).Association("Claims").Append(dbClaims); err != nil {
			return nil, fmt.Errorf("failed to associate claims with role: %w", err)
		}
	}

	return role, nil
}

func (s *RoleDataStore) UpdateRole(ctx *appctx.AppContext, tenantID string, role *entities.Role) error {
	role.UpdatedAt = time.Now()
	if role.Slug != "" {
		role.Slug = pkg_utils.Slugify(role.Slug)
	}
	// check if the role exists in the database
	existingRole, err := s.GetRoleBySlugOrID(ctx, tenantID, role.Slug)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if existingRole == nil {
		return fmt.Errorf("role not found")
	}

	// using the partial update map to update the role
	updates := utils.PartialUpdateMap(existingRole, role, "updated_at", "slug")
	if err := s.GetDB().Model(&entities.Role{}).Where("id = ?", role.ID).Updates(updates).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (s *RoleDataStore) DeleteRole(ctx *appctx.AppContext, tenantID string, id string) error {
	return s.GetDB().Where("tenant_id = ?", tenantID).Delete(&entities.Role{}, "id = ?", id).Error
}

func (s *RoleDataStore) GetUserRoles(ctx *appctx.AppContext, tenantID string, userID string) ([]entities.Role, error) {
	var user entities.User
	result := s.GetDB().
		Preload("Roles").
		Where("users.id = ?", userID).
		Find(&user)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get user roles: %w", result.Error)
	}
	if user.ID == "" {
		return nil, fmt.Errorf("user not found")
	}

	return user.Roles, nil
}

func (s *RoleDataStore) AddUserToRole(ctx *appctx.AppContext, tenantID string, userID string, roleIdOrSlug string) error {
	var user entities.User
	result := s.GetDB().
		Where("id = ? AND tenant_id = ?", userID, tenantID).
		First(&user)
	if result.Error != nil {
		return fmt.Errorf("failed to get user: %w", result.Error)
	}
	if user.ID == "" {
		return fmt.Errorf("user not found")
	}

	// check if the roles exist
	existingRole, err := s.GetRoleBySlugOrID(ctx, tenantID, roleIdOrSlug)
	if err != nil {
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get role: %w", err)
		}
	}
	if existingRole.ID == "" {
		return fmt.Errorf("role not found")
	}

	// check if the user is already in the role
	var userRole entities.UserRoles
	result = s.GetDB().Where("user_id = ? AND role_id = ?", user.ID, existingRole.ID).First(&userRole)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get user role: %w", result.Error)
		}
	}
	if userRole.RoleID != "" {
		return fmt.Errorf("user already in role")
	}

	// add the role to the user
	userRole.UserID = user.ID
	userRole.RoleID = existingRole.ID
	result = s.GetDB().Create(&userRole)
	if result.Error != nil {
		return fmt.Errorf("failed to add user to role: %w", result.Error)
	}

	return nil
}

func (s *RoleDataStore) RemoveUserFromRole(ctx *appctx.AppContext, tenantID string, userID string, roleIdOrSlug string) error {
	var user entities.User
	result := s.GetDB().
		Preload("Roles").
		Where("users.id = ?", userID).
		Find(&user)
	if result.Error != nil {
		return fmt.Errorf("failed to get user: %w", result.Error)
	}
	if user.ID == "" {
		return fmt.Errorf("user not found")
	}

	// check if the role exists
	role, err := s.GetRoleBySlugOrID(ctx, tenantID, roleIdOrSlug)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if role == nil {
		return fmt.Errorf("role not found")
	}

	// remove the role from the user
	var userRole entities.UserRoles
	result = s.GetDB().Where("user_id = ? AND role_id = ?", user.ID, role.ID).First(&userRole)
	if result.Error != nil {
		return fmt.Errorf("failed to get user role: %w", result.Error)
	}

	if userRole.RoleID == "" {
		return fmt.Errorf("user role not found")
	}

	result = s.GetDB().Where("user_id = ? AND role_id = ?", user.ID, role.ID).Delete(&userRole)
	if result.Error != nil {
		return fmt.Errorf("failed to remove user from role: %w", result.Error)
	}

	return nil
}

func (s *RoleDataStore) AddClaimToRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) error {
	var role entities.Role
	result := s.GetDB().
		Where("id = ? AND tenant_id = ?", roleID, tenantID).
		First(&role)
	if result.Error != nil {
		return fmt.Errorf("failed to get role: %w", result.Error)
	}
	if role.ID == "" {
		return fmt.Errorf("role not found")
	}

	// check if the claim exists
	var claim entities.Claim
	result = s.GetDB().Where("id = ?", claimID).First(&claim)
	if result.Error != nil {
		return fmt.Errorf("failed to get claim: %w", result.Error)
	}
	if claim.ID == "" {
		return fmt.Errorf("claim not found")
	}

	// check if the claim is already in the role
	var roleClaim entities.RoleClaims
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, claim.ID).First(&roleClaim)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get role claim: %w", result.Error)
		}
	}
	if roleClaim.RoleID != "" {
		return fmt.Errorf("claim already in role")
	}

	// add the claim to the role
	roleClaim.RoleID = role.ID
	roleClaim.ClaimID = claim.ID
	result = s.GetDB().Create(&roleClaim)
	if result.Error != nil {
		return fmt.Errorf("failed to add claim to role: %w", result.Error)
	}

	return nil
}

func (s *RoleDataStore) RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) error {
	var role entities.Role
	result := s.GetDB().
		Where("id = ? AND tenant_id = ?", roleID, tenantID).
		First(&role)
	if result.Error != nil {
		return fmt.Errorf("failed to get role: %w", result.Error)
	}
	if role.ID == "" {
		return fmt.Errorf("role not found")
	}

	// check if the claim exists
	var claim entities.Claim
	result = s.GetDB().Where("id = ?", claimID).First(&claim)
	if result.Error != nil {
		return fmt.Errorf("failed to get claim: %w", result.Error)
	}
	if claim.ID == "" {
		return fmt.Errorf("claim not found")
	}

	// check if the claim is in the role
	var roleClaim entities.RoleClaims
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, claim.ID).First(&roleClaim)
	if result.Error != nil {
		return fmt.Errorf("failed to get role claim: %w", result.Error)
	}

	// remove the claim from the role
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, claim.ID).Delete(&roleClaim)
	if result.Error != nil {
		return fmt.Errorf("failed to remove claim from role: %w", result.Error)
	}

	return nil
}
