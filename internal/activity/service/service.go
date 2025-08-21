// Package service provides the activity service implementation.
package service

import (
	"sync"

	"github.com/cjlapao/locally-cli/internal/activity/interfaces"
	"github.com/cjlapao/locally-cli/internal/activity/types"
	api_models "github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/database/filters"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/mappers"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
)

var (
	globalActivityService *ActivityService
	activityServiceOnce   sync.Once
	activityServiceMutex  sync.Mutex
)

type ActivityService struct {
	activityStore stores.ActivityDataStoreInterface
}

func Initialize(activityStore stores.ActivityDataStoreInterface) interfaces.ActivityServiceInterface {
	activityServiceMutex.Lock()
	defer activityServiceMutex.Unlock()

	activityServiceOnce.Do(func() {
		globalActivityService = new(activityStore)
	})
	return globalActivityService
}

func GetInstance() interfaces.ActivityServiceInterface {
	if globalActivityService == nil {
		panic("activity service not initialized")
	}
	return globalActivityService
}

// Reset resets the singleton for testing purposes
func Reset() {
	activityServiceMutex.Lock()
	defer activityServiceMutex.Unlock()
	globalActivityService = nil
	activityServiceOnce = sync.Once{}
}

func new(activityStore stores.ActivityDataStoreInterface) *ActivityService {
	return &ActivityService{
		activityStore: activityStore,
	}
}

func (s *ActivityService) GetName() string {
	return "activity"
}

func (s *ActivityService) GetActivities(ctx *appctx.AppContext, tenantID string, pagination *api_models.PaginationRequest) (*api_models.PaginationResponse[pkg_models.Activity], *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_activities")
	defer diag.Complete()
	var query *filters.QueryBuilder
	if pagination != nil {
		query = pagination.ToQueryBuilder()
	}

	activities, activitiesDiag := s.activityStore.GetActivitiesByQuery(ctx, tenantID, query)
	if activitiesDiag.HasErrors() {
		diag.Append(activitiesDiag)
		return nil, diag
	}

	activitiesDto := make([]pkg_models.Activity, len(activities.Items))
	for i, activity := range activities.Items {
		activitiesDto[i] = *mappers.MapActivityToDto(&activity)
	}

	response := api_models.PaginationResponse[pkg_models.Activity]{
		TotalCount: activities.Total,
		Pagination: api_models.Pagination{
			Page:       activities.Page,
			PageSize:   activities.PageSize,
			TotalPages: activities.TotalPages,
		},
		Data: activitiesDto,
	}

	return &response, diag
}

func (s *ActivityService) GetActivity(ctx *appctx.AppContext, tenantID string, activityID string) (*pkg_models.Activity, *diagnostics.Diagnostics) {
	diag := diagnostics.New("get_activity")
	defer diag.Complete()

	activity, activityDiag := s.activityStore.GetActivityByID(ctx, tenantID, activityID)
	if activityDiag.HasErrors() {
		diag.Append(activityDiag)
		return nil, diag
	}

	return mappers.MapActivityToDto(activity), diag
}

func (s *ActivityService) CreateActivity(ctx *appctx.AppContext, tenantID string, activity *pkg_models.CreateActivityRequest) (*pkg_models.Activity, *diagnostics.Diagnostics) {
	diag := diagnostics.New("create_activity")
	defer diag.Complete()

	activityEntity := mappers.MapCreateActivityRequestToEntity(activity)

	createdActivity, createdActivityDiag := s.activityStore.CreateActivity(ctx, tenantID, activityEntity)
	if createdActivityDiag.HasErrors() {
		diag.Append(createdActivityDiag)
		return nil, diag
	}

	return mappers.MapActivityToDto(createdActivity), diag
}

func (s *ActivityService) UpdateActivity(ctx *appctx.AppContext, tenantID string, activityID string, activity *pkg_models.UpdateActivityRequest) (*pkg_models.Activity, *diagnostics.Diagnostics) {
	diag := diagnostics.New("update_activity")
	defer diag.Complete()

	activityEntity, activityEntityDiag := s.activityStore.GetActivityByID(ctx, tenantID, activityID)
	if activityEntityDiag.HasErrors() {
		diag.Append(activityEntityDiag)
		return nil, diag
	}

	updatedActivityEntity := mappers.MapUpdateActivityRequestToEntity(activity, activityEntity)

	updatedActivityEntityDiag := s.activityStore.UpdateActivity(ctx, tenantID, updatedActivityEntity)
	if updatedActivityEntityDiag.HasErrors() {
		diag.Append(updatedActivityEntityDiag)
		return nil, diag
	}

	return mappers.MapActivityToDto(updatedActivityEntity), diag
}

