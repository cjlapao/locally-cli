package auth

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/validation"
)

type Credentials struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	TenantID string `json:"tenant_id"`
}

type Handler struct {
	authService *AuthService
	store       *stores.AuthDataStore
}

func NewApiHandler(authService *AuthService, store *stores.AuthDataStore) *Handler {
	return &Handler{authService: authService, store: store}
}

func (h *Handler) Routes() []api.Route {
	return []api.Route{
		{
			Method:       http.MethodPost,
			Path:         "/v1/auth/login",
			Handler:      h.HandleLogin,
			Description:  "Login to the system",
			AuthRequired: false,
		},
		{
			Method:       http.MethodPost,
			Path:         "/v1/auth/refresh",
			Handler:      h.HandleRefresh,
			Description:  "Refresh the token",
			AuthRequired: true,
		},
	}
}

func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	var creds Credentials
	if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
		api.WriteBadRequest(w, r, "Invalid request body", "Failed to parse JSON: "+err.Error())
		return
	}

	// Validate request
	if errors := validation.Validate(creds); len(errors) > 0 {
		api.WriteValidationError(w, r, "Invalid request", fmt.Sprintf("%v", errors))
		return
	}

	// Authenticate
	token, err := h.authService.Authenticate(creds)
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

func (h *Handler) HandleRefresh(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.Header.Get("Authorization")
	if refreshToken == "" {
		api.WriteUnauthorized(w, r, "No refresh token found", "")
		return
	}
	token, err := h.authService.RefreshToken(refreshToken)
	if err != nil {
		api.WriteUnauthorized(w, r, "Invalid refresh token", "")
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(token); err != nil {
		api.WriteInternalError(w, r, "Failed to encode response", err.Error())
	}
}
