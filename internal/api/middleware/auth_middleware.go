package middleware

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	activity_interfaces "github.com/cjlapao/locally-cli/internal/activity/interfaces"
	activity_types "github.com/cjlapao/locally-cli/internal/activity/types"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	auth_interfaces "github.com/cjlapao/locally-cli/internal/auth/interfaces"
	authctx "github.com/cjlapao/locally-cli/internal/authctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/cjlapao/locally-cli/pkg/utils"

	"github.com/sirupsen/logrus"
)

func NewAuthorizationPreMiddleware(authService auth_interfaces.AuthServiceInterface, activityService activity_interfaces.ActivityServiceInterface, route *api_types.Route) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		// Debug logging to see if auth middleware is being called
		debugCtx := appctx.FromContext(r.Context())
		debugCtx.LogInfo("Auth middleware: Starting authentication")
		diag := diagnostics.New("auth_middleware")
		defer diag.Complete()

		tenantID := debugCtx.GetTenantID()
		if tenantID == "" {
			tenantID = r.Header.Get(config.TenantIDHeader)
		}

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
			// check if the security level is api key, meaning a user token is not allowed
			if route.SecurityRequirement.SecurityLevel == models.ApiKeySecurityLevelApiKey {
				writeUnauthorizedError(w, r, "User token not allowed", "User token not allowed on this endpoint, you should use an API key instead")
				return MiddlewareResult{Continue: false, Diagnostics: diag}
			}
			valid, err := validateBearerToken(debugCtx, r, authService, activityService, route, authHeader.Token)
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
			// check if the security level is api key, meaning a user token is not allowed
			if route.SecurityRequirement.SecurityLevel == models.ApiKeySecurityLevelBearer {
				writeUnauthorizedError(w, r, "User token not allowed", "User token not allowed on this endpoint, you should use a bearer token instead")
				return MiddlewareResult{Continue: false, Diagnostics: diag}
			}
			valid, validateDiag := validateApiKey(debugCtx, r, tenantID, authService, activityService, route, authHeader.Token)
			if validateDiag != nil && validateDiag.HasErrors() {
				diag.Append(validateDiag)
				getLastError := validateDiag.Errors[0].Message
				getLastErrorCode := validateDiag.Errors[0].Code
				writeUnauthorizedError(w, r, getLastErrorCode, getLastError)
				return MiddlewareResult{Continue: false, Diagnostics: diag}
			}
			if !valid {
				writeUnauthorizedError(w, r, "Invalid API key", "Invalid API key")
				return MiddlewareResult{Continue: false, Diagnostics: diag}
			}
			authService.UpdateApiKeyLastUsed(debugCtx, tenantID, authHeader.Token)
			activityService.RecordSuccessActivity(debugCtx, "api_key_used", &activity_types.ActivityRecord{
				TenantID:      tenantID,
				ActorID:       authHeader.Token,
				ActorName:     debugCtx.GetUsername(),
				Module:        "api",
				Service:       "auth_middleware",
				Success:       true,
				ActorType:     activity_types.ActorTypeUser,
				ActivityType:  activity_types.ActivityTypeAudit,
				ActivityLevel: activity_types.ActivityLevelInfo,
				Data: &activity_types.ActivityData{
					Metadata: map[string]interface{}{
						"api_key": authHeader.Token,
					},
				},
			})
		}

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

