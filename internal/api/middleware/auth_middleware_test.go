package middleware

// import (
// 	"context"
// 	"net/http"
// 	"net/http/httptest"
// 	"testing"
// 	"time"

// 	"github.com/golang-jwt/jwt/v5"
// 	"github.com/stretchr/testify/assert"

// 	"github.com/cjlapao/locally-cli/internal/appctx"
// 	"github.com/cjlapao/locally-cli/internal/auth"
// 	"github.com/cjlapao/locally-cli/internal/config"
// 	"github.com/cjlapao/locally-cli/internal/database/mocks"
// 	"github.com/cjlapao/locally-cli/pkg/models"
// 	"github.com/cjlapao/locally-cli/pkg/types"
// )

// type MockAuthDataStore struct {
// 	*mocks.BaseMockStore
// }

// func TestNewRequireAuthPreMiddleware_PreservesAppContext(t *testing.T) {
// 	// Create a mock auth service
// 	mockStore := &MockAuthDataStore{BaseMockStore: mocks.NewBaseMockStore()}
// 	authService, diag := auth.Initialize(auth.AuthServiceConfig{
// 		SecretKey: []byte("test-secret-key"),
// 		Issuer:    "test-issuer",
// 	}, mockStore, mockStore, mockStore)
// 	assert.False(t, diag.HasErrors())

// 	// Apply the auth middleware
// 	middleware := NewRequireAuthPreMiddleware(authService)

// 	// Verify the middleware can be created
// 	assert.NotNil(t, middleware)

// 	// Note: Full testing of context preservation would require a real JWT token
// 	// and integration testing with the actual middleware chain
// }

// func TestAuthMiddleware_WithValidToken(t *testing.T) {
// 	mockStore := &MockAuthDataStore{BaseMockStore: mocks.NewBaseMockStore()}
// 	// Create a mock auth service
// 	authService, diag := auth.Initialize(auth.AuthServiceConfig{
// 		SecretKey: []byte("test-secret-key"),
// 		Issuer:    "test-issuer",
// 	}, mockStore, mockStore, mockStore)
// 	assert.False(t, diag.HasErrors())

// 	// Create the auth middleware
// 	middleware := NewRequireAuthPreMiddleware(authService)

// 	// Verify the middleware can be created
// 	assert.NotNil(t, middleware)

// 	// Note: Full testing of context preservation would require a real JWT token
// 	// and integration testing with the actual middleware chain
// }

// func TestAuthMiddleware_PreservesContextThroughoutRequest(t *testing.T) {
// 	// Create a mock auth service
// 	mockStore := &MockAuthDataStore{BaseMockStore: mocks.NewBaseMockStore()}
// 	cfgSvc, _ := config.Initialize()
// 	cfg := cfgSvc.Get()
// 	cfg.Set(config.JwtAuthSecretKey, "test-secret-key")
// 	authService, diag := auth.Initialize(auth.AuthServiceConfig{
// 		SecretKey: []byte("test-secret-key"),
// 		Issuer:    "test-issuer",
// 	}, mockStore, mockStore, mockStore)
// 	assert.False(t, diag.HasErrors())

// 	// Create a valid token for testing
// 	claims := &auth.AuthClaims{
// 		Username:  "test-user",
// 		TenantID:  "test-tenant",
// 		Roles:     []string{"user"},
// 		ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
// 		IssuedAt:  time.Now().Unix(),
// 		Issuer:    "test-issuer",
// 	}

// 	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 		"username":  claims.Username,
// 		"tenant_id": claims.TenantID,
// 		"roles":     claims.Roles,
// 		"exp":       claims.ExpiresAt,
// 		"iat":       claims.IssuedAt,
// 		"iss":       claims.Issuer,
// 	})

// 	tokenString, err := token.SignedString([]byte(cfg.GetString(config.JwtAuthSecretKey, "")))
// 	assert.NoError(t, err)

// 	// Create middleware
// 	middleware := NewRequireAuthPreMiddleware(authService)

