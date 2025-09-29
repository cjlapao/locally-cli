package stores

import (
	"errors"
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
	GetClaims(ctx *appctx.AppContext, tenantID string) ([]entities.Claim, *diagnostics.Diagnostics)
	GetClaimsByQuery(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *diagnostics.Diagnostics)
	GetClaimBySlugOrID(ctx *appctx.AppContext, tenantID string, slugOrID string) (*entities.Claim, *diagnostics.Diagnostics)
	GetClaimUsers(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.User, *diagnostics.Diagnostics)
	GetClaimUsersByQuery(ctx *appctx.AppContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *diagnostics.Diagnostics)
	CreateClaim(ctx *appctx.AppContext, tenantID string, claim *entities.Claim) (*entities.Claim, *diagnostics.Diagnostics)
	UpdateClaim(ctx *appctx.AppContext, tenantID string, claim *entities.Claim) *diagnostics.Diagnostics
	DeleteClaim(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics
	GetClaimsByLevel(ctx *appctx.AppContext, tenantID string, level models.SecurityLevel) ([]entities.Claim, *diagnostics.Diagnostics)
	AddClaimToUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics
	RemoveClaimFromUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics
	GetClaimApiKeys(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.ApiKey, *diagnostics.Diagnostics)
	GetClaimApiKeysByQuery(ctx *appctx.AppContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.ApiKey], *diagnostics.Diagnostics)
	AddClaimToApiKey(ctx *appctx.AppContext, tenantID string, claimID string, apiKeyID string) *diagnostics.Diagnostics
	RemoveClaimFromApiKey(ctx *appctx.AppContext, tenantID string, claimID string, apiKeyID string) *diagnostics.Diagnostics
	GetClaimRoles(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.Role, *diagnostics.Diagnostics)
	GetClaimRolesByQuery(ctx *appctx.AppContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *diagnostics.Diagnostics)
	AddClaimToRole(ctx *appctx.AppContext, tenantID string, claimID string, roleID string) *diagnostics.Diagnostics
	RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, claimID string, roleID string) *diagnostics.Diagnostics
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

func (s *ClaimDataStore) GetClaims(ctx *appctx.AppContext, tenantID string) ([]entities.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_claims")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}
	db := s.GetDB()

	var claims []entities.Claim
	result := db.Where("tenant_id = ?", tenantID).Find(&claims)
	if result.Error != nil {
		diag.AddError("failed_to_get_claims", "failed to get claims", "claim_data_store", nil)
		return nil, diag
	}
	return claims, diag
}

func (s *ClaimDataStore) GetClaimsByQuery(ctx *appctx.AppContext, tenantID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Claim], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_claims_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}

	db := s.GetDB()

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := utils.QueryDatabase[entities.Claim](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_claims_by_query", "failed to get claims by query", "claim_data_store", nil)
		return nil, diag
	}

	return result, diag
}

func (s *ClaimDataStore) GetClaimBySlugOrID(ctx *appctx.AppContext, tenantID string, slugOrID string) (*entities.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_claim_by_slug_or_id")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}

	db := s.GetDB()
	db = db.Where("tenant_id = ?", tenantID)

	var claim entities.Claim
	result := db.First(&claim, "(slug = ? OR id = ?)", slugOrID, slugOrID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, diag
		}
		diag.AddError("failed_to_get_claim", "failed to get claim", "claim_data_store", nil)
		return nil, diag
	}
	return &claim, diag
}

func (s *ClaimDataStore) GetClaimUsers(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_claim_users")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}

	var users []entities.User
	claim, userDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if userDiag.HasErrors() {
		diag.Append(userDiag)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}
	query := s.GetDB().
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN user_claims ON users.id = user_claims.user_id").
		Where("user_claims.claim_id = ?", claim.ID)
	if tenantID != "" {
		query = query.Where("users.tenant_id = ?", tenantID)
	}
	result := query.Find(&users)
	if result.Error != nil {
		diag.AddError("failed_to_get_claim_users", "failed to get claim users", "claim_data_store", nil)
		return nil, diag
	}
	return users, diag
}

func (s *ClaimDataStore) GetClaimUsersByQuery(ctx *appctx.AppContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.User], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_claim_users_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}

	claim, userDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if userDiag.HasErrors() {
		diag.Append(userDiag)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}
	db := s.GetDB().
		Preload("Roles", func(db *gorm.DB) *gorm.DB {
			return db.Order("roles.created_at DESC")
		}).
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN user_claims ON users.id = user_claims.user_id").
		Where("user_claims.claim_id = ?", claim.ID)

	result, err := utils.QueryDatabase[entities.User](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_claim_users_by_query", "failed to get claim users by query", "claim_data_store", nil)
		return nil, diag
	}

	return result, diag
}

