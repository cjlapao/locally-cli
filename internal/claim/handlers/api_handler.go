// Package handlers provides the API handlers for the claim service
package handlers

import (
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/claim/interfaces"
	claim_models "github.com/cjlapao/locally-cli/internal/claim/models"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/gorilla/mux"
)

type ClaimsApiHandler struct {
	claimService interfaces.ClaimServiceInterface
}

func NewClaimsApiHandler(claimService interfaces.ClaimServiceInterface) *ClaimsApiHandler {
	return &ClaimsApiHandler{claimService: claimService}
}

func (h *ClaimsApiHandler) Routes() []api_types.Route {
	return []api_types.Route{
		{
			Method:      http.MethodGet,
			Path:        "/v1/claims",
			Handler:     h.HandleGetClaims,
			Description: "Get all claims",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/claims/{id}",
			Handler:     h.HandleGetClaim,
			Description: "Get a claim by ID",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/claims",
			Handler:     h.HandleCreateClaim,
			Description: "Create a new claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/claims/{id}",
			Handler:     h.HandleUpdateClaim,
			Description: "Update a claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/claims/{id}",
			Handler:     h.HandleDeleteClaim,
			Description: "Delete a claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/claims/{id}/users",
			Handler:     h.HandleGetClaimUsers,
			Description: "Get users by claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/claims/{id}/users/{user_id}",
			Handler:     h.HandleAddUserToClaim,
			Description: "Add users to a claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/claims/{id}/users/{user_id}",
			Handler:     h.HandleRemoveUserFromClaim,
			Description: "Remove users from a claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/claims/{id}/roles",
			Handler:     h.HandleGetClaimRoles,
			Description: "Get roles by claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/claims/{id}/roles/{role_id}",
			Handler:     h.HandleAddRoleToClaim,
			Description: "Add roles to a claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/claims/{id}/roles/{role_id}",
			Handler:     h.HandleRemoveRoleFromClaim,
			Description: "Remove roles from a claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/claims/{id}/api-keys",
			Handler:     h.HandleGetClaimApiKeys,
			Description: "Get api keys by claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/claims/{id}/api-keys/{api_key_id}",
			Handler:     h.HandleAddApiKeyToClaim,
			Description: "Add api keys to a claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/claims/{id}/api-keys/{api_key_id}",
			Handler:     h.HandleRemoveApiKeyFromClaim,
			Description: "Remove api keys from a claim",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/claims/all/superuser",
			Handler:     h.HandleGetAllSuperUserLevelClaims,
			Description: "Get all superuser level claims",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/claims/all/admin",
			Handler:     h.HandleGetAllAdminLevelClaims,
			Description: "Get all admin level claims",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/claims/all/manager",
			Handler:     h.HandleGetAllManagerLevelClaims,
			Description: "Get all manager level claims",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/claims/all/user",
			Handler:     h.HandleGetAllUserLevelClaims,
			Description: "Get all user level claims",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/claims/all/guest",
			Handler:     h.HandleGetAllGuestLevelClaims,
			Description: "Get all guest level claims",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "claim", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
	}
}

func (h *ClaimsApiHandler) HandleGetClaims(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	pagination := utils.ParseQueryRequest(r)

	claims, diag := h.claimService.GetClaims(ctx, tenantID, pagination)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_claims", "Failed to get claims", diag)
		return
	}

	api.WriteObjectResponse(w, r, claims)
}

func (h *ClaimsApiHandler) HandleGetClaim(w http.ResponseWriter, r *http.Request) {
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

	claim, diag := h.claimService.GetClaimByIDOrSlug(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_claim", "Failed to get claim", diag)
		return
	}
	if claim == nil {
		api.WriteError(w, r, http.StatusNotFound, "claim not found", "claim not found", "")
		return
	}

	api.WriteObjectResponse(w, r, claim)
}

