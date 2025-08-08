package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	activity_interfaces "github.com/cjlapao/locally-cli/internal/activity/interfaces"
	activity_types "github.com/cjlapao/locally-cli/internal/activity/types"
	"github.com/cjlapao/locally-cli/internal/api"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/auth"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/validation"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/google/uuid"
)

type Handler struct {
	authService     *auth.AuthService
	store           stores.ApiKeyStoreInterface
	activityService activity_interfaces.ActivityServiceInterface
}

func NewApiHandler(authService *auth.AuthService, store stores.ApiKeyStoreInterface, activityService activity_interfaces.ActivityServiceInterface) *Handler {
	return &Handler{authService: authService, store: store, activityService: activityService}
}

func (h *Handler) Routes() []api_types.Route {
	return []api_types.Route{
		{
			Method:        http.MethodPost,
			Path:          "/v1/auth/login",
			Handler:       h.HandleLogin,
			Description:   "Login to the system with username/password",
			SecurityLevel: models.ApiKeySecurityLevelNone,
		},
		{
			Method:        http.MethodPost,
			Path:          "/v1/auth/login/api-key",
			Handler:       h.HandleAPIKeyLogin,
			Description:   "Login to the system with API key",
			SecurityLevel: models.ApiKeySecurityLevelNone,
		},
		{
			Method:        http.MethodPost,
			Path:          "/v1/auth/refresh",
			Handler:       h.HandleRefresh,
			Description:   "Refresh the token",
			SecurityLevel: models.ApiKeySecurityLevelAny,
		},
	}
}

// @Summary      Login with username/password
// @Description  Authenticate a user with username and password
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body auth.AuthCredentials true "Login credentials"
// @Success      200  {object}  auth.AuthToken
// @Failure      400  {object}  api.ErrorResponse
// @Failure      401  {object}  api.ErrorResponse
// @Router       /auth/login [post]
func (h *Handler) HandleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	cfg := config.GetInstance().Get()
	suUser := cfg.GetString(config.RootUserUsernameKey, "")

	var creds auth.AuthCredentials
	tenantID, _ := api.GetTenantIDFromRequest(r)

	creds, parseDiags := api.ParseAndValidateBody[auth.AuthCredentials](r)
	if parseDiags.HasErrors() {
		h.activityService.RecordErrorActivity(ctx, activity_types.ActivityTypeLogin, activity_types.ActivityErrorData{
			ErrorCode:    "invalid_request_body",
			ErrorMessage: "Failed to parse JSON: " + parseDiags.GetSummary(),
			StatusCode:   http.StatusBadRequest,
		}, &activity_types.ActivityRecord{
			TenantID:  creds.TenantID,
			ActorID:   config.UnknownUserID,
			ActorName: creds.Username,
			ActorType: activity_types.ActorTypeUser,
			Module:    "auth",
			Service:   "login",
			Message:   "Failed to parse JSON: " + parseDiags.GetSummary(),
			Success:   false,
		})
		api.WriteBadRequest(w, r, "Invalid request body", "Failed to parse JSON: "+parseDiags.GetSummary())
		return
	}

	if creds.TenantID == "" {
		creds.TenantID = tenantID
	}
	if creds.TenantID == "" && strings.EqualFold(creds.Username, suUser) {
		creds.TenantID = config.GlobalTenantID
	}

	invalidTenantID := uuid.Validate(creds.TenantID)

	if creds.TenantID == "" || invalidTenantID != nil {
		message := "Attempted to login with invalid or missing tenant ID"
		if invalidTenantID != nil {
			message = "Attempted to login with invalid or missing tenant ID: " + invalidTenantID.Error()
		} else {
			message = "Attempted to login with missing tenant ID"
		}
		h.activityService.RecordErrorActivity(ctx, activity_types.ActivityTypeLogin, activity_types.ActivityErrorData{
			ErrorCode:    "invalid_tenant_id",
			ErrorMessage: message,
			StatusCode:   http.StatusBadRequest,
		}, &activity_types.ActivityRecord{
			TenantID:  creds.TenantID,
			ActorName: creds.Username,
			ActorType: activity_types.ActorTypeUser,
			Module:    "auth",
			Service:   "login",
			Message:   message,
			Success:   false,
		})
		api.WriteBadRequest(w, r, "Tenant ID is required", "Tenant ID is required")
		return
	}

	// Authenticate
	token, authenticateDiags := h.authService.AuthenticateWithPassword(ctx, creds)
	if authenticateDiags != nil {
		tenantID := creds.TenantID
		actorID := config.UnknownUserID
		actorName := creds.Username
		tokenError := authenticateDiags.GetSummary()
		if token != nil {
			tenantID = token.TenantID
			actorID = token.UserID
			actorName = token.Username
			tokenError = token.Error
		}
		h.activityService.RecordErrorActivity(ctx, activity_types.ActivityTypeLogin, activity_types.ActivityErrorData{
			ErrorCode:    "invalid_credentials",
			ErrorMessage: tokenError,
			StatusCode:   http.StatusBadRequest,
		}, &activity_types.ActivityRecord{
			TenantID:  tenantID,
			ActorID:   actorID,
			ActorName: actorName,
			ActorType: activity_types.ActorTypeUser,
			Module:    "auth",
			Service:   "login",
			Message:   "Invalid credentials",
			Success:   false,
		})
		api.WriteUnauthorized(w, r, "Invalid credentials", "Authentication failed")
		return
	}

	h.activityService.RecordSuccessActivity(ctx, activity_types.ActivityTypeLogin, &activity_types.ActivityRecord{
		TenantID:  token.TenantID,
		ActorID:   creds.Username,
		ActorName: creds.Username,
		ActorType: activity_types.ActorTypeUser,
		Module:    "auth",
		Service:   "login",
		Message:   "Login successful for user: " + creds.Username,
		Success:   true,
	})

	// Return token
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(token); err != nil {
		api.WriteInternalError(w, r, "Failed to encode response", err.Error())
	}
}

// @Summary      Login with API key
// @Description  Authenticate a user with API key
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Param        request body auth.AuthCredentials true "API key credentials"
// @Success      200  {object}  auth.AuthToken
// @Failure      400  {object}  api.ErrorResponse
// @Failure      401  {object}  api.ErrorResponse
// @Router       /auth/login/api-key [post]
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

// @Summary      Refresh token
// @Description  Refresh the current authentication token
// @Tags         Authentication
// @Accept       json
// @Produce      json
// @Security     BearerAuth
// @Success      200  {object}  auth.AuthToken
// @Failure      400  {object}  api.ErrorResponse
// @Failure      401  {object}  api.ErrorResponse
// @Router       /auth/refresh [post]
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
