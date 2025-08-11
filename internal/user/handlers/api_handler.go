// Package handlers provides the API handlers for the user service
package handlers

import (
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/internal/user/interfaces"
	"github.com/cjlapao/locally-cli/internal/user/models"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/gorilla/mux"
)

type ApiHandler struct {
	userService interfaces.UserServiceInterface
}

func NewUsersApiHandler(userService interfaces.UserServiceInterface) *ApiHandler {
	return &ApiHandler{userService: userService}
}

func (h *ApiHandler) Routes() []api_types.Route {
	return []api_types.Route{
		{
			Method:      http.MethodGet,
			Path:        "/v1/users",
			Handler:     h.HandleGetUsers,
			Description: "Get all users",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "user", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/users/self",
			Handler:     h.HandleGetSelfUser,
			Description: "Get the current user",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/users/{id}",
			Handler:     h.HandleGetUser,
			Description: "Get a user by ID",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "user", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},

		{
			Method:      http.MethodPost,
			Path:        "/v1/users",
			Handler:     h.HandleCreateUser,
			Description: "Create a user",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "user", Module: "api", Action: pkg_models.AccessLevelCreate}},
				},
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/users/self",
			Handler:     h.HandleUpdateSelfUser,
			Description: "Update the current user",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/users/self/password",
			Handler:     h.HandleUpdateSelfUserPassword,
			Description: "Update the current user password",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/users/{id}",
			Handler:     h.HandleUpdateUser,
			Description: "Update a user",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "user", Module: "api", Action: pkg_models.AccessLevelUpdate}},
				},
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/users/{id}",
			Handler:     h.HandleDeleteUser,
			Description: "Delete a user",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "user", Module: "api", Action: pkg_models.AccessLevelDelete}},
				},
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/users/{id}/password",
			Handler:     h.HandleUpdateUserPassword,
			Description: "Update a user password",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "user", Module: "api", Action: pkg_models.AccessLevelUpdate}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/users/self/claims",
			Handler:     h.HandleGetSelfUserClaims,
			Description: "Get the current user claims",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "user", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/users/{id}/claims",
			Handler:     h.HandleGetUserClaims,
			Description: "Get a user claims",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "user", Module: "api", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
	}
}

// @Summary      Get current user
// @Description  Get the currently authenticated user's information
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  pkg.models.User
// @Failure      400  {object}  api.ErrorResponse
// @Failure      401  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Router       /users/self [get]
func (h *ApiHandler) HandleGetSelfUser(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}
	id := ctx.GetUserID()
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "User ID is required", "User ID is required", "")
		return
	}

	user, diag := h.userService.GetUserByID(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to get user", "Failed to get user", diag)
		return
	}
	if user == nil {
		api.WriteError(w, r, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	response := mappers.MapSelfUserToDto(user)

	api.WriteObjectResponse(w, r, response)
}

// @Summary      Get all users
// @Description  Get a paginated list of all users in the current tenant
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        page query int false "Page number (default: 1)"
// @Param        limit query int false "Items per page (default: 10)"
// @Param        sort query string false "Sort field"
// @Param        order query string false "Sort order (asc/desc)"
// @Success      200  {object}  api.PaginatedResponse{pkg.models.User}
// @Failure      400  {object}  api.ErrorResponse
// @Failure      401  {object}  api.ErrorResponse
// @Router       /users [get]
func (h *ApiHandler) HandleGetUsers(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	filter, err := utils.GetFilterFromRequest(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, "Invalid filter", "Invalid filter", err.Error())
		return
	}
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}
	users, diag := h.userService.GetUsersByFilter(ctx, tenantID, filter)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to get users", "Failed to get users", diag)
		return
	}
	api.WritePaginatedResponse(w, r, users.Data, users.Pagination, users.TotalCount)
}

