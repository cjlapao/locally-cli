package mappers

import (
	"time"

	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

// MapCreateActivityRequestToEntity converts a create activity request to an entity
func MapCreateActivityRequestToEntity(request *models.CreateActivityRequest) *entities.Activity {
	if request == nil {
		return nil
	}

	activity := &entities.Activity{
		ActivityType:  request.ActivityType,
		ActivityLevel: request.ActivityLevel,
		Message:       request.Message,
		Module:        request.Module,
		Service:       request.Service,
		ActorType:     request.ActorType,
		ActorID:       request.ActorID,
		ActorName:     request.ActorName,
		ActorIP:       request.ActorIP,
		UserAgent:     request.UserAgent,
		RequestID:     request.RequestID,
		CorrelationID: request.CorrelationID,
		DurationMs:    request.DurationMs,
		Success:       request.Success,
		ErrorCode:     request.ErrorCode,
		ErrorMessage:  request.ErrorMessage,
		StatusCode:    request.StatusCode,
		IsSensitive:   request.IsSensitive,
		RetentionDays: request.RetentionDays,
	}

	// Set started at time
	if request.StartedAt != nil {
		activity.StartedAt = request.StartedAt
	} else {
		now := time.Now()
		activity.StartedAt = &now
	}

	// Set completed at time
	activity.CompletedAt = request.CompletedAt

	// Map metadata
	if request.Metadata != nil {
		activity.Metadata.Set(request.Metadata)
	}

	// Map tags
	if request.Tags != nil {
		activity.Tags = request.Tags
	}

	return activity
}

// MapUpdateActivityRequestToEntity converts an update activity request to an entity
func MapUpdateActivityRequestToEntity(request *models.UpdateActivityRequest, existingActivity *entities.Activity) *entities.Activity {
	if request == nil || existingActivity == nil {
		return nil
	}

	// Create a copy of the existing activity
	updatedActivity := *existingActivity

	// Update fields if provided
	if request.Message != "" {
		updatedActivity.Message = request.Message
	}

	if request.CompletedAt != nil {
		updatedActivity.CompletedAt = request.CompletedAt
	}

	if request.DurationMs > 0 {
		updatedActivity.DurationMs = request.DurationMs
	}

	updatedActivity.Success = request.Success

	if request.ErrorCode != "" {
		updatedActivity.ErrorCode = request.ErrorCode
	}

	if request.ErrorMessage != "" {
		updatedActivity.ErrorMessage = request.ErrorMessage
	}

	if request.StatusCode > 0 {
		updatedActivity.StatusCode = request.StatusCode
	}

	updatedActivity.IsSensitive = request.IsSensitive

	if request.RetentionDays > 0 {
		updatedActivity.RetentionDays = request.RetentionDays
	}

	// Map metadata
	if request.Metadata != nil {
		updatedActivity.Metadata.Set(request.Metadata)
	}

	// Map tags
	if request.Tags != nil {
		updatedActivity.Tags = request.Tags
	}

	return &updatedActivity
}

// MapActivityDtoToEntity converts an activity DTO to an entity
func MapActivityToEntity(dto *models.Activity) *entities.Activity {
	if dto == nil {
		return nil
	}

	activity := &entities.Activity{
		ActivityType:  dto.ActivityType,
		ActivityLevel: dto.ActivityLevel,
		Message:       dto.Message,
		Module:        dto.Module,
		Service:       dto.Service,
		ActorType:     dto.ActorType,
		ActorID:       dto.ActorID,
		ActorName:     dto.ActorName,
		ActorIP:       dto.ActorIP,
		UserAgent:     dto.UserAgent,
		RequestID:     dto.RequestID,
		CorrelationID: dto.CorrelationID,
		StartedAt:     dto.StartedAt,
		CompletedAt:   dto.CompletedAt,
		DurationMs:    dto.DurationMs,
		Success:       dto.Success,
		ErrorCode:     dto.ErrorCode,
		ErrorMessage:  dto.ErrorMessage,
		StatusCode:    dto.StatusCode,
		IsSensitive:   dto.IsSensitive,
		RetentionDays: dto.RetentionDays,
	}

	if dto.TenantID != "" {
		activity.TenantID = dto.TenantID
	}

	// Map metadata
	if dto.Metadata != nil {
		activity.Metadata.Set(dto.Metadata)
	}

	// Map tags
	if dto.Tags != nil {
		activity.Tags = dto.Tags
	}

	return activity
}

// MapActivitySummaryDtoToEntity converts an activity summary DTO to an entity
func MapActivitySummaryDtoToEntity(dto *models.ActivitySummary) *entities.ActivitySummary {
	if dto == nil {
		return nil
	}

	summary := &entities.ActivitySummary{
		SummaryType:     dto.SummaryType,
		SummaryDate:     dto.SummaryDate,
		Module:          dto.Module,
		Service:         dto.Service,
		TotalActivities: dto.TotalActivities,
		SuccessCount:    dto.SuccessCount,
		ErrorCount:      dto.ErrorCount,
		UniqueActors:    dto.UniqueActors,
		AvgDurationMs:   dto.AvgDurationMs,
		MaxDurationMs:   dto.MaxDurationMs,
		MinDurationMs:   dto.MinDurationMs,
	}

	if dto.TenantID != "" {
		summary.TenantID = dto.TenantID
	}

	// Map top actors
	if dto.TopActors != nil {
		summary.TopActors.Set(dto.TopActors)
	}

	// Map activity breakdown
	if dto.ActivityBreakdown != nil {
		summary.ActivityBreakdown.Set(dto.ActivityBreakdown)
	}

	return summary
}

// MapActivityFilterDtoToEntity converts an activity filter DTO to an entity filter
func MapActivityFilterDtoToEntity(filter *models.ActivityFilter) *entities.ActivityFilter {
	if filter == nil {
		return nil
	}

	return &entities.ActivityFilter{
		Module:        filter.Module,
		Service:       filter.Service,
		ActivityType:  filter.ActivityType,
		ActivityLevel: filter.ActivityLevel,
		ActorType:     filter.ActorType,
		ActorID:       filter.ActorID,
		TargetType:    filter.TargetType,
		TargetID:      filter.TargetID,
		TenantID:      filter.TenantID,
		Success:       filter.Success,
		IsSensitive:   filter.IsSensitive,
		Tags:          filter.Tags,
		StartedAtFrom: filter.StartedAtFrom,
		StartedAtTo:   filter.StartedAtTo,
		CreatedAtFrom: filter.CreatedAtFrom,
		CreatedAtTo:   filter.CreatedAtTo,
	}
}
