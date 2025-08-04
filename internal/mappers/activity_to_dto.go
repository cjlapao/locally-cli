package mappers

import (
	"github.com/cjlapao/locally-cli/internal/database/entities"
	"github.com/cjlapao/locally-cli/pkg/models"
)

// MapActivityToDto converts an activity entity to a DTO for API responses
func MapActivityToDto(activity *entities.Activity) *models.Activity {
	if activity == nil {
		return nil
	}

	result := &models.Activity{
		ID:            activity.ID,
		Slug:          activity.Slug,
		ActivityType:  activity.ActivityType,
		ActivityLevel: activity.ActivityLevel,
		Description:   activity.Description,
		Module:        activity.Module,
		Service:       activity.Service,
		ActorType:     activity.ActorType,
		ActorID:       activity.ActorID,
		ActorName:     activity.ActorName,
		ActorIP:       activity.ActorIP,
		UserAgent:     activity.UserAgent,
		TargetType:    activity.TargetType,
		TargetID:      activity.TargetID,
		TargetName:    activity.TargetName,
		TenantID:      activity.TenantID,
		SessionID:     activity.SessionID,
		RequestID:     activity.RequestID,
		CorrelationID: activity.CorrelationID,
		StartedAt:     activity.StartedAt,
		CompletedAt:   activity.CompletedAt,
		DurationMs:    activity.DurationMs,
		Success:       activity.Success,
		ErrorCode:     activity.ErrorCode,
		ErrorMessage:  activity.ErrorMessage,
		StatusCode:    activity.StatusCode,
		IsSensitive:   activity.IsSensitive,
		RetentionDays: activity.RetentionDays,
		CreatedAt:     activity.CreatedAt,
		UpdatedAt:     activity.UpdatedAt,
	}

	// Map metadata
	if activity.Metadata.Get() != nil {
		result.Metadata = activity.Metadata.Get()
	} else {
		result.Metadata = make(map[string]interface{})
	}

	// Map tags
	if activity.Tags != nil {
		result.Tags = activity.Tags
	} else {
		result.Tags = []string{}
	}

	return result
}

// MapActivitiesToDto converts a slice of activity entities to DTOs
func MapActivitiesToDto(activities []entities.Activity) []models.Activity {
	result := make([]models.Activity, len(activities))
	for i, activity := range activities {
		dto := MapActivityToDto(&activity)
		result[i] = *dto
	}
	return result
}

// MapActivitySummaryToDto converts an activity summary entity to a DTO
func MapActivitySummaryToDto(summary *entities.ActivitySummary) *models.ActivitySummary {
	if summary == nil {
		return nil
	}

	result := &models.ActivitySummary{
		ID:              summary.ID,
		Slug:            summary.Slug,
		SummaryType:     summary.SummaryType,
		SummaryDate:     summary.SummaryDate,
		Module:          summary.Module,
		Service:         summary.Service,
		TenantID:        summary.TenantID,
		TotalActivities: summary.TotalActivities,
		SuccessCount:    summary.SuccessCount,
		ErrorCount:      summary.ErrorCount,
		UniqueActors:    summary.UniqueActors,
		AvgDurationMs:   summary.AvgDurationMs,
		MaxDurationMs:   summary.MaxDurationMs,
		MinDurationMs:   summary.MinDurationMs,
		CreatedAt:       summary.CreatedAt,
		UpdatedAt:       summary.UpdatedAt,
	}

	// Map top actors
	if summary.TopActors.Get() != nil {
		result.TopActors = summary.TopActors.Get()
	} else {
		result.TopActors = []map[string]interface{}{}
	}

	// Map activity breakdown
	if summary.ActivityBreakdown.Get() != nil {
		result.ActivityBreakdown = summary.ActivityBreakdown.Get()
	} else {
		result.ActivityBreakdown = make(map[string]int64)
	}

	return result
}

// MapActivitySummariesToDto converts a slice of activity summary entities to DTOs
func MapActivitySummariesToDto(summaries []entities.ActivitySummary) []models.ActivitySummary {
	result := make([]models.ActivitySummary, len(summaries))
	for i, summary := range summaries {
		dto := MapActivitySummaryToDto(&summary)
		result[i] = *dto
	}
	return result
}

// MapActivityFilterToEntity converts an activity filter DTO to an entity filter
func MapActivityFilterToEntity(filter *models.ActivityFilter) *entities.ActivityFilter {
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
