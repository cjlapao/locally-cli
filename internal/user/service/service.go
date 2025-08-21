// Package service provides a service for managing users.
package service

import (
	"sync"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	claimsvc "github.com/cjlapao/locally-cli/internal/claim/interfaces"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	rolesvc "github.com/cjlapao/locally-cli/internal/role/interfaces"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
	"github.com/cjlapao/locally-cli/internal/user/interfaces"
	"github.com/cjlapao/locally-cli/internal/user/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

var (
	globalUserService *UserService
	userServiceOnce   sync.Once
	userServiceMutex  sync.Mutex
)

type UserService struct {
	userStore     stores.UserDataStoreInterface
	claimService  claimsvc.ClaimServiceInterface
	roleService   rolesvc.RoleServiceInterface
	systemService system_interfaces.SystemServiceInterface
}

func Initialize(userStore stores.UserDataStoreInterface, roleService rolesvc.RoleServiceInterface, claimService claimsvc.ClaimServiceInterface, systemService system_interfaces.SystemServiceInterface) interfaces.UserServiceInterface {
	userServiceMutex.Lock()
	defer userServiceMutex.Unlock()

	userServiceOnce.Do(func() {
		globalUserService = new(userStore, roleService, claimService, systemService)
	})
	return globalUserService
}

func GetInstance() interfaces.UserServiceInterface {
	if globalUserService == nil {
		panic("user service not initialized")
	}
	return globalUserService
}

// Reset resets the singleton for testing purposes
func Reset() {
	userServiceMutex.Lock()
	defer userServiceMutex.Unlock()
	globalUserService = nil
	userServiceOnce = sync.Once{}
}

func new(userStore stores.UserDataStoreInterface, roleService rolesvc.RoleServiceInterface, claimService claimsvc.ClaimServiceInterface, systemService system_interfaces.SystemServiceInterface) *UserService {
	return &UserService{
		userStore:     userStore,
		roleService:   roleService,
		claimService:  claimService,
		systemService: systemService,
	}
}

func (s *UserService) GetName() string {
	return "user"
}

func (s *UserService) GetUsers(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.User], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_users")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	dbUsers, getUsersDiag := s.userStore.GetUsersByQuery(ctx, tenantID, query)
	if getUsersDiag.HasErrors() {
		diag.Append(getUsersDiag)
		return nil, diag
	}

	users := mappers.MapUsersToDto(dbUsers.Items)

	response := api_models.PaginationResponse[pkg_models.User]{
		Data:       users,
		TotalCount: dbUsers.Total,
		Pagination: api_models.Pagination{
			Page:       dbUsers.Page,
			PageSize:   dbUsers.PageSize,
			TotalPages: dbUsers.TotalPages,
		},
	}

	return &response, diag
}

func (s *UserService) GetUserByID(ctx *appctx.AppContext, tenantID string, id string) (*pkg_models.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_by_id")
	defer diag.Complete()

	user, getUserDiag := s.userStore.GetUserByID(ctx, tenantID, id)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return nil, diag
	}

	return mappers.MapUserToDto(user), nil
}

func (s *UserService) GetUserByUsername(ctx *appctx.AppContext, tenantID string, username string) (*pkg_models.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_by_username")
	defer diag.Complete()

	user, getUserDiag := s.userStore.GetUserByUsername(ctx, tenantID, username)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return nil, diag
	}

	return mappers.MapUserToDto(user), diag
}

