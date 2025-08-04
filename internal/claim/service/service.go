// Package service provides the claim service implementation.
package service

import (
	"sync"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/claim/interfaces"
	"github.com/cjlapao/locally-cli/internal/claim/models"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

var (
	globalClaimService *ClaimService
	claimServiceOnce   sync.Once
	claimServiceMutex  sync.Mutex
)

type ClaimService struct {
	claimStore stores.ClaimDataStoreInterface
}

func Initialize(claimStore stores.ClaimDataStoreInterface) interfaces.ClaimServiceInterface {
	claimServiceMutex.Lock()
	defer claimServiceMutex.Unlock()

	claimServiceOnce.Do(func() {
		globalClaimService = new(claimStore)
	})
	return globalClaimService
}

func GetInstance() interfaces.ClaimServiceInterface {
	if globalClaimService == nil {
		panic("claim service not initialized")
	}
	return globalClaimService
}

// Reset resets the singleton for testing purposes
func Reset() {
	claimServiceMutex.Lock()
	defer claimServiceMutex.Unlock()
	globalClaimService = nil
	claimServiceOnce = sync.Once{}
}

func new(claimStore stores.ClaimDataStoreInterface) *ClaimService {
	return &ClaimService{
		claimStore: claimStore,
	}
}

func (s *ClaimService) GetName() string {
	return "claim"
}

func (s *ClaimService) GetClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_claims")
	defer diag.Complete()

	dbClaims, err := s.claimStore.GetClaims(ctx, tenantID)
	if err != nil {
		diag.AddError("failed_to_get_claims", "failed to get claims", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *ClaimService) GetClaimsByFilter(ctx *appctx.AppContext, tenantID string, filter *filters.Filter) (*api.PaginatedResponse[pkg_models.Claim], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_claims_by_filter")
	defer diag.Complete()

	dbClaims, err := s.claimStore.GetClaimsByFilter(ctx, tenantID, filter)
	if err != nil {
		diag.AddError("failed_to_get_claims_by_filter", "failed to get claims by filter", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims.Items)
	pagination := api.Pagination{
		Page:       dbClaims.Page,
		PageSize:   dbClaims.PageSize,
		TotalPages: dbClaims.TotalPages,
	}

	response := api.PaginatedResponse[pkg_models.Claim]{
		Data:       claims,
		TotalCount: int(dbClaims.Total),
		Pagination: pagination,
	}

	return &response, diag
}

func (s *ClaimService) GetClaimByIDOrSlug(ctx *appctx.AppContext, tenantID string, idOrSlug string) (*pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_claim_by_id_or_slug")
	defer diag.Complete()

	claim, err := s.claimStore.GetClaimBySlugOrID(ctx, tenantID, idOrSlug)
	if err != nil {
		diag.AddError("failed_to_get_claim_by_id", "failed to get claim by id", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}
	if claim == nil {
		return nil, diag
	}

	return mappers.MapClaimToDto(claim), diag
}

func (s *ClaimService) CreateClaim(ctx *appctx.AppContext, tenantID string, claim *models.CreateClaimRequest) (*pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_claim")
	defer diag.Complete()

	claimEntity := MapCreateClaimRequestToEntity(claim)
	claimEntity, err := s.claimStore.CreateClaim(ctx, tenantID, claimEntity)
	if err != nil {
		diag.AddError("failed_to_create_claim", "failed to create claim", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	return mappers.MapClaimToDto(claimEntity), diag
}

func (s *ClaimService) UpdateClaim(ctx *appctx.AppContext, tenantID string, request *models.UpdateClaimRequest) (string, *diagnostics.Diagnostics) {
	diag := diagnostics.New("update_claim")
	defer diag.Complete()

	claimEntity := MapUpdateClaimRequestToEntity(request)
	err := s.claimStore.UpdateClaim(ctx, tenantID, claimEntity)
	if err != nil {
		diag.AddError("failed_to_update_claim", "failed to update claim", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return "", diag
	}

	return claimEntity.Slug, diag
}

func (s *ClaimService) DeleteClaim(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_claim")
	defer diag.Complete()

	// checking if the claim is used by any user
	users, err := s.claimStore.GetClaimUsers(ctx, tenantID, id)
	if err != nil {
		diag.AddError("failed_to_check_claim_usage", "failed to check claim usage", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	if len(users) > 0 {
		diag.AddError("claim_in_use", "claim is in use", "claim", map[string]interface{}{
			"claim_id": id,
			"users":    users,
		})
		return diag
	}

	err = s.claimStore.DeleteClaim(ctx, tenantID, id)
	if err != nil {
		diag.AddError("failed_to_delete_claim", "failed to delete claim", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}

func (s *ClaimService) GetClaimUsers(ctx *appctx.AppContext, tenantID string, claimID string, pagination *pkg_models.Pagination) (*api.PaginatedResponse[pkg_models.User], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_claim_users")
	defer diag.Complete()
	cfg := config.GetInstance().Get()

	if pagination == nil {
		pagination = &pkg_models.Pagination{
			Page:     1,
			PageSize: cfg.GetInt(config.PaginationDefaultPageSizeKey, config.DefaultPageSizeInt),
		}
	}
	dbPagination := mappers.MapPaginationToEntity(pagination)

	dbUsers, err := s.claimStore.GetPaginatedClaimUsers(ctx, tenantID, claimID, dbPagination)
	if err != nil {
		diag.AddError("failed_to_get_claim_users", "failed to get claim users", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	response := api.PaginatedResponse[pkg_models.User]{
		Data:       mappers.MapUsersToDto(dbUsers.Items),
		TotalCount: int(dbUsers.Total),
		Pagination: api.Pagination{
			Page:       dbUsers.Page,
			PageSize:   dbUsers.PageSize,
			TotalPages: dbUsers.TotalPages,
		},
	}

	return &response, diag
}

func (s *ClaimService) AddClaimToUser(ctx *appctx.AppContext, tenantID string, userID string, claimSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_user_claim")
	defer diag.Complete()

	// check if the claim exists
	existingClaim, existingClaimDiag := s.GetClaimByIDOrSlug(ctx, tenantID, claimSlug)
	if existingClaimDiag.HasErrors() {
		diag.Append(existingClaimDiag)
		return diag
	}

	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim", map[string]interface{}{
			"claim_slug": claimSlug,
		})
		return diag
	}

	err := s.claimStore.AddClaimToUser(ctx, tenantID, userID, claimSlug)
	if err != nil {
		diag.AddError("failed_to_add_user_claim", "failed to add user claim", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}

func (s *ClaimService) RemoveClaimFromUser(ctx *appctx.AppContext, tenantID string, userID string, claimSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_user_claim")
	defer diag.Complete()

	// check if the claim exists
	existingClaim, existingClaimDiag := s.GetClaimByIDOrSlug(ctx, tenantID, claimSlug)
	if existingClaimDiag.HasErrors() {
		diag.Append(existingClaimDiag)
		return diag
	}

	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim", map[string]interface{}{
			"claim_slug": claimSlug,
		})
		return diag
	}

	err := s.claimStore.RemoveClaimFromUser(ctx, tenantID, userID, existingClaim.ID)
	if err != nil {
		diag.AddError("failed_to_remove_user_claim", "failed to remove user claim", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}

func (s *ClaimService) GetAllSuperUserLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_all_superuser_level_claims")
	defer diag.Complete()

	dbClaims, err := s.claimStore.GetClaimsByLevel(ctx, tenantID, pkg_models.SecurityLevelSuperUser)
	if err != nil {
		diag.AddError("failed_to_get_all_superuser_level_claims", "failed to get all superuser level claims", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *ClaimService) GetAllUserLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_all_user_level_claims")
	defer diag.Complete()

	dbClaims, err := s.claimStore.GetClaimsByLevel(ctx, tenantID, pkg_models.SecurityLevelUser)
	if err != nil {
		diag.AddError("failed_to_get_all_user_level_claims", "failed to get all user level claims", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *ClaimService) GetAllAdminLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_all_admin_level_claims")
	defer diag.Complete()

	dbClaims, err := s.claimStore.GetClaimsByLevel(ctx, tenantID, pkg_models.SecurityLevelAdmin)
	if err != nil {
		diag.AddError("failed_to_get_all_admin_level_claims", "failed to get all admin level claims", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *ClaimService) GetAllManagerLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_all_manager_level_claims")
	defer diag.Complete()

	dbClaims, err := s.claimStore.GetClaimsByLevel(ctx, tenantID, pkg_models.SecurityLevelManager)
	if err != nil {
		diag.AddError("failed_to_get_all_manager_level_claims", "failed to get all manager level claims", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *ClaimService) GetAllGuestLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_all_guest_level_claims")
	defer diag.Complete()

	dbClaims, err := s.claimStore.GetClaimsByLevel(ctx, tenantID, pkg_models.SecurityLevelGuest)
	if err != nil {
		diag.AddError("failed_to_get_all_guest_level_claims", "failed to get all guest level claims", "claim", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}