// 	// Create a request with the token
// 	req := httptest.NewRequest("GET", "/api/v1/environment/vaults", nil)
// 	req.Header.Set("Authorization", "Bearer "+tokenString)
// 	req.Header.Set("X-Tenant-ID", claims.TenantID)

// 	// Create a response recorder
// 	w := httptest.NewRecorder()

// 	// Track context values at different stages
// 	var requestContext context.Context

// 	// Create a handler that captures the context
// 	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		requestContext = r.Context()

// 		// Simulate some processing
// 		time.Sleep(1 * time.Millisecond)

// 		// Check context values in the handler
// 		appCtx := appctx.FromContext(r.Context())
// 		assert.Equal(t, "test-tenant", appCtx.GetTenantID())
// 		assert.Equal(t, "test-user", appCtx.GetUserID())

// 		w.WriteHeader(http.StatusOK)
// 		w.Write([]byte("[]"))
// 	})

// 	// Create a middleware chain to test the auth middleware
// 	chain := NewMiddlewareChain()
// 	chain.AddPreMiddleware(middleware)

// 	// Apply the middleware chain
// 	chain.Execute(handler).ServeHTTP(w, req)

// 	// Verify the response
// 	assert.Equal(t, http.StatusOK, w.Code)

// 	// Verify the context was properly set
// 	assert.NotNil(t, requestContext)

// 	// Extract AppContext from the request context
// 	appCtx := appctx.FromContext(requestContext)

// 	// Verify tenant_id and user_id are preserved
// 	assert.Equal(t, "test-tenant", appCtx.GetTenantID())
// 	assert.Equal(t, "test-user", appCtx.GetUserID())

// 	// Verify the underlying context also has the values
// 	assert.Equal(t, "test-tenant", requestContext.Value(types.TenantIDKey))
// 	assert.Equal(t, "test-user", requestContext.Value(types.UserIDKey))

// 	// Verify claims are also accessible
// 	authClaims := auth.GetClaims(requestContext)
// 	assert.NotNil(t, authClaims)
// 	assert.Equal(t, "test-tenant", authClaims.TenantID)
// 	assert.Equal(t, "test-user", authClaims.Username)
// }

