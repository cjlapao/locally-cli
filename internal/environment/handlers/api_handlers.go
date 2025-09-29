// Package handlers provides the API handlers for the environment service.
package handlers

import (
	"net/http"

	"github.com/cjlapao/locally-cli/internal/api"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/environment/interfaces"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
)

type EnvironmentApiHandler struct {
	environmentService interfaces.EnvironmentServiceInterface
}

func NewEnvironmentApiHandler(environmentService interfaces.EnvironmentServiceInterface) *EnvironmentApiHandler {
	return &EnvironmentApiHandler{environmentService: environmentService}
}

func (h *EnvironmentApiHandler) Routes() []api_types.Route {
	return []api_types.Route{
		{
			Method:      http.MethodGet,
			Path:        "/v1/environments",
			Handler:     h.HandleGetEnvironments,
			Description: "Get all vaults from the environment",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: models.ApiKeySecurityLevelAny,
			},
		},
	}
}

func (h *EnvironmentApiHandler) HandleGetEnvironments(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	tenantID := ctx.GetTenantID()
	if tenantID == "" {
		api.WriteError(w, r, http.StatusBadRequest, "tenant id is required", "tenant id is required", "")
		return
	}

	pagination := utils.ParseQueryRequest(r)

	environments, errDiags := h.environmentService.GetPaginatedEnvironments(ctx, tenantID, pagination)
	if errDiags.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed to get environments", "failed to get environments", errDiags)
		return
	}
	api.WriteObjectResponse(w, r, environments)
}
