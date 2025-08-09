// Package models contains the DTOs used by the auth service
package models

import (
	"time"

	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

type AuthCredentials struct {
	Username string `json:"username" validate:"required"`
	Password string `json:"password" validate:"required"`
	TenantID string `json:"tenant_id"`
}

type AuthClaims struct {
	Username      string                   `json:"username"`
	UserID        string                   `json:"user_id"`
	ExpiresAt     int64                    `json:"exp"`
	IssuedAt      int64                    `json:"iat"`
	Issuer        string                   `json:"iss"`
	Roles         []string                 `json:"roles"`
	TenantID      string                   `json:"tenant_id"`
	AuthType      string                   `json:"auth_type"` // "password" or "api_key"
	APIKeyID      string                   `json:"api_key_id,omitempty"`
	SecurityLevel pkg_models.SecurityLevel `json:"security_level"`
}

type TokenResponse struct {
	TenantID     string    `json:"-" yaml:"-"`
	UserID       string    `json:"-" yaml:"-"`
	Username     string    `json:"-" yaml:"-"`
	Error        string    `json:"-" yaml:"-"`
	Token        string    `json:"token" yaml:"token"`
	RefreshToken string    `json:"refresh_token" yaml:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at" yaml:"expires_at"`
}

// APIKeyCredentials represents API key authentication
type APIKeyCredentials struct {
	APIKey   string `json:"api_key" validate:"required"`
	TenantID string `json:"tenant_id"`
}
