// Package auth implements a service that authenticates users and manages tokens.
package auth

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/encryption"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/golang-jwt/jwt/v5"
)

var (
	globalAuthService *AuthService
	authServiceOnce   sync.Once
	authServiceMutex  sync.Mutex
)

// AuthServiceConfig represents the authentication service configuration
type AuthServiceConfig struct {
	SecretKey []byte
	Issuer    string
}

func (s *AuthServiceConfig) Validate() error {
	if len(s.SecretKey) == 0 {
		return fmt.Errorf("secret key is required")
	}
	if s.Issuer == "" {
		return fmt.Errorf("issuer is required")
	}
	return nil
}

type AuthService struct {
	authConfig    *AuthServiceConfig
	AuthDataStore stores.ApiKeyStoreInterface
	UserStore     stores.UserDataStoreInterface
	TenantStore   stores.TenantDataStoreInterface
}

func Initialize(cfg AuthServiceConfig, authDataStore stores.ApiKeyStoreInterface, userStore stores.UserDataStoreInterface, tenantStore stores.TenantDataStoreInterface) (*AuthService, *diagnostics.Diagnostics) {
	diag := diagnostics.New("auth_service")
	var newDiag *diagnostics.Diagnostics
	authServiceOnce.Do(func() {
		globalAuthService, newDiag = new(cfg, authDataStore, userStore, tenantStore)
		if newDiag.HasErrors() {
			diag.AddError("auth_service", "failed to initialize auth service", newDiag.GetSummary())
		}
	})

	return globalAuthService, diag
}

func GetInstance() *AuthService {
	if globalAuthService == nil {
		panic("auth service not initialized")
	}
	return globalAuthService
}

// Reset resets the singleton for testing purposes
func Reset() {
	authServiceMutex.Lock()
	defer authServiceMutex.Unlock()
	globalAuthService = nil
	authServiceOnce = sync.Once{}
}

// NewService creates a new authentication service
func new(cfg AuthServiceConfig, authDataStore stores.ApiKeyStoreInterface, userStore stores.UserDataStoreInterface, tenantStore stores.TenantDataStoreInterface) (*AuthService, *diagnostics.Diagnostics) {
	diag := diagnostics.New("auth_service")
	if err := cfg.Validate(); err != nil {
		diag.AddError("auth_service", "failed to validate auth service configuration", err.Error())
		return nil, diag
	}

	return &AuthService{
		authConfig:    &cfg,
		AuthDataStore: authDataStore,
		UserStore:     userStore,
		TenantStore:   tenantStore,
	}, diag
}

func (s *AuthService) GetUserByID(ctx *appctx.AppContext, tenantID string, userID string) (*models.User, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_user_by_id")
	user, err := s.UserStore.GetUserByID(ctx, tenantID, userID)
	if err != nil {
		diag.AddError("get_user_by_id", "failed to get user by id", err.Error())
		return nil, diag
	}
	if user == nil {
		diag.AddError("get_user_by_id", "user not found", "")
		return nil, diag
	}
	return mappers.MapUserToDto(user), nil
}

// GenerateSecureAPIKey generates a cryptographically secure API key
func (s *AuthService) GenerateSecureAPIKey() (string, *diagnostics.Diagnostics) {
	// Generate 32 bytes of random data
	diag := diagnostics.New("generate_secure_api_key")
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		diag.AddError("generate_secure_api_key", "failed to generate random bytes", err.Error())
		return "", diag
	}

	// Encode as base64 and remove padding
	key := base64.URLEncoding.WithPadding(base64.NoPadding).EncodeToString(bytes)

	// Add prefix for identification
	prefix := config.ApiKeyPrefix
	return prefix + key, diag
}