// func TestHasClaims(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		user          *models.User
// 		requiredClaim models.Claim
// 		expected      bool
// 		description   string
// 	}{
// 		{
// 			name: "user with no claims should not have access",
// 			user: &models.User{
// 				Claims: []models.Claim{},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    false,
// 			description: "User with no claims should not have access to any resource",
// 		},
// 		{
// 			name: "user with exact matching claim should have access",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "test",
// 						Service: "test",
// 						Action:  models.ClaimActionRead,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    true,
// 			description: "User with exact matching claim should have access",
// 		},
// 		{
// 			name: "user with wildcard module should have access",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "*",
// 						Service: "test",
// 						Action:  models.ClaimActionRead,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "any-module",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    true,
// 			description: "User with wildcard module should have access to any module",
// 		},
// 		{
// 			name: "user with wildcard service should have access",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "test",
// 						Service: "*",
// 						Action:  models.ClaimActionRead,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "any-service",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    true,
// 			description: "User with wildcard service should have access to any service",
// 		},
// 		{
// 			name: "user with wildcard action should have access",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "test",
// 						Service: "test",
// 						Action:  models.ClaimActionAll,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionWrite,
// 			},
// 			expected:    true,
// 			description: "User with wildcard action should have access to any action",
// 		},
// 		{
// 			name: "user with read action should have access to read",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "test",
// 						Service: "test",
// 						Action:  models.ClaimActionRead,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    true,
// 			description: "User with read action should have access to read",
// 		},
// 		{
// 			name: "user with all action should have access to read",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "test",
// 						Service: "test",
// 						Action:  models.ClaimActionAll,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    true,
// 			description: "User with all action should have access to read",
// 		},
// 		{
// 			name: "user with read action should not have access to write",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "test",
// 						Service: "test",
// 						Action:  models.ClaimActionRead,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionWrite,
// 			},
// 			expected:    false,
// 			description: "User with read action should not have access to write",
// 		},
// 		{
// 			name: "user with multiple claims should have access if any match",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "test",
// 						Service: "test",
// 						Action:  models.ClaimActionRead,
// 					},
// 					{
// 						Module:  "other",
// 						Service: "other",
// 						Action:  models.ClaimActionWrite,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    true,
// 			description: "User with multiple claims should have access if any match",
// 		},
// 		{
// 			name: "user with multiple claims should not have access if none match",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "test",
// 						Service: "test",
// 						Action:  models.ClaimActionRead,
// 					},
// 					{
// 						Module:  "other",
// 						Service: "other",
// 						Action:  models.ClaimActionWrite,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "different",
// 				Service: "different",
// 				Action:  models.ClaimActionDelete,
// 			},
// 			expected:    false,
// 			description: "User with multiple claims should not have access if none match",
// 		},
// 		{
// 			name: "user with complete wildcard claim should have access to everything",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "*",
// 						Service: "*",
// 						Action:  models.ClaimActionAll,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "any-module",
// 				Service: "any-service",
// 				Action:  models.ClaimActionDelete,
// 			},
// 			expected:    true,
// 			description: "User with complete wildcard claim should have access to everything",
// 		},
// 		{
// 			name: "user with partial wildcard should have access to matching resources",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "test",
// 						Service: "*",
// 						Action:  models.ClaimActionWrite,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "specific-service",
// 				Action:  models.ClaimActionWrite,
// 			},
// 			expected:    true,
// 			description: "User with partial wildcard should have access to matching resources",
// 		},
// 		{
// 			name: "user with partial wildcard should not have access to non-matching resources",
// 			user: &models.User{
// 				Claims: []models.Claim{
// 					{
// 						Module:  "test",
// 						Service: "*",
// 						Action:  models.ClaimActionWrite,
// 					},
// 				},
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "different-module",
// 				Service: "specific-service",
// 				Action:  models.ClaimActionWrite,
// 			},
// 			expected:    false,
// 			description: "User with partial wildcard should not have access to non-matching resources",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := hasClaims(tt.user, tt.requiredClaim)
// 			assert.Equal(t, tt.expected, result, tt.description)
// 		})
// 	}
// }

// func TestMatchesField(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		userField     string
// 		requiredField string
// 		expected      bool
// 		description   string
// 	}{
// 		{
// 			name:          "exact match should return true",
// 			userField:     "test",
// 			requiredField: "test",
// 			expected:      true,
// 			description:   "Exact field match should return true",
// 		},
// 		{
// 			name:          "different fields should return false",
// 			userField:     "test",
// 			requiredField: "different",
// 			expected:      false,
// 			description:   "Different fields should return false",
// 		},
// 		{
// 			name:          "user wildcard should match anything",
// 			userField:     "*",
// 			requiredField: "any-field",
// 			expected:      true,
// 			description:   "User wildcard should match any required field",
// 		},
// 		{
// 			name:          "required wildcard should match anything",
// 			userField:     "any-field",
// 			requiredField: "*",
// 			expected:      true,
// 			description:   "Required wildcard should match any user field",
// 		},
// 		{
// 			name:          "both wildcards should match",
// 			userField:     "*",
// 			requiredField: "*",
// 			expected:      true,
// 			description:   "Both wildcards should match",
// 		},
// 		{
// 			name:          "empty strings should not match",
// 			userField:     "",
// 			requiredField: "test",
// 			expected:      false,
// 			description:   "Empty user field should not match non-empty required field",
// 		},
// 		{
// 			name:          "empty required field should not match non-empty user field",
// 			userField:     "test",
// 			requiredField: "",
// 			expected:      false,
// 			description:   "Empty required field should not match non-empty user field",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := matchesField(tt.userField, tt.requiredField)
// 			assert.Equal(t, tt.expected, result, tt.description)
// 		})
// 	}
// }

