package auth

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/entities"
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

// AuthDataStoreInterface defines the interface for auth data store operations
type AuthDataStoreInterface interface {
	CreateAPIKey(ctx *appctx.AppContext, apiKey *entities.APIKey) (*entities.APIKey, error)
	GetAPIKeyByHash(ctx *appctx.AppContext, keyHash string) (*entities.APIKey, error)
	GetAPIKeyByPrefix(ctx *appctx.AppContext, keyPrefix string) (*entities.APIKey, error)
	GetAPIKeyByID(ctx *appctx.AppContext, id string) (*entities.APIKey, error)
	GetUserByID(ctx *appctx.AppContext, id string) (*entities.User, error)
	GetUserByUsername(ctx *appctx.AppContext, username string) (*entities.User, error)
	UpdateAPIKeyLastUsed(ctx *appctx.AppContext, id string) error
	ListAPIKeysByUserID(ctx *appctx.AppContext, userID string) ([]entities.APIKey, error)
	RevokeAPIKey(ctx *appctx.AppContext, id string, revokedBy string, reason string) error
	DeleteAPIKey(ctx *appctx.AppContext, id string) error
	SetRefreshToken(ctx *appctx.AppContext, id string, refreshToken string) error
}

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
	AuthDataStore AuthDataStoreInterface
}

