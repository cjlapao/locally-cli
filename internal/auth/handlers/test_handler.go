package handlers

import (
	"net/http"

	activity_interfaces "github.com/cjlapao/locally-cli/internal/activity/interfaces"
	"github.com/cjlapao/locally-cli/internal/api"
	api_types "github.com/cjlapao/locally-cli/internal/api/types"
	"github.com/cjlapao/locally-cli/internal/appctx"
	auth_interfaces "github.com/cjlapao/locally-cli/internal/auth/interfaces"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

type AuthTestHandler struct {
	authService     auth_interfaces.AuthServiceInterface
	store           stores.ApiKeyStoreInterface
	activityService activity_interfaces.ActivityServiceInterface
}

func NewTestHandler(authService auth_interfaces.AuthServiceInterface, store stores.ApiKeyStoreInterface, activityService activity_interfaces.ActivityServiceInterface) *AuthTestHandler {
	return &AuthTestHandler{authService: authService, store: store, activityService: activityService}
}

func (h *AuthTestHandler) Routes() []api_types.Route {
	return []api_types.Route{
		{
			Method:      http.MethodGet,
			Path:        "/v1/auth/test/no-auth-nil",
			Handler:     h.HandleAuthTest,
			Description: "Test endpoint with no authentication",
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/auth/test/no-auth",
			Handler:     h.HandleAuthTest,
			Description: "Test endpoint with open access",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelNone,
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/auth/test/superuser",
			Handler:     h.HandleAuthTest,
			Description: "Test endpoint with superuser access",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelSuperUser,
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/auth/test/claim",
			Handler:     h.HandleAuthTest,
			Description: "Test endpoint with claim access",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "auth", Module: "test", Action: pkg_models.AccessLevelRead}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/auth/test/claims",
			Handler:     h.HandleAuthTest,
			Description: "Test endpoint with claims access",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items: []pkg_models.Claim{
						{Service: "auth", Module: "test", Action: pkg_models.AccessLevelRead},
						{Service: "auth", Module: "test", Action: pkg_models.AccessLevelApprove},
						{Service: "auth", Module: "test", Action: pkg_models.AccessLevelLock},
					},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/auth/test/role",
			Handler:     h.HandleAuthTest,
			Description: "Test endpoint with role access",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Roles: &api_types.SecurityRequirementRoles{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Role{{Name: "test"}},
				},
			},
		},
		{
			Method:      http.MethodGet,
			Path:        "/v1/auth/test/role-and-claim",
			Handler:     h.HandleAuthTest,
			Description: "Test endpoint with role and claim access",
			SecurityRequirement: &api_types.SecurityRequirement{
				SecurityLevel: pkg_models.ApiKeySecurityLevelAny,
				Claims: &api_types.SecurityRequirementClaims{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Claim{{Service: "auth", Module: "test", Action: pkg_models.AccessLevelWrite}},
				},
				Roles: &api_types.SecurityRequirementRoles{
					Relation: api_types.SecurityRequirementRelationAnd,
					Items:    []pkg_models.Role{{Name: "test"}},
				},
			},
		},
	}
}

func (h *AuthTestHandler) HandleAuthTest(w http.ResponseWriter, r *http.Request) {
	ctx := appctx.FromContext(r.Context())
	response := map[string]interface{}{
		"message":    "OK",
		"tenant_id":  ctx.GetTenantID(),
		"user_id":    ctx.GetUserID(),
		"username":   ctx.GetUsername(),
		"request_id": ctx.GetRequestID(),
	}

	api.WriteObjectResponse(w, r, response)
}
