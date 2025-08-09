package ncontext

import (
	"encoding/json"
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/gorilla/mux"
)

type ApiHandler struct {
	ncontext *NContext
}

func NewApiHandler(ncontext *NContext) *ApiHandler {
	return &ApiHandler{ncontext: ncontext}
}

func (h *ApiHandler) Routes() []api_types.Route {
	return []api_types.Route{
		{
			Method:      http.MethodGet,
			Path:        "/v1/ncontext/status",
			Handler:     h.HandleGetStatus,
			Description: "Get the status of the NContext service",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/ncontext/contexts",
			Handler:     h.HandleGetContexts,
			Description: "Get all contexts",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/ncontext/contexts/{id}",
			Handler:     h.HandleGetContext,
			Description: "Get a specific context by ID",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodPost,
			Path:        "/v1/ncontext/contexts",
			Handler:     h.HandleCreateContext,
			Description: "Create a new context",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/ncontext/contexts/{id}",
			Handler:     h.HandleUpdateContext,
			Description: "Update a context",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodDelete,
			Path:        "/v1/ncontext/contexts/{id}",
			Handler:     h.HandleDeleteContext,
			Description: "Delete a context",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/ncontext/services",
			Handler:     h.HandleGetServices,
			Description: "Get all registered services",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/ncontext/default-context",
			Handler:     h.HandleGetDefaultContext,
			Description: "Get the default context",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
		{
			Method:      http.MethodPut,
			Path:        "/v1/ncontext/default-context/{id}",
			Handler:     h.HandleSetDefaultContext,
			Description: "Set the default context",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
	}
}

func (h *ApiHandler) HandleGetStatus(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting NContext service status")

	status := h.ncontext.GetStatus(ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
	ctx.LogWithField("loaded", status["loaded"]).Info("NContext status retrieved successfully")
}

func (h *ApiHandler) HandleGetContexts(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting all contexts")

	contextIDs := h.ncontext.ListContexts(ctx)
	contexts := make([]*Context, 0, len(contextIDs))

	for _, id := range contextIDs {
		if context, exists := h.ncontext.GetContext(ctx, id); exists {
			contexts = append(contexts, context)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(contexts)
	ctx.LogWithField("count", len(contexts)).Info("Contexts retrieved successfully")
}

func (h *ApiHandler) HandleGetContext(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())

	// Extract context ID from URL path
	vars := mux.Vars(r)
	contextID := vars["id"]
	if contextID == "" {
		api.WriteBadRequest(w, r, "Context ID is required", "Missing context ID in URL path")
		return
	}

	ctx.LogWithField("context_id", contextID).Info("Getting context")

	context, exists := h.ncontext.GetContext(ctx, contextID)
	if !exists {
		api.WriteNotFound(w, r, "Context not found", "Context with ID "+contextID+" not found")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(context)
	ctx.LogWithField("context_id", contextID).Info("Context retrieved successfully")
}

func (h *ApiHandler) HandleCreateContext(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Creating new context")

	var context Context
	if err := json.NewDecoder(r.Body).Decode(&context); err != nil {
		api.WriteBadRequest(w, r, "Invalid request body", err.Error())
		return
	}

	diag := h.ncontext.AddContext(ctx, &context)
	if diag.HasErrors() {
		api.WriteBadRequest(w, r, "Failed to create context", diag.GetSummary())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(context)
	ctx.LogWithField("context_id", context.ID).Info("Context created successfully")
}

func (h *ApiHandler) HandleUpdateContext(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())

	// Extract context ID from URL path
	vars := mux.Vars(r)
	contextID := vars["id"]
	if contextID == "" {
		api.WriteBadRequest(w, r, "Context ID is required", "Missing context ID in URL path")
		return
	}

	ctx.LogWithField("context_id", contextID).Info("Updating context")

	// Check if context exists
	existingContext, exists := h.ncontext.GetContext(ctx, contextID)
	if !exists {
		api.WriteNotFound(w, r, "Context not found", "Context with ID "+contextID+" not found")
		return
	}

	var updatedContext Context
	if err := json.NewDecoder(r.Body).Decode(&updatedContext); err != nil {
		api.WriteBadRequest(w, r, "Invalid request body", err.Error())
		return
	}

	// Preserve the ID
	updatedContext.ID = contextID

	// TODO: Implement actual update logic
	// For now, just return the existing context
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(existingContext)
	ctx.LogWithField("context_id", contextID).Info("Context update requested")
}

func (h *ApiHandler) HandleDeleteContext(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())

	// Extract context ID from URL path
	vars := mux.Vars(r)
	contextID := vars["id"]
	if contextID == "" {
		api.WriteBadRequest(w, r, "Context ID is required", "Missing context ID in URL path")
		return
	}

	ctx.LogWithField("context_id", contextID).Info("Deleting context")

	// Check if context exists
	_, exists := h.ncontext.GetContext(ctx, contextID)
	if !exists {
		api.WriteNotFound(w, r, "Context not found", "Context with ID "+contextID+" not found")
		return
	}

	// TODO: Implement actual delete logic
	// For now, just return success
	w.WriteHeader(http.StatusNoContent)
	ctx.LogWithField("context_id", contextID).Info("Context deletion requested")
}

func (h *ApiHandler) HandleGetServices(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting all registered services")

	services := h.ncontext.ListServices(ctx)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(services)
	ctx.LogWithField("count", len(services)).Info("Services retrieved successfully")
}

func (h *ApiHandler) HandleGetDefaultContext(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	ctx.LogInfo("Getting default context")

	defaultContext, exists := h.ncontext.GetDefaultContext(ctx)
	if !exists {
		api.WriteNotFound(w, r, "Default context not found", "No default context is set")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(defaultContext)
	ctx.LogWithField("context_id", defaultContext.ID).Info("Default context retrieved successfully")
}

func (h *ApiHandler) HandleSetDefaultContext(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())

	// Extract context ID from URL path
	vars := mux.Vars(r)
	contextID := vars["id"]
	if contextID == "" {
		api.WriteBadRequest(w, r, "Context ID is required", "Missing context ID in URL path")
		return
	}

	ctx.LogWithField("context_id", contextID).Info("Setting default context")

	diag := h.ncontext.SetDefaultContext(ctx, contextID)
	if diag.HasErrors() {
		api.WriteBadRequest(w, r, "Failed to set default context", diag.GetSummary())
		return
	}

	w.WriteHeader(http.StatusNoContent)
	ctx.LogWithField("context_id", contextID).Info("Default context set successfully")
}