// @Summary      Get user by ID
// @Description  Get a specific user by their ID
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        id path string true "User ID"
// @Success      200  {object}  pkg.models.User
// @Failure      400  {object}  api.ErrorResponse
// @Failure      401  {object}  api.ErrorResponse
// @Failure      404  {object}  api.ErrorResponse
// @Router       /users/{id} [get]
func (h *ApiHandler) HandleGetUser(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "User ID is required", "User ID is required", "")
		return
	}
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}

	user, diag := h.userService.GetUserByID(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to get user", "Failed to get user", diag)
		return
	}
	if user == nil {
		api.WriteError(w, r, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	api.WriteObjectResponse(w, r, user)
}

// @Summary      Create user
// @Description  Create a new user in the current tenant
// @Tags         Users
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Param        request body CreateUserRequest true "User creation data"
// @Success      200  {object}  pkg.models.User
// @Failure      400  {object}  api.ErrorResponse
// @Failure      401  {object}  api.ErrorResponse
// @Router       /users [post]
func (h *ApiHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}
	request, parseDiags := api.ParseAndValidateBody[models.CreateUserRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "Invalid request", "Invalid request", parseDiags)
		return
	}

	response, respDiags := h.userService.CreateUser(ctx, tenantID, request.Role, &request)
	if respDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to create user", "Failed to create user", respDiags)
		return
	}

	api.WriteObjectResponse(w, r, response)
}

func (h *ApiHandler) HandleUpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "User ID is required", "User ID is required", "")
		return
	}

	user, diag := h.userService.GetUserByID(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to get user", "Failed to get user", diag)
		return
	}
	if user == nil {
		api.WriteError(w, r, http.StatusNotFound, "User not found", "User not found", "")
		return
	}

	request, parseDiags := api.ParseAndValidateBody[models.UpdateUserRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "Invalid request", "Invalid request", parseDiags)
		return
	}

	updatedUser, respDiags := h.userService.UpdateUser(ctx, tenantID, id, &request)
	if respDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to update user", "Failed to update user", respDiags)
		return
	}

	response := &models.UpdateUserResponse{
		ID:     updatedUser.ID,
		Name:   updatedUser.Name,
		Status: updatedUser.Status,
	}

	api.WriteObjectResponse(w, r, response)
}

func (h *ApiHandler) HandleDeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "User ID is required", "User ID is required", "")
		return
	}

	diag := h.userService.DeleteUser(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to delete user", "Failed to delete user", diag)
		return
	}

	api.WriteSuccessResponse(w, r, id, "User deleted successfully")
}

func (h *ApiHandler) HandleUpdateSelfUser(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}
	id := ctx.GetUserID()
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "User ID is required", "User ID is required", "")
		return
	}

	request, parseDiags := api.ParseAndValidateBody[models.UpdateUserRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "Invalid request", "Invalid request", parseDiags)
		return
	}

	updatedUser, respDiags := h.userService.UpdateUser(ctx, tenantID, id, &request)
	if respDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to update user", "Failed to update user", respDiags)
		return
	}

	response := &models.UpdateUserResponse{
		ID:     updatedUser.ID,
		Name:   updatedUser.Name,
		Status: updatedUser.Status,
	}

	api.WriteObjectResponse(w, r, response)
}

func (h *ApiHandler) HandleUpdateSelfUserPassword(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}

	id := ctx.GetUserID()
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "User ID is required", "User ID is required", "")
		return
	}

	request, parseDiags := api.ParseAndValidateBody[models.UpdateUserPasswordRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "Invalid request", "Invalid request", parseDiags)
		return
	}

	diag := h.userService.UpdateUserPassword(ctx, tenantID, id, &request)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to update user password", "Failed to update user password", diag)
		return
	}
}

func (h *ApiHandler) HandleUpdateUserPassword(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "User ID is required", "User ID is required", "")
		return
	}

	request, parseDiags := api.ParseAndValidateBody[models.UpdateUserPasswordRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "Invalid request", "Invalid request", parseDiags)
		return
	}

	diag := h.userService.UpdateUserPassword(ctx, tenantID, id, &request)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to update user password", "Failed to update user password", diag)
		return
	}

	api.WriteSuccessResponse(w, r, id, "User password updated successfully")
}

func (h *ApiHandler) HandleGetUserClaims(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}

	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "User ID is required", "User ID is required", "")
		return
	}

	claims, diag := h.userService.GetUserClaims(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to get user claims", "Failed to get user claims", diag)
		return
	}

	api.WriteObjectResponse(w, r, claims)
}

func (h *ApiHandler) HandleGetSelfUserClaims(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}

	id := ctx.GetUserID()
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "User ID is required", "User ID is required", "")
		return
	}

	claims, diag := h.userService.GetUserClaims(ctx, tenantID, id)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to get user claims", "Failed to get user claims", diag)
		return
	}

	api.WriteObjectResponse(w, r, claims)
}
