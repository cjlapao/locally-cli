package api

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	auth_interfaces "github.com/cjlapao/locally-cli/internal/auth/interfaces"
	auth_models "github.com/cjlapao/locally-cli/internal/auth/models"
	authctx "github.com/cjlapao/locally-cli/internal/authctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/sirupsen/logrus"
)

func NewAuthorizationPreMiddleware(authService auth_interfaces.AuthServiceInterface, route *api_types.Route) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		// Debug logging to see if auth middleware is being called
		debugCtx := appctx.FromContext(r.Context())
		debugCtx.LogInfo("Auth middleware: Starting authentication")
		diag := diagnostics.New("auth_middleware")
		defer diag.Complete()

		if route == nil {
			diag.AddError("security_requirement_nil", "Security requirement is nil, this should not happen", "Authorization_middleware_security_requirement_nil")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// If the route has no security requirement, we will skip the authentication
		if route.SecurityRequirement == nil {
			return MiddlewareResult{Continue: true}
		}

		// checking if there is a need for checking security in the route
		if !route.SecurityRequirement.SecurityLevel.RequiresAuthentication() {
			return MiddlewareResult{Continue: true}
		}

		// Get the Authorization header
		authHeader := extractAuthorizationHeaders(debugCtx, r)
		if authHeader.AuthorizationType == api_types.AuthorizationHeaderTypeNone {
			errorMessage := "Failed to extract authorization headers"
			errorDetails := "No known authorization header found in the request"
			writeUnauthorizedError(w, r, errorMessage, errorDetails)
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		if authHeader.AuthorizationType == api_types.AuthorizationHeaderTypeBearer {
			valid, err := validateBearerToken(debugCtx, r, authService, route, authHeader.Token)
			if err != nil {
				writeUnauthorizedError(w, r, "Invalid token", err.Error())
				return MiddlewareResult{Continue: false, Diagnostics: diag}
			}
			if !valid {
				writeUnauthorizedError(w, r, "Invalid token", "Invalid token")
				return MiddlewareResult{Continue: false, Diagnostics: diag}
			}
		}

		if authHeader.AuthorizationType == api_types.AuthorizationHeaderTypeApiKey {
			valid, err := validateApiKey(debugCtx, r, authService, route, authHeader.Token)
			if err != nil {
				writeUnauthorizedError(w, r, "Invalid API key", err.Error())
				return MiddlewareResult{Continue: false, Diagnostics: diag}
			}
			if !valid {
				writeUnauthorizedError(w, r, "Invalid API key", "Invalid API key")
				return MiddlewareResult{Continue: false, Diagnostics: diag}
			}
		}

		return MiddlewareResult{Continue: true}
	})
}

// NewRequireAuthPreMiddleware creates a pre-middleware that validates JWT tokens
func NewRequireAuthPreMiddleware(authService auth_interfaces.AuthServiceInterface) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		diag := diagnostics.New("auth_middleware")
		defer diag.Complete()
		// Debug logging to see if auth middleware is being called
		debugCtx := appctx.FromContext(r.Context())
		debugCtx.LogInfo("Auth middleware: Starting authentication")

		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			debugCtx.LogError("Auth middleware: Missing Authorization header")
			writeUnauthorizedError(w, r, "Authorization header required", "Missing Authorization header")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeUnauthorizedError(w, r, "Invalid authorization header format", "Expected 'Bearer <token>' format")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			writeUnauthorizedError(w, r, "Empty token", "Token cannot be empty")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Validate the token using the provided auth service
		debugCtx.LogInfo("Auth middleware: Validating token")
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			debugCtx.LogWithError(err).Error("Auth middleware: Token validation failed")
			writeInvalidTokenError(w, r, "Token validation failed", err.Error())
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}
		debugCtx.LogInfo("Auth middleware: Token validation successful")

		// Add claims to context using AppContext
		appCtx := appctx.FromContext(r.Context())
		appCtx = appCtx.WithTenantID(claims.TenantID)
		appCtx = appCtx.WithUserID(claims.UserID)
		appCtx = appCtx.WithUsername(claims.Username)

		// Add claims to the underlying context for backward compatibility
		// We need to update the AppContext's underlying context directly
		appCtx.Context = context.WithValue(appCtx.Context, authctx.ClaimsKey, claims)
		appCtx.Context = context.WithValue(appCtx.Context, types.TenantIDKey, claims.TenantID)
		appCtx.Context = context.WithValue(appCtx.Context, types.UserIDKey, claims.Username)

		*r = *r.WithContext(appCtx)

		// Debug logging to verify claims are set
		appCtx.LogWithFields(logrus.Fields{
			"tenant_id":         claims.TenantID,
			"user_id":           claims.UserID,
			"username":          claims.Username,
			"roles":             claims.Roles,
			"auth_context_addr": fmt.Sprintf("%p", appCtx),
		}).Info("Auth middleware: Claims set in context")

		return MiddlewareResult{Continue: true}
	})
}

