// Package service provides a service for managing roles.
package service

import (
	"sync"

	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	claim_interfaces "github.com/cjlapao/locally-cli/internal/claim/interfaces"
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

func (s *RoleService) GetRoles(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Role], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_roles")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	dbRoles, getRolesDiag := s.roleStore.GetRolesByQuery(ctx, tenantID, query)
	if getRolesDiag.HasErrors() {
		diag.Append(getRolesDiag)
		return &api_models.PaginationResponse[pkg_models.Role]{}, diag
	}

	roles := mappers.MapRolesToDto(dbRoles.Items)

	response := api_models.PaginationResponse[pkg_models.Role]{
		Data:       roles,
		TotalCount: dbRoles.Total,
		Pagination: api_models.Pagination{
			Page:       dbRoles.Page,
			PageSize:   dbRoles.PageSize,
			TotalPages: dbRoles.TotalPages,
		},
	}

	return &response, diag
}

func (s *RoleService) GetRoleByIDorSlug(ctx *appctx.AppContext, tenantID string, idOrSlug string) (*pkg_models.Role, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_role_by_id_or_slug")
	defer diag.Complete()

	role, getRoleDiag := s.roleStore.GetRoleBySlugOrID(ctx, tenantID, idOrSlug)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
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

	role, getRoleDiag := s.roleStore.GetRoleBySlugOrID(ctx, tenantID, idOrSlug)
	if getRoleDiag.HasErrors() {
		diag.Append(getRoleDiag)
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
	roleEntity, createRoleDiag := s.roleStore.CreateRole(ctx, tenantID, roleEntity)
	if createRoleDiag.HasErrors() {
		diag.Append(createRoleDiag)
		return nil, diag
	}

	return mappers.MapRoleToDto(roleEntity), diag
}

func (s *RoleService) UpdateRole(ctx *appctx.AppContext, tenantID string, role *models.UpdateRoleRequest) (string, *diagnostics.Diagnostics) {
	diag := diagnostics.New("update_role")
	defer diag.Complete()

	roleEntity := MapUpdateRoleRequestToEntity(role)
	updateRoleDiag := s.roleStore.UpdateRole(ctx, tenantID, roleEntity)
	if updateRoleDiag.HasErrors() {
		diag.Append(updateRoleDiag)
		return "", diag
	}

	return roleEntity.Slug, diag
}

func (s *RoleService) DeleteRole(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_role")
	defer diag.Complete()

	// checking if the role is used by any user
	users, getRoleUsersDiag := s.roleStore.GetRoleUsers(ctx, tenantID, id)
	if getRoleUsersDiag.HasErrors() {
		diag.Append(getRoleUsersDiag)
		return diag
	}

	if len(users) > 0 {
		diag.AddError("role_in_use", "role is in use", "role", map[string]interface{}{
			"role_id": id,
			"users":   users,
		})
		return diag
	}

	deleteRoleDiag := s.roleStore.DeleteRole(ctx, tenantID, id)
	if deleteRoleDiag.HasErrors() {
		diag.Append(deleteRoleDiag)
		return diag
	}

	return diag
}

func (s *RoleService) GetRoleUsers(ctx *appctx.AppContext, tenantID string, id string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.User], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_role_users")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	dbUsers, getUsersDiag := s.roleStore.GetRoleUsersByQuery(ctx, tenantID, id, query)
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

func (s *RoleService) AddUserToRole(ctx *appctx.AppContext, tenantID string, userID string, roleSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_user_role")
	defer diag.Complete()

	addUserToRoleDiag := s.roleStore.AddUserToRole(ctx, tenantID, userID, roleSlug)
	if addUserToRoleDiag.HasErrors() {
		diag.Append(addUserToRoleDiag)
		return diag
	}

	return diag
}

func (s *RoleService) RemoveUserFromRole(ctx *appctx.AppContext, tenantID string, userID string, roleSlug string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_user_role")
	defer diag.Complete()

	removeUserFromRoleDiag := s.roleStore.RemoveUserFromRole(ctx, tenantID, userID, roleSlug)
	if removeUserFromRoleDiag.HasErrors() {
		diag.Append(removeUserFromRoleDiag)
		return diag
	}

	return diag
}

func (s *RoleService) GetRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Claim], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_role_claims")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	dbClaims, getRoleClaimsDiag := s.roleStore.GetRoleClaimsByQuery(ctx, tenantID, roleID, query)
	if getRoleClaimsDiag.HasErrors() {
		diag.Append(getRoleClaimsDiag)
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims.Items)
	response := api_models.PaginationResponse[pkg_models.Claim]{
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

func (s *RoleService) GetPaginatedRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Claim], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_paginated_role_claims")
	defer diag.Complete()

	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	dbClaims, getRoleClaimsDiag := s.roleStore.GetRoleClaimsByQuery(ctx, tenantID, roleID, query)
	if getRoleClaimsDiag.HasErrors() {
		diag.Append(getRoleClaimsDiag)
		return nil, diag
	}

	claims := mappers.MapClaimsToDto(dbClaims.Items)
	response := api_models.PaginationResponse[pkg_models.Claim]{
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

	addClaimToRoleDiag := s.roleStore.AddClaimToRole(ctx, tenantID, roleID, claimID)
	if addClaimToRoleDiag.HasErrors() {
		diag.Append(addClaimToRoleDiag)
		return diag
	}

	return diag
}

func (s *RoleService) RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_claim_from_role")
	defer diag.Complete()

	removeClaimFromRoleDiag := s.roleStore.RemoveClaimFromRole(ctx, tenantID, roleID, claimID)
	if removeClaimFromRoleDiag.HasErrors() {
		diag.Append(removeClaimFromRoleDiag)
		return diag
	}

	return diag
}
