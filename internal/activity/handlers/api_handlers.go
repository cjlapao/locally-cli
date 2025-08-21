// Package handlers provides the API handlers for the activity service
package handlers

import (
	"net/http"

	"github.com/cjlapao/locally-cli/internal/activity/interfaces"
	"github.com/cjlapao/locally-cli/internal/api"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
)

type ActivityApiHandler struct {
	activityService interfaces.ActivityServiceInterface
	systemService   system_interfaces.SystemServiceInterface
}

func NewActivityApiHandler(activityService interfaces.ActivityServiceInterface, systemService system_interfaces.SystemServiceInterface) *ActivityApiHandler {
	return &ActivityApiHandler{activityService: activityService, systemService: systemService}
}

func (h *ActivityApiHandler) Routes() []api_types.Route {
	module := h.systemService.GetServiceModuleByName("activity", "audits")
	adminRole := h.systemService.GetRoleByName("admin")
	auditorRole := h.systemService.GetRoleByName("auditor")

	// If the module is not found, return an empty route
	if module == nil {
		return []api_types.Route{}
	}
	if adminRole == nil {
		return []api_types.Route{}
	}
	if auditorRole == nil {
		return []api_types.Route{}
	}

	return []api_types.Route{
		{
			Method:      http.MethodGet,
			Path:        "/v1/audits",
			Handler:     h.HandleGetAudits,
			Description: "Get all audits",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Roles: &api_types.SecurityRequirementRoles{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Role{*adminRole, *auditorRole},
				},
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: module.System, Module: module.Name, Action: pkg_models.AccessLevelRead}},
				},
			},
		},
	}
}

func (h *ActivityApiHandler) HandleGetAudits(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	requestContext := utils.GetRequestContextFromRequest(r)
	ctx.Log().Info("Getting audits")

	response, diag := h.activityService.GetActivities(ctx, requestContext.TenantID, requestContext.Pagination)
	if diag.HasErrors() {
		api.WriteErrorWithDiagnostics(w, r, http.StatusInternalServerError, "failed_to_get_audits", "Failed to get audits", diag)
		return
	}

	api.WriteObjectResponse(w, r, response)
}