func (s *ClaimDataStore) CreateClaim(ctx *appctx.AppContext, tenantID string, claim *entities.Claim) (*entities.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_create_claim")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}

	claim.ID = uuid.New().String()
	claim.CreatedAt = time.Now()
	claim.UpdatedAt = time.Now()
	claim.TenantID = tenantID
	if claim.Slug != "" {
		claim.Slug = claim.GetSlug()
	}

	result := s.GetDB().Create(claim)
	if result.Error != nil {
		diag.AddError("failed_to_create_claim", "failed to create claim", "claim_data_store", nil)
		return nil, diag
	}
	return claim, diag
}

func (s *ClaimDataStore) UpdateClaim(ctx *appctx.AppContext, tenantID string, claim *entities.Claim) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_update_claim")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return diag
	}

	claim.UpdatedAt = time.Now()
	if claim.Slug != "" {
		claim.Slug = pkg_utils.Slugify(claim.Slug)
	}

	// check if the claim exists in the database
	claim.UpdatedAt = time.Now()
	if claim.Slug != "" {
		claim.Slug = pkg_utils.Slugify(claim.Slug)
	}

	// check if the claim exists in the database
	existingClaim, getClaimDiag := s.GetClaimBySlugOrID(ctx, tenantID, claim.ID)
	if getClaimDiag.HasErrors() {
		diag.Append(getClaimDiag)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// using the partial update map to update the claim
	updates := utils.PartialUpdateMap(existingClaim, claim, "updated_at", "slug")
	if err := s.GetDB().
		Model(&entities.Claim{}).
		Where("id = ?", claim.ID).
		Updates(updates).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		diag.AddError("failed_to_update_claim", "failed to update claim", "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) DeleteClaim(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_delete_claim")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return diag
	}

	err := s.GetDB().
		Where("tenant_id = ?", tenantID).
		Where("id = ? OR slug = ?", id, id).
		Delete(&entities.Claim{}).Error
	if err != nil {
		diag.AddError("failed_to_delete_claim", "failed to delete claim", "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) GetClaimsByLevel(ctx *appctx.AppContext, tenantID string, level models.SecurityLevel) ([]entities.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_claims_by_level")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}

	var claims []entities.Claim
	result := s.GetDB().
		Where("tenant_id = ?", tenantID).
		Where("security_level = ?", level).
		Find(&claims)
	if result.Error != nil {
		diag.AddError("failed_to_get_claims_by_level", "failed to get claims by level", "claim_data_store", nil)
		return nil, diag
	}
	return claims, diag
}

func (s *ClaimDataStore) AddClaimToUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_add_claim_to_user")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return diag
	}

	var user entities.User
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, userID).First(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user", "failed to get user", "claim_data_store", nil)
		return diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimIdOrSlug)
	if err != nil {
		diag.AddError("failed_to_get_claim", "failed to get claim", "claim_data_store", nil)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is already assigned to the user
	var userClaims entities.UserClaims
	result = s.GetDB().Where("user_id = ? AND claim_id = ?", user.ID, existingClaim.ID).First(&userClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_user_claim", "failed to get user claim", "claim_data_store", nil)
			return diag
		}
	}
	if userClaims.ClaimID != "" {
		diag.AddError("claim_already_assigned_to_user", "claim already assigned to user", "claim_data_store", nil)
		return diag
	}

	// create the user claim
	userClaims.UserID = user.ID
	userClaims.ClaimID = existingClaim.ID
	result = s.GetDB().Create(&userClaims)
	if result.Error != nil {
		diag.AddError("failed_to_create_user_claim", "failed to create user claim", "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) RemoveClaimFromUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_remove_claim_from_user")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return diag
	}

	var user entities.User
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, userID).First(&user)
	if result.Error != nil {
		diag.AddError("failed_to_get_user", "failed to get user", "claim_data_store", nil)
		return diag
	}
	if user.ID == "" {
		diag.AddError("user_not_found", "user not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimIdOrSlug)
	if err != nil {
		diag.AddError("failed_to_get_claim", "failed to get claim", "claim_data_store", nil)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is assigned to the user
	var userClaims entities.UserClaims
	result = s.GetDB().Where("user_id = ? AND claim_id = ?", user.ID, existingClaim.ID).First(&userClaims)
	if result.Error != nil {
		diag.AddError("failed_to_get_user_claim", "failed to get user claim", "claim_data_store", nil)
		return diag
	}
	if userClaims.ClaimID == "" {
		diag.AddError("claim_not_assigned_to_user", "claim not assigned to user", "claim_data_store", nil)
		return diag
	}

	// delete the user claim
	result = s.GetDB().Where("user_id = ? AND claim_id = ?", user.ID, existingClaim.ID).Delete(&userClaims)
	if result.Error != nil {
		diag.AddError("failed_to_delete_user_claim", "failed to delete user claim", "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) GetClaimApiKeys(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.ApiKey, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_claim_api_keys")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}

	var apiKeys []entities.ApiKey
	claim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		diag.AddError("failed_to_get_claim", "failed to get claim", "claim_data_store", nil)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
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
		diag.AddError("failed_to_get_api_keys", "failed to get API keys", "claim_data_store", nil)
		return nil, diag
	}

	return apiKeys, diag
}

func (s *ClaimDataStore) GetClaimApiKeysByQuery(ctx *appctx.AppContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.ApiKey], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_claim_api_keys_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}

	claim, getClaimDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if getClaimDiag.HasErrors() {
		diag.Append(getClaimDiag)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}

	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN api_key_claims ON api_keys.id = api_key_claims.api_key_id").
		Where("api_key_claims.claim_id = ?", claim.ID)

	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := utils.QueryDatabase[entities.ApiKey](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_api_keys", "failed to get API keys", "claim_data_store", nil)
		return nil, diag
	}

	return result, diag
}

