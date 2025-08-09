// Package interfaces provides the interfaces for the auth service
package interfaces

import (
	"github.com/cjlapao/locally-cli/internal/appctx"
	auth_models "github.com/cjlapao/locally-cli/internal/auth/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

// AuthServiceInterface defines the contract for the auth service
type AuthServiceInterface interface {
	GetName() string
	GetUserByID(ctx *appctx.AppContext, tenantID string, userID string) (*pkg_models.User, *diagnostics.Diagnostics)

	GenerateSecureAPIKey() (string, *diagnostics.Diagnostics)
	GenerateToken(ctx *appctx.AppContext, user *pkg_models.User, tenantID string, authType string, apiKeyID string) (*auth_models.TokenResponse, *diagnostics.Diagnostics)
	RefreshToken(ctx *appctx.AppContext, refreshTokenString string) (*auth_models.TokenResponse, error)

	Authenticate(ctx *appctx.AppContext, creds auth_models.AuthCredentials) (*auth_models.TokenResponse, *diagnostics.Diagnostics)
	AuthenticateWithPassword(ctx *appctx.AppContext, creds auth_models.AuthCredentials) (*auth_models.TokenResponse, *diagnostics.Diagnostics)
	AuthenticateWithAPIKey(ctx *appctx.AppContext, creds auth_models.APIKeyCredentials) (*auth_models.TokenResponse, *diagnostics.Diagnostics)

	ValidateToken(tokenString string) (*auth_models.AuthClaims, error)
}
