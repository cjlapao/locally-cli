// Package service provides a service for managing roles.
package service

import (
	"sync"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	claim_interfaces "github.com/cjlapao/locally-cli/internal/claim/interfaces"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/internal/role/interfaces"
	"github.com/cjlapao/locally-cli/internal/role/models"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

var (
	globalRoleService *RoleService
	roleServiceOnce   sync.Once
	roleServiceMutex  sync.Mutex
)

type RoleService struct {
	roleStore     stores.RoleDataStoreInterface
	claimService  claim_interfaces.ClaimServiceInterface
	systemService system_interfaces.SystemServiceInterface
}

func Initialize(roleStore stores.RoleDataStoreInterface, systemService system_interfaces.SystemServiceInterface, claimService claim_interfaces.ClaimServiceInterface) interfaces.RoleServiceInterface {
	roleServiceMutex.Lock()
	defer roleServiceMutex.Unlock()

	roleServiceOnce.Do(func() {
		globalRoleService = new(roleStore, systemService, claimService)
	})
	return globalRoleService
}

func GetInstance() interfaces.RoleServiceInterface {
	if globalRoleService == nil {
		panic("role service not initialized")
	}
	return globalRoleService
}

// Reset resets the singleton for testing purposes
func Reset() {
	roleServiceMutex.Lock()
	defer roleServiceMutex.Unlock()
	globalRoleService = nil
	roleServiceOnce = sync.Once{}
}

func new(roleStore stores.RoleDataStoreInterface, systemService system_interfaces.SystemServiceInterface, claimService claim_interfaces.ClaimServiceInterface) *RoleService {
	return &RoleService{
		roleStore:     roleStore,
		systemService: systemService,
		claimService:  claimService,
	}
}

func (s *RoleService) GetName() string {
	return "role"
}

func (s *RoleService) GetRoles(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_roles")
	defer diag.Complete()

	dbRoles, err := s.roleStore.GetRoles(ctx, tenantID)
	if err != nil {
		diag.AddError("failed_to_get_roles", "failed to get roles", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	roles := mappers.MapRolesToDto(dbRoles)

	return roles, diag
}

func (s *RoleService) GetRolesByFilter(ctx *appctx.AppContext, tenantID string, filter *filters.Filter) (*api_models.PaginatedResponse[pkg_models.Role], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_roles_by_filter")
	defer diag.Complete()

	dbRoles, err := s.roleStore.GetRolesByFilter(ctx, tenantID, filter)
	if err != nil {
		diag.AddError("failed_to_get_roles_by_filter", "failed to get roles by filter", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	roles := mappers.MapRolesToDto(dbRoles.Items)
	pagination := api_models.Pagination{
		Page:       dbRoles.Page,
		PageSize:   dbRoles.PageSize,
		TotalPages: dbRoles.TotalPages,
	}

	response := api_models.PaginatedResponse[pkg_models.Role]{
		Data:       roles,
		TotalCount: dbRoles.Total,
		Pagination: pagination,
	}

	return &response, diag
}

func (s *RoleService) GetRoleByIDorSlug(ctx *appctx.AppContext, tenantID string, idOrSlug string) (*pkg_models.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_role_by_id_or_slug")
	defer diag.Complete()

	role, err := s.roleStore.GetRoleBySlugOrID(ctx, tenantID, idOrSlug)
	if err != nil {
		diag.AddError("failed_to_get_role_by_id", "failed to get role by id", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	if role == nil {
		return nil, diag
	}

	return mappers.MapRoleToDto(role), diag
}

func (s *RoleService) GetRoleByIDorSlugWithClaims(ctx *appctx.AppContext, tenantID string, idOrSlug string) (*pkg_models.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_role_by_id_or_slug_with_claims")
	defer diag.Complete()

	role, err := s.roleStore.GetRoleBySlugOrID(ctx, tenantID, idOrSlug)
	if err != nil {
		diag.AddError("failed_to_get_role_by_id", "failed to get role by id", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	if role == nil {
		return nil, diag
	}

	return mappers.MapRoleToDtoWithClaims(role), diag
}

func (s *RoleService) CreateRole(ctx *appctx.AppContext, tenantID string, role *models.CreateRoleRequest) (*pkg_models.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_role")
	defer diag.Complete()

	roleEntity := MapCreateRoleRequestToEntity(role)
	roleSystemClaims := s.systemService.GenerateDefaultClaimsForSecurityLevel(roleEntity.SecurityLevel)
	dbClaims := make([]entities.Claim, len(roleSystemClaims))
	for i, claim := range roleSystemClaims {
		claimSlug := claim.GetSlug()
		claimDto, getDiag := s.claimService.GetClaimByIDOrSlug(ctx, tenantID, claimSlug)
		if getDiag.HasErrors() {
			diag.Append(getDiag)
			return nil, diag
		}
		dbClaims[i] = *mappers.MapClaimToEntity(claimDto)
	}

	roleEntity.Claims = dbClaims
	roleEntity, err := s.roleStore.CreateRole(ctx, tenantID, roleEntity)
	if err != nil {
		diag.AddError("failed_to_create_role", "failed to create role", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	return mappers.MapRoleToDto(roleEntity), diag
}

func (s *RoleService) UpdateRole(ctx *appctx.AppContext, tenantID string, role *models.UpdateRoleRequest) (string, *diagnostics.Diagnostics) {
	diag := diagnostics.New("update_role")
	defer diag.Complete()

	roleEntity := MapUpdateRoleRequestToEntity(role)
	err := s.roleStore.UpdateRole(ctx, tenantID, roleEntity)
	if err != nil {
		diag.AddError("failed_to_update_role", "failed to update role", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return "", diag
	}

	return roleEntity.Slug, diag
}

func (s *RoleService) DeleteRole(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_role")
	defer diag.Complete()

	// checking if the role is used by any user
	users, err := s.roleStore.GetRoleUsers(ctx, tenantID, id)
	if err != nil {
		diag.AddError("failed_to_check_role_usage", "failed to check role usage", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	if len(users) > 0 {
		diag.AddError("role_in_use", "role is in use", "role", map[string]interface{}{
			"role_id": id,
			"users":   users,
		})
		return diag
	}

	err = s.roleStore.DeleteRole(ctx, tenantID, id)
	if err != nil {
		diag.AddError("failed_to_delete_role", "failed to delete role", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}

func (s *RoleService) GetRoleUsers(ctx *appctx.AppContext, tenantID string, id string, pagination *pkg_models.Pagination) (*api_models.PaginatedResponse[pkg_models.User], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_role_users")
	defer diag.Complete()
	cfg := config.GetInstance().Get()

	if pagination == nil {
		pagination = &pkg_models.Pagination{
			Page:     1,
			PageSize: cfg.GetInt(config.PaginationDefaultPageSizeKey, config.DefaultPageSizeInt),
		}
	}

	dbPagination := mappers.MapPaginationToEntity(pagination)

	dbUsers, err := s.roleStore.GetPaginatedRoleUsers(ctx, tenantID, id, dbPagination)
	if err != nil {
		diag.AddError("failed_to_get_role_users", "failed to get role users", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	response := api_models.PaginatedResponse[pkg_models.User]{
		Data:       mappers.MapUsersToDto(dbUsers.Items),
		TotalCount: dbUsers.Total,
		Pagination: api_models.Pagination{
			Page:       dbUsers.Page,
			PageSize:   dbUsers.PageSize,
			TotalPages: dbUsers.TotalPages,
		},
	}

	return &response, diag
}

func (s *RoleService) AddUserToRole(ctx *appctx.AppContext, tenantID string, userID string, roleSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_user_role")
	defer diag.Complete()

	err := s.roleStore.AddUserToRole(ctx, tenantID, userID, roleSlug)
	if err != nil {
		diag.AddError("failed_to_add_user_role", "failed to add user role", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}

func (s *RoleService) RemoveUserFromRole(ctx *appctx.AppContext, tenantID string, userID string, roleSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_user_role")
	defer diag.Complete()

	err := s.roleStore.RemoveUserFromRole(ctx, tenantID, userID, roleSlug)
	if err != nil {
		diag.AddError("failed_to_remove_user_role", "failed to remove user role", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}

func (s *RoleService) GetUserRoles(ctx *appctx.AppContext, tenantID string, userID string) ([]pkg_models.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_roles")
	defer diag.Complete()

	dbRoles, err := s.roleStore.GetUserRoles(ctx, tenantID, userID)
	if err != nil {
		diag.AddError("failed_to_get_user_roles", "failed to get user roles", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	roles := mappers.MapRolesToDto(dbRoles)

	return roles, diag
}

func (s *RoleService) GetRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string) ([]pkg_models.Claim, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_role_claims")
	defer diag.Complete()

	dbClaims, err := s.roleStore.GetRoleClaims(ctx, tenantID, roleID)
	if err != nil {
		diag.AddError("failed_to_get_role_claims", "failed to get role claims", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims)

	return claims, diag
}

func (s *RoleService) GetPaginatedRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string, pagination *pkg_models.Pagination) (*api_models.PaginatedResponse[pkg_models.Claim], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_paginated_role_claims")
	defer diag.Complete()

	// Convert pkg_models.Pagination to filters.Pagination
	filterPagination := &filters.Pagination{
		Page:     pagination.Page,
		PageSize: pagination.PageSize,
	}

	dbClaims, err := s.roleStore.GetPaginatedRoleClaims(ctx, tenantID, roleID, filterPagination)
	if err != nil {
		diag.AddError("failed_to_get_paginated_role_claims", "failed to get paginated role claims", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims.Items)
	response := api_models.PaginatedResponse[pkg_models.Claim]{
		Data:       claims,
		TotalCount: dbClaims.Total,
		Pagination: api_models.Pagination{
			Page:       dbClaims.Page,
			PageSize:   dbClaims.PageSize,
			TotalPages: dbClaims.TotalPages,
		},
	}

	return &response, diag
}

func (s *RoleService) AddClaimToRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_claim_to_role")
	defer diag.Complete()

	err := s.roleStore.AddClaimToRole(ctx, tenantID, roleID, claimID)
	if err != nil {
		diag.AddError("failed_to_add_claim_to_role", "failed to add claim to role", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}

func (s *RoleService) RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_claim_from_role")
	defer diag.Complete()

	err := s.roleStore.RemoveClaimFromRole(ctx, tenantID, roleID, claimID)
	if err != nil {
		diag.AddError("failed_to_remove_claim_from_role", "failed to remove claim from role", "role", map[string]interface{}{
			"error": err.Error(),
		})
		return diag
	}

	return diag
}