func Initialize(cfg AuthServiceConfig, authDataStore AuthDataStoreInterface) (*AuthService, *diagnostics.Diagnostics) {
	diag := diagnostics.New("auth_service")
	var newDiag *diagnostics.Diagnostics
	authServiceOnce.Do(func() {
		globalAuthService, newDiag = new(cfg, authDataStore)
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
func new(cfg AuthServiceConfig, authDataStore AuthDataStoreInterface) (*AuthService, *diagnostics.Diagnostics) {
	diag := diagnostics.New("auth_service")
	if err := cfg.Validate(); err != nil {
		diag.AddError("auth_service", "failed to validate auth service configuration", err.Error())
		return nil, diag
	}

	return &AuthService{
		authConfig:    &cfg,
		AuthDataStore: authDataStore,
	}, diag
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
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    s.authConfig.Issuer,
		Roles:     roles,
		TenantID:  tenantID,
		AuthType:  authType,
		APIKeyID:  apiKeyID,
	}

	tokenClaims := jwt.MapClaims{
		"username":  claims.Username,
		"exp":       claims.ExpiresAt,
		"iat":       claims.IssuedAt,
		"iss":       claims.Issuer,
		"roles":     claims.Roles,
		"tenant_id": claims.TenantID,
		"auth_type": claims.AuthType,
		"is_admin":  isAdminUser(user.Roles),
		"is_su":     isSuperUser(user.Roles),
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

		err = s.AuthDataStore.SetRefreshToken(ctx, user.ID, refreshTokenString)
		if err != nil {
			diag.AddError("generate_token", "failed to set refresh token", err.Error())
			return nil, diag
		}
	}

	return &TokenResponse{
		Token:        tokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
	}, diag
}

// AuthenticateWithPassword authenticates a user with username/password
func (s *AuthService) AuthenticateWithPassword(ctx *appctx.AppContext, creds AuthCredentials) (*TokenResponse, *diagnostics.Diagnostics) {
	diag := diagnostics.New("authenticate_with_password")
	// Check if user exists and password matches
	if creds.TenantID == "" {
		creds.TenantID = config.GlobalTenantID
	} else {
		creds.TenantID = strings.TrimSpace(creds.TenantID)
	}
	// Get the user by username
	dbUser, err := s.AuthDataStore.GetUserByUsername(ctx, creds.Username)
	if err != nil {
		diag.AddError("authenticate_with_password", "failed to get user", err.Error())
		return nil, diag
	}
	if dbUser == nil {
		diag.AddError("authenticate_with_password", "invalid credentials", "")
		return nil, diag
	}
	// Map the user to a DTO
	user := mappers.MapUserToDto(dbUser)
	encryptionService := encryption.GetInstance()
	err = encryptionService.VerifyPassword(creds.Password, dbUser.Password)
	if err != nil {
		diag.AddError("authenticate_with_password", "invalid credentials", err.Error())
		return nil, diag
	}
	if dbUser.Blocked {
		diag.AddError("authenticate_with_password", "user is blocked", "")
		return nil, diag
	}

	if dbUser.TenantID != creds.TenantID {
		if !isSuperUser(user.Roles) {
			diag.AddError("authenticate_with_password", "invalid tenant", "")
			return nil, diag
		}
	}

	token, diag := s.GenerateToken(ctx, user, creds.TenantID, "password", "")
	if diag.HasErrors() {
		return nil, diag
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
	dbUser, err := s.AuthDataStore.GetUserByID(ctx, apiKey.UserID)
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

	// Update last used timestamp
	err = s.AuthDataStore.UpdateAPIKeyLastUsed(ctx, apiKey.ID)
	if err != nil {
		// Log but don't fail authentication
		fmt.Printf("Warning: failed to update API key last used: %v\n", err)
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

// CreateAPIKey creates a new API key for a user
func (s *AuthService) CreateAPIKey(ctx *appctx.AppContext, userID string, req CreateAPIKeyRequest, createdBy string) (*CreateAPIKeyResponse, error) {
	// Generate a secure API key
	apiKeyString, diag := s.GenerateSecureAPIKey()
	if diag.HasErrors() {
		return nil, fmt.Errorf("failed to generate API key: %v", diag.GetSummary())
	}

	// Serialize permissions to JSON
	permissionsJSON, err := json.Marshal(req.Permissions)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize permissions: %v", err.Error())
	}

	// Create API key record
	apiKey := &entities.APIKey{
		UserID:      userID,
		Name:        req.Name,
		KeyHash:     apiKeyString,                              // Will be hashed in CreateAPIKey
		KeyPrefix:   apiKeyString[:8+len(config.ApiKeyPrefix)], // First 8 chars after prefix
		Permissions: string(permissionsJSON),
		TenantID:    req.TenantID,
		ExpiresAt:   req.ExpiresAt,
		IsActive:    true,
		CreatedBy:   createdBy,
	}

	// Save to database
	dbAPIKey, err := s.AuthDataStore.CreateAPIKey(ctx, apiKey)
	if err != nil {
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return &CreateAPIKeyResponse{
		ID:          dbAPIKey.ID,
		Name:        dbAPIKey.Name,
		APIKey:      apiKeyString, // Return the original key (only shown once)
		KeyPrefix:   dbAPIKey.KeyPrefix,
		Permissions: req.Permissions,
		TenantID:    dbAPIKey.TenantID,
		ExpiresAt:   dbAPIKey.ExpiresAt,
		CreatedAt:   dbAPIKey.CreatedAt,
		CreatedBy:   dbAPIKey.CreatedBy,
	}, nil
}

// ListAPIKeys lists all API keys for a user
func (s *AuthService) ListAPIKeys(ctx *appctx.AppContext, userID string) (*ListAPIKeysResponse, error) {
	apiKeys, err := s.AuthDataStore.ListAPIKeysByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to list API keys: %w", err)
	}

	var items []APIKeyListItem
	for _, key := range apiKeys {
		var permissions entities.APIKeyPermissions
		if key.Permissions != "" {
			if err := json.Unmarshal([]byte(key.Permissions), &permissions); err != nil {
				// Log error but continue
				fmt.Printf("Warning: failed to unmarshal permissions for key %s: %v\n", key.ID, err)
			}
		}

		items = append(items, APIKeyListItem{
			ID:          key.ID,
			Name:        key.Name,
			KeyPrefix:   key.KeyPrefix,
			Permissions: permissions,
			TenantID:    key.TenantID,
			ExpiresAt:   key.ExpiresAt,
			LastUsedAt:  key.LastUsedAt,
			IsActive:    key.IsActive,
			CreatedAt:   key.CreatedAt,
			CreatedBy:   key.CreatedBy,
			RevokedAt:   key.RevokedAt,
			RevokedBy:   key.RevokedBy,
		})
	}

	return &ListAPIKeysResponse{
		APIKeys: items,
		Total:   int64(len(items)),
	}, nil
}

// RevokeAPIKey revokes an API key
func (s *AuthService) RevokeAPIKey(ctx *appctx.AppContext, apiKeyID string, revokedBy string, reason string) error {
	return s.AuthDataStore.RevokeAPIKey(ctx, apiKeyID, revokedBy, reason)
}

// DeleteAPIKey permanently deletes an API key
func (s *AuthService) DeleteAPIKey(ctx *appctx.AppContext, apiKeyID string) error {
	return s.AuthDataStore.DeleteAPIKey(ctx, apiKeyID)
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
		isSuperUser := false
		if isSuperUserClaim, exists := claims["is_su"]; exists {
			if isSuperUserBool, ok := isSuperUserClaim.(bool); ok {
				isSuperUser = isSuperUserBool
			}
		}
		isAdmin := false
		if isAdminClaim, exists := claims["is_admin"]; exists {
			if isAdminBool, ok := isAdminClaim.(bool); ok {
				isAdmin = isAdminBool
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
			Username:    claims["username"].(string),
			ExpiresAt:   int64(claims["exp"].(float64)),
			IssuedAt:    int64(claims["iat"].(float64)),
			Issuer:      claims["iss"].(string),
			Roles:       roles,
			TenantID:    claims["tenant_id"].(string),
			AuthType:    authType,
			APIKeyID:    apiKeyID,
			IsSuperUser: isSuperUser,
			IsAdmin:     isAdmin,
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
		dbUser, userErr := s.AuthDataStore.GetUserByID(ctx, claims["id"].(string))
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
		if role.IsSuperUser {
			return true
		}
	}
	return false
}

func isAdminUser(roles []models.Role) bool {
	for _, role := range roles {
		if role.IsAdmin {
			return true
		}
	}
	return false
}