func (h *ClaimsApiHandler) HandleCreateClaim(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	request, parseDiags := api.ParseAndValidateBody[claim_models.CreateClaimRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "invalid_request", "Invalid request data", parseDiags)
		return
	}

	// check if the claim already exists
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	existingClaim, diag := h.claimService.GetClaimByIDOrSlug(ctx, tenantID, request.GetSlug())
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_claim", "Failed to get claim", diag)
		return
	}

	if existingClaim != nil {
		api.WriteError(w, r, http.StatusBadRequest, "claim already exists", "claim already exists", "")
		return
	}

	resultClaim, diag := h.claimService.CreateClaim(ctx, tenantID, &request)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_create_claim", "Failed to create claim", diag)
		return
	}

	api.WriteObjectResponse(w, r, resultClaim)
}

func (h *ClaimsApiHandler) HandleUpdateClaim(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}

	request, parseDiags := api.ParseAndValidateBody[claim_models.UpdateClaimRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "invalid_request", "Invalid request data", parseDiags)
		return
	}

	// check if the claim exists
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	existingClaim, diag := h.claimService.GetClaimByIDOrSlug(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_claim", "Failed to get claim", diag)
		return
	}

	if existingClaim == nil {
		api.WriteError(w, r, http.StatusNotFound, "claim not found", "claim not found", "")
		return
	}

	request.ID = existingClaim.ID

	slug, diag := h.claimService.UpdateClaim(ctx, tenantID, &request)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_update_claim", "Failed to update claim", diag)
		return
	}

	api.WriteSuccessResponse(w, r, slug, "Claim updated successfully")
}

func (h *ClaimsApiHandler) HandleDeleteClaim(w http.ResponseWriter, r *http.Request) {
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

	diag := h.claimService.DeleteClaim(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_delete_claim", "Failed to delete claim", diag)
		return
	}

	api.WriteSuccessResponse(w, r, id, "Claim deleted successfully")
}

func (h *ClaimsApiHandler) HandleGetClaimUsers(w http.ResponseWriter, r *http.Request) {
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

	pagination := utils.ParseQueryRequest(r)

	users, diag := h.claimService.GetClaimUsers(ctx, tenantID, id, pagination)

	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_claim_users", "Failed to get claim users", diag)
		return
	}

	api.WriteObjectResponse(w, r, users)
}

func (h *ClaimsApiHandler) HandleAddUserToClaim(w http.ResponseWriter, r *http.Request) {
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

	userID := mux.Vars(r)["user_id"]
	if userID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "user_id is required", "user_id is required", "")
		return
	}

	diag := h.claimService.AddClaimToUser(ctx, tenantID, userID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_add_user_to_claim", "Failed to add user to claim", diag)
		return
	}

	api.WriteSuccessResponse(w, r, id, "User added to claim successfully")
}

func (h *ClaimsApiHandler) HandleRemoveUserFromClaim(w http.ResponseWriter, r *http.Request) {
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

	userID := mux.Vars(r)["user_id"]
	if userID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "user_id is required", "user_id is required", "")
		return
	}

	diag := h.claimService.RemoveClaimFromUser(ctx, tenantID, userID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_remove_user_from_claim", "Failed to remove user from claim", diag)
		return
	}

	api.WriteSuccessResponse(w, r, id, "User removed from claim successfully")
}

func (h *ClaimsApiHandler) HandleGetAllSuperUserLevelClaims(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	claims, diag := h.claimService.GetAllSuperUserLevelClaims(ctx, tenantID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_all_superuser_level_claims", "Failed to get all superuser level claims", diag)
		return
	}

	api.WriteObjectResponse(w, r, claims)
}

func (h *ClaimsApiHandler) HandleGetAllAdminLevelClaims(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	claims, diag := h.claimService.GetAllAdminLevelClaims(ctx, tenantID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_all_admin_level_claims", "Failed to get all admin level claims", diag)
		return
	}

	api.WriteObjectResponse(w, r, claims)
}

