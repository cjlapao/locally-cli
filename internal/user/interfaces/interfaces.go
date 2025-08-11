// Package interfaces provides the user service interface.
package interfaces

import (
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/user/models"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

type UserServiceInterface interface {
	GetName() string
	GetUsersByFilter(ctx *appctx.AppContext, tenantID string, filter *filters.Filter) (*api_models.PaginatedResponse[pkg_models.User], *diagnostics.Diagnostics)
	GetUserByID(ctx *appctx.AppContext, tenantID string, id string) (*pkg_models.User, *diagnostics.Diagnostics)
	GetUserByUsername(ctx *appctx.AppContext, tenantID string, username string) (*pkg_models.User, *diagnostics.Diagnostics)
	CreateUser(ctx *appctx.AppContext, tenantID string, role string, user *models.CreateUserRequest) (*models.CreateUserResponse, *diagnostics.Diagnostics)
	UpdateUser(ctx *appctx.AppContext, tenantID string, userId string, user *models.UpdateUserRequest) (*models.UpdateUserResponse, *diagnostics.Diagnostics)
	DeleteUser(ctx *appctx.AppContext, tenantID string, userId string) *diagnostics.Diagnostics
	UpdateUserPassword(ctx *appctx.AppContext, tenantID string, id string, request *models.UpdateUserPasswordRequest) *diagnostics.Diagnostics
	GetUserClaims(ctx *appctx.AppContext, tenantID string, userID string) ([]pkg_models.Claim, *diagnostics.Diagnostics)
	AddClaimToUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics
	RemoveClaimFromUser(ctx *appctx.AppContext, tenantID string, userID string, claimIdOrSlug string) *diagnostics.Diagnostics
}