func NewRequireSuperUserPreMiddleware(authService auth_interfaces.AuthServiceInterface) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		// Debug logging to see if auth middleware is being called
		diag := diagnostics.New("auth_middleware")
		defer diag.Complete()
		debugCtx := appctx.FromContext(r.Context())
		debugCtx.LogInfo("Auth middleware: Starting authentication")

		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			debugCtx.LogError("Auth middleware: Missing Authorization header")
			writeUnauthorizedError(w, r, "Authorization header required", "Missing Authorization header")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeUnauthorizedError(w, r, "Invalid authorization header format", "Expected 'Bearer <token>' format")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			writeUnauthorizedError(w, r, "Empty token", "Token cannot be empty")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Validate the token using the provided auth service
		debugCtx.LogInfo("Auth middleware: Validating token")
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			debugCtx.LogWithError(err).Error("Auth middleware: Token validation failed")
			writeInvalidTokenError(w, r, "Token validation failed", err.Error())
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		if claims.SecurityLevel != models.SecurityLevelSuperUser {
			writeForbiddenError(w, r, "Forbidden", "User is not a super user")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		return MiddlewareResult{Continue: true}
	})
}

// NewRequireRolePreMiddleware creates a middleware that requires a specific role
func NewRequireRolePreMiddleware(authService auth_interfaces.AuthServiceInterface, requiredRoles []models.Role) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		diag := diagnostics.New("auth_middleware")
		defer diag.Complete()
		//
		// Debug logging
		debugCtx := appctx.FromContext(r.Context())
		debugCtx.LogInfo("Role middleware: Starting role validation")

		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			debugCtx.LogError("Role middleware: Missing Authorization header")
			writeUnauthorizedError(w, r, "Authorization header required", "Missing Authorization header")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeUnauthorizedError(w, r, "Invalid authorization header format", "Expected 'Bearer <token>' format")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			writeUnauthorizedError(w, r, "Empty token", "Token cannot be empty")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Validate the token using the provided auth service
		debugCtx.LogInfo("Role middleware: Validating token")
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			debugCtx.LogWithError(err).Error("Role middleware: Token validation failed")
			writeInvalidTokenError(w, r, "Token validation failed", err.Error())
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Super users can do anything
		if claims.SecurityLevel == models.SecurityLevelSuperUser {
			return MiddlewareResult{Continue: true}
		}

		// Get user to validate roles
		appCtx := appctx.FromContext(r.Context())
		userModel, getUserDiag := authService.GetUserByID(appCtx, claims.TenantID, claims.UserID)
		if getUserDiag != nil && getUserDiag.HasErrors() {
			debugCtx.LogError("Role middleware: Failed to get user from service")
			writeForbiddenError(w, r, "Forbidden", "Failed to get user information")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}
		if userModel == nil {
			debugCtx.LogError("Role middleware: User not found")
			writeForbiddenError(w, r, "Forbidden", "User not found")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Check if user has any of the required roles
		hasRequiredRole := false
		unmatchedRoles := make([]string, 0)
		for i, role := range requiredRoles {
			if hasRole(userModel, role.Name) {
				hasRequiredRole = true
				requiredRoles[i].Matched = true
				break
			}
		}

		for _, role := range requiredRoles {
			if !role.Matched {
				unmatchedRoles = append(unmatchedRoles, role.Name)
			}
		}

		if !hasRequiredRole {
			debugCtx.LogWithFields(logrus.Fields{
				"user_id":        claims.Username,
				"required_roles": requiredRoles,
			}).Error("Role middleware: User does not have some of the required roles")
			writeForbiddenError(w, r, "Forbidden", fmt.Sprintf("User does not have some of the required roles, missing: %v", strings.Join(unmatchedRoles, ", ")))
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		debugCtx.LogWithFields(logrus.Fields{
			"user_id":        claims.Username,
			"required_roles": requiredRoles,
		}).Info("Role middleware: Role validation successful")

		return MiddlewareResult{Continue: true}
	})
}

