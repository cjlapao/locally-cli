package interfaces

import (
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
)

type EnvironmentsInterface interface {
	GetName() string
	GetEnvironments(ctx *appctx.AppContext, tenantID string) ([]models.Environment, *diagnostics.Diagnostics)
	GetPaginatedEnvironments(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[models.Environment], *diagnostics.Diagnostics)
}
