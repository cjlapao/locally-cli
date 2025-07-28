package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/auth"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/validation"
	"github.com/cjlapao/locally-cli/pkg/types"
)

type Handler struct {
	authService *auth.AuthService
	store       stores.AuthDataStoreInterface
}

func NewApiHandler(authService *auth.AuthService, store stores.AuthDataStoreInterface) *Handler {
	return &Handler{authService: authService, store: store}
}

func (h *Handler) Routes() []api.Route {
	return []api.Route{
		{
			Method:       http.MethodPost,
			Path:         "/v1/auth/login",
			Handler:      h.HandleLogin,
			Description:  "Login to the system with username/password",
			AuthRequired: false,
		},
		{
			Method:       http.MethodPost,
			Path:         "/v1/auth/login/api-key",
			Handler:      h.HandleAPIKeyLogin,
			Description:  "Login to the system with API key",
			AuthRequired: false,
		},
		{
			Method:       http.MethodPost,
			Path:         "/v1/auth/refresh",
			Handler:      h.HandleRefresh,
			Description:  "Refresh the token",
			AuthRequired: true,
		},
		{
			Method:       http.MethodPost,
			Path:         "/v1/auth/api-keys",
			Handler:      h.HandleCreateAPIKey,
			Description:  "Create a new API key",
			AuthRequired: true,
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/auth/api-keys",
			Handler:      h.HandleListAPIKeys,
			Description:  "List all API keys for the current user",
			AuthRequired: true,
		},
		{
			Method:       http.MethodGet,
			Path:         "/v1/auth/api-keys/{id}",
			Handler:      h.HandleGetAPIKey,
			Description:  "Get details of a specific API key",
			AuthRequired: true,
		},
		{
			Method:       http.MethodDelete,
			Path:         "/v1/auth/api-keys/{id}",
			Handler:      h.HandleDeleteAPIKey,
			Description:  "Delete an API key",
			AuthRequired: true,
		},
		{
			Method:       http.MethodPost,
			Path:         "/v1/auth/api-keys/{id}/revoke",
			Handler:      h.HandleRevokeAPIKey,
			Description:  "Revoke an API key",
			AuthRequired: true,
		},
	}
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	var creds auth.AuthCredentials
	tenantID, _ := api.GetTenantIDFromRequest(r)

	creds, parseDiags := api.ParseAndValidateBody[auth.AuthCredentials](r)
	if parseDiags.HasErrors() {
		api.WriteBadRequest(w, r, "Invalid request body", "Failed to parse JSON: "+parseDiags.GetSummary())
		return
	}

	if creds.TenantID == "" {
		creds.TenantID = tenantID
	}
	if creds.TenantID == "" {
		api.WriteBadRequest(w, r, "Tenant ID is required", "Tenant ID is required")
		return
	}

	// Authenticate
	token, err := h.authService.AuthenticateWithPassword(ctx, creds)
	if err != nil {
		api.WriteUnauthorized(w, r, "Invalid credentials", "Authentication failed")
		return
	}

	// Return token
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(token); err != nil {
		api.WriteInternalError(w, r, "Failed to encode response", err.Error())
	}
}

func (h *Handler) HandleAPIKeyLogin(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	var creds auth.APIKeyCredentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		api.WriteBadRequest(w, r, "Invalid request body", "Failed to parse JSON: "+err.Error())
		return
	}

	// Validate request
	if errors := validation.Validate(creds); len(errors) > 0 {
		api.WriteValidationError(w, r, "Invalid request", fmt.Sprintf("%v", errors))
		return
	}

	// Authenticate with API key
	token, err := h.authService.AuthenticateWithAPIKey(ctx, creds)
	if err != nil {
		api.WriteUnauthorized(w, r, "Invalid API key", "Authentication failed")
		return
	}

	// Return token
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(token); err != nil {
		api.WriteInternalError(w, r, "Failed to encode response", err.Error())
	}
}

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	refreshToken := r.Header.Get("Authorization")
	if refreshToken == "" {
		api.WriteUnauthorized(w, r, "No refresh token found", "")
		return
	}
	token, err := h.authService.RefreshToken(ctx, refreshToken)
	if err != nil {
		api.WriteUnauthorized(w, r, "Invalid refresh token", "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(token); err != nil {
		api.WriteInternalError(w, r, "Failed to encode response", err.Error())
	}
}

func (h *Handler) HandleCreateAPIKey(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	// Get user from context
	userID := r.Context().Value(types.UserIDKey)
	if userID == nil {
		api.WriteUnauthorized(w, r, "User not found in context", "")
		return
	}

	var req auth.CreateAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteBadRequest(w, r, "Invalid request body", "Failed to parse JSON: "+err.Error())
		return
	}

	// Validate request
	if errors := validation.Validate(req); len(errors) > 0 {
		api.WriteValidationError(w, r, "Invalid request", fmt.Sprintf("%v", errors))
		return
	}

	// Create API key
	response, err := h.authService.CreateAPIKey(ctx, userID.(string), req, userID.(string))
	if err != nil {
		api.WriteInternalError(w, r, "Failed to create API key", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		api.WriteInternalError(w, r, "Failed to encode response", err.Error())
	}
}