// func TestMatchesAction(t *testing.T) {
// 	tests := []struct {
// 		name           string
// 		userAction     models.ClaimAction
// 		requiredAction models.ClaimAction
// 		expected       bool
// 		description    string
// 	}{
// 		{
// 			name:           "exact action match should return true",
// 			userAction:     models.ClaimActionRead,
// 			requiredAction: models.ClaimActionRead,
// 			expected:       true,
// 			description:    "Exact action match should return true",
// 		},
// 		{
// 			name:           "different actions should return false",
// 			userAction:     models.ClaimActionRead,
// 			requiredAction: models.ClaimActionWrite,
// 			expected:       false,
// 			description:    "Different actions should return false",
// 		},
// 		{
// 			name:           "user wildcard action should match anything",
// 			userAction:     models.ClaimActionAll,
// 			requiredAction: models.ClaimActionDelete,
// 			expected:       true,
// 			description:    "User wildcard action should match any required action",
// 		},
// 		{
// 			name:           "required wildcard action should match anything",
// 			userAction:     models.ClaimActionWrite,
// 			requiredAction: models.ClaimActionAll,
// 			expected:       true,
// 			description:    "Required wildcard action should match any user action",
// 		},
// 		{
// 			name:           "both wildcard actions should match",
// 			userAction:     models.ClaimActionAll,
// 			requiredAction: models.ClaimActionAll,
// 			expected:       true,
// 			description:    "Both wildcard actions should match",
// 		},
// 		{
// 			name:           "user with all action should have access to read",
// 			userAction:     models.ClaimActionAll,
// 			requiredAction: models.ClaimActionRead,
// 			expected:       true,
// 			description:    "User with all action should have access to read",
// 		},
// 		{
// 			name:           "user with read action should have access to read",
// 			userAction:     models.ClaimActionRead,
// 			requiredAction: models.ClaimActionRead,
// 			expected:       true,
// 			description:    "User with read action should have access to read",
// 		},
// 		{
// 			name:           "user with read action should not have access to write",
// 			userAction:     models.ClaimActionRead,
// 			requiredAction: models.ClaimActionWrite,
// 			expected:       false,
// 			description:    "User with read action should not have access to write",
// 		},
// 		{
// 			name:           "user with write action should not have access to read",
// 			userAction:     models.ClaimActionWrite,
// 			requiredAction: models.ClaimActionRead,
// 			expected:       false,
// 			description:    "User with write action should not have access to read",
// 		},
// 		{
// 			name:           "user with delete action should not have access to read",
// 			userAction:     models.ClaimActionDelete,
// 			requiredAction: models.ClaimActionRead,
// 			expected:       false,
// 			description:    "User with delete action should not have access to read",
// 		},
// 		{
// 			name:           "user with update action should not have access to read",
// 			userAction:     models.ClaimActionUpdate,
// 			requiredAction: models.ClaimActionRead,
// 			expected:       false,
// 			description:    "User with update action should not have access to read",
// 		},
// 		{
// 			name:           "user with create action should not have access to read",
// 			userAction:     models.ClaimActionCreate,
// 			requiredAction: models.ClaimActionRead,
// 			expected:       false,
// 			description:    "User with create action should not have access to read",
// 		},
// 		{
// 			name:           "user with none action should not have access to read",
// 			userAction:     models.ClaimActionNone,
// 			requiredAction: models.ClaimActionRead,
// 			expected:       false,
// 			description:    "User with none action should not have access to read",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := matchesAction(tt.userAction, tt.requiredAction)
// 			assert.Equal(t, tt.expected, result, tt.description)
// 		})
// 	}
// }

