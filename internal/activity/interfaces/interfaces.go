// Package interfaces provides the activity service interfaces.
package interfaces

import (
	"github.com/cjlapao/locally-cli/internal/activity/types"
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

type ActivityServiceInterface interface {
	GetName() string
	GetActivities(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Activity], *diagnostics.Diagnostics)
	GetActivity(ctx *appctx.AppContext, tenantID string, activityID string) (*pkg_models.Activity, *diagnostics.Diagnostics)
	CreateActivity(ctx *appctx.AppContext, tenantID string, activity *pkg_models.CreateActivityRequest) (*pkg_models.Activity, *diagnostics.Diagnostics)
	UpdateActivity(ctx *appctx.AppContext, tenantID string, activityID string, activity *pkg_models.UpdateActivityRequest) (*pkg_models.Activity, *diagnostics.Diagnostics)
	DeleteActivity(ctx *appctx.AppContext, tenantID string, activityID string) *diagnostics.Diagnostics
	RecordInfoActivity(ctx *appctx.AppContext, activityType types.ActivityType, record *types.ActivityRecord) *diagnostics.Diagnostics
	RecordWarningActivity(ctx *appctx.AppContext, activityType types.ActivityType, record *types.ActivityRecord) *diagnostics.Diagnostics
	RecordErrorActivity(ctx *appctx.AppContext, activityType types.ActivityType, err types.ActivityErrorData, record *types.ActivityRecord) *diagnostics.Diagnostics
	RecordSuccessActivity(ctx *appctx.AppContext, activityType types.ActivityType, record *types.ActivityRecord) *diagnostics.Diagnostics
	RecordFailureActivity(ctx *appctx.AppContext, activityType types.ActivityType, err types.ActivityErrorData, record *types.ActivityRecord) *diagnostics.Diagnostics
}
