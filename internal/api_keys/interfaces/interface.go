// Package interfaces provides the API keys service interface.
package interfaces

import (
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/api_keys/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

type ApiKeysServiceInterface interface {
	GetName() string
	GetApiKeys(ctx *appctx.AppContext, tenantID string, pagination *api_models.Pagination) (*api_models.PaginatedResponse[pkg_models.ApiKey], *diagnostics.Diagnostics)
	GetFilteredApiKeys(ctx *appctx.AppContext, tenantID string, filter *filters.Filter) (*api_models.PaginatedResponse[pkg_models.ApiKey], *diagnostics.Diagnostics)
	GetApiKeyByID(ctx *appctx.AppContext, tenantID string, id string) (*pkg_models.ApiKey, *diagnostics.Diagnostics)
	CreateApiKey(ctx *appctx.AppContext, tenantID string, request *models.CreateApiKeyRequest) (*pkg_models.ApiKey, *diagnostics.Diagnostics)
	DeleteApiKey(ctx *appctx.AppContext, tenantID string, id string) *diagnostics.Diagnostics
	RevokeApiKey(ctx *appctx.AppContext, tenantID string, revokedBy string, id string, request *models.RevokeApiKeyRequest) *diagnostics.Diagnostics
}
