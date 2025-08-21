// Package interfaces provides the claim service interface.
package interfaces

import (
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	claim_models "github.com/cjlapao/locally-cli/internal/claim/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

type ClaimServiceInterface interface {
	GetName() string
	GetClaims(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Claim], *diagnostics.Diagnostics)
	GetClaimByIDOrSlug(ctx *appctx.AppContext, tenantID string, idOrSlug string) (*pkg_models.Claim, *diagnostics.Diagnostics)
	CreateClaim(ctx *appctx.AppContext, tenantID string, claim *claim_models.CreateClaimRequest) (*pkg_models.Claim, *diagnostics.Diagnostics)
	UpdateClaim(ctx *appctx.AppContext, tenantID string, request *claim_models.UpdateClaimRequest) (string, *diagnostics.Diagnostics)
	DeleteClaim(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics
	GetClaimUsers(ctx *appctx.AppContext, tenantID string, claimID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.User], *diagnostics.Diagnostics)
	AddClaimToUser(ctx *appctx.AppContext, tenantID string, userID string, claimSlug string) *diagnostics.Diagnostics
	RemoveClaimFromUser(ctx *appctx.AppContext, tenantID string, userID string, claimSlug string) *diagnostics.Diagnostics
	GetClaimRoles(ctx *appctx.AppContext, tenantID string, claimID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Role], *diagnostics.Diagnostics)
	AddRoleToClaim(ctx *appctx.AppContext, tenantID string, claimID string, roleSlug string) *diagnostics.Diagnostics
	RemoveRoleFromClaim(ctx *appctx.AppContext, tenantID string, claimID string, roleSlug string) *diagnostics.Diagnostics
	GetClaimApiKeys(ctx *appctx.AppContext, tenantID string, claimID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.ApiKey], *diagnostics.Diagnostics)
	AddApiKeyToClaim(ctx *appctx.AppContext, tenantID string, claimID string, apiKeySlug string) *diagnostics.Diagnostics
	RemoveApiKeyFromClaim(ctx *appctx.AppContext, tenantID string, claimID string, apiKeySlug string) *diagnostics.Diagnostics
	GetAllSuperUserLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics)
	GetAllAdminLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics)
	GetAllManagerLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics)
	GetAllUserLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics)
	GetAllGuestLevelClaims(ctx *appctx.AppContext, tenantID string) ([]pkg_models.Claim, *diagnostics.Diagnostics)
}