// func TestMatchesClaim(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		userClaim     models.Claim
// 		requiredClaim models.Claim
// 		expected      bool
// 		description   string
// 	}{
// 		{
// 			name: "exact claim match should return true",
// 			userClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    true,
// 			description: "Exact claim match should return true",
// 		},
// 		{
// 			name: "different module should return false",
// 			userClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "different",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    false,
// 			description: "Different module should return false",
// 		},
// 		{
// 			name: "different service should return false",
// 			userClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "different",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    false,
// 			description: "Different service should return false",
// 		},
// 		{
// 			name: "different action should return false",
// 			userClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionWrite,
// 			},
// 			expected:    false,
// 			description: "Different action should return false",
// 		},
// 		{
// 			name: "wildcard module should match",
// 			userClaim: models.Claim{
// 				Module:  "*",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "any-module",
// 				Service: "test",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    true,
// 			description: "Wildcard module should match any module",
// 		},
// 		{
// 			name: "wildcard service should match",
// 			userClaim: models.Claim{
// 				Module:  "test",
// 				Service: "*",
// 				Action:  models.ClaimActionRead,
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "any-service",
// 				Action:  models.ClaimActionRead,
// 			},
// 			expected:    true,
// 			description: "Wildcard service should match any service",
// 		},
// 		{
// 			name: "wildcard action should match",
// 			userClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionAll,
// 			},
// 			requiredClaim: models.Claim{
// 				Module:  "test",
// 				Service: "test",
// 				Action:  models.ClaimActionWrite,
// 			},
// 			expected:    true,
// 			description: "Wildcard action should match any action",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			result := matchesClaim(tt.userClaim, tt.requiredClaim)
// 			assert.Equal(t, tt.expected, result, tt.description)
// 		})
// 	}
// }

// // Test middleware creation (basic compilation tests)
// func TestNewRequireRolePreMiddleware_Creation(t *testing.T) {
// 	middleware := NewRequireRolePreMiddleware([]models.Role{{Name: "admin"}})
// 	assert.NotNil(t, middleware, "Middleware should be created successfully")
// }

// func TestNewRequireClaimPreMiddleware_Creation(t *testing.T) {
// 	requiredClaim := models.Claim{
// 		Module:  "users",
// 		Service: "auth",
// 		Action:  models.ClaimActionRead,
// 	}

// 	middleware := NewRequireClaimPreMiddleware([]models.Claim{requiredClaim})
// 	assert.NotNil(t, middleware, "Middleware should be created successfully")
// }

// // Test middleware with missing auth header
// func TestNewRequireRolePreMiddleware_MissingAuthHeader(t *testing.T) {
// 	middleware := NewRequireRolePreMiddleware([]models.Role{{Name: "admin"}})
// 	_, diag := auth.Initialize(auth.AuthServiceConfig{
// 		SecretKey: []byte("test-secret-key"),
// 		Issuer:    "test-issuer",
// 	}, nil, nil, nil)
// 	assert.False(t, diag.HasErrors())
// 	req := httptest.NewRequest("GET", "/test", nil)
// 	w := httptest.NewRecorder()

// 	result := middleware.Execute(w, req)

// 	assert.False(t, result.Continue)
// 	assert.NotNil(t, result.Error)
// 	assert.Equal(t, http.StatusUnauthorized, w.Code)
// }

// func TestNewRequireClaimPreMiddleware_MissingAuthHeader(t *testing.T) {
// 	_, diag := auth.Initialize(auth.AuthServiceConfig{
// 		SecretKey: []byte("test-secret-key"),
// 		Issuer:    "test-issuer",
// 	}, nil, nil, nil)
// 	assert.False(t, diag.HasErrors())

// 	requiredClaim := models.Claim{
// 		Module:  "users",
// 		Service: "auth",
// 		Action:  models.ClaimActionRead,
// 	}

// 	middleware := NewRequireClaimPreMiddleware([]models.Claim{requiredClaim})
// 	req := httptest.NewRequest("GET", "/test", nil)
// 	w := httptest.NewRecorder()

// 	result := middleware.Execute(w, req)

// 	assert.False(t, result.Continue)
// 	assert.NotNil(t, result.Error)
// 	assert.Equal(t, http.StatusUnauthorized, w.Code)
// }

