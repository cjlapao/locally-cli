// Package user provides a service for managing users.
package user

import (
	"sync"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
)

var (
	globalUserService *UserService
	userServiceOnce   sync.Once
	userServiceMutex  sync.Mutex
)

type UserServiceInterface interface {
	GetName() string
	GetUsersByFilter(ctx *appctx.AppContext, tenantID string, filter *filters.Filter) (*api.PaginatedResponse[models.User], *diagnostics.Diagnostics)
	GetUserByID(ctx *appctx.AppContext, tenantID string, id string) (*models.User, *diagnostics.Diagnostics)
	GetUserByUsername(ctx *appctx.AppContext, tenantID string, username string) (*models.User, *diagnostics.Diagnostics)
	CreateUser(ctx *appctx.AppContext, tenantID string, role string, user *CreateUserRequest) (*CreateUserResponse, *diagnostics.Diagnostics)
	UpdateUser(ctx *appctx.AppContext, tenantID string, userId string, user *UpdateUserRequest) (*UpdateUserResponse, *diagnostics.Diagnostics)
	DeleteUser(ctx *appctx.AppContext, tenantID string, userId string) *diagnostics.Diagnostics
	UpdateUserPassword(ctx *appctx.AppContext, tenantID string, id string, request *UpdateUserPasswordRequest) *diagnostics.Diagnostics
}

type UserService struct {
	userStore stores.UserDataStoreInterface
}

func Initialize(userStore stores.UserDataStoreInterface) UserServiceInterface {
	userServiceMutex.Lock()
	defer userServiceMutex.Unlock()

	userServiceOnce.Do(func() {
		globalUserService = new(userStore)
	})
	return globalUserService
}

func GetInstance() UserServiceInterface {
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

func new(userStore stores.UserDataStoreInterface) *UserService {
	return &UserService{
		userStore: userStore,
	}
}

func (s *UserService) GetName() string {
	return "user"
}

func (s *UserService) GetUsersByFilter(ctx *appctx.AppContext, tenantID string, filter *filters.Filter) (*api.PaginatedResponse[models.User], *diagnostics.Diagnostics) {
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
	pagination := api.Pagination{
		Page:       dbUsers.Page,
		PageSize:   dbUsers.PageSize,
		TotalPages: dbUsers.TotalPages,
	}

	response := api.PaginatedResponse[models.User]{
		Data:       users,
		TotalCount: int(dbUsers.Total),
		Pagination: pagination,
	}

	return &response, diag
}

func (s *UserService) GetUserByID(ctx *appctx.AppContext, tenantID string, id string) (*models.User, *diagnostics.Diagnostics) {
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

func (s *UserService) GetUserByUsername(ctx *appctx.AppContext, tenantID string, username string) (*models.User, *diagnostics.Diagnostics) {
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

func (s *UserService) CreateUser(ctx *appctx.AppContext, tenantID string, role string, request *CreateUserRequest) (*CreateUserResponse, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_user")
	defer diag.Complete()

	userEntity := MapCreateUserRequestToEntity(request)

	userEntity, err := s.userStore.CreateUser(ctx, tenantID, userEntity)
	if err != nil {
		diag.AddError("failed_to_create_user", "failed to create user", "user", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	result := &CreateUserResponse{
		ID:     userEntity.ID,
		Name:   userEntity.Name,
		Status: userEntity.Status,
	}

	return result, diag
}

func (s *UserService) UpdateUser(ctx *appctx.AppContext, tenantID string, userId string, request *UpdateUserRequest) (*UpdateUserResponse, *diagnostics.Diagnostics) {
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

	result := &UpdateUserResponse{
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

func (s *UserService) UpdateUserPassword(ctx *appctx.AppContext, tenantID string, id string, request *UpdateUserPasswordRequest) *diagnostics.Diagnostics {
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
