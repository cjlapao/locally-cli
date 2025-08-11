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
	"github.com/cjlapao/locally-cli/pkg/models"
	pkg_utils "github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var (
	claimDataStoreInstance *ClaimDataStore
	claimDataStoreOnce     sync.Once
)

type ClaimDataStoreInterface interface {
	GetClaims(ctx *appctx.AppContext, tenantID string) ([]entities.Claim, error)
	GetClaimsByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Claim], error)
	GetClaimBySlugOrID(ctx *appctx.AppContext, tenantID string, slugOrID string) (*entities.Claim, error)
	GetClaimUsers(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.User, error)
	GetClaimUsersByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.User], error)
	GetPaginatedClaimUsers(ctx *appctx.AppContext, tenantID string, claimID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.User], error)
	CreateClaim(ctx *appctx.AppContext, tenantID string, claim *entities.Claim) (*entities.Claim, error)
	UpdateClaim(ctx *appctx.AppContext, tenantID string, claim *entities.Claim) error
	DeleteClaim(ctx *appctx.AppContext, tenantID string, id string) error
	GetClaimsByLevel(ctx *appctx.AppContext, tenantID string, level models.SecurityLevel) ([]entities.Claim, error)
	AddClaimToUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) error
	RemoveClaimFromUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) error
	GetClaimApiKeys(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.ApiKey, error)
	GetPaginatedClaimApiKeys(ctx *appctx.AppContext, tenantID string, claimID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.ApiKey], error)
	AddClaimToApiKey(ctx *appctx.AppContext, tenantID string, claimID string, apiKeyID string) error
	RemoveClaimFromApiKey(ctx *appctx.AppContext, tenantID string, claimID string, apiKeyID string) error
	GetClaimRoles(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.Role, error)
	GetPaginatedClaimRoles(ctx *appctx.AppContext, tenantID string, claimID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.Role], error)
	AddClaimToRole(ctx *appctx.AppContext, tenantID string, claimID string, roleID string) error
	RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, claimID string, roleID string) error
}

type ClaimDataStore struct {
	database.BaseDataStore
}

func GetClaimDataStoreInstance() ClaimDataStoreInterface {
	return claimDataStoreInstance
}

func InitializeClaimDataStore() (ClaimDataStoreInterface, *diagnostics.Diagnostics) {
	diag := diagnostics.New("initialize_claim_data_store")
	cfg := config.GetInstance().Get()
	logging.Info("Initializing claim store...")

	claimDataStoreOnce.Do(func() {
		dbService := database.GetInstance()
		if dbService == nil {
			diag.AddError("database_service_not_initialized", "database service not initialized", "claim_data_store", nil)
			return
		}

		store := &ClaimDataStore{
			BaseDataStore: *database.NewBaseDataStore(dbService.GetDB()),
		}

		if cfg.Get(config.DatabaseMigrateKey).GetBool() {
			logging.Info("Running claim migrations")
			if migrateDiag := store.Migrate(); migrateDiag.HasErrors() {
				diag.Append(migrateDiag)
				return
			}
			logging.Info("Claim migrations completed")
		}

		claimDataStoreInstance = store
	})

	logging.Info("Claim store initialized successfully")
	return claimDataStoreInstance, diag
}

func (s *ClaimDataStore) Migrate() *diagnostics.Diagnostics {
	diag := diagnostics.New("migrate_claim_data_store")
	if err := s.GetDB().AutoMigrate(&entities.Claim{}); err != nil {
		diag.AddError("failed_to_migrate_claim_table", "failed to migrate claim table", "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) GetClaims(ctx *appctx.AppContext, tenantID string) ([]entities.Claim, error) {
	var claims []entities.Claim
	result := s.GetDB().Where("tenant_id = ?", tenantID).Find(&claims)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get claims: %w", result.Error)
	}
	return claims, nil
}

func (s *ClaimDataStore) GetClaimsByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.Claim], error) {
	return utils.PaginatedFilteredQuery(s.GetDB(), tenantID, filterObj, entities.Claim{})
}

