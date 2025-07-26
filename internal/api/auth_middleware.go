package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/auth"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/sirupsen/logrus"
)

// NewRequireAuthPreMiddleware creates a pre-middleware that validates JWT tokens
func NewRequireAuthPreMiddleware(authService *auth.AuthService) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		// Debug logging to see if auth middleware is being called
		debugCtx := appctx.FromContext(r.Context())
		debugCtx.LogInfo("Auth middleware: Starting authentication")

		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			debugCtx.LogError("Auth middleware: Missing Authorization header")
			writeUnauthorizedError(w, r, "Authorization header required", "Missing Authorization header")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("missing authorization header")}
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeUnauthorizedError(w, r, "Invalid authorization header format", "Expected 'Bearer <token>' format")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("invalid authorization header format")}
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			writeUnauthorizedError(w, r, "Empty token", "Token cannot be empty")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("empty token")}
		}

		// Validate the token using the provided auth service
		debugCtx.LogInfo("Auth middleware: Validating token")
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			debugCtx.LogWithError(err).Error("Auth middleware: Token validation failed")
			writeInvalidTokenError(w, r, "Token validation failed", err.Error())
			return MiddlewareResult{Continue: false, Error: err}
		}
		debugCtx.LogInfo("Auth middleware: Token validation successful")

		// Add claims to context using AppContext
		appCtx := appctx.FromContext(r.Context())
		appCtx = appCtx.WithTenantID(claims.TenantID)
		appCtx = appCtx.WithUserID(claims.Username)

		// Add claims to the underlying context for backward compatibility
		// We need to update the AppContext's underlying context directly
		appCtx.Context = context.WithValue(appCtx.Context, auth.ClaimsKey, claims)
		appCtx.Context = context.WithValue(appCtx.Context, types.TenantIDKey, claims.TenantID)
		appCtx.Context = context.WithValue(appCtx.Context, types.UserIDKey, claims.Username)

		*r = *r.WithContext(appCtx)

		// Debug logging to verify claims are set
		appCtx.LogWithFields(logrus.Fields{
			"tenant_id":         claims.TenantID,
			"user_id":           claims.Username,
			"roles":             claims.Roles,
			"auth_context_addr": fmt.Sprintf("%p", appCtx),
		}).Info("Auth middleware: Claims set in context")

		return MiddlewareResult{Continue: true}
	})
}

func NewRequireSuperUserPreMiddleware(authService *auth.AuthService) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		// Debug logging to see if auth middleware is being called
		debugCtx := appctx.FromContext(r.Context())
		debugCtx.LogInfo("Auth middleware: Starting authentication")

		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			debugCtx.LogError("Auth middleware: Missing Authorization header")
			writeUnauthorizedError(w, r, "Authorization header required", "Missing Authorization header")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("missing authorization header")}
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeUnauthorizedError(w, r, "Invalid authorization header format", "Expected 'Bearer <token>' format")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("invalid authorization header format")}
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			writeUnauthorizedError(w, r, "Empty token", "Token cannot be empty")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("empty token")}
		}

		// Validate the token using the provided auth service
		debugCtx.LogInfo("Auth middleware: Validating token")
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			debugCtx.LogWithError(err).Error("Auth middleware: Token validation failed")
			writeInvalidTokenError(w, r, "Token validation failed", err.Error())
			return MiddlewareResult{Continue: false, Error: err}
		}

		if !claims.IsSuperUser {
			writeForbiddenError(w, r, "Forbidden", "User is not a super user")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("user is not a super user")}
		}

		return MiddlewareResult{Continue: true}
	})
}

// NewRequireRolePreMiddleware creates a middleware that requires a specific role
func NewRequireRolePreMiddleware(requiredRoles []models.Role) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		authService := auth.GetInstance()
		// Debug logging
		debugCtx := appctx.FromContext(r.Context())
		debugCtx.LogInfo("Role middleware: Starting role validation")

		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			debugCtx.LogError("Role middleware: Missing Authorization header")
			writeUnauthorizedError(w, r, "Authorization header required", "Missing Authorization header")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("missing authorization header")}
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeUnauthorizedError(w, r, "Invalid authorization header format", "Expected 'Bearer <token>' format")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("invalid authorization header format")}
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			writeUnauthorizedError(w, r, "Empty token", "Token cannot be empty")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("empty token")}
		}

		// Validate the token using the provided auth service
		debugCtx.LogInfo("Role middleware: Validating token")
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			debugCtx.LogWithError(err).Error("Role middleware: Token validation failed")
			writeInvalidTokenError(w, r, "Token validation failed", err.Error())
			return MiddlewareResult{Continue: false, Error: err}
		}

		// Super users can do anything
		if claims.IsSuperUser {
			return MiddlewareResult{Continue: true}
		}

		// Get user from database to validate roles
		appCtx := appctx.FromContext(r.Context())
		user, err := authService.AuthDataStore.GetUserByUsername(appCtx, claims.Username)
		if err != nil {
			debugCtx.LogWithError(err).Error("Role middleware: Failed to get user from database")
			writeForbiddenError(w, r, "Forbidden", "Failed to get user information")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("failed to get user: %w", err)}
		}

		if user == nil {
			debugCtx.LogError("Role middleware: User not found in database")
			writeForbiddenError(w, r, "Forbidden", "User not found")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("user not found")}
		}

		// Convert entity to model for role checking
		userModel := mappers.MapUserToDto(user)

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
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("user does not have some of the required roles, missing: %v", strings.Join(unmatchedRoles, ", "))}
		}

		debugCtx.LogWithFields(logrus.Fields{
			"user_id":        claims.Username,
			"required_roles": requiredRoles,
		}).Info("Role middleware: Role validation successful")

		return MiddlewareResult{Continue: true}
	})
}