// GenerateToken generates a JWT token for a user or API key
func (s *AuthService) GenerateToken(ctx *appctx.AppContext, user *models.User, tenantID string, authType string, apiKeyID string) (*TokenResponse, *diagnostics.Diagnostics) {
	diag := diagnostics.New("generate_token")
	// Create token expiry time (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)

	roles := []string{}
	for _, role := range user.Roles {
		roles = append(roles, role.Slug)
	}

	// Create the Claims
	claims := AuthClaims{
		Username:  user.Username,
		UserID:    user.ID,
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    s.authConfig.Issuer,
		Roles:     roles,
		TenantID:  tenantID,
		AuthType:  authType,
		APIKeyID:  apiKeyID,
	}

	highestSecurityLevel := models.SecurityLevelGuest
	for _, role := range user.Roles {
		if role.SecurityLevel.IsHigherThan(highestSecurityLevel) {
			highestSecurityLevel = role.SecurityLevel
		}
	}

	tokenClaims := jwt.MapClaims{
		"username":       claims.Username,
		"user_id":        claims.UserID,
		"exp":            claims.ExpiresAt,
		"iat":            claims.IssuedAt,
		"iss":            claims.Issuer,
		"roles":          claims.Roles,
		"tenant_id":      claims.TenantID,
		"auth_type":      claims.AuthType,
		"security_level": highestSecurityLevel,
	}
	if authType == "api_key" {
		tokenClaims["api_key_id"] = claims.APIKeyID
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, tokenClaims)

	tokenString, err := token.SignedString(s.authConfig.SecretKey)
	if err != nil {
		diag.AddError("generate_token", "failed to sign token", err.Error())
		return nil, diag
	}

	// Only generate refresh tokens for password-based authentication
	var refreshTokenString string
	if authType == "password" {
		refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username":  claims.Username,
			"user_id":   claims.UserID,
			"exp":       expiresAt.Add(24 * time.Hour).Unix(),
			"iat":       time.Now().Unix(),
			"iss":       s.authConfig.Issuer,
			"roles":     claims.Roles,
			"tenant_id": claims.TenantID,
		})

		refreshTokenString, err = refreshToken.SignedString(s.authConfig.SecretKey)
		if err != nil {
			diag.AddError("generate_token", "failed to sign refresh token", err.Error())
			return nil, diag
		}

		err = s.UserStore.SetRefreshToken(ctx, tenantID, user.ID, refreshTokenString)
		if err != nil {
			diag.AddError("generate_token", "failed to set refresh token", err.Error())
			return nil, diag
		}
	}

	return &TokenResponse{
		TenantID:     tenantID,
		UserID:       user.ID,
		Username:     user.Username,
		Token:        tokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
	}, diag
}