func (s *ClaimDataStore) GetClaimBySlugOrID(ctx *appctx.AppContext, tenantID string, slugOrID string) (*entities.Claim, error) {
	var claim entities.Claim
	result := s.GetDB().Where("tenant_id = ?", tenantID).First(&claim, "(slug = ? OR id = ?)", slugOrID, slugOrID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to get claim: %w", result.Error)
	}
	return &claim, nil
}

func (s *ClaimDataStore) GetClaimUsers(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.User, error) {
	var users []entities.User
	claim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("claim not found")
	}
	query := s.GetDB().
		Preload("Roles").
		Preload("Claims").
		Joins("JOIN user_claims ON users.id = user_claims.user_id").
		Where("user_claims.claim_id = ?", claim.ID)
	if tenantID != "" {
		query = query.Where("users.tenant_id = ?", tenantID)
	}
	result := query.Find(&users)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get claim users: %w", result.Error)
	}
	return users, nil
}

func (s *ClaimDataStore) GetClaimUsersByFilter(ctx *appctx.AppContext, tenantID string, filterObj *filters.Filter) (*filters.FilterResponse[entities.User], error) {
	return utils.PaginatedFilteredQueryWithPreload(s.GetDB(), tenantID, filterObj, entities.User{}, "Roles", "Claims")
}

func (s *ClaimDataStore) GetPaginatedClaimUsers(ctx *appctx.AppContext, tenantID string, claimID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.User], error) {
	claim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("claim not found")
	}
	query := s.GetDB().
		Preload("Roles").
		Preload("Claims").
		Joins("JOIN user_claims ON users.id = user_claims.user_id").
		Where("user_claims.claim_id = ?", claim.ID)
	result, err := utils.PaginatedQuery(query, tenantID, pagination, entities.User{})
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated claim users: %w", err)
	}

	return result, nil
}

func (s *ClaimDataStore) CreateClaim(ctx *appctx.AppContext, tenantID string, claim *entities.Claim) (*entities.Claim, error) {
	claim.ID = uuid.New().String()
	claim.CreatedAt = time.Now()
	claim.UpdatedAt = time.Now()
	claim.TenantID = tenantID
	if claim.Slug != "" {
		claim.Slug = claim.GetSlug()
	}

	result := s.GetDB().Create(claim)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to create claim: %w", result.Error)
	}
	return claim, nil
}

func (s *ClaimDataStore) UpdateClaim(ctx *appctx.AppContext, tenantID string, claim *entities.Claim) error {
	claim.UpdatedAt = time.Now()
	if claim.Slug != "" {
		claim.Slug = pkg_utils.Slugify(claim.Slug)
	}

	// check if the claim exists in the database
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claim.ID)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}
	if existingClaim == nil {
		return fmt.Errorf("claim not found")
	}

	// using the partial update map to update the claim
	updates := utils.PartialUpdateMap(existingClaim, claim, "updated_at", "slug")
	if err := s.GetDB().Model(&entities.Claim{}).Where("id = ?", claim.ID).Updates(updates).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	return nil
}

func (s *ClaimDataStore) DeleteClaim(ctx *appctx.AppContext, tenantID string, id string) error {
	return s.GetDB().Where("tenant_id = ?", tenantID).Delete(&entities.Claim{}, "id = ? OR slug = ?", id, id).Error
}

func (s *ClaimDataStore) GetClaimsByLevel(ctx *appctx.AppContext, tenantID string, level models.SecurityLevel) ([]entities.Claim, error) {
	var claims []entities.Claim
	result := s.GetDB().Where("tenant_id = ?", tenantID).Where("security_level = ?", level).Find(&claims)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get all user level claims: %w", result.Error)
	}
	return claims, nil
}

func (s *ClaimDataStore) AddClaimToUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) error {
	var user entities.User
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, userID).First(&user)
	if result.Error != nil {
		return fmt.Errorf("failed to get user: %w", result.Error)
	}
	if user.ID == "" {
		return fmt.Errorf("user not found")
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimIdOrSlug)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}
	if existingClaim == nil {
		return fmt.Errorf("claim not found")
	}

	// check if the claim is already assigned to the user
	var userClaims entities.UserClaims
	result = s.GetDB().Where("user_id = ? AND claim_id = ?", user.ID, existingClaim.ID).First(&userClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get user claim: %w", result.Error)
		}
	}
	if userClaims.ClaimID != "" {
		return fmt.Errorf("claim already assigned to user")
	}

	// create the user claim
	userClaims.UserID = user.ID
	userClaims.ClaimID = existingClaim.ID
	result = s.GetDB().Create(&userClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to create user claim: %w", result.Error)
	}

	return nil
}