func (s *ClaimDataStore) AddClaimToApiKey(ctx *appctx.AppContext, tenantID string, claimID string, apiKeyID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_add_claim_to_api_key")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return diag
	}

	var apiKey entities.ApiKey
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, apiKeyID).First(&apiKey)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_key", "failed to get API key", "claim_data_store", nil)
		return diag
	}
	if apiKey.ID == "" {
		diag.AddError("api_key_not_found", "API key not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		diag.AddError("failed_to_get_claim", "failed to get claim", "claim_data_store", nil)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is already assigned to the api key
	var apiKeyClaims entities.ApiKeyClaims
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ?", apiKey.ID, existingClaim.ID).First(&apiKeyClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_api_key_claim", "failed to get API key claim", "claim_data_store", nil)
			return diag
		}
	}
	if apiKeyClaims.ClaimID != "" {
		diag.AddError("claim_already_assigned_to_api_key", "claim already assigned to API key", "claim_data_store", nil)
		return diag
	}

	// create the api key claim
	apiKeyClaims.ApiKeyID = apiKey.ID
	apiKeyClaims.ClaimID = existingClaim.ID
	result = s.GetDB().Create(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_create_api_key_claim", "failed to create API key claim", "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) RemoveClaimFromApiKey(ctx *appctx.AppContext, tenantID string, claimID string, apiKeyID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_remove_claim_from_api_key")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return diag
	}

	var apiKey entities.ApiKey
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, apiKeyID).First(&apiKey)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_key", "failed to get API key", "claim_data_store", nil)
		return diag
	}
	if apiKey.ID == "" {
		diag.AddError("api_key_not_found", "API key not found", "claim_data_store", nil)
	}

	// check if the claim exists
	existingClaim, err := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if err != nil {
		diag.AddError("failed_to_get_claim", "failed to get claim", "claim_data_store", nil)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is assigned to the api key
	var apiKeyClaims entities.ApiKeyClaims
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ?", apiKey.ID, existingClaim.ID).First(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_get_api_key_claim", "failed to get API key claim", "claim_data_store", nil)
		return diag
	}
	if apiKeyClaims.ClaimID == "" {
		diag.AddError("claim_not_assigned_to_api_key", "claim not assigned to API key", "claim_data_store", nil)
		return diag
	}

	// delete the api key claim
	result = s.GetDB().Where("api_key_id = ? AND claim_id = ?", apiKey.ID, existingClaim.ID).Delete(&apiKeyClaims)
	if result.Error != nil {
		diag.AddError("failed_to_delete_api_key_claim", "failed to delete API key claim", "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) GetClaimRolesByQuery(ctx *appctx.AppContext, tenantID string, claimID string, queryBuilder *filters.QueryBuilder) (*filters.QueryBuilderResponse[entities.Role], *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_claim_roles_by_query")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}

	claim, getDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if getDiag.HasErrors() {
		diag.Append(getDiag)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}

	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN role_claims ON roles.id = role_claims.role_id").
		Where("role_claims.claim_id = ?", claim.ID)
	if queryBuilder == nil {
		queryBuilder = filters.NewQueryBuilder("")
	}

	result, err := utils.QueryDatabase[entities.Role](db, tenantID, queryBuilder)
	if err != nil {
		diag.AddError("failed_to_get_roles", "failed to get roles", "claim_data_store", nil)
		return nil, diag
	}

	return result, diag
}