// AuthenticateWithPassword authenticates a user with username/password
func (s *AuthService) AuthenticateWithPassword(ctx *appctx.AppContext, creds AuthCredentials) (*TokenResponse, *diagnostics.Diagnostics) {
	diag := diagnostics.New("authenticate_with_password")
	errorToken := &TokenResponse{
		TenantID: creds.TenantID,
		UserID:   creds.Username,
		Username: creds.Username,
		Error:    "Invalid credentials",
		Token:    "",
	}

	// Check if user exists and password matches
	if creds.TenantID == "" {
		creds.TenantID = config.GlobalTenantID
	} else {
		creds.TenantID = strings.TrimSpace(creds.TenantID)
	}
	if strings.EqualFold(creds.TenantID, config.GlobalTenantID) {
		tenant, err := s.TenantStore.GetTenantByIdOrSlug(ctx, creds.TenantID)
		if err != nil {
			diag.AddError("authenticate_with_password", "failed to get tenant", err.Error())
			errorToken.Error = fmt.Sprintf("failed to get tenant: %s", err.Error())
			return errorToken, diag
		}
		if tenant == nil {
			diag.AddError("authenticate_with_password", "tenant not found", "")
			errorToken.Error = "tenant not found"
			return errorToken, diag
		}

		creds.TenantID = tenant.ID
	}
	errorToken.TenantID = creds.TenantID

	// Get the user by username
	dbUser, err := s.UserStore.GetUserByUsername(ctx, creds.TenantID, creds.Username)
	if err != nil {
		diag.AddError("authenticate_with_password", "failed to get user", err.Error())
		errorToken.Error = fmt.Sprintf("failed to get user: %s", err.Error())
		errorToken.UserID = config.UnknownUserID
		return errorToken, diag
	}
	if dbUser == nil {
		diag.AddError("authenticate_with_password", "invalid credentials", "")
		errorToken.Error = "invalid credentials"
		errorToken.UserID = config.UnknownUserID
		return errorToken, diag
	}
	// Map the user to a DTO
	errorToken.UserID = dbUser.ID
	user := mappers.MapUserToDto(dbUser)
	encryptionService := encryption.GetInstance()
	err = encryptionService.VerifyPassword(creds.Password, dbUser.Password)
	if err != nil {
		diag.AddError("authenticate_with_password", "invalid credentials", err.Error())
		errorToken.Error = "invalid credentials"
		return errorToken, diag
	}
	if dbUser.Blocked {
		diag.AddError("authenticate_with_password", "user is blocked", "")
		errorToken.Error = "user is blocked"
		return errorToken, diag
	}

	if dbUser.TenantID != creds.TenantID {
		if !isSuperUser(user.Roles) {
			diag.AddError("authenticate_with_password", "invalid tenant", "")
			errorToken.Error = "user not authorized to access this tenant"
			return errorToken, diag
		}
	}

	token, diag := s.GenerateToken(ctx, user, creds.TenantID, "password", "")
	if diag.HasErrors() {
		errorToken.Error = fmt.Sprintf("failed to generate token: %s", diag.GetSummary())
		return errorToken, diag
	}
	return token, nil
}

// AuthenticateWithAPIKey authenticates a user with an API key
func (s *AuthService) AuthenticateWithAPIKey(ctx *appctx.AppContext, creds APIKeyCredentials) (*TokenResponse, *diagnostics.Diagnostics) {
	diag := diagnostics.New("authenticate_with_api_key")
	if creds.TenantID == "" {
		creds.TenantID = "global"
	} else {
		creds.TenantID = strings.TrimSpace(creds.TenantID)
	}

	// Extract prefix from API key for quick lookup
	if !strings.HasPrefix(creds.APIKey, "sk-locally-") {
		diag.AddError("authenticate_with_api_key", "invalid API key format", "")
		return nil, diag
	}

	// Hash the provided API key for comparison
	encryptionService := encryption.GetInstance()
	keyHash, err := encryptionService.HashPassword(creds.APIKey)
	if err != nil {
		diag.AddError("authenticate_with_api_key", "failed to hash API key", err.Error())
		return nil, diag
	}

	// Find the API key in the database
	apiKey, err := s.AuthDataStore.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		diag.AddError("authenticate_with_api_key", "failed to get API key", err.Error())
		return nil, diag
	}
	if apiKey == nil {
		diag.AddError("authenticate_with_api_key", "invalid API key", "")
		return nil, diag
	}

	// Check if API key is active
	if !apiKey.IsActive {
		diag.AddError("authenticate_with_api_key", "API key is revoked", "")
		return nil, diag
	}

	// Check if API key is expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		diag.AddError("authenticate_with_api_key", "API key has expired", "")
		return nil, diag
	}

	// Get the dbUser associated with this API key
	dbUser, err := s.UserStore.GetUserByID(ctx, apiKey.TenantID, apiKey.CreatedBy)
	if err != nil {
		diag.AddError("authenticate_with_api_key", "failed to get user", err.Error())
		return nil, diag
	}
	if dbUser == nil {
		diag.AddError("authenticate_with_api_key", "user not found", "")
		return nil, diag
	}
	if dbUser.Blocked {
		diag.AddError("authenticate_with_api_key", "user is blocked", "")
		return nil, diag
	}

	user := mappers.MapUserToDto(dbUser)

	// Check tenant access
	if apiKey.TenantID != "" && apiKey.TenantID != creds.TenantID {
		diag.AddError("authenticate_with_api_key", "API key not valid for this tenant", "")
		return nil, diag
	}

	// Use API key's tenant ID if not specified
	tenantID := creds.TenantID
	if tenantID == "" {
		tenantID = apiKey.TenantID
	}
	if tenantID == "" {
		tenantID = config.GlobalTenantID
	}

	token, diag := s.GenerateToken(ctx, user, tenantID, "api_key", apiKey.ID)
	if diag.HasErrors() {
		return nil, diag
	}
	return token, nil
}