// NewRequireClaimPreMiddleware creates a middleware that requires a specific claim
func NewRequireClaimPreMiddleware(requiredClaims []models.Claim) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		authService := auth.GetInstance()
		// Debug logging
		debugCtx := appctx.FromContext(r.Context())
		debugCtx.LogInfo("Claim middleware: Starting claim validation")

		// Get the Authorization header
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			debugCtx.LogError("Claim middleware: Missing Authorization header")
			writeUnauthorizedError(w, r, "Authorization header required", "Missing Authorization header")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("missing authorization header")}
		}

		// Check if it starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			writeUnauthorizedError(w, r, "Invalid authorization header format", "Expected 'Bearer <token>' format")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("invalid authorization header format")}
		}

		// Extract the token
		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == "" {
			writeUnauthorizedError(w, r, "Empty token", "Token cannot be empty")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("empty token")}
		}

		// Validate the token using the provided auth service
		debugCtx.LogInfo("Claim middleware: Validating token")
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			debugCtx.LogWithError(err).Error("Claim middleware: Token validation failed")
			writeInvalidTokenError(w, r, "Token validation failed", err.Error())
			return MiddlewareResult{Continue: false, Error: err}
		}

		// Super users can do anything
		if claims.IsSuperUser {
			return MiddlewareResult{Continue: true}
		}

		// Get user from database to validate claims
		appCtx := appctx.FromContext(r.Context())
		user, err := authService.AuthDataStore.GetUserByUsername(appCtx, claims.Username)
		if err != nil {
			debugCtx.LogWithError(err).Error("Claim middleware: Failed to get user from database")
			writeForbiddenError(w, r, "Forbidden", "Failed to get user information")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("failed to get user: %w", err)}
		}

		if user == nil {
			debugCtx.LogError("Claim middleware: User not found in database")
			writeForbiddenError(w, r, "Forbidden", "User not found")
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("user not found")}
		}

		// Convert entity to model for claim checking
		userModel := mappers.MapUserToDto(user)

		// Check if user has any of the required claims
		hasRequiredClaim := false
		unmatchedClaims := make([]string, 0)
		for _, claim := range requiredClaims {
			if hasClaims(userModel, claim) {
				hasRequiredClaim = true
				claim.Matched = true
				break
			}
		}

		for _, claim := range requiredClaims {
			if !claim.Matched {
				unmatchedClaims = append(unmatchedClaims, claim.GetName())
			}
		}

		if !hasRequiredClaim {
			debugCtx.LogWithFields(logrus.Fields{
				"user_id":         claims.Username,
				"required_claims": requiredClaims,
			}).Error("Claim middleware: User does not have some of the required claims")
			writeForbiddenError(w, r, "Forbidden", fmt.Sprintf("User does not have some of the required claims, missing: %v", strings.Join(unmatchedClaims, ", ")))
			return MiddlewareResult{Continue: false, Error: fmt.Errorf("user does not have some of the required claims, missing: %v", strings.Join(unmatchedClaims, ", "))}
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
func hasClaims(user *models.User, requiredClaim models.Claim) bool {
	// If user has no claims, they can't access anything
	if len(user.Claims) == 0 {
		return false
	}

	// Check if user has any claim that matches the required claim

	for _, userClaim := range user.Claims {
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
func matchesAction(userAction, requiredAction models.ClaimAction) bool {
	// If required action is wildcard, it matches anything
	if requiredAction == models.ClaimActionAll {
		return true
	}

	// If user action is wildcard, it matches anything
	if userAction == models.ClaimActionAll {
		return true
	}

	// Special case: if required action is read, user can have read or all
	if requiredAction == models.ClaimActionRead {
		return userAction == models.ClaimActionRead || userAction == models.ClaimActionAll
	}

	// Exact match
	return strings.EqualFold(string(userAction), string(requiredAction))
}
