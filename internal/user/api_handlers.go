package user

import (
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/gorilla/mux"
)

type ApiHandler struct {
	userService UserServiceInterface
}

func NewApiHandler(userService UserServiceInterface) *ApiHandler {
	return &ApiHandler{userService: userService}
}

func (h *ApiHandler) Routes() []api.Route {
	return []api.Route{
		{
			Method:       http.MethodGet,
			Path:         "/v1/users",
			Handler:      h.HandleGetUsers,
			Description:  "Get all users",
			AuthRequired: true,
			Roles:        []models.Role{{Name: "admin"}},
			Claims:       []models.Claim{{Service: "user", Module: "records", Action: models.ClaimActionRead}},
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/users/self",
			Handler:      h.HandleGetSelfUser,
			Description:  "Get the current user",
			AuthRequired: true,
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/users/{id}",
			Handler:      h.HandleGetUser,
			Description:  "Get a user by ID",
			AuthRequired: true,
			Roles:        []models.Role{{Name: "admin"}},
			Claims:       []models.Claim{{Service: "user", Module: "records", Action: models.ClaimActionRead}},
		},
		{
			Method:       http.MethodPost,
			Path:         "/v1/users",
			Handler:      h.HandleCreateUser,
			Description:  "Create a user",
			AuthRequired: true,
			Roles:        []models.Role{{Name: "admin"}},
			Claims:       []models.Claim{{Service: "user", Module: "records", Action: models.ClaimActionCreate}},
		},
		{
			Method:       http.MethodPut,
			Path:         "/v1/users/self",
			Handler:      h.HandleUpdateSelfUser,
			Description:  "Update the current user",
			AuthRequired: true,
		},
		{
			Method:       http.MethodPut,
			Path:         "/v1/users/self/password",
			Handler:      h.HandleUpdateSelfUserPassword,
			Description:  "Update the current user password",
			AuthRequired: true,
		},
		{
			Method:       http.MethodPut,
			Path:         "/v1/users/{id}",
			Handler:      h.HandleUpdateUser,
			Description:  "Update a user",
			AuthRequired: true,
			Roles:        []models.Role{{Name: "admin"}},
			Claims:       []models.Claim{{Service: "user", Module: "records", Action: models.ClaimActionUpdate}},
		},
		{
			Method:       http.MethodDelete,
			Path:         "/v1/users/{id}",
			Handler:      h.HandleDeleteUser,
			Description:  "Delete a user",
			AuthRequired: true,
			Roles:        []models.Role{{Name: "admin"}},
			Claims:       []models.Claim{{Service: "user", Module: "records", Action: models.ClaimActionDelete}},
		},
		{
			Method:       http.MethodPut,
			Path:         "/v1/users/{id}/password",
			Handler:      h.HandleUpdateUserPassword,
			Description:  "Update a user password",
			AuthRequired: true,
			Roles:        []models.Role{{Name: "admin"}},
			Claims:       []models.Claim{{Service: "user", Module: "records", Action: models.ClaimActionUpdate}},
		},
	}
}

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

	api.WriteObjectResponse(w, r, user)
}

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

func (h *ApiHandler) HandleCreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}
	request, parseDiags := api.ParseAndValidateBody[CreateUserRequest](r)
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

	request, parseDiags := api.ParseAndValidateBody[UpdateUserRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "Invalid request", "Invalid request", parseDiags)
		return
	}

	updatedUser, respDiags := h.userService.UpdateUser(ctx, tenantID, id, &request)
	if respDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to update user", "Failed to update user", respDiags)
		return
	}

	response := &UpdateUserResponse{
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

	request, parseDiags := api.ParseAndValidateBody[UpdateUserRequest](r)
	if parseDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "Invalid request", "Invalid request", parseDiags)
		return
	}

	updatedUser, respDiags := h.userService.UpdateUser(ctx, tenantID, id, &request)
	if respDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to update user", "Failed to update user", respDiags)
		return
	}

	response := &UpdateUserResponse{
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

	request, parseDiags := api.ParseAndValidateBody[UpdateUserPasswordRequest](r)
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

	request, parseDiags := api.ParseAndValidateBody[UpdateUserPasswordRequest](r)
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