// Authenticate is the main authentication method that supports both password and API key
func (s *AuthService) Authenticate(ctx *appctx.AppContext, creds AuthCredentials) (*TokenResponse, *diagnostics.Diagnostics) {
	return s.AuthenticateWithPassword(ctx, creds)
}

// ValidateToken validates a JWT token and returns claims
func (s *AuthService) ValidateToken(tokenString string) (*AuthClaims, error) {
	// Parse takes the token string and a function for looking up the key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.authConfig.SecretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		authType := "password" // default
		if authTypeClaim, exists := claims["auth_type"]; exists {
			if authTypeStr, ok := authTypeClaim.(string); ok {
				authType = authTypeStr
			}
		}

		apiKeyID := ""
		if apiKeyIDClaim, exists := claims["api_key_id"]; exists {
			if apiKeyIDStr, ok := apiKeyIDClaim.(string); ok {
				apiKeyID = apiKeyIDStr
			}
		}

		roles := []string{}
		if rolesClaim, exists := claims["roles"]; exists {
			if rolesSlice, ok := rolesClaim.([]interface{}); ok {
				for _, role := range rolesSlice {
					if roleStr, ok := role.(string); ok {
						roles = append(roles, roleStr)
					}
				}
			}
		}

		return &AuthClaims{
			Username:      claims["username"].(string),
			UserID:        claims["user_id"].(string),
			ExpiresAt:     int64(claims["exp"].(float64)),
			IssuedAt:      int64(claims["iat"].(float64)),
			Issuer:        claims["iss"].(string),
			Roles:         roles,
			TenantID:      claims["tenant_id"].(string),
			AuthType:      authType,
			APIKeyID:      apiKeyID,
			SecurityLevel: models.SecurityLevel(claims["security_level"].(string)),
		}, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (s *AuthService) RefreshToken(ctx *appctx.AppContext, refreshTokenString string) (*TokenResponse, error) {
	refreshToken, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.authConfig.SecretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if !refreshToken.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}
	// Extract claims
	var user *models.User
	var tenantID string
	if claims, ok := refreshToken.Claims.(jwt.MapClaims); ok {
		dbUser, userErr := s.UserStore.GetUserByID(ctx, claims["tenant_id"].(string), claims["id"].(string))
		if userErr != nil {
			return nil, fmt.Errorf("failed to get user: %w", userErr)
		}
		if dbUser == nil {
			return nil, fmt.Errorf("user not found")
		}
		user = mappers.MapUserToDto(dbUser)
		if user.RefreshToken != refreshTokenString {
			return nil, fmt.Errorf("invalid refresh token")
		}
		if user.RefreshTokenExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("refresh token expired")
		}
		tenantID = claims["tenant_id"].(string)
	}

	token, diag := s.GenerateToken(ctx, user, tenantID, "password", "") // Assuming refresh token is for password auth
	if diag.HasErrors() {
		return nil, fmt.Errorf("failed to generate token: %v", diag.GetSummary())
	}
	return token, nil
}

func isSuperUser(roles []models.Role) bool {
	for _, role := range roles {
		if role.SecurityLevel == models.SecurityLevelSuperUser {
			return true
		}
	}
	return false
}

func isAdminUser(roles []models.Role) bool {
	for _, role := range roles {
		if role.SecurityLevel == models.SecurityLevelAdmin {
			return true
		}
	}
	return false
}