// // Test middleware with invalid token
// func TestNewRequireRolePreMiddleware_InvalidToken(t *testing.T) {
// 	_, diag := auth.Initialize(auth.AuthServiceConfig{
// 		SecretKey: []byte("test-secret-key"),
// 		Issuer:    "test-issuer",
// 	}, nil, nil, nil)
// 	assert.False(t, diag.HasErrors())
// 	middleware := NewRequireRolePreMiddleware([]models.Role{{Name: "admin"}})
// 	req := httptest.NewRequest("GET", "/test", nil)
// 	req.Header.Set("Authorization", "Bearer invalid-token")
// 	w := httptest.NewRecorder()

// 	result := middleware.Execute(w, req)

// 	assert.False(t, result.Continue)
// 	assert.NotNil(t, result.Error)
// 	assert.Equal(t, http.StatusUnauthorized, w.Code)
// }

// func TestNewRequireClaimPreMiddleware_InvalidToken(t *testing.T) {
// 	_, diag := auth.Initialize(auth.AuthServiceConfig{
// 		SecretKey: []byte("test-secret-key"),
// 		Issuer:    "test-issuer",
// 	}, nil, nil, nil)
// 	assert.False(t, diag.HasErrors())
// 	requiredClaim := models.Claim{
// 		Module:  "users",
// 		Service: "auth",
// 		Action:  models.ClaimActionRead,
// 	}

// 	middleware := NewRequireClaimPreMiddleware([]models.Claim{requiredClaim})
// 	req := httptest.NewRequest("GET", "/test", nil)
// 	req.Header.Set("Authorization", "Bearer invalid-token")
// 	w := httptest.NewRecorder()

// 	result := middleware.Execute(w, req)

// 	assert.False(t, result.Continue)
// 	assert.NotNil(t, result.Error)
// 	assert.Equal(t, http.StatusUnauthorized, w.Code)
// }

// func TestNewRequireRolePreMiddleware(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		requiredRole  string
// 		userRoles     []models.Role
// 		hasAuthHeader bool
// 		validToken    bool
// 		userExists    bool
// 		expectedCode  int
// 		description   string
// 	}{
// 		{
// 			name:         "user with required role should pass",
// 			requiredRole: "admin",
// 			userRoles: []models.Role{
// 				{Name: "admin", Slug: "admin"},
// 				{Name: "user", Slug: "user"},
// 			},
// 			hasAuthHeader: true,
// 			validToken:    true,
// 			userExists:    true,
// 			expectedCode:  http.StatusOK,
// 			description:   "User with required role should pass middleware",
// 		},
// 		{
// 			name:         "user without required role should fail",
// 			requiredRole: "admin",
// 			userRoles: []models.Role{
// 				{Name: "user", Slug: "user"},
// 			},
// 			hasAuthHeader: true,
// 			validToken:    true,
// 			userExists:    true,
// 			expectedCode:  http.StatusForbidden,
// 			description:   "User without required role should fail middleware",
// 		},
// 		{
// 			name:          "missing auth header should fail",
// 			requiredRole:  "admin",
// 			userRoles:     []models.Role{},
// 			hasAuthHeader: false,
// 			validToken:    false,
// 			userExists:    false,
// 			expectedCode:  http.StatusUnauthorized,
// 			description:   "Missing auth header should fail middleware",
// 		},
// 		{
// 			name:          "invalid token should fail",
// 			requiredRole:  "admin",
// 			userRoles:     []models.Role{},
// 			hasAuthHeader: true,
// 			validToken:    false,
// 			userExists:    false,
// 			expectedCode:  http.StatusUnauthorized,
// 			description:   "Invalid token should fail middleware",
// 		},
// 		{
// 			name:          "user not found should fail",
// 			requiredRole:  "admin",
// 			userRoles:     []models.Role{},
// 			hasAuthHeader: true,
// 			validToken:    true,
// 			userExists:    false,
// 			expectedCode:  http.StatusForbidden,
// 			description:   "User not found should fail middleware",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			cfgSvc, _ := config.Initialize()
// 			cfg := cfgSvc.Get()
// 			cfg.Set(config.JwtAuthSecretKey, "test-secret-key")
// 			// Create mock auth service
// 			mockStore := &MockAuthDataStore{BaseMockStore: mocks.NewBaseMockStore()}
// 			_, diag := auth.Initialize(auth.AuthServiceConfig{
// 				SecretKey: []byte("test-secret-key"),
// 				Issuer:    "test-issuer",
// 			}, mockStore, mockStore, mockStore)
// 			assert.False(t, diag.HasErrors())