func (s *UserService) CreateUser(ctx *appctx.AppContext, tenantID string, role string, request *models.CreateUserRequest) (*models.CreateUserResponse, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_user")
	defer diag.Complete()

	userEntity := MapCreateUserRequestToEntity(request)

	// checking if role exists
	roleEntity, roleDiag := s.roleService.GetRoleByIDorSlugWithClaims(ctx, tenantID, role)
	if roleDiag.HasErrors() {
		diag.Append(roleDiag)
		return nil, diag
	}
	if roleEntity == nil {
		diag.AddError("role_not_found", "role not found", "user", map[string]interface{}{
			"role": role,
		})
		return nil, diag
	}

	roleModel := mappers.MapRoleToEntity(roleEntity)
	// checking if the user already exists
	existingUser, getUserDiag := s.userStore.GetUserByUsername(ctx, tenantID, userEntity.Username)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return nil, diag
	}
	if existingUser != nil {
		diag.AddError("user_already_exists", "user already exists", "user", map[string]interface{}{
			"username": userEntity.Username,
		})
		return nil, diag
	}

	// now we going to add the claims that are associated with the role
	claims := make([]entities.Claim, len(roleModel.Claims))
	for i, claim := range roleModel.Claims {
		dbClaim, dbClaimDiag := s.claimService.GetClaimByIDOrSlug(ctx, tenantID, claim.GetSlug())
		if dbClaimDiag.HasErrors() {
			diag.Append(dbClaimDiag)
			return nil, diag
		}
		claims[i] = *mappers.MapClaimToEntity(dbClaim)
	}

	userEntity.Roles = []entities.Role{*roleModel}
	userEntity.Claims = claims

	userEntity, createUserDiag := s.userStore.CreateUser(ctx, tenantID, userEntity)
	if createUserDiag.HasErrors() {
		diag.Append(createUserDiag)
		return nil, diag
	}

	result := &models.CreateUserResponse{
		ID:     userEntity.ID,
		Name:   userEntity.Name,
		Status: userEntity.Status,
	}

	return result, diag
}

func (s *UserService) UpdateUser(ctx *appctx.AppContext, tenantID string, userId string, request *models.UpdateUserRequest) (*models.UpdateUserResponse, *diagnostics.Diagnostics) {
	diag := diagnostics.New("update_user")
	defer diag.Complete()

	userEntity := MapUpdateUserRequestToEntity(request)
	userEntity.ID = userId
	updateUserDiag := s.userStore.UpdateUser(ctx, tenantID, userEntity)
	if updateUserDiag.HasErrors() {
		diag.Append(updateUserDiag)
		return nil, diag
	}

	result := &models.UpdateUserResponse{
		ID:     userEntity.ID,
		Name:   userEntity.Name,
		Status: userEntity.Status,
	}

	return result, diag
}

func (s *UserService) DeleteUser(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_user")
	defer diag.Complete()

	deleteUserDiag := s.userStore.DeleteUser(ctx, tenantID, id)
	if deleteUserDiag.HasErrors() {
		diag.Append(deleteUserDiag)
		return diag
	}

	return diag
}

func (s *UserService) UpdateUserPassword(ctx *appctx.AppContext, tenantID string, id string, request *models.UpdateUserPasswordRequest) *diagnostics.Diagnostics {
	diag := diagnostics.New("update_user_password")
	defer diag.Complete()

	updateUserPasswordDiag := s.userStore.UpdateUserPassword(ctx, tenantID, id, request.Password)
	if updateUserPasswordDiag.HasErrors() {
		diag.Append(updateUserPasswordDiag)
		return diag
	}

	return diag
}

func (s *UserService) GetUserClaims(ctx *appctx.AppContext, tenantID string, userID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Claim], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_claims")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	claimsQuery, getUserClaimsDiag := s.userStore.GetUserClaimsByQuery(ctx, tenantID, userID, query)
	if getUserClaimsDiag.HasErrors() {
		diag.Append(getUserClaimsDiag)
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(claimsQuery.Items)

	result := api_models.PaginationResponse[pkg_models.Claim]{
		Data:       claims,
		TotalCount: claimsQuery.Total,
		Pagination: api_models.Pagination{
			Page:       claimsQuery.Page,
			PageSize:   claimsQuery.PageSize,
			TotalPages: claimsQuery.TotalPages,
		},
	}

	return &result, diag
}