func validateBearerToken(ctx *appctx.AppContext, r *http.Request, authService auth_interfaces.AuthServiceInterface, activityService activity_interfaces.ActivityServiceInterface, route *api_types.Route, authHeader string) (bool, error) {
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
		activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
			ErrorCode:    "UNAUTHORIZED",
			ErrorMessage: "this endpoint is only available to superusers",
			StatusCode:   401,
		}, &activity_types.ActivityRecord{
			TenantID:     claims.TenantID,
			ActorID:      claims.UserID,
			ActorName:    claims.Username,
			Module:       "api",
			Service:      "auth_middleware",
			Success:      false,
			ActorType:    activity_types.ActorTypeUser,
			ActivityType: activity_types.ActivityTypeAudit,
		})
		return false, errors.New("this endpoint is only available to superusers")
	}

	// now we will check the required claims and role if they exist
	var currentUser *models.User
	var currentUserDiag *diagnostics.Diagnostics
	if route.SecurityRequirement.Claims != nil || route.SecurityRequirement.Roles != nil {
		currentUser, currentUserDiag = authService.GetUserByID(ctx, claims.TenantID, claims.UserID)
		if currentUserDiag.HasErrors() {
			activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
				ErrorCode:    "INTERNAL_SERVER_ERROR",
				ErrorMessage: "failed to get user by id",
				StatusCode:   500,
			}, &activity_types.ActivityRecord{
				TenantID:     claims.TenantID,
				ActorID:      claims.UserID,
				ActorName:    claims.Username,
				Module:       "api",
				Service:      "auth_middleware",
				Success:      false,
				ActorType:    activity_types.ActorTypeUser,
				ActivityType: activity_types.ActivityTypeAudit,
			})
			return false, errors.New("failed to get user by id")
		}
		if currentUser == nil {
			activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
				ErrorCode:    "UNAUTHORIZED",
				ErrorMessage: "user not found",
				StatusCode:   401,
			}, &activity_types.ActivityRecord{
				TenantID:     claims.TenantID,
				ActorID:      claims.UserID,
				ActorName:    claims.Username,
				Module:       "api",
				Service:      "auth_middleware",
				Success:      false,
				ActorType:    activity_types.ActorTypeUser,
				ActivityType: activity_types.ActivityTypeAudit,
			})
			return false, errors.New("user not found")
		}
	}

	// first we will check if the user has any of the required roles
	// roles are always a or relation, you just need to have one of the roles
	if route.SecurityRequirement.Roles != nil {
		for _, role := range route.SecurityRequirement.Roles.Items {
			// if the user is nil, we will return an error and we will not continue
			if !hasRole(currentUser, role.Name) {
				activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
					ErrorCode:    "UNAUTHORIZED",
					ErrorMessage: "user does not have the required role",
					StatusCode:   401,
				}, &activity_types.ActivityRecord{
					TenantID:     claims.TenantID,
					ActorID:      claims.UserID,
					ActorName:    claims.Username,
					Module:       "api",
					Service:      "auth_middleware",
					Success:      false,
					ActorType:    activity_types.ActorTypeUser,
					ActivityType: activity_types.ActivityTypeAudit,
				})
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
					activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
						ErrorCode:    "UNAUTHORIZED",
						ErrorMessage: "user does not have the required claim",
						StatusCode:   401,
					}, &activity_types.ActivityRecord{
						TenantID:     claims.TenantID,
						ActorID:      claims.UserID,
						ActorName:    claims.Username,
						Module:       "api",
						Service:      "auth_middleware",
						Success:      false,
						ActorType:    activity_types.ActorTypeUser,
						ActivityType: activity_types.ActivityTypeAudit,
					})
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
				activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
					ErrorCode:    "UNAUTHORIZED",
					ErrorMessage: "user does not have the required claim",
					StatusCode:   401,
				}, &activity_types.ActivityRecord{
					TenantID:     claims.TenantID,
					ActorID:      claims.UserID,
					ActorName:    claims.Username,
					Module:       "api",
					Service:      "auth_middleware",
					Success:      false,
					ActorType:    activity_types.ActorTypeUser,
					ActivityType: activity_types.ActivityTypeAudit,
				})
				return false, errors.New("user does not have the required claim")
			}
		default:
			activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
				ErrorCode:    "INTERNAL_SERVER_ERROR",
				ErrorMessage: "invalid claim relation",
				StatusCode:   500,
			}, &activity_types.ActivityRecord{
				TenantID:     claims.TenantID,
				ActorID:      claims.UserID,
				ActorName:    claims.Username,
				Module:       "api",
				Service:      "auth_middleware",
				Success:      false,
				ActorType:    activity_types.ActorTypeUser,
				ActivityType: activity_types.ActivityTypeAudit,
			})
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
func validateApiKey(ctx *appctx.AppContext, r *http.Request, tenantID string, authService auth_interfaces.AuthServiceInterface, activityService activity_interfaces.ActivityServiceInterface, route *api_types.Route, apiKey string) (bool, *diagnostics.Diagnostics) {
	diag := diagnostics.New("validate_api_key")
	defer diag.Complete()

	// Validate raw api key
	apiKeyDto, validateDiag := authService.ValidateApiKey(ctx, tenantID, apiKey)
	if validateDiag != nil && validateDiag.HasErrors() {
		diag.Append(validateDiag)
		activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
			ErrorCode:    "INTERNAL_SERVER_ERROR",
			ErrorMessage: "failed to validate api key",
			StatusCode:   500,
		}, &activity_types.ActivityRecord{
			TenantID:     tenantID,
			ActorID:      apiKey,
			ActorName:    "api_key",
			Module:       "api",
			Service:      "auth_middleware",
			Success:      false,
			ActorType:    activity_types.ActorTypeUser,
			ActivityType: activity_types.ActivityTypeAudit,
		})
		return false, diag
	}
	if apiKeyDto == nil {
		diag.AddError("invalid_api_key", "Invalid API key", "api_key", nil)
		activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
			ErrorCode:    "INVALID_API_KEY",
			ErrorMessage: "Invalid API key",
			StatusCode:   401,
		}, &activity_types.ActivityRecord{
			TenantID:     tenantID,
			ActorID:      apiKey,
			ActorName:    "api_key",
			Module:       "api",
			Service:      "auth_middleware",
			Success:      false,
			ActorType:    activity_types.ActorTypeUser,
			ActivityType: activity_types.ActivityTypeAudit,
		})
		return false, diag
	}

	// Enforce tenant scoping: header X-Tenant-ID must match the key's tenant
	headerTenant := r.Header.Get(config.TenantIDHeader)
	if headerTenant != "" && headerTenant != apiKeyDto.TenantID {
		diag.AddError("tenant_mismatch", "Tenant mismatch", "api_key", nil)
		activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
			ErrorCode:    "UNAUTHORIZED",
			ErrorMessage: "Tenant mismatch",
			StatusCode:   401,
		}, &activity_types.ActivityRecord{
			TenantID:     tenantID,
			ActorID:      apiKey,
			ActorName:    "api_key",
			Module:       "api",
			Service:      "auth_middleware",
			Success:      false,
			ActorType:    activity_types.ActorTypeUser,
			ActivityType: activity_types.ActivityTypeAudit,
		})
		return false, diag
	}

	// Roles: API keys always pass role assignment per requirements

	// Claims: enforce route claims
	if route.SecurityRequirement != nil && route.SecurityRequirement.Claims != nil {
		switch route.SecurityRequirement.Claims.Relation {
		case api_types.SecurityRequirementRelationAnd:
			for _, claim := range route.SecurityRequirement.Claims.Items {
				if !hasClaims(apiKeyDto.Claims, claim) {
					diag.AddError("api_key_does_not_have_required_claim", "API key does not have the required claim", "api_key", nil)
					activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
						ErrorCode:    "UNAUTHORIZED",
						ErrorMessage: "API key does not have the required claim",
						StatusCode:   401,
					}, &activity_types.ActivityRecord{
						TenantID:     tenantID,
						ActorID:      apiKey,
						ActorName:    "api_key",
						Module:       "api",
						Service:      "auth_middleware",
						Success:      false,
						ActorType:    activity_types.ActorTypeUser,
						ActivityType: activity_types.ActivityTypeAudit,
					})
					return false, diag
				}
			}
		case api_types.SecurityRequirementRelationOr:
			hasAnyClaim := false
			for _, claim := range route.SecurityRequirement.Claims.Items {
				if hasClaims(apiKeyDto.Claims, claim) {
					hasAnyClaim = true
					break
				}
			}
			if !hasAnyClaim {
				diag.AddError("api_key_does_not_have_required_claim", "API key does not have the required claim", "api_key", nil)
				activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
					ErrorCode:    "UNAUTHORIZED",
					ErrorMessage: "API key does not have the required claim",
					StatusCode:   401,
				}, &activity_types.ActivityRecord{
					TenantID:     tenantID,
					ActorID:      apiKey,
					ActorName:    "api_key",
					Module:       "api",
					Service:      "auth_middleware",
					Success:      false,
					ActorType:    activity_types.ActorTypeUser,
					ActivityType: activity_types.ActivityTypeAudit,
				})
				return false, diag
			}
		default:
			diag.AddError("invalid_claim_relation", "Invalid claim relation", "api_key", nil)
			activityService.RecordFailureActivity(ctx, activity_types.ActivityTypeAudit, activity_types.ActivityErrorData{
				ErrorCode:    "INTERNAL_SERVER_ERROR",
				ErrorMessage: "Invalid claim relation",
				StatusCode:   500,
			}, &activity_types.ActivityRecord{
				TenantID:     tenantID,
				ActorID:      apiKey,
				ActorName:    "api_key",
				Module:       "api",
				Service:      "auth_middleware",
				Success:      false,
				ActorType:    activity_types.ActorTypeUser,
				ActivityType: activity_types.ActivityTypeAudit,
			})
			return false, diag
		}
	}

	// Set context: tenant from key overrides everything; username uses api key name with prefix
	appCtx := appctx.FromContext(ctx.Context)
	appCtx = appCtx.WithTenantID(apiKeyDto.TenantID)
	appCtx = appCtx.WithUserID("")
	appCtx = appCtx.WithUsername("api_key:" + apiKeyDto.Name)
	// Optionally add metadata for principal type and key prefix
	appCtx = appCtx.WithMetadata("principal_type", "api_key")
	appCtx = appCtx.WithMetadata("api_key_prefix", apiKeyDto.KeyPrefix)
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
