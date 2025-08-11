// Package handlers provides the API handlers for the API keys service
package handlers

import (
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/api_keys/interfaces"
	api_keys_models "github.com/cjlapao/locally-cli/internal/api_keys/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/mappers"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/gorilla/mux"
)

type ApiKeysApiHandler struct {
	apiKeysService interfaces.ApiKeysServiceInterface
}

func NewApiKeysApiHandler(apiKeysService interfaces.ApiKeysServiceInterface) *ApiKeysApiHandler {
	return &ApiKeysApiHandler{apiKeysService: apiKeysService}
}

func (h *ApiKeysApiHandler) Routes() []api_types.Route {
	return []api_types.Route{
		{
			Method:      http.MethodGet,
			Path:        "/v1/api-keys",
			Handler:     h.HandleGetApiKeys,
			Description: "Get all API keys",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "auth", Module: "api_keys", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/api-keys/{id}",
			Handler:     h.HandleGetApiKey,
			Description: "Get an API key by ID",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "auth", Module: "api_keys", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/api-keys",
			Handler:     h.HandleCreateApiKey,
			Description: "Create a new API key",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelSuperUser,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "auth", Module: "api_keys", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/api-keys/{id}",
			Handler:     h.HandleDeleteApiKey,
			Description: "Delete an API key",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelSuperUser,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "auth", Module: "api_keys", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/api-keys/{id}/revoke",
			Handler:     h.HandleRevokeApiKey,
			Description: "Revoke an API key",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelSuperUser,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "auth", Module: "api_keys", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/api-keys/{id}/claim/{claim_id}",
			Handler:     h.HandleAddClaimToApiKey,
			Description: "Add a claim to an API key",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelSuperUser,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "auth", Module: "api_keys", Action: pkg_models.AccessLevelWrite}},
				},
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/api-keys/{id}/claim/{claim_id}",
			Handler:     h.HandleRemoveClaimFromApiKey,
			Description: "Remove a claim from an API key",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelSuperUser,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "auth", Module: "api_keys", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
	}
}

func (h *ApiKeysApiHandler) HandleGetApiKeys(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	filter, err := utils.GetFilterFromRequest(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, "Invalid filter", "Invalid filter", err.Error())
		return
	}
	page, pageSize := utils.GetPaginationFromRequest(r)
	pagination := &api_models.Pagination{
		Page:     page,
		PageSize: pageSize,
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	// if a filter is provided, return filtered api keys
	if filter != nil {
		apiKeys, diag := h.apiKeysService.GetFilteredApiKeys(ctx, tenantID, filter)
		if diag.HasErrors() {
			api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_api_keys", "Failed to get API keys", diag)
			return
		}
		api.WriteObjectResponse(w, r, apiKeys)
		return
	}

	// if no filter is provided, return all api keys
	apiKeys, diag := h.apiKeysService.GetApiKeys(ctx, tenantID, pagination)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_api_keys", "Failed to get API keys", diag)
		return
	}

	api.WriteObjectResponse(w, r, apiKeys)
}

func (h *ApiKeysApiHandler) HandleGetApiKey(w http.ResponseWriter, r *http.Request) {
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

	apiKey, diag := h.apiKeysService.GetApiKeyByID(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_api_key", "Failed to get API key", diag)
		return
	}
	if apiKey == nil {
		api.WriteError(w, r, http.StatusNotFound, "api key not found", "api key not found", "")
		return
	}

	api.WriteObjectResponse(w, r, apiKey)
}

func (h *ApiKeysApiHandler) HandleCreateApiKey(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	request, parseDiags := api.ParseAndValidateBody[api_keys_models.CreateApiKeyRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "invalid_request", "Invalid request data", parseDiags)
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	resultApiKey, diag := h.apiKeysService.CreateApiKey(ctx, tenantID, &request)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_create_api_key", "Failed to create API key", diag)
		return
	}

	response := mappers.MapApiKeyDtoToCreateResponse(resultApiKey, resultApiKey.PlaintextKey)
	api.WriteObjectResponse(w, r, response)
}

func (h *ApiKeysApiHandler) HandleDeleteApiKey(w http.ResponseWriter, r *http.Request) {
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

	diag := h.apiKeysService.DeleteApiKey(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_delete_api_key", "Failed to delete API key", diag)
		return
	}

	api.WriteObjectResponse(w, r, api_models.SuccessResponse{Message: "API key deleted successfully"})
}

func (h *ApiKeysApiHandler) HandleRevokeApiKey(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}
	userID := ctx.GetUserID()
	if userID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "user_id is required", "user_id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	request, parseDiags := api.ParseAndValidateBody[api_keys_models.RevokeApiKeyRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "invalid_request", "Invalid request data", parseDiags)
		return
	}

	diag := h.apiKeysService.RevokeApiKey(ctx, tenantID, userID, id, &request)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_revoke_api_key", "Failed to revoke API key", diag)
		return
	}

	api.WriteObjectResponse(w, r, api_models.SuccessResponse{Message: "API key revoked successfully"})
}

func (h *ApiKeysApiHandler) HandleAddClaimToApiKey(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}
	claimID := mux.Vars(r)["claim_id"]
	if claimID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "claim_id is required", "claim_id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	diag := h.apiKeysService.AddClaimToApiKey(ctx, tenantID, id, claimID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "failed_to_add_claim_to_api_key", "Failed to add claim to API key", diag)
		return
	}

	api.WriteObjectResponse(w, r, api_models.SuccessResponse{Message: "Claim added to API key successfully"})
}

func (h *ApiKeysApiHandler) HandleRemoveClaimFromApiKey(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "id is required", "id is required", "")
		return
	}
	claimID := mux.Vars(r)["claim_id"]
	if claimID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "claim_id is required", "claim_id is required", "")
		return
	}

	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant_id is required", "tenant_id is required", "")
		return
	}

	diag := h.apiKeysService.RemoveClaimFromApiKey(ctx, tenantID, id, claimID)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_remove_claim_from_api_key", "Failed to remove claim from API key", diag)
		return
	}

	api.WriteObjectResponse(w, r, api_models.SuccessResponse{Message: "Claim removed from API key successfully"})
}
