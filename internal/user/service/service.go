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

func (s *UserService) GetUsersByFilter(ctx *appctx.AppContext, tenantID string, filter *filters.Filter) (*api_models.PaginatedResponse[pkg_models.User], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_users")
	defer diag.Complete()

	dbUsers, err := s.userStore.GetUsersByFilter(ctx, tenantID, filter)
	if err != nil {
		diag.AddError("failed_to_get_users", "failed to get users", "user", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	users := mappers.MapUsersToDto(dbUsers.Items)
	pagination := api_models.Pagination{
		Page:       dbUsers.Page,
		PageSize:   dbUsers.PageSize,
		TotalPages: dbUsers.TotalPages,
	}

	response := api_models.PaginatedResponse[pkg_models.User]{
		Data:       users,
		TotalCount: dbUsers.Total,
		Pagination: pagination,
	}

	return &response, diag
}

func (s *UserService) GetUserByID(ctx *appctx.AppContext, tenantID string, id string) (*pkg_models.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_by_id")
	defer diag.Complete()

	user, err := s.userStore.GetUserByID(ctx, tenantID, id)
	if err != nil {
		diag.AddError("failed_to_get_user_by_id", "failed to get user by id", "user", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	return mappers.MapUserToDto(user), nil
}

func (s *UserService) GetUserByUsername(ctx *appctx.AppContext, tenantID string, username string) (*pkg_models.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_by_username")
	defer diag.Complete()

	user, err := s.userStore.GetUserByUsername(ctx, tenantID, username)
	if err != nil {
		diag.AddError("failed_to_get_user_by_username", "failed to get user by username", "user", map[string]interface{}{
			"error": err.Error(),
		})
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
	existingUser, err := s.userStore.GetUserByUsername(ctx, tenantID, userEntity.Username)
	if err != nil {
		diag.AddError("failed_to_get_user", "failed to get user", "user", map[string]interface{}{
			"error": err.Error(),
		})
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

	userEntity, err = s.userStore.CreateUser(ctx, tenantID, userEntity)
	if err != nil {
		diag.AddError("failed_to_create_user", "failed to create user", "user", map[string]interface{}{
			"error": err.Error(),
		})
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
	err := s.userStore.UpdateUser(ctx, tenantID, userEntity)
	if err != nil {
		diag.AddError("failed_to_update_user", "failed to update user", "user", map[string]interface{}{
			"error": err.Error(),
		})
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

	err := s.userStore.DeleteUser(ctx, tenantID, id)
	if err != nil {
		diag.AddError("failed_to_delete_user", "failed to delete user", "user", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}

func (s *UserService) UpdateUserPassword(ctx *appctx.AppContext, tenantID string, id string, request *models.UpdateUserPasswordRequest) *diagnostics.Diagnostics {
	diag := diagnostics.New("update_user_password")
	defer diag.Complete()

	err := s.userStore.UpdateUserPassword(ctx, tenantID, id, request.Password)
	if err != nil {
		diag.AddError("failed_to_update_user_password", "failed to update user password", "user", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}

func (s *UserService) GetUserClaims(ctx *appctx.AppContext, tenantID string, userID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_claims")
	defer diag.Complete()

	claims, err := s.userStore.GetUserClaims(ctx, tenantID, userID)
	if err != nil {
		diag.AddError("failed_to_get_user_claims", "failed to get user claims", "user", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	return mappers.MapClaimsToDto(claims), diag
}