func (s *UserService) AddClaimToUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_claim_to_user")
	defer diag.Complete()

	user, getUserDiag := s.userStore.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}

	claim, claimDiag := s.claimService.GetClaimByIDOrSlug(ctx, tenantID, claimIdOrSlug)
	if claimDiag.HasErrors() {
		diag.Append(claimDiag)
		return diag
	}
	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim", map[string]interface{}{
			"claim_id": claimIdOrSlug,
		})
		return diag
	}

	addClaimDiag := s.claimService.AddClaimToUser(ctx, tenantID, userID, claimIdOrSlug)
	if addClaimDiag.HasErrors() {
		diag.Append(addClaimDiag)
		return diag
	}

	return diag
}

func (s *UserService) RemoveClaimFromUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_claim_from_user")
	defer diag.Complete()

	user, getUserDiag := s.userStore.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}

	claim, claimDiag := s.claimService.GetClaimByIDOrSlug(ctx, tenantID, claimIdOrSlug)
	if claimDiag.HasErrors() {
		diag.Append(claimDiag)
		return diag
	}

	if claim == nil {
		diag.AddError("claim_not_found", "claim not found", "claim", map[string]interface{}{
			"claim_id": claimIdOrSlug,
		})
		return diag
	}

	removeClaimDiag := s.claimService.RemoveClaimFromUser(ctx, tenantID, userID, claimIdOrSlug)
	if removeClaimDiag.HasErrors() {
		diag.Append(removeClaimDiag)
		return diag
	}

	return diag
}

func (s *UserService) GetUserRoles(ctx *appctx.AppContext, tenantID string, userID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Role], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_roles")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	rolesQuery, getUserRolesDiag := s.userStore.GetUserRolesByQuery(ctx, tenantID, userID, query)
	if getUserRolesDiag.HasErrors() {
		diag.Append(getUserRolesDiag)
		return nil, diag
	}

	roles := mappers.MapRolesToDto(rolesQuery.Items)

	result := api_models.PaginationResponse[pkg_models.Role]{
		Data:       roles,
		TotalCount: rolesQuery.Total,
		Pagination: api_models.Pagination{
			Page:       rolesQuery.Page,
			PageSize:   rolesQuery.PageSize,
			TotalPages: rolesQuery.TotalPages,
		},
	}

	return &result, diag
}

func (s *UserService) AddRoleToUser(ctx *appctx.AppContext, tenantID string, userID string, roleIdOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_role_to_user")
	defer diag.Complete()

	user, getUserDiag := s.userStore.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}

	role, roleDiag := s.roleService.GetRoleByIDorSlugWithClaims(ctx, tenantID, roleIdOrSlug)
	if roleDiag.HasErrors() {
		diag.Append(roleDiag)
		return diag
	}

	if role == nil {
		diag.AddError("role_not_found", "role not found", "role", map[string]interface{}{
			"role_id": roleIdOrSlug,
		})
		return diag
	}

	addRoleDiag := s.userStore.AddUserToRole(ctx, tenantID, userID, role.ID)
	if addRoleDiag.HasErrors() {
		diag.Append(addRoleDiag)
		return diag
	}

	return diag
}

func (s *UserService) RemoveRoleFromUser(ctx *appctx.AppContext, tenantID string, userID string, roleIdOrSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_role_from_user")
	defer diag.Complete()

	user, getUserDiag := s.userStore.GetUserByID(ctx, tenantID, userID)
	if getUserDiag.HasErrors() {
		diag.Append(getUserDiag)
		return diag
	}
	if user == nil {
		diag.AddError("user_not_found", "user not found", "user", map[string]interface{}{
			"user_id": userID,
		})
		return diag
	}

	role, roleDiag := s.roleService.GetRoleByIDorSlugWithClaims(ctx, tenantID, roleIdOrSlug)
	if roleDiag.HasErrors() {
		diag.Append(roleDiag)
		return diag
	}
	if role == nil {
		diag.AddError("role_not_found", "role not found", "role", map[string]interface{}{
			"role_id": roleIdOrSlug,
		})
		return diag
	}

	removeRoleDiag := s.userStore.RemoveUserFromRole(ctx, tenantID, userID, role.ID)
	if removeRoleDiag.HasErrors() {
		diag.Append(removeRoleDiag)
		return diag
	}

	return diag
}