// 			// Create middleware
// 			middleware := NewRequireRolePreMiddleware([]models.Role{{Name: tt.requiredRole}})

// 			// Create request
// 			req := httptest.NewRequest("GET", "/test", nil)
// 			if tt.hasAuthHeader {
// 				if tt.validToken {
// 					// Create a valid token
// 					claims := &auth.AuthClaims{
// 						Username:  "test-user",
// 						TenantID:  "test-tenant",
// 						Roles:     []string{"user"},
// 						ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
// 						IssuedAt:  time.Now().Unix(),
// 						Issuer:    "test-issuer",
// 					}

// 					token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 						"username":  claims.Username,
// 						"tenant_id": claims.TenantID,
// 						"roles":     claims.Roles,
// 						"exp":       claims.ExpiresAt,
// 						"iat":       claims.IssuedAt,
// 						"iss":       claims.Issuer,
// 					})

// 					tokenString, err := token.SignedString([]byte(cfg.GetString(config.JwtAuthSecretKey, "")))
// 					assert.NoError(t, err)
// 					req.Header.Set("Authorization", "Bearer "+tokenString)
// 				} else {
// 					req.Header.Set("Authorization", "Bearer invalid-token")
// 				}
// 			}

// 			// Create response recorder
// 			w := httptest.NewRecorder()

// 			// Execute middleware
// 			result := middleware.Execute(w, req)

// 			// Verify result
// 			if tt.expectedCode == http.StatusOK {
// 				assert.True(t, result.Continue, tt.description)
// 				assert.Nil(t, result.Error, tt.description)
// 			} else {
// 				assert.False(t, result.Continue, tt.description)
// 				assert.NotNil(t, result.Error, tt.description)
// 				assert.Equal(t, tt.expectedCode, w.Code, tt.description)
// 			}
// 		})
// 	}
// }