func (h *Handler) HandleListAPIKeys(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	// Get user from context
	userID := r.Context().Value(types.UserIDKey)
	if userID == nil {
		api.WriteUnauthorized(w, r, "User not found in context", "")
		return
	}

	// List API keys
	response, err := h.authService.ListAPIKeys(ctx, userID.(string))
	if err != nil {
		api.WriteInternalError(w, r, "Failed to list API keys", err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		api.WriteInternalError(w, r, "Failed to encode response", err.Error())
	}
}

func (h *Handler) HandleGetAPIKey(w http.ResponseWriter, r *http.Request) {
	// Extract API key ID from URL
	ctx := appctx.FromContext(r.Context())
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		api.WriteBadRequest(w, r, "Invalid API key ID", "Missing API key ID in URL")
		return
	}
	apiKeyID := pathParts[len(pathParts)-1]

	// Get user from context
	userID := r.Context().Value(types.UserIDKey)
	if userID == nil {
		api.WriteUnauthorized(w, r, "User not found in context", "")
		return
	}

	// Get API key
	apiKey, err := h.store.GetAPIKeyByID(ctx, apiKeyID)
	if err != nil {
		api.WriteInternalError(w, r, "Failed to get API key", err.Error())
		return
	}
	if apiKey == nil {
		api.WriteNotFound(w, r, "API key not found", "")
		return
	}

	// Check if user owns this API key
	if apiKey.UserID != userID.(string) {
		api.WriteForbidden(w, r, "Access denied", "You can only access your own API keys")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(apiKey); err != nil {
		api.WriteInternalError(w, r, "Failed to encode response", err.Error())
	}
}

func (h *Handler) HandleDeleteAPIKey(w http.ResponseWriter, r *http.Request) {
	// Extract API key ID from URL
	ctx := appctx.FromContext(r.Context())
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 5 {
		api.WriteBadRequest(w, r, "Invalid API key ID", "Missing API key ID in URL")
		return
	}
	apiKeyID := pathParts[len(pathParts)-1]

	// Get user from context
	userID := r.Context().Value(types.UserIDKey)
	if userID == nil {
		api.WriteUnauthorized(w, r, "User not found in context", "")
		return
	}

	// Check if user owns this API key
	apiKey, err := h.store.GetAPIKeyByID(ctx, apiKeyID)
	if err != nil {
		api.WriteInternalError(w, r, "Failed to get API key", err.Error())
		return
	}
	if apiKey == nil {
		api.WriteNotFound(w, r, "API key not found", "")
		return
	}
	if apiKey.UserID != userID.(string) {
		api.WriteForbidden(w, r, "Access denied", "You can only delete your own API keys")
		return
	}

	// Delete API key
	err = h.authService.DeleteAPIKey(ctx, apiKeyID)
	if err != nil {
		api.WriteInternalError(w, r, "Failed to delete API key", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HandleRevokeAPIKey(w http.ResponseWriter, r *http.Request) {
	// Extract API key ID from URL
	ctx := appctx.FromContext(r.Context())
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		api.WriteBadRequest(w, r, "Invalid API key ID", "Missing API key ID in URL")
		return
	}
	apiKeyID := pathParts[len(pathParts)-2]

	// Get user from context
	userID := r.Context().Value(types.UserIDKey)
	if userID == nil {
		api.WriteUnauthorized(w, r, "User not found in context", "")
		return
	}

	// Check if user owns this API key
	apiKey, err := h.store.GetAPIKeyByID(ctx, apiKeyID)
	if err != nil {
		api.WriteInternalError(w, r, "Failed to get API key", err.Error())
		return
	}
	if apiKey == nil {
		api.WriteNotFound(w, r, "API key not found", "")
		return
	}
	if apiKey.UserID != userID.(string) {
		api.WriteForbidden(w, r, "Access denied", "You can only revoke your own API keys")
		return
	}

	var req auth.RevokeAPIKeyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		api.WriteBadRequest(w, r, "Invalid request body", "Failed to parse JSON: "+err.Error())
		return
	}

	// Validate request
	if errors := validation.Validate(req); len(errors) > 0 {
		api.WriteValidationError(w, r, "Invalid request", fmt.Sprintf("%v", errors))
		return
	}

	// Revoke API key
	err = h.authService.RevokeAPIKey(ctx, apiKeyID, userID.(string), req.Reason)
	if err != nil {
		api.WriteInternalError(w, r, "Failed to revoke API key", err.Error())
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