func (s *ActivityService) DeleteActivity(ctx *appctx.AppContext, tenantID string, activityID string) *diagnostics.Diagnostics {
	diag := diagnostics.New("delete_activity")
	defer diag.Complete()

	activity, activityDiag := s.activityStore.GetActivityByID(ctx, tenantID, activityID)
	if activityDiag.HasErrors() {
		diag.Append(activityDiag)
		return diag
	}
	if activity == nil {
		diag.AddError("activity_not_found", "Activity not found", "activity_id", map[string]interface{}{
			"activity_id": activityID,
		})
		return diag
	}

	deleteActivityDiag := s.activityStore.DeleteActivity(ctx, tenantID, activityID)
	if deleteActivityDiag.HasErrors() {
		diag.Append(deleteActivityDiag)
		return diag
	}

	return diag
}

func (s *ActivityService) RecordInfoActivity(ctx *appctx.AppContext, activityType types.ActivityType, record *types.ActivityRecord) *diagnostics.Diagnostics {
	diag := diagnostics.New("record_info_activity")
	defer diag.Complete()

	return s.RecordActivity(ctx, record, activityType, types.ActivityLevelInfo)
}

func (s *ActivityService) RecordWarningActivity(ctx *appctx.AppContext, activityType types.ActivityType, record *types.ActivityRecord) *diagnostics.Diagnostics {
	diag := diagnostics.New("record_warning_activity")
	defer diag.Complete()

	return s.RecordActivity(ctx, record, types.ActivityTypeWarning, types.ActivityLevelWarning)
}

func (s *ActivityService) RecordErrorActivity(ctx *appctx.AppContext, activityType types.ActivityType, err types.ActivityErrorData, record *types.ActivityRecord) *diagnostics.Diagnostics {
	diag := diagnostics.New("record_error_activity")
	defer diag.Complete()

	if record == nil {
		diag.AddError("failed_to_record_error_activity", "Failed to record error activity: record is required", "activity")
		return diag
	}

	record.Error = &err

	return s.RecordActivity(ctx, record, activityType, types.ActivityLevelError)
}

func (s *ActivityService) RecordSuccessActivity(ctx *appctx.AppContext, activityType types.ActivityType, record *types.ActivityRecord) *diagnostics.Diagnostics {
	diag := diagnostics.New("record_success_activity")
	defer diag.Complete()

	record.Success = true

	return s.RecordActivity(ctx, record, activityType, types.ActivityLevelInfo)
}

func (s *ActivityService) RecordFailureActivity(ctx *appctx.AppContext, activityType types.ActivityType, err types.ActivityErrorData, record *types.ActivityRecord) *diagnostics.Diagnostics {
	diag := diagnostics.New("record_failure_activity")
	defer diag.Complete()

	record.Success = false

	return s.RecordActivity(ctx, record, activityType, types.ActivityLevelError)
}

func (s *ActivityService) RecordActivity(ctx *appctx.AppContext, record *types.ActivityRecord, activityType types.ActivityType, activityLevel types.ActivityLevel) *diagnostics.Diagnostics {
	diag := diagnostics.New("record_info_activity")
	defer diag.Complete()

	activity := utils.NewActivityFromContext(ctx)
	// if the activity tenant id is not set, use the record tenant id
	if activity.TenantID == "" || activity.TenantID == config.UnknownTenantID {
		activity.TenantID = record.TenantID
	}
	if (activity.ActorID == "" || activity.ActorID == config.UnknownUserID) && record.ActorID != "" {
		activity.ActorID = record.ActorID
	}
	if activity.ActorName == "" || activity.ActorName == "unknown" {
		activity.ActorName = record.ActorName
	}
	activity.ActivityType = activityType
	activity.ActivityLevel = activityLevel
	activity.Message = record.Message
	activity.Module = record.Module
	activity.Service = record.Service
	activity.ActorType = record.ActorType
	activity.Success = record.Success
	if record.Data != nil {
		activity.IsSensitive = record.Data.IsSensitive
		activity.Metadata = record.Data.Metadata
		activity.Tags = record.Data.Tags
		activity.StartedAt = record.Data.StartedAt
		activity.CompletedAt = record.Data.CompletedAt
	}
	if record.Error != nil {
		activity.ErrorCode = record.Error.ErrorCode
		activity.ErrorMessage = record.Error.ErrorMessage
		activity.StatusCode = record.Error.StatusCode
	}

	// checking if the activity tenant id is set if this is not a error level activity
	if activityLevel != types.ActivityLevelError {
		if activity.TenantID == "" || activity.TenantID == "unknown" {
			diag.AddError("failed_to_record_info_activity", "Failed to record info activity: tenant ID is required", "activity")
			return diag
		}
		if err := uuid.Validate(activity.TenantID); err != nil {
			diag.AddError("failed_to_record_info_activity", "Failed to record info activity: tenant ID is not a valid UUID", "activity")
			return diag
		}
	}

	dbActivity := mappers.MapActivityToEntity(activity)

	createdActivity, createdActivityDiag := s.activityStore.CreateActivity(ctx, activity.TenantID, dbActivity)
	if createdActivityDiag.HasErrors() {
		diag.Append(createdActivityDiag)
		return diag
	}

	ctx.Log().WithField("activity_id", createdActivity.ID).Info("Activity created")

	return diag
}
