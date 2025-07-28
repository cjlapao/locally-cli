package auth

import (
	"time"

	"github.com/cjlapao/locally-cli/internal/database/entities"
)

type AuthCredentials struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	TenantID string `json:"tenant_id"`
}

type AuthClaims struct {
	Username    string   `json:"username"`
	UserID      string   `json:"user_id"`
	ExpiresAt   int64    `json:"exp"`
	IssuedAt    int64    `json:"iat"`
	Issuer      string   `json:"iss"`
	Roles       []string `json:"roles"`
	TenantID    string   `json:"tenant_id"`
	AuthType    string   `json:"auth_type"` // "password" or "api_key"
	APIKeyID    string   `json:"api_key_id,omitempty"`
	IsSuperUser bool     `json:"is_su"`
	IsAdmin     bool     `json:"is_admin"`
}

type TokenResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// APIKeyCredentials represents API key authentication
type APIKeyCredentials struct {
	APIKey   string `json:"api_key" validate:"required"`
	TenantID string `json:"tenant_id"`
}

// CreateAPIKeyRequest represents a request to create a new API key
type CreateAPIKeyRequest struct {
	Name        string                     `json:"name" validate:"required"`
	Permissions entities.APIKeyPermissions `json:"permissions"`
	TenantID    string                     `json:"tenant_id"`
	ExpiresAt   *time.Time                 `json:"expires_at"`
}

// CreateAPIKeyResponse represents the response when creating an API key
type CreateAPIKeyResponse struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	APIKey      string                     `json:"api_key"` // Only shown once
	KeyPrefix   string                     `json:"key_prefix"`
	Permissions entities.APIKeyPermissions `json:"permissions"`
	TenantID    string                     `json:"tenant_id"`
	ExpiresAt   *time.Time                 `json:"expires_at"`
	CreatedAt   time.Time                  `json:"created_at"`
	CreatedBy   string                     `json:"created_by"`
}

// ListAPIKeysResponse represents the response when listing API keys
type ListAPIKeysResponse struct {
	APIKeys []APIKeyListItem `json:"api_keys"`
	Total   int64            `json:"total"`
}

// APIKeyListItem represents an API key in a list (without the actual key)
type APIKeyListItem struct {
	ID          string                     `json:"id"`
	Name        string                     `json:"name"`
	KeyPrefix   string                     `json:"key_prefix"`
	Permissions entities.APIKeyPermissions `json:"permissions"`
	TenantID    string                     `json:"tenant_id"`
	ExpiresAt   *time.Time                 `json:"expires_at"`
	LastUsedAt  *time.Time                 `json:"last_used_at"`
	IsActive    bool                       `json:"is_active"`
	CreatedAt   time.Time                  `json:"created_at"`
	CreatedBy   string                     `json:"created_by"`
	RevokedAt   *time.Time                 `json:"revoked_at"`
	RevokedBy   string                     `json:"revoked_by"`
}

// RevokeAPIKeyRequest represents a request to revoke an API key
type RevokeAPIKeyRequest struct {
	Reason string `json:"reason" validate:"required"`
}