func (s *ClaimDataStore) RemoveClaimFromUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) error {
	var user entities.User
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, userID).First(&user)
	if result.Error != nil {
		return fmt.Errorf("failed to get user: %w", result.Error)
	}
	if user.ID == "" {
		return fmt.Errorf("user not found")
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimIdOrSlug)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}
	if existingClaim == nil {
		return fmt.Errorf("claim not found")
	}

	// check if the claim is assigned to the user
	var userClaims entities.UserClaims
	result = s.GetDB().Where("user_id = ? AND claim_id = ?", user.ID, existingClaim.ID).First(&userClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to get user claim: %w", result.Error)
	}
	if userClaims.ClaimID == "" {
		return fmt.Errorf("claim not assigned to user")
	}

	// delete the user claim
	result = s.GetDB().Where("user_id = ? AND claim_id = ?", user.ID, existingClaim.ID).Delete(&userClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to delete user claim: %w", result.Error)
	}

	return nil
}

func (s *ClaimDataStore) GetClaimApiKeys(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.ApiKey, error) {
	var apiKeys []entities.ApiKey
	claim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("claim not found")
	}

	query := s.GetDB().
		Preload("Claims").
		Joins("JOIN api_key_claims ON api_keys.id = api_key_claims.api_key_id").
		Where("api_key_claims.claim_id = ?", claim.ID)
	if tenantID != "" {
		query = query.Where("api_keys.tenant_id = ?", tenantID)
	}

	result := query.Find(&apiKeys)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get API keys: %w", result.Error)
	}

	return apiKeys, nil
}

func (s *ClaimDataStore) GetPaginatedClaimApiKeys(ctx *appctx.AppContext, tenantID string, claimID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.ApiKey], error) {
	claim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("claim not found")
	}

	query := s.GetDB().
		Preload("Claims").
		Joins("JOIN api_key_claims ON api_keys.id = api_key_claims.api_key_id").
		Where("api_key_claims.claim_id = ?", claim.ID)
	if tenantID != "" {
		query = query.Where("api_keys.tenant_id = ?", tenantID)
	}

	result, err := utils.PaginatedQuery(query, tenantID, pagination, entities.ApiKey{})
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated API keys: %w", err)
	}

	return result, nil
}

func (s *ClaimDataStore) AddClaimToApiKey(ctx *appctx.AppContext, tenantID string, claimID string, apiKeyID string) error {
	var apiKey entities.ApiKey
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, apiKeyID).First(&apiKey)
	if result.Error != nil {
		return fmt.Errorf("failed to get API key: %w", result.Error)
	}
	if apiKey.ID == "" {
		return fmt.Errorf("API key not found")
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}
	if existingClaim == nil {
		return fmt.Errorf("claim not found")
	}

	// check if the claim is already assigned to the api key
	var apiKeyClaims entities.ApiKeyClaims
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ?", apiKey.ID, existingClaim.ID).First(&apiKeyClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get API key claim: %w", result.Error)
		}
	}
	if apiKeyClaims.ClaimID != "" {
		return fmt.Errorf("claim already assigned to API key")
	}

	// create the api key claim
	apiKeyClaims.ApiKeyID = apiKey.ID
	apiKeyClaims.ClaimID = existingClaim.ID
	result = s.GetDB().Create(&apiKeyClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to create API key claim: %w", result.Error)
	}

	return nil
}

