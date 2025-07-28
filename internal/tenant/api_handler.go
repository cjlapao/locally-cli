package tenant

import (
	"fmt"
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/gorilla/mux"
)

type APIHandler struct {
	tenantService TenantServiceInterface
}

// NewApiHandler creates a new API handler for events
func NewApiHandler(tenantService TenantServiceInterface) *APIHandler {
	return &APIHandler{
		tenantService: tenantService,
	}
}

// Routes implements the RouteRegistrar interface
func (h *APIHandler) Routes() []api.Route {
	return []api.Route{
		{
			Method:            http.MethodGet,
			Path:              "/v1/tenants",
			Handler:           h.HandleGetTenants,
			Description:       "Get all tenants",
			AuthRequired:      true,
			SuperUserRequired: true,
		},
		{
			Method:            http.MethodGet,
			Path:              "/v1/tenants/{id}",
			Handler:           h.HandleGetTenant,
			Description:       "Get a tenant by ID",
			AuthRequired:      true,
			SuperUserRequired: true,
		},
		{
			Method:            http.MethodPost,
			Path:              "/v1/tenants",
			Handler:           h.HandleCreateTenant,
			Description:       "Create a new tenant",
			AuthRequired:      true,
			SuperUserRequired: true,
		},
		{
			Method:            http.MethodPut,
			Path:              "/v1/tenants/{id}",
			Handler:           h.HandleUpdateTenant,
			Description:       "Update a tenant",
			AuthRequired:      true,
			SuperUserRequired: true,
		},
		{
			Method:            http.MethodDelete,
			Path:              "/v1/tenants/{id}",
			Handler:           h.HandleDeleteTenant,
			Description:       "Delete a tenant",
			AuthRequired:      true,
			SuperUserRequired: true,
		},
	}
}

// HandleGetTenants handles the request to get all tenants
func (h *APIHandler) HandleGetTenants(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	filter, err := utils.GetFilterFromRequest(r)
	if err != nil {
		api.WriteError(w, r, http.StatusBadRequest, "Invalid filter", "Invalid filter", err.Error())
		return
	}
	tenants, diag := h.tenantService.GetTenantsByFilter(ctx, filter)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to get tenants", "Failed to get tenants", diag)
		return
	}
	api.WritePaginatedResponse(w, r, tenants.Data, tenants.Pagination, tenants.TotalCount)
}

// HandleGetTenant handles the request to get a tenant by ID

func (h *APIHandler) HandleGetTenant(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	diag := diagnostics.New("get_tenant")
	defer diag.Complete()
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}
	tenant, getDiag := h.tenantService.GetTenantByID(ctx, id)
	if getDiag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to get tenant", "Failed to get tenant", getDiag)
		return
	}
	if tenant == nil {
		api.WriteNotFound(w, r, "Tenant not found", fmt.Sprintf("Tenant with id or slug %v not found", id))
		return
	}
	api.WriteObjectResponse(w, r, tenant)
}

// HandleCreateTenant handles the request to create a new tenant
func (h *APIHandler) HandleCreateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	diag := diagnostics.New("create_tenant")
	defer diag.Complete()

	request, validateDiag := api.ParseAndValidateBody[models.Tenant](r)
	if validateDiag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "Invalid request body", "Invalid request body", validateDiag)
		return
	}
	tenant, createDiag := h.tenantService.CreateTenant(ctx, &request)
	if createDiag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to create tenant", "Failed to create tenant", diag)
		return
	}

	api.WriteObjectResponse(w, r, tenant)
}

// HandleUpdateTenant handles the request to update a tenant
func (h *APIHandler) HandleUpdateTenant(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	diag := diagnostics.New("update_tenant")
	defer diag.Complete()
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}

	request, validateDiag := api.ParseAndValidateBody[TenantUpdateRequest](r)
	if validateDiag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusBadRequest, "Invalid request body", "Invalid request body", validateDiag)
		return
	}
	request.ID = id

	tenant, updateDiag := h.tenantService.UpdateTenant(ctx, &request)
	if updateDiag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to update tenant", "Failed to update tenant", updateDiag)
		return
	}

	api.WriteSuccessResponse(w, r, tenant.ID, "Tenant updated successfully")
}

// HandleDeleteTenant handles the request to delete a tenant
func (h *APIHandler) HandleDeleteTenant(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())

	diag := diagnostics.New("delete_tenant")
	defer diag.Complete()
	id := mux.Vars(r)["id"]
	if id == "" {
		api.WriteError(w, r, http.StatusBadRequest, "Tenant ID is required", "Tenant ID is required", "")
		return
	}
	deleteDiag := h.tenantService.DeleteTenant(ctx, id)
	if deleteDiag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "Failed to delete tenant", "Failed to delete tenant", deleteDiag)
		return
	}
	api.WriteSuccessResponse(w, r, id, "Tenant deleted successfully")
}
