// Package handlers provides the API handlers for the role service.
package handlers

import (
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/role/interfaces"
	"github.com/cjlapao/locally-cli/internal/role/models"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/gorilla/mux"
)

type RolesApiHandler struct {
	roleService interfaces.RoleServiceInterface
}

func NewRolesApiHandler(roleService interfaces.RoleServiceInterface) *RolesApiHandler {
	return &RolesApiHandler{roleService: roleService}
}

func (h *RolesApiHandler) Routes() []api_types.Route {
	return []api_types.Route{
		{
			Method:      http.MethodGet,
			Path:        "/v1/roles",
			Handler:     h.HandleGetRoles,
			Description: "Get all roles",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/roles/{id}",
			Handler:     h.HandleGetRole,
			Description: "Get a role by ID",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/roles/{id}/claims",
			Handler:     h.HandleGetRoleClaims,
			Description: "Get claims by role",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/roles",
			Handler:     h.HandleCreateRole,
			Description: "Create a new role",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/roles/{id}",
			Handler:     h.HandleUpdateRole,
			Description: "Update a role",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/roles/{id}",
			Handler:     h.HandleDeleteRole,
			Description: "Delete a role",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/roles/{id}/users",
			Handler:     h.HandleGetRoleUsers,
			Description: "Get users by role",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/roles/{id}/users/{user_id}",
			Handler:     h.HandleAddUserToRole,
			Description: "Add users to a role",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/roles/{id}/users/{user_id}",
			Handler:     h.HandleRemoveUserFromRole,
			Description: "Remove users from a role",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/roles/{id}/claims/{claim_id}",
			Handler:     h.HandleAddClaimToRole,
			Description: "Add claims to a role",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/roles/{id}/claims/{claim_id}",
			Handler:     h.HandleRemoveClaimFromRole,
			Description: "Remove claims from a role",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "role", Module: "api", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
	}
}

func (h *RolesApiHandler) HandleGetRoles(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	filter, err := utils.GetFilterFromRequest(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, "Invalid filter", "Invalid filter", err.Error())
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	roles, diag := h.roleService.GetRolesByFilter(ctx, tenantID, filter)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_roles", "Failed to get roles", diag)
		return
	}

	api.WriteObjectResponse(w, r, roles)
}

func (h *RolesApiHandler) HandleGetRole(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	role, diag := h.roleService.GetRoleByIDorSlug(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_roles", "Failed to get roles", diag)
		return
	}

	if role == nil {
		api.WriteError(w, r, http.StatusNotFound, "role not found", "role not found", "")
		return
	}

	api.WriteObjectResponse(w, r, role)
}

func (h *RolesApiHandler) HandleCreateRole(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	request, parseDiags := api.ParseAndValidateBody[models.CreateRoleRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "invalid_request", "Invalid request data", parseDiags)
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	role, diag := h.roleService.CreateRole(ctx, tenantID, &request)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_create_role", "Failed to create role", diag)
		return
	}

	api.WriteObjectResponse(w, r, role)
}

func (h *RolesApiHandler) HandleUpdateRole(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}

	request, parseDiags := api.ParseAndValidateBody[models.UpdateRoleRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "invalid_request", "Invalid request data", parseDiags)
		return
	}
	// check if the role exists
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	existingRole, diag := h.roleService.GetRoleByIDorSlug(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_role", "Failed to get role", diag)
		return
	}
	if existingRole == nil {
		api.WriteError(w, r, http.StatusBadRequest, "role not found", "role not found", "")
		return
	}

	request.ID = existingRole.ID
	request.Slug = existingRole.Slug

	slug, diag := h.roleService.UpdateRole(ctx, tenantID, &request)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_update_role", "Failed to update role", diag)
		return
	}

	api.WriteSuccessResponse(w, r, slug, "Role updated successfully")
}