func (s *ClaimDataStore) RemoveClaimFromApiKey(ctx *appctx.AppContext, tenantID string, claimID string, apiKeyID string) error {
	var apiKey entities.ApiKey
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, apiKeyID).First(&apiKey)
	if result.Error != nil {
		return fmt.Errorf("failed to get API key: %w", result.Error)
	}
	if apiKey.ID == "" {
		return fmt.Errorf("API key not found")
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}
	if existingClaim == nil {
		return fmt.Errorf("claim not found")
	}

	// check if the claim is assigned to the api key
	var apiKeyClaims entities.ApiKeyClaims
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ?", apiKey.ID, existingClaim.ID).First(&apiKeyClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to get API key claim: %w", result.Error)
	}
	if apiKeyClaims.ClaimID == "" {
		return fmt.Errorf("claim not assigned to API key")
	}

	// delete the api key claim
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ?", apiKey.ID, existingClaim.ID).Delete(&apiKeyClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to delete API key claim: %w", result.Error)
	}

	return nil
}

func (s *ClaimDataStore) GetPaginatedClaimRoles(ctx *appctx.AppContext, tenantID string, claimID string, pagination *filters.Pagination) (*filters.PaginationResponse[entities.Role], error) {
	claim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("claim not found")
	}

	query := s.GetDB().
		Preload("Claims").
		Joins("JOIN role_claims ON roles.id = role_claims.role_id").
		Where("role_claims.claim_id = ?", claim.ID)
	if tenantID != "" {
		query = query.Where("roles.tenant_id = ?", tenantID)
	}

	result, err := utils.PaginatedQuery(query, tenantID, pagination, entities.Role{})
	if err != nil {
		return nil, fmt.Errorf("failed to get paginated roles: %w", err)
	}

	return result, nil
}

func (s *ClaimDataStore) GetClaimRoles(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.Role, error) {
	var roles []entities.Role
	claim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		return nil, fmt.Errorf("failed to get claim: %w", err)
	}
	if claim == nil {
		return nil, fmt.Errorf("claim not found")
	}

	query := s.GetDB().
		Preload("Claims").
		Joins("JOIN role_claims ON roles.id = role_claims.role_id").
		Where("role_claims.claim_id = ?", claim.ID)
	if tenantID != "" {
		query = query.Where("roles.tenant_id = ?", tenantID)
	}

	result := query.Find(&roles)
	if result.Error != nil {
		return nil, fmt.Errorf("failed to get roles: %w", result.Error)
	}

	return roles, nil
}

func (s *ClaimDataStore) AddClaimToRole(ctx *appctx.AppContext, tenantID string, claimID string, roleID string) error {
	var role entities.Role
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, roleID).First(&role)
	if result.Error != nil {
		return fmt.Errorf("failed to get role: %w", result.Error)
	}
	if role.ID == "" {
		return fmt.Errorf("role not found")
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}
	if existingClaim == nil {
		return fmt.Errorf("claim not found")
	}

	// check if the claim is assigned to the role
	var roleClaims entities.RoleClaims
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, existingClaim.ID).First(&roleClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to get role claim: %w", result.Error)
		}
	}
	if roleClaims.ClaimID != "" {
		return fmt.Errorf("claim already assigned to role")
	}

	// create the role claim

	roleClaims.RoleID = role.ID
	roleClaims.ClaimID = existingClaim.ID
	result = s.GetDB().Create(&roleClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to create role claim: %w", result.Error)
	}

	return nil
}

func (s *ClaimDataStore) RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, claimID string, roleID string) error {
	var role entities.Role
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, roleID).First(&role)
	if result.Error != nil {
		return fmt.Errorf("failed to get role: %w", result.Error)
	}
	if role.ID == "" {
		return fmt.Errorf("role not found")
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		return fmt.Errorf("failed to get claim: %w", err)
	}
	if existingClaim == nil {
		return fmt.Errorf("claim not found")
	}

	// check if the claim is assigned to the role
	var roleClaims entities.RoleClaims
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, existingClaim.ID).First(&roleClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to get role claim: %w", result.Error)
	}
	if roleClaims.ClaimID == "" {
		return fmt.Errorf("claim not assigned to role")
	}

	// delete the role claim
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, existingClaim.ID).Delete(&roleClaims)
	if result.Error != nil {
		return fmt.Errorf("failed to delete role claim: %w", result.Error)
	}

	return nil
}
