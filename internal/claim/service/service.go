// Package service provides the claim service implementation.
package service

import (
	"sync"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/claim/interfaces"
	"github.com/cjlapao/locally-cli/internal/claim/models"
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

func (s *ClaimService) GetClaims(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Claim], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_claims")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	dbClaims, getClaimsDiag := s.claimStore.GetClaimsByQuery(ctx, tenantID, query)
	if getClaimsDiag.HasErrors() {
		diag.Append(getClaimsDiag)
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims.Items)

	return &api_models.PaginationResponse[pkg_models.Claim]{
		Data:       claims,
		TotalCount: dbClaims.Total,
		Pagination: api_models.Pagination{
			Page:       dbClaims.Page,
			PageSize:   dbClaims.PageSize,
			TotalPages: dbClaims.TotalPages,
		},
	}, diag
}

func (s *ClaimService) GetClaimByIDOrSlug(ctx *appctx.AppContext, tenantID string, idOrSlug string) (*pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_claim_by_id_or_slug")
	defer diag.Complete()

	claim, getClaimDiag := s.claimStore.GetClaimBySlugOrID(ctx, tenantID, idOrSlug)
	if getClaimDiag.HasErrors() {
		diag.Append(getClaimDiag)
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
	claimEntity, createClaimDiag := s.claimStore.CreateClaim(ctx, tenantID, claimEntity)
	if createClaimDiag.HasErrors() {
		diag.Append(createClaimDiag)
		return nil, diag
	}

	return mappers.MapClaimToDto(claimEntity), diag
}

func (s *ClaimService) UpdateClaim(ctx *appctx.AppContext, tenantID string, request *models.UpdateClaimRequest) (string, *diagnostics.Diagnostics) {
	diag := diagnostics.New("update_claim")
	defer diag.Complete()

	claimEntity := MapUpdateClaimRequestToEntity(request)
	updateClaimDiag := s.claimStore.UpdateClaim(ctx, tenantID, claimEntity)
	if updateClaimDiag.HasErrors() {
		diag.Append(updateClaimDiag)
		return "", diag
	}

	return claimEntity.Slug, diag
}

func (s *ClaimService) DeleteClaim(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_claim")
	defer diag.Complete()

	// checking if the claim is used by any user
	users, getClaimUsersDiag := s.claimStore.GetClaimUsers(ctx, tenantID, id)
	if getClaimUsersDiag.HasErrors() {
		diag.Append(getClaimUsersDiag)
		return diag
	}

	if len(users) > 0 {
		diag.AddError("claim_in_use", "claim is in use", "claim", map[string]interface{}{
			"claim_id": id,
			"users":    users,
		})
		return diag
	}

	deleteClaimDiag := s.claimStore.DeleteClaim(ctx, tenantID, id)
	if deleteClaimDiag.HasErrors() {
		diag.Append(deleteClaimDiag)
		return diag
	}

	return diag
}

func (s *ClaimService) GetClaimUsers(ctx *appctx.AppContext, tenantID string, claimID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.User], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_claim_users")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	dbUsers, getClaimUsersDiag := s.claimStore.GetClaimUsersByQuery(ctx, tenantID, claimID, query)
	if getClaimUsersDiag.HasErrors() {
		diag.Append(getClaimUsersDiag)
		return nil, diag
	}

	users := mappers.MapUsersToDto(dbUsers.Items)

	response := &api_models.PaginationResponse[pkg_models.User]{
		Data:       users,
		TotalCount: dbUsers.Total,
		Pagination: api_models.Pagination{
			Page:       dbUsers.Page,
			PageSize:   dbUsers.PageSize,
			TotalPages: dbUsers.TotalPages,
		},
	}

	return response, diag
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

	addClaimToUserDiag := s.claimStore.AddClaimToUser(ctx, tenantID, userID, existingClaim.ID)
	if addClaimToUserDiag.HasErrors() {
		diag.Append(addClaimToUserDiag)
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

	removeClaimFromUserDiag := s.claimStore.RemoveClaimFromUser(ctx, tenantID, userID, existingClaim.ID)
	if removeClaimFromUserDiag.HasErrors() {
		diag.Append(removeClaimFromUserDiag)
		return diag
	}

	return diag
}

func (s *ClaimService) GetAllSuperUserLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_all_superuser_level_claims")
	defer diag.Complete()

	dbClaims, getClaimsDiag := s.claimStore.GetClaimsByLevel(ctx, tenantID, pkg_models.SecurityLevelSuperUser)
	if getClaimsDiag.HasErrors() {
		diag.Append(getClaimsDiag)
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *ClaimService) GetAllUserLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_all_user_level_claims")
	defer diag.Complete()

	dbClaims, getClaimsDiag := s.claimStore.GetClaimsByLevel(ctx, tenantID, pkg_models.SecurityLevelUser)
	if getClaimsDiag.HasErrors() {
		diag.Append(getClaimsDiag)
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *ClaimService) GetAllAdminLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_all_admin_level_claims")
	defer diag.Complete()

	dbClaims, getClaimsDiag := s.claimStore.GetClaimsByLevel(ctx, tenantID, pkg_models.SecurityLevelAdmin)
	if getClaimsDiag.HasErrors() {
		diag.Append(getClaimsDiag)
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *ClaimService) GetAllManagerLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_all_manager_level_claims")
	defer diag.Complete()

	dbClaims, getClaimsDiag := s.claimStore.GetClaimsByLevel(ctx, tenantID, pkg_models.SecurityLevelManager)
	if getClaimsDiag.HasErrors() {
		diag.Append(getClaimsDiag)
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *ClaimService) GetAllGuestLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_all_guest_level_claims")
	defer diag.Complete()

	dbClaims, getClaimsDiag := s.claimStore.GetClaimsByLevel(ctx, tenantID, pkg_models.SecurityLevelGuest)
	if getClaimsDiag.HasErrors() {
		diag.Append(getClaimsDiag)
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *ClaimService) GetClaimRoles(ctx *appctx.AppContext, tenantID string, claimID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Role], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_claim_roles")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	dbRoles, getClaimRolesDiag := s.claimStore.GetClaimRolesByQuery(ctx, tenantID, claimID, query)
	if getClaimRolesDiag.HasErrors() {
		diag.Append(getClaimRolesDiag)
		return nil, diag
	}

	roles := mappers.MapRolesToDto(dbRoles.Items)

	response := &api_models.PaginationResponse[pkg_models.Role]{
		Data:       roles,
		TotalCount: dbRoles.Total,
		Pagination: api_models.Pagination{
			Page:       dbRoles.Page,
			PageSize:   dbRoles.PageSize,
			TotalPages: dbRoles.TotalPages,
		},
	}

	return response, diag
}