// func TestNewRequireClaimPreMiddleware(t *testing.T) {
// 	tests := []struct {
// 		name          string
// 		requiredClaim models.Claim
// 		userClaims    []models.Claim
// 		hasAuthHeader bool
// 		validToken    bool
// 		userExists    bool
// 		expectedCode  int
// 		description   string
// 	}{
// 		{
// 			name: "user with required claim should pass",
// 			requiredClaim: models.Claim{
// 				Module:  "users",
// 				Service: "auth",
// 				Action:  models.ClaimActionRead,
// 			},
// 			userClaims: []models.Claim{
// 				{
// 					Module:  "users",
// 					Service: "auth",
// 					Action:  models.ClaimActionRead,
// 				},
// 			},
// 			hasAuthHeader: true,
// 			validToken:    true,
// 			userExists:    true,
// 			expectedCode:  http.StatusOK,
// 			description:   "User with required claim should pass middleware",
// 		},
// 		{
// 			name: "user without required claim should fail",
// 			requiredClaim: models.Claim{
// 				Module:  "users",
// 				Service: "auth",
// 				Action:  models.ClaimActionWrite,
// 			},
// 			userClaims: []models.Claim{
// 				{
// 					Module:  "users",
// 					Service: "auth",
// 					Action:  models.ClaimActionRead,
// 				},
// 			},
// 			hasAuthHeader: true,
// 			validToken:    true,
// 			userExists:    true,
// 			expectedCode:  http.StatusForbidden,
// 			description:   "User without required claim should fail middleware",
// 		},
// 		{
// 			name: "user with wildcard claim should pass",
// 			requiredClaim: models.Claim{
// 				Module:  "users",
// 				Service: "auth",
// 				Action:  models.ClaimActionRead,
// 			},
// 			userClaims: []models.Claim{
// 				{
// 					Module:  "*",
// 					Service: "*",
// 					Action:  models.ClaimActionAll,
// 				},
// 			},
// 			hasAuthHeader: true,
// 			validToken:    true,
// 			userExists:    true,
// 			expectedCode:  http.StatusOK,
// 			description:   "User with wildcard claim should pass middleware",
// 		},
// 		{
// 			name: "missing auth header should fail",
// 			requiredClaim: models.Claim{
// 				Module:  "users",
// 				Service: "auth",
// 				Action:  models.ClaimActionRead,
// 			},
// 			userClaims:    []models.Claim{},
// 			hasAuthHeader: false,
// 			validToken:    false,
// 			userExists:    false,
// 			expectedCode:  http.StatusUnauthorized,
// 			description:   "Missing auth header should fail middleware",
// 		},
// 		{
// 			name: "invalid token should fail",
// 			requiredClaim: models.Claim{
// 				Module:  "users",
// 				Service: "auth",
// 				Action:  models.ClaimActionRead,
// 			},
// 			userClaims:    []models.Claim{},
// 			hasAuthHeader: true,
// 			validToken:    false,
// 			userExists:    false,
// 			expectedCode:  http.StatusUnauthorized,
// 			description:   "Invalid token should fail middleware",
// 		},
// 		{
// 			name: "user not found should fail",
// 			requiredClaim: models.Claim{
// 				Module:  "users",
// 				Service: "auth",
// 				Action:  models.ClaimActionRead,
// 			},
// 			userClaims:    []models.Claim{},
// 			hasAuthHeader: true,
// 			validToken:    true,
// 			userExists:    false,
// 			expectedCode:  http.StatusForbidden,
// 			description:   "User not found should fail middleware",
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			// Create mock auth service
// 			cfgSvc, _ := config.Initialize()
// 			cfg := cfgSvc.Get()
// 			cfg.Set(config.JwtAuthSecretKey, "test-secret-key")
// 			_, diag := auth.Initialize(auth.AuthServiceConfig{
// 				SecretKey: []byte("test-secret-key"),
// 				Issuer:    "test-issuer",
// 			}, nil, nil, nil)
// 			assert.False(t, diag.HasErrors())

// 			// Create middleware
// 			middleware := NewRequireClaimPreMiddleware([]models.Claim{tt.requiredClaim})

// 			// Create request
// 			req := httptest.NewRequest("GET", "/test", nil)
// 			if tt.hasAuthHeader {
// 				if tt.validToken {
// 					// Create a valid token
// 					claims := &auth.AuthClaims{
// 						Username:  "test-user",
// 						TenantID:  "test-tenant",
// 						Roles:     []string{"user"},
// 						ExpiresAt: time.Now().Add(1 * time.Hour).Unix(),
// 						IssuedAt:  time.Now().Unix(),
// 						Issuer:    "test-issuer",
// 					}

// 					token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
// 						"username":  claims.Username,
// 						"tenant_id": claims.TenantID,
// 						"roles":     claims.Roles,
// 						"exp":       claims.ExpiresAt,
// 						"iat":       claims.IssuedAt,
// 						"iss":       claims.Issuer,
// 					})

// 					tokenString, err := token.SignedString([]byte(cfg.GetString(config.JwtAuthSecretKey, "")))
// 					assert.NoError(t, err)
// 					req.Header.Set("Authorization", "Bearer "+tokenString)
// 				} else {
// 					req.Header.Set("Authorization", "Bearer invalid-token")
// 				}
// 			}

// 			// Create response recorder
// 			w := httptest.NewRecorder()

// 			// Execute middleware
// 			result := middleware.Execute(w, req)

// 			// Verify result
// 			if tt.expectedCode == http.StatusOK {
// 				assert.True(t, result.Continue, tt.description)
// 				assert.Nil(t, result.Error, tt.description)
// 			} else {
// 				assert.False(t, result.Continue, tt.description)
// 				assert.NotNil(t, result.Error, tt.description)
// 				assert.Equal(t, tt.expectedCode, w.Code, tt.description)
// 			}
// 		})
// 	}
// }
