package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/internal/encryption"
	pkg_types "github.com/cjlapao/locally-cli/pkg/types"
	"github.com/golang-jwt/jwt/v5"
)

type AuthService struct {
	secretKey     []byte
	authDataStore *stores.AuthDataStore
}

// AuthServiceConfig represents the authentication service configuration
type AuthServiceConfig struct {
	// SecretKey is used to sign JWT tokens
	SecretKey string
}

// NewService creates a new authentication service
func NewService(cfg AuthServiceConfig, authDataStore *stores.AuthDataStore) *AuthService {
	return &AuthService{
		secretKey:     []byte(cfg.SecretKey),
		authDataStore: authDataStore,
	}
}

func (s *AuthService) GenerateToken(user *types.User, tenantID string) (*TokenResponse, error) {
	// Create token expiry time (24 hours from now)
	expiresAt := time.Now().Add(24 * time.Hour)

	// Create the Claims
	claims := Claims{
		Username:  user.Username,
		ExpiresAt: expiresAt.Unix(),
		IssuedAt:  time.Now().Unix(),
		Issuer:    "jamf-integrator",
		Role:      user.Role,
		TenantID:  tenantID,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username":  claims.Username,
		"exp":       claims.ExpiresAt,
		"iat":       claims.IssuedAt,
		"iss":       claims.Issuer,
		"role":      claims.Role,
		"tenant_id": claims.TenantID,
		"is_admin":  claims.Role == "root",
	})

	tokenString, err := token.SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign token: %w", err)
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username":  claims.Username,
		"exp":       expiresAt.Add(24 * time.Hour).Unix(),
		"iat":       time.Now().Unix(),
		"iss":       "jamf-integrator",
		"role":      claims.Role,
		"tenant_id": claims.TenantID,
	})

	refreshTokenString, err := refreshToken.SignedString(s.secretKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	err = s.authDataStore.SetRefreshToken(context.Background(), user.ID, refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to set refresh token: %w", err)
	}

	return &TokenResponse{
		Token:        tokenString,
		RefreshToken: refreshTokenString,
		ExpiresAt:    expiresAt,
	}, nil
}

func (s *AuthService) Authenticate(creds Credentials) (*TokenResponse, error) {
	// Check if user exists and password matches
	if creds.TenantID == "" {
		creds.TenantID = "global"
	} else {
		creds.TenantID = strings.TrimSpace(creds.TenantID)
	}
	user, err := s.authDataStore.GetUserByUsername(context.Background(), creds.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	encryptionService := encryption.GetInstance()
	err = encryptionService.VerifyPassword(creds.Password, user.Password)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	if user.Blocked {
		return nil, fmt.Errorf("user is blocked")
	}

	var tenantID string
	if creds.TenantID == config.GlobalTenantID {
		if user.Role != config.SuperUserRole {
			return nil, fmt.Errorf("tenant not found")
		}
		// checking if the root is logging in with a valid tenant or a global one
		if creds.TenantID != config.GlobalTenantID {
			return nil, fmt.Errorf("tenant not found")
		}

		tenantID = creds.TenantID
	} else {
		tenantID = config.GlobalTenantID
	}

	return s.GenerateToken(user, tenantID)
}

func (s *AuthService) ValidateToken(tokenString string) (*Claims, error) {
	// Parse takes the token string and a function for looking up the key
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate the algorithm is what we expect
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		return &Claims{
			Username:  claims["username"].(string),
			ExpiresAt: int64(claims["exp"].(float64)),
			IssuedAt:  int64(claims["iat"].(float64)),
			Issuer:    claims["iss"].(string),
			Role:      claims["role"].(string),
			TenantID:  claims["tenant_id"].(string),
		}, nil
	}

	return nil, fmt.Errorf("invalid token claims")
}

func (s *AuthService) RefreshToken(refreshTokenString string) (*TokenResponse, error) {
	refreshToken, err := jwt.Parse(refreshTokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.secretKey, nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to parse refresh token: %w", err)
	}

	if !refreshToken.Valid {
		return nil, fmt.Errorf("invalid refresh token")
	}
	// Extract claims
	var user *types.User
	var tenantID string
	if claims, ok := refreshToken.Claims.(jwt.MapClaims); ok {
		user, userErr := s.authDataStore.GetUserByID(context.Background(), claims["id"].(string))
		if userErr != nil {
			return nil, fmt.Errorf("failed to get user: %w", userErr)
		}
		if user == nil {
			return nil, fmt.Errorf("user not found")
		}
		if user.RefreshToken != refreshTokenString {
			return nil, fmt.Errorf("invalid refresh token")
		}
		if user.RefreshTokenExpiresAt.Before(time.Now()) {
			return nil, fmt.Errorf("refresh token expired")
		}
		tenantID = claims["tenant_id"].(string)
	}

	return s.GenerateToken(user, tenantID)
}

// NewRequireAuthMiddleware creates a middleware that validates JWT tokens using the provided auth service
func NewRequireAuthMiddleware(authService *AuthService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// Get the Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				// Import api package for error handling
				writeUnauthorizedError(w, r, "Authorization header required", "Missing Authorization header")
				return
			}

			// Check if it starts with "Bearer "
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeUnauthorizedError(w, r, "Invalid authorization header format", "Expected 'Bearer <token>' format")
				return
			}

			// Extract the token
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == "" {
				writeUnauthorizedError(w, r, "Empty token", "Token cannot be empty")
				return
			}

			// Validate the token using the provided auth service
			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				writeInvalidTokenError(w, r, "Token validation failed", err.Error())
				return
			}

			// Add claims to context
			ctx := WithClaims(r.Context(), claims)
			ctx = context.WithValue(ctx, pkg_types.TenantIDKey, claims.TenantID)
			ctx = context.WithValue(ctx, pkg_types.UserIDKey, claims.Username)
			r = r.WithContext(ctx)

			// Call the next handler
			next(w, r)
		}
	}
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