func (s *ClaimService) AddRoleToClaim(ctx *appctx.AppContext, tenantID string, claimID string, roleID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_role_to_claim")
	defer diag.Complete()

	// check if the claim exists
	existingClaim, existingClaimDiag := s.GetClaimByIDOrSlug(ctx, tenantID, claimID)
	if existingClaimDiag.HasErrors() {
		diag.Append(existingClaimDiag)
		return diag
	}

	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}

	addClaimToRoleDiag := s.claimStore.AddClaimToRole(ctx, tenantID, claimID, roleID)
	if addClaimToRoleDiag.HasErrors() {
		diag.Append(addClaimToRoleDiag)
		return diag
	}

	return diag
}

func (s *ClaimService) RemoveRoleFromClaim(ctx *appctx.AppContext, tenantID string, claimID string, roleID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_role_from_claim")
	defer diag.Complete()

	// check if the claim exists
	existingClaim, existingClaimDiag := s.GetClaimByIDOrSlug(ctx, tenantID, claimID)
	if existingClaimDiag.HasErrors() {
		diag.Append(existingClaimDiag)
		return diag
	}

	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}

	removeClaimFromRoleDiag := s.claimStore.RemoveClaimFromRole(ctx, tenantID, claimID, roleID)
	if removeClaimFromRoleDiag.HasErrors() {
		diag.Append(removeClaimFromRoleDiag)
		return diag
	}

	return diag
}

func (s *ClaimService) GetClaimApiKeys(ctx *appctx.AppContext, tenantID string, claimID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.ApiKey], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_claim_api_keys")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	dbApiKeys, getClaimApiKeysDiag := s.claimStore.GetClaimApiKeysByQuery(ctx, tenantID, claimID, query)
	if getClaimApiKeysDiag.HasErrors() {
		diag.Append(getClaimApiKeysDiag)
		return nil, diag
	}

	apiKeys := mappers.MapApiKeysToDto(dbApiKeys.Items)

	response := &api_models.PaginationResponse[pkg_models.ApiKey]{
		Data:       apiKeys,
		TotalCount: dbApiKeys.Total,
		Pagination: api_models.Pagination{
			Page:       dbApiKeys.Page,
			PageSize:   dbApiKeys.PageSize,
			TotalPages: dbApiKeys.TotalPages,
		},
	}

	return response, diag
}

func (s *ClaimService) AddApiKeyToClaim(ctx *appctx.AppContext, tenantID string, claimID string, apiKeySlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_api_key_to_claim")
	defer diag.Complete()

	// check if the claim exists
	existingClaim, existingClaimDiag := s.GetClaimByIDOrSlug(ctx, tenantID, claimID)
	if existingClaimDiag.HasErrors() {
		diag.Append(existingClaimDiag)
		return diag
	}

	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}

	addClaimToApiKeyDiag := s.claimStore.AddClaimToApiKey(ctx, tenantID, claimID, apiKeySlug)
	if addClaimToApiKeyDiag.HasErrors() {
		diag.Append(addClaimToApiKeyDiag)
		return diag
	}

	return diag
}

func (s *ClaimService) RemoveApiKeyFromClaim(ctx *appctx.AppContext, tenantID string, claimID string, apiKeySlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_api_key_from_claim")
	defer diag.Complete()

	// check if the claim exists
	existingClaim, existingClaimDiag := s.GetClaimByIDOrSlug(ctx, tenantID, claimID)
	if existingClaimDiag.HasErrors() {
		diag.Append(existingClaimDiag)
		return diag
	}

	if existingClaim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim", map[string]interface{}{
			"claim_id": claimID,
		})
		return diag
	}

	removeClaimFromApiKeyDiag := s.claimStore.RemoveClaimFromApiKey(ctx, tenantID, claimID, apiKeySlug)
	if removeClaimFromApiKeyDiag.HasErrors() {
		diag.Append(removeClaimFromApiKeyDiag)
		return diag
	}

	return diag
}