// NewRequireClaimPreMiddleware creates a middleware that requires a specific claim
func NewRequireClaimPreMiddleware(authService auth_interfaces.AuthServiceInterface, requiredClaims []models.Claim) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		diag := diagnostics.New("auth_middleware")
		defer diag.Complete()
		// using injected authService
		// Debug logging
		debugCtx := appctx.FromContext(r.Context())
		debugCtx.LogInfo("Claim middleware: Starting claim validation")

		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			debugCtx.LogError("Claim middleware: Missing Authorization header")
			writeUnauthorizedError(w, r, "Authorization header required", "Missing Authorization header")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeUnauthorizedError(w, r, "Invalid authorization header format", "Expected 'Bearer <token>' format")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			writeUnauthorizedError(w, r, "Empty token", "Token cannot be empty")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Validate the token using the provided auth service
		debugCtx.LogInfo("Claim middleware: Validating token")
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			debugCtx.LogWithError(err).Error("Claim middleware: Token validation failed")
			writeInvalidTokenError(w, r, "Token validation failed", err.Error())
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		// Super users can do anything
		if claims.SecurityLevel == models.SecurityLevelSuperUser {
			return MiddlewareResult{Continue: true}
		}

		// Get user to validate claims
		appCtx := appctx.FromContext(r.Context())
		currentUser, getUserDiag := authService.GetUserByID(appCtx, claims.TenantID, claims.UserID)
		if getUserDiag != nil && getUserDiag.HasErrors() {
			debugCtx.LogError("Claim middleware: Failed to get user from service")
			writeForbiddenError(w, r, "Forbidden", "Failed to get user information")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}
		if currentUser == nil {
			debugCtx.LogError("Claim middleware: User not found")
			writeForbiddenError(w, r, "Forbidden", "User not found")
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}
		userClaims := currentUser.Claims

		// Check if user has any of the required claims
		hasRequiredClaim := false
		unmatchedClaims := make([]string, 0)
		for _, claim := range requiredClaims {
			if hasClaims(userClaims, claim) {
				hasRequiredClaim = true
				claim.Matched = true
				break
			}
		}

		for _, claim := range requiredClaims {
			if !claim.Matched {
				unmatchedClaims = append(unmatchedClaims, models.GetClaimName(&claim))
			}
		}

		if !hasRequiredClaim {
			debugCtx.LogWithFields(logrus.Fields{
				"user_id":         claims.Username,
				"required_claims": requiredClaims,
			}).Error("Claim middleware: User does not have some of the required claims")
			writeForbiddenError(w, r, "Forbidden", fmt.Sprintf("User does not have some of the required claims, missing: %v", strings.Join(unmatchedClaims, ", ")))
			return MiddlewareResult{Continue: false, Diagnostics: diag}
		}

		debugCtx.LogWithFields(logrus.Fields{
			"user_id":         claims.Username,
			"required_claims": requiredClaims,
		}).Info("Claim middleware: Claim validation successful")

		return MiddlewareResult{Continue: true}
	})
}