func (h *RolesApiHandler) HandleDeleteRole(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	diag := h.roleService.DeleteRole(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_delete_role", "Failed to delete role", diag)
		return
	}

	api.WriteSuccessResponse(w, r, id, "Role deleted successfully")
}

func (h *RolesApiHandler) HandleGetRoleUsers(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	page, pageSize := utils.GetPaginationFromRequest(r)

	users, diag := h.roleService.GetRoleUsers(ctx, tenantID, id, &pkg_models.Pagination{
		Page:     page,
		PageSize: pageSize,
	})

	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_role_users", "Failed to get role users", diag)
		return
	}

	api.WriteObjectResponse(w, r, users)
}

func (h *RolesApiHandler) HandleAddUserToRole(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	roleID := mux.Vars(r)["id"]
	if roleID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	userID := mux.Vars(r)["user_id"]
	if userID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "user_id is required", "user_id is required", "")
		return
	}

	// check if the role exists
	existingRole, diag := h.roleService.GetRoleByIDorSlug(ctx, tenantID, roleID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_role", "Failed to get role", diag)
		return
	}
	if existingRole == nil {
		api.WriteError(w, r, http.StatusBadRequest, "role not found", "role not found", "")
		return
	}

	diag = h.roleService.AddUserToRole(ctx, tenantID, userID, roleID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_add_user_to_role", "Failed to add user to role", diag)
		return
	}

	api.WriteSuccessResponse(w, r, userID, "User added to role successfully")
}

func (h *RolesApiHandler) HandleRemoveUserFromRole(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	roleID := mux.Vars(r)["id"]
	if roleID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}

	userID := mux.Vars(r)["user_id"]
	if userID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "user_id is required", "user_id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	// check if the role exists
	existingRole, diag := h.roleService.GetRoleByIDorSlug(ctx, tenantID, roleID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_role", "Failed to get role", diag)
		return
	}
	if existingRole == nil {
		api.WriteError(w, r, http.StatusBadRequest, "role not found", "role not found", "")
		return
	}

	diag = h.roleService.RemoveUserFromRole(ctx, tenantID, userID, roleID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_remove_user_from_role", "Failed to remove user from role", diag)
		return
	}

	api.WriteSuccessResponse(w, r, userID, "User removed from role successfully")
}

func (h *RolesApiHandler) HandleGetRoleClaims(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	roleID := mux.Vars(r)["id"]
	if roleID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	existingRole, diag := h.roleService.GetRoleByIDorSlug(ctx, tenantID, roleID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_role", "Failed to get role", diag)
		return
	}
	if existingRole == nil {
		api.WriteError(w, r, http.StatusBadRequest, "role not found", "role not found", "")
		return
	}

	claims, diag := h.roleService.GetRoleClaims(ctx, tenantID, existingRole.ID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_role_claims", "Failed to get role claims", diag)
		return
	}

	api.WriteObjectResponse(w, r, claims)
}

func (h *RolesApiHandler) HandleAddClaimToRole(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	roleID := mux.Vars(r)["id"]
	if roleID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	claimID := mux.Vars(r)["claim_id"]
	if claimID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "claim_id is required", "claim_id is required", "")
		return
	}

	diag := h.roleService.AddClaimToRole(ctx, tenantID, roleID, claimID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_add_claim_to_role", "Failed to add claim to role", diag)
		return
	}
	api.WriteSuccessResponse(w, r, claimID, "Claim added to role successfully")
}

func (h *RolesApiHandler) HandleRemoveClaimFromRole(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	roleID := mux.Vars(r)["id"]
	if roleID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	claimID := mux.Vars(r)["claim_id"]
	if claimID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "claim_id is required", "claim_id is required", "")
		return
	}

	diag := h.roleService.RemoveClaimFromRole(ctx, tenantID, roleID, claimID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_remove_claim_from_role", "Failed to remove claim from role", diag)
		return
	}
	api.WriteSuccessResponse(w, r, claimID, "Claim removed from role successfully")
}