func (h *ClaimsApiHandler) HandleGetAllManagerLevelClaims(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	claims, diag := h.claimService.GetAllManagerLevelClaims(ctx, tenantID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_all_manager_level_claims", "Failed to get all manager level claims", diag)
		return
	}

	api.WriteObjectResponse(w, r, claims)
}

func (h *ClaimsApiHandler) HandleGetAllUserLevelClaims(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	claims, diag := h.claimService.GetAllUserLevelClaims(ctx, tenantID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_all_user_level_claims", "Failed to get all user level claims", diag)
		return
	}

	api.WriteObjectResponse(w, r, claims)
}

func (h *ClaimsApiHandler) HandleGetAllGuestLevelClaims(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	claims, diag := h.claimService.GetAllGuestLevelClaims(ctx, tenantID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_all_guest_level_claims", "Failed to get all guest level claims", diag)
		return
	}

	api.WriteObjectResponse(w, r, claims)
}

func (h *ClaimsApiHandler) HandleGetClaimRoles(w http.ResponseWriter, r *http.Request) {
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

	pagination := utils.ParseQueryRequest(r)

	roles, diag := h.claimService.GetClaimRoles(ctx, tenantID, id, pagination)

	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_claim_roles", "Failed to get claim roles", diag)
		return
	}

	api.WriteObjectResponse(w, r, roles)
}

func (h *ClaimsApiHandler) HandleAddRoleToClaim(w http.ResponseWriter, r *http.Request) {
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

	roleID := mux.Vars(r)["role_id"]
	if roleID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "role_id is required", "role_id is required", "")
		return
	}

	diag := h.claimService.AddRoleToClaim(ctx, tenantID, id, roleID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_add_role_to_claim", "Failed to add role to claim", diag)
		return
	}

	api.WriteSuccessResponse(w, r, id, "Role added to claim successfully")
}

func (h *ClaimsApiHandler) HandleRemoveRoleFromClaim(w http.ResponseWriter, r *http.Request) {
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

	roleID := mux.Vars(r)["role_id"]
	if roleID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "role_id is required", "role_id is required", "")
		return
	}

	diag := h.claimService.RemoveRoleFromClaim(ctx, tenantID, id, roleID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_remove_role_from_claim", "Failed to remove role from claim", diag)
		return
	}

	api.WriteSuccessResponse(w, r, id, "Role removed from claim successfully")
}

func (h *ClaimsApiHandler) HandleGetClaimApiKeys(w http.ResponseWriter, r *http.Request) {
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

	pagination := utils.ParseQueryRequest(r)

	apiKeys, diag := h.claimService.GetClaimApiKeys(ctx, tenantID, id, pagination)

	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_claim_api_keys", "Failed to get claim api keys", diag)
		return
	}

	api.WriteObjectResponse(w, r, apiKeys)
}

func (h *ClaimsApiHandler) HandleAddApiKeyToClaim(w http.ResponseWriter, r *http.Request) {
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

	apiKeyID := mux.Vars(r)["api_key_id"]
	if apiKeyID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "api_key_id is required", "api_key_id is required", "")
		return
	}

	diag := h.claimService.AddApiKeyToClaim(ctx, tenantID, id, apiKeyID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_add_api_key_to_claim", "Failed to add api key to claim", diag)
		return
	}

	api.WriteSuccessResponse(w, r, id, "Api key added to claim successfully")
}

func (h *ClaimsApiHandler) HandleRemoveApiKeyFromClaim(w http.ResponseWriter, r *http.Request) {
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

	apiKeyID := mux.Vars(r)["api_key_id"]
	if apiKeyID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "api_key_id is required", "api_key_id is required", "")
		return
	}

	diag := h.claimService.RemoveApiKeyFromClaim(ctx, tenantID, id, apiKeyID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_remove_api_key_from_claim", "Failed to remove api key from claim", diag)
		return
	}

	api.WriteSuccessResponse(w, r, id, "Api key removed from claim successfully")
}
