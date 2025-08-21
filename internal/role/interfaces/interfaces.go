// Package interfaces provides the role service interface.
package interfaces

import (
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/role/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

type RoleServiceInterface interface {
	GetName() string
	GetRoles(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Role], *diagnostics.Diagnostics)
	GetRoleByIDorSlug(ctx *appctx.AppContext, tenantID string, idOrSlug string) (*pkg_models.Role, *diagnostics.Diagnostics)
	GetRoleByIDorSlugWithClaims(ctx *appctx.AppContext, tenantID string, idOrSlug string) (*pkg_models.Role, *diagnostics.Diagnostics)
	CreateRole(ctx *appctx.AppContext, tenantID string, role *models.CreateRoleRequest) (*pkg_models.Role, *diagnostics.Diagnostics)
	UpdateRole(ctx *appctx.AppContext, tenantID string, role *models.UpdateRoleRequest) (string, *diagnostics.Diagnostics)
	DeleteRole(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics
	GetRoleUsers(ctx *appctx.AppContext, tenantID string, id string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.User], *diagnostics.Diagnostics)
	AddUserToRole(ctx *appctx.AppContext, tenantID string, userID string, roleSlug string) *diagnostics.Diagnostics
	RemoveUserFromRole(ctx *appctx.AppContext, tenantID string, userID string, roleSlug string) *diagnostics.Diagnostics
	GetRoleClaims(ctx *appctx.AppContext, tenantID string, roleID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Claim], *diagnostics.Diagnostics)
	AddClaimToRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) *diagnostics.Diagnostics
	RemoveClaimFromRole(ctx *appctx.AppContext, tenantID string, roleID string, claimID string) *diagnostics.Diagnostics
}