func (s *ClaimDataStore) GetClaimRoles(ctx *appctx.AppContext, tenantID string, claimID string) ([]entities.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("store_get_claim_roles")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return nil, diag
	}

	claim, getDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if getDiag.HasErrors() {
		diag.Append(getDiag)
		return nil, diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return nil, diag
	}

	var roles []entities.Role
	db := s.GetDB().
		Preload("Claims", func(db *gorm.DB) *gorm.DB {
			return db.Order("claims.created_at DESC")
		}).
		Joins("JOIN role_claims ON roles.id = role_claims.role_id").
		Where("role_claims.claim_id = ?", claim.ID)

	err := db.Find(&roles).Error
	if err != nil {
		diag.AddError("failed_to_get_roles", "failed to get roles", "claim_data_store", nil)
		return nil, diag
	}

	return roles, diag
}

func (s *ClaimDataStore) AddClaimToRole(ctx *appctx.AppContext, tenantID string, claimID string, roleID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_add_claim_to_role")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return diag
	}

	var role entities.Role
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, roleID).First(&role)
	if result.Error != nil {
		diag.AddError("failed_to_get_role", "failed to get role", "claim_data_store", nil)
		return diag
	}
	if role.ID == "" {
		diag.AddError("role_not_found", "role not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim exists
	existingClaim, existingClaimDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if existingClaimDiag.HasErrors() {
		diag.Append(existingClaimDiag)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is assigned to the role
	var roleClaims entities.RoleClaims
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, existingClaim.ID).First(&roleClaims)
	if result.Error != nil {
		if !errors.Is(result.Error, gorm.ErrRecordNotFound) {
			diag.AddError("failed_to_get_role_claim", "failed to get role claim", "claim_data_store", nil)
			return diag
		}
	}
	if roleClaims.ClaimID != "" {
		diag.AddError("claim_already_assigned_to_role", "claim already assigned to role", "claim_data_store", nil)
		return diag
	}

	// create the role claim

	roleClaims.RoleID = role.ID
	roleClaims.ClaimID = existingClaim.ID
	result = s.GetDB().Create(&roleClaims)
	if result.Error != nil {
		diag.AddError("failed_to_create_role_claim", "failed to create role claim", "claim_data_store", nil)
		return diag
	}

	return diag
}

func (s *ClaimDataStore) RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, claimID string, roleID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("store_remove_claim_from_role")
	if tenantID == "" {
		diag.AddError("tenant_id_cannot_be_empty", "tenant ID cannot be empty", "claim_data_store")
		return diag
	}

	var role entities.Role
	result := s.GetDB().Where("tenant_id = ? AND id = ?", tenantID, roleID).First(&role)
	if result.Error != nil {
		diag.AddError("failed_to_get_role", "failed to get role", "claim_data_store", nil)
		return diag
	}
	if role.ID == "" {
		diag.AddError("role_not_found", "role not found", "claim_data_store", nil)
	}

	// check if the claim exists
	existingClaim, existingClaimDiag := s.GetClaimBySlugOrID(ctx, tenantID, claimID)
	if existingClaimDiag.HasErrors() {
		diag.Append(existingClaimDiag)
		return diag
	}
	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim_data_store", nil)
		return diag
	}

	// check if the claim is assigned to the role
	var roleClaims entities.RoleClaims
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, existingClaim.ID).First(&roleClaims)
	if result.Error != nil {
		diag.AddError("failed_to_get_role_claim", "failed to get role claim", "claim_data_store", nil)
		return diag
	}
	if roleClaims.ClaimID == "" {
		diag.AddError("claim_not_assigned_to_role", "claim not assigned to role", "claim_data_store", nil)
	}

	// delete the role claim
	result = s.GetDB().Where("role_id = ? AND claim_id = ?", role.ID, existingClaim.ID).Delete(&roleClaims)
	if result.Error != nil {
		diag.AddError("failed_to_delete_role_claim", "failed to delete role claim", "claim_data_store", nil)
		return diag
	}

	return diag
}