// Helper functions for middleware errors (to avoid circular import)
func writeUnauthorizedError(w http.ResponseWriter, r *http.Request, message, details string) {
	errorResponse := map[string]interface{}{
		"error": map[string]string{
			"code":    "UNAUTHORIZED",
			"message": message,
			"details": details,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"path":      r.URL.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(errorResponse)
}

func writeInvalidTokenError(w http.ResponseWriter, r *http.Request, message, details string) {
	errorResponse := map[string]interface{}{
		"error": map[string]string{
			"code":    "INVALID_TOKEN",
			"message": message,
			"details": details,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"path":      r.URL.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusUnauthorized)
	_ = json.NewEncoder(w).Encode(errorResponse)
}

// Helper functions for middleware errors (to avoid circular import)
func writeForbiddenError(w http.ResponseWriter, r *http.Request, message, details string) {
	errorResponse := map[string]interface{}{
		"error": map[string]string{
			"code":    "FORBIDDEN",
			"message": message,
			"details": details,
		},
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"path":      r.URL.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusForbidden)
	_ = json.NewEncoder(w).Encode(errorResponse)
}

// This function will be used to indicate if the user has the necessary role to access the resource
func hasRole(user *models.User, role string) bool {
	for _, r := range user.Roles {
		if strings.EqualFold(r.Slug, role) {
			return true
		}
	}

	return false
}

// this function will be used to indicate if the user has the necessary claims to access the resource
// We will need to break it into the module, service and action and take into account the wildcard
func hasClaims(userClaims []models.Claim, requiredClaim models.Claim) bool {
	// If user has no claims, they can't access anything
	if len(userClaims) == 0 {
		return false
	}

	// Check if user has any claim that matches the required claim

	for _, userClaim := range userClaims {
		if matchesClaim(userClaim, requiredClaim) {
			return true
		}
	}

	return false
}

// matchesClaim checks if a user claim matches a required claim
// It handles wildcards (*) for module, service, and action
func matchesClaim(userClaim, requiredClaim models.Claim) bool {
	// Check module match
	if !matchesField(userClaim.Module, requiredClaim.Module) {
		return false
	}

	// Check service match
	if !matchesField(userClaim.Service, requiredClaim.Service) {
		return false
	}

	// Check action match
	if !matchesAction(userClaim.Action, requiredClaim.Action) {
		return false
	}

	return true
}

// matchesField checks if a user field matches a required field, handling wildcards
func matchesField(userField, requiredField string) bool {
	// If required field is wildcard, it matches anything
	if requiredField == "*" {
		return true
	}

	// If user field is wildcard, it matches anything
	if userField == "*" {
		return true
	}

	// Exact match
	return strings.EqualFold(userField, requiredField)
}

// matchesAction checks if a user action matches a required action, handling special cases
func matchesAction(userAction, requiredAction models.AccessLevel) bool {
	// If required action is wildcard, it matches anything
	if requiredAction == models.AccessLevelAll {
		return true
	}

	// If user action is wildcard, it matches anything
	if userAction == models.AccessLevelAll {
		return true
	}

	// Special case: if required action is read, user can have read or all
	if requiredAction == models.AccessLevelRead {
		return userAction == models.AccessLevelRead || userAction == models.AccessLevelAll
	}

	// Exact match
	return strings.EqualFold(string(userAction), string(requiredAction))
}

// extractAuthorizationHeaders extracts the authorization header from the request
// it will return the authorization type and the token
// if no authorization header is found, it will return an error
func extractAuthorizationHeaders(ctx *appctx.AppContext, r *http.Request) *api_types.AuthorizationHeader {
	authHeader := r.Header.Get("Authorization")
	if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
		token := strings.TrimPrefix(authHeader, "Bearer ")
		response := api_types.AuthorizationHeader{
			AuthorizationType: api_types.AuthorizationHeaderTypeBearer,
			Token:             token,
		}
		ctx.LogInfof("Found Bearer authorization header: %s", utils.ObfuscateString(token))

		return &response
	}

	// we did not find a bearer token so trying to extract the api key
	apiKey := r.Header.Get(config.ApiKeyAuthorizationHeader)
	if apiKey != "" {
		response := api_types.AuthorizationHeader{
			AuthorizationType: api_types.AuthorizationHeaderTypeApiKey,
			Token:             apiKey,
		}
		ctx.LogInfof("Found API key authorization header: %s", utils.ObfuscateString(apiKey))

		return &response
	}

	// we did not find a bearer token or api key so returning an error
	response := api_types.AuthorizationHeader{
		AuthorizationType: api_types.AuthorizationHeaderTypeNone,
		Token:             "",
	}
	ctx.LogInfo("No authorization header found")

	return &response
}

func validateBearerToken(ctx *appctx.AppContext, r *http.Request, authService auth_interfaces.AuthServiceInterface, route *api_types.Route, authHeader string) (bool, error) {
	// Validate the token using the provided auth service
	ctx.LogInfo("Auth middleware: Validating token")
	claims, err := authService.ValidateToken(authHeader)
	if err != nil {
		ctx.LogWithError(err).Error("Auth middleware: Token validation failed")
		return false, err
	}
	ctx.LogInfo("Auth middleware: Token validation successful")

	// Super users can do anything
	if claims.SecurityLevel == models.SecurityLevelSuperUser {
		// Add claims to context using AppContext
		appCtx := appctx.FromContext(ctx.Context)
		appCtx = appCtx.WithTenantID(claims.TenantID)
		appCtx = appCtx.WithUserID(claims.UserID)
		appCtx = appCtx.WithUsername(claims.Username)

		// Add claims to the underlying context for backward compatibility
		// We need to update the AppContext's underlying context directly
		appCtx.Context = context.WithValue(appCtx.Context, authctx.ClaimsKey, claims)
		appCtx.Context = context.WithValue(appCtx.Context, types.TenantIDKey, claims.TenantID)
		appCtx.Context = context.WithValue(appCtx.Context, types.UserIDKey, claims.Username)

		*r = *r.WithContext(appCtx)

		return true, nil
	}

	// checking if this is just a superuser endpoint and we are not superuser
	if route.SecurityRequirement.SecurityLevel == models.ApiKeySecurityLevelSuperUser &&
		claims.SecurityLevel != models.SecurityLevelSuperUser {
		return false, errors.New("this endpoint is only available to superusers")
	}

	// now we will check the required claims and role if they exist
	var currentUser *models.User
	var currentUserDiag *diagnostics.Diagnostics
	if route.SecurityRequirement.Claims != nil || route.SecurityRequirement.Roles != nil {
		currentUser, currentUserDiag = authService.GetUserByID(ctx, claims.TenantID, claims.UserID)
		if currentUserDiag.HasErrors() {
			return false, errors.New("failed to get user by id")
		}
		if currentUser == nil {
			return false, errors.New("user not found")
		}
	}

	// first we will check if the user has any of the required roles
	// roles are always a or relation, you just need to have one of the roles
	if route.SecurityRequirement.Roles != nil {
		for _, role := range route.SecurityRequirement.Roles.Items {
			// if the user is nil, we will return an error and we will not continue
			if !hasRole(currentUser, role.Name) {
				return false, errors.New("user does not have the required role")
			}
		}
	}

	// now we will check if the user has any of the required claims
	// here we need to check what type of relation we have set, if none is set, we will assume and
	// if we have a relation set, we will check if the user has all the claims
	if route.SecurityRequirement.Claims != nil {
		switch route.SecurityRequirement.Claims.Relation {
		case api_types.SecurityRequirementRelationAnd:
			for _, claim := range route.SecurityRequirement.Claims.Items {
				if !userHasClaims(currentUser, claim) {
					return false, errors.New("user does not have the required claim")
				}
			}
		case api_types.SecurityRequirementRelationOr:
			hasAnyClaim := false
			for _, claim := range route.SecurityRequirement.Claims.Items {
				if userHasClaims(currentUser, claim) {
					hasAnyClaim = true
					break
				}
			}
			if !hasAnyClaim {
				return false, errors.New("user does not have the required claim")
			}
		default:
			return false, errors.New("invalid claim relation")
		}
	}

	// Add claims to context using AppContext
	appCtx := appctx.FromContext(ctx.Context)
	appCtx = appCtx.WithTenantID(claims.TenantID)
	appCtx = appCtx.WithUserID(claims.UserID)
	appCtx = appCtx.WithUsername(claims.Username)

	// Add claims to the underlying context for backward compatibility
	// We need to update the AppContext's underlying context directly
	appCtx.Context = context.WithValue(appCtx.Context, authctx.ClaimsKey, claims)
	appCtx.Context = context.WithValue(appCtx.Context, types.TenantIDKey, claims.TenantID)
	appCtx.Context = context.WithValue(appCtx.Context, types.UserIDKey, claims.Username)

	*r = *r.WithContext(appCtx)

	// Debug logging to verify claims are set
	appCtx.LogWithFields(logrus.Fields{
		"tenant_id":         claims.TenantID,
		"user_id":           claims.UserID,
		"username":          claims.Username,
		"roles":             claims.Roles,
		"auth_context_addr": fmt.Sprintf("%p", appCtx),
	}).Info("Auth middleware: Claims set in context")

	return true, nil
}

// validateApiKey validates an API key header and sets context if valid
func validateApiKey(ctx *appctx.AppContext, r *http.Request, authService auth_interfaces.AuthServiceInterface, route *api_types.Route, apiKey string) (bool, error) {
	// Build credentials
	tenantID := ctx.GetTenantID()
	creds := auth_models.APIKeyCredentials{
		APIKey:   strings.TrimSpace(apiKey),
		TenantID: tenantID,
	}

	// Authenticate via service
	token, diag := authService.AuthenticateWithAPIKey(ctx, creds)
	if diag != nil && diag.HasErrors() {
		return false, fmt.Errorf(diag.GetSummary())
	}
	if token == nil || token.Token == "" {
		return false, errors.New("invalid API key")
	}

	// Create synthetic claims from token
	claims := &auth_models.AuthClaims{
		Username:  token.Username,
		UserID:    token.UserID,
		ExpiresAt: token.ExpiresAt.Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "api", // informational
		Roles:     []string{},
		TenantID:  token.TenantID,
		AuthType:  "api_key",
		APIKeyID:  "",
	}

	// Super users can do anything; note security checked via bearer tokens typically. For API keys we rely on route checks below.

	// If route requires superuser explicitly, we need to load user and check security level
	if route.SecurityRequirement != nil && route.SecurityRequirement.SecurityLevel == models.ApiKeySecurityLevelSuperUser {
		currentUser, userDiag := authService.GetUserByID(ctx, token.TenantID, token.UserID)
		if userDiag != nil && userDiag.HasErrors() {
			return false, errors.New("failed to get user by id")
		}
		if currentUser == nil {
			return false, errors.New("user not found")
		}
		// Require superuser level
		isSuper := false
		for _, r := range currentUser.Roles {
			if r.SecurityLevel == models.SecurityLevelSuperUser {
				isSuper = true
				break
			}
		}
		if !isSuper {
			return false, errors.New("this endpoint is only available to superusers")
		}
	}

	// Check role/claim requirements if present
	var currentUser *models.User
	var userDiag *diagnostics.Diagnostics
	if route.SecurityRequirement != nil && (route.SecurityRequirement.Claims != nil || route.SecurityRequirement.Roles != nil) {
		currentUser, userDiag = authService.GetUserByID(ctx, token.TenantID, token.UserID)
		if userDiag != nil && userDiag.HasErrors() {
			return false, errors.New("failed to get user by id")
		}
		if currentUser == nil {
			return false, errors.New("user not found")
		}
	}

	if route.SecurityRequirement != nil && route.SecurityRequirement.Roles != nil {
		for _, role := range route.SecurityRequirement.Roles.Items {
			if !hasRole(currentUser, role.Name) {
				return false, errors.New("user does not have the required role")
			}
		}
	}

	if route.SecurityRequirement != nil && route.SecurityRequirement.Claims != nil {
		switch route.SecurityRequirement.Claims.Relation {
		case api_types.SecurityRequirementRelationAnd:
			for _, claim := range route.SecurityRequirement.Claims.Items {
				if !userHasClaims(currentUser, claim) {
					return false, errors.New("user does not have the required claim")
				}
			}
		case api_types.SecurityRequirementRelationOr:
			hasAnyClaim := false
			for _, claim := range route.SecurityRequirement.Claims.Items {
				if userHasClaims(currentUser, claim) {
					hasAnyClaim = true
					break
				}
			}
			if !hasAnyClaim {
				return false, errors.New("user does not have the required claim")
			}
		default:
			return false, errors.New("invalid claim relation")
		}
	}

	// Set context with claims for downstream handlers
	appCtx := appctx.FromContext(ctx.Context)
	appCtx = appCtx.WithTenantID(token.TenantID)
	appCtx = appCtx.WithUserID(token.UserID)
	appCtx = appCtx.WithUsername(token.Username)
	appCtx.Context = context.WithValue(appCtx.Context, authctx.ClaimsKey, claims)
	appCtx.Context = context.WithValue(appCtx.Context, types.TenantIDKey, token.TenantID)
	appCtx.Context = context.WithValue(appCtx.Context, types.UserIDKey, token.Username)
	*r = *r.WithContext(appCtx)

	return true, nil
}

// this function will be used to check if the user has the necessary claims to access the resource
// it will return true if the user has the necessary claims, false otherwise
func userHasClaims(user *models.User, claim models.Claim) bool {
	if user == nil {
		return false
	}

	for _, userClaim := range user.Claims {
		if userClaim.CanAccess(&claim) {
			return true
		}
	}

	return false
}
