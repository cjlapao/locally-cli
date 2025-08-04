package models

import (
	"time"
)

// ActivityType constants
const (
	ActivityTypeUserLogin            = "user_login"
	ActivityTypeUserLogout           = "user_logout"
	ActivityTypeUserCreate           = "user_create"
	ActivityTypeUserUpdate           = "user_update"
	ActivityTypeUserDelete           = "user_delete"
	ActivityTypeUserBlock            = "user_block"
	ActivityTypeUserUnblock          = "user_unblock"
	ActivityTypeAPICall              = "api_call"
	ActivityTypeAPICreate            = "api_create"
	ActivityTypeAPIUpdate            = "api_update"
	ActivityTypeAPIDelete            = "api_delete"
	ActivityTypeCertificateCreate    = "certificate_create"
	ActivityTypeCertificateUpdate    = "certificate_update"
	ActivityTypeCertificateDelete    = "certificate_delete"
	ActivityTypePipelineStart        = "pipeline_start"
	ActivityTypePipelineComplete     = "pipeline_complete"
	ActivityTypePipelineFail         = "pipeline_fail"
	ActivityTypeInfrastructureCreate = "infrastructure_create"
	ActivityTypeInfrastructureUpdate = "infrastructure_update"
	ActivityTypeInfrastructureDelete = "infrastructure_delete"
	ActivityTypeSystemEvent          = "system_event"
	ActivityTypeSecurityEvent        = "security_event"
)

// ActivityLevel constants
const (
	ActivityLevelInfo     = "info"
	ActivityLevelWarning  = "warning"
	ActivityLevelError    = "error"
	ActivityLevelCritical = "critical"
)

// ActorType constants
const (
	ActorTypeUser    = "user"
	ActorTypeSystem  = "system"
	ActorTypeAPIKey  = "api_key"
	ActorTypeService = "service"
)

// TargetType constants
const (
	TargetTypeUser           = "user"
	TargetTypeCertificate    = "certificate"
	TargetTypePipeline       = "pipeline"
	TargetTypeInfrastructure = "infrastructure"
	TargetTypeAPIKey         = "api_key"
	TargetTypeTenant         = "tenant"
	TargetTypeConfiguration  = "configuration"
)

// Activity represents a user or system activity for API responses
type Activity struct {
	ID            string                 `json:"id" yaml:"id"`
	Slug          string                 `json:"slug" yaml:"slug"`
	ActivityType  string                 `json:"activity_type" yaml:"activity_type"`
	ActivityLevel string                 `json:"activity_level" yaml:"activity_level"`
	Description   string                 `json:"description" yaml:"description"`
	Module        string                 `json:"module" yaml:"module"`
	Service       string                 `json:"service" yaml:"service"`
	ActorType     string                 `json:"actor_type" yaml:"actor_type"`
	ActorID       string                 `json:"actor_id" yaml:"actor_id"`
	ActorName     string                 `json:"actor_name" yaml:"actor_name"`
	ActorIP       string                 `json:"actor_ip" yaml:"actor_ip"`
	UserAgent     string                 `json:"user_agent" yaml:"user_agent"`
	TargetType    string                 `json:"target_type" yaml:"target_type"`
	TargetID      string                 `json:"target_id" yaml:"target_id"`
	TargetName    string                 `json:"target_name" yaml:"target_name"`
	TenantID      string                 `json:"tenant_id" yaml:"tenant_id"`
	SessionID     string                 `json:"session_id" yaml:"session_id"`
	RequestID     string                 `json:"request_id" yaml:"request_id"`
	CorrelationID string                 `json:"correlation_id" yaml:"correlation_id"`
	Metadata      map[string]interface{} `json:"metadata" yaml:"metadata"`
	Tags          []string               `json:"tags" yaml:"tags"`
	StartedAt     time.Time              `json:"started_at" yaml:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at" yaml:"completed_at"`
	DurationMs    int64                  `json:"duration_ms" yaml:"duration_ms"`
	Success       bool                   `json:"success" yaml:"success"`
	ErrorCode     string                 `json:"error_code" yaml:"error_code"`
	ErrorMessage  string                 `json:"error_message" yaml:"error_message"`
	StatusCode    int                    `json:"status_code" yaml:"status_code"`
	IsSensitive   bool                   `json:"is_sensitive" yaml:"is_sensitive"`
	RetentionDays int                    `json:"retention_days" yaml:"retention_days"`
	CreatedAt     time.Time              `json:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time              `json:"updated_at" yaml:"updated_at"`
}

// ActivitySummary represents aggregated activity data for API responses
type ActivitySummary struct {
	ID                string                   `json:"id" yaml:"id"`
	Slug              string                   `json:"slug" yaml:"slug"`
	SummaryType       string                   `json:"summary_type" yaml:"summary_type"`
	SummaryDate       time.Time                `json:"summary_date" yaml:"summary_date"`
	Module            string                   `json:"module" yaml:"module"`
	Service           string                   `json:"service" yaml:"service"`
	TenantID          string                   `json:"tenant_id" yaml:"tenant_id"`
	TotalActivities   int64                    `json:"total_activities" yaml:"total_activities"`
	SuccessCount      int64                    `json:"success_count" yaml:"success_count"`
	ErrorCount        int64                    `json:"error_count" yaml:"error_count"`
	UniqueActors      int64                    `json:"unique_actors" yaml:"unique_actors"`
	TopActors         []map[string]interface{} `json:"top_actors" yaml:"top_actors"`
	AvgDurationMs     float64                  `json:"avg_duration_ms" yaml:"avg_duration_ms"`
	MaxDurationMs     int64                    `json:"max_duration_ms" yaml:"max_duration_ms"`
	MinDurationMs     int64                    `json:"min_duration_ms" yaml:"min_duration_ms"`
	ActivityBreakdown map[string]int64         `json:"activity_breakdown" yaml:"activity_breakdown"`
	CreatedAt         time.Time                `json:"created_at" yaml:"created_at"`
	UpdatedAt         time.Time                `json:"updated_at" yaml:"updated_at"`
}

// ActivityFilter represents filtering options for activity queries
type ActivityFilter struct {
	Module        []string   `json:"module" yaml:"module"`
	Service       []string   `json:"service" yaml:"service"`
	ActivityType  []string   `json:"activity_type" yaml:"activity_type"`
	ActivityLevel []string   `json:"activity_level" yaml:"activity_level"`
	ActorType     []string   `json:"actor_type" yaml:"actor_type"`
	ActorID       []string   `json:"actor_id" yaml:"actor_id"`
	TargetType    []string   `json:"target_type" yaml:"target_type"`
	TargetID      []string   `json:"target_id" yaml:"target_id"`
	TenantID      []string   `json:"tenant_id" yaml:"tenant_id"`
	Success       *bool      `json:"success" yaml:"success"`
	IsSensitive   *bool      `json:"is_sensitive" yaml:"is_sensitive"`
	Tags          []string   `json:"tags" yaml:"tags"`
	StartedAtFrom *time.Time `json:"started_at_from" yaml:"started_at_from"`
	StartedAtTo   *time.Time `json:"started_at_to" yaml:"started_at_to"`
	CreatedAtFrom *time.Time `json:"created_at_from" yaml:"created_at_from"`
	CreatedAtTo   *time.Time `json:"created_at_to" yaml:"created_at_to"`
}

// CreateActivityRequest represents a request to create a new activity
type CreateActivityRequest struct {
	ActivityType  string                 `json:"activity_type" yaml:"activity_type" validate:"required"`
	ActivityLevel string                 `json:"activity_level" yaml:"activity_level" validate:"required"`
	Description   string                 `json:"description" yaml:"description" validate:"required"`
	Module        string                 `json:"module" yaml:"module" validate:"required"`
	Service       string                 `json:"service" yaml:"service" validate:"required"`
	ActorType     string                 `json:"actor_type" yaml:"actor_type" validate:"required"`
	ActorID       string                 `json:"actor_id" yaml:"actor_id"`
	ActorName     string                 `json:"actor_name" yaml:"actor_name"`
	ActorIP       string                 `json:"actor_ip" yaml:"actor_ip"`
	UserAgent     string                 `json:"user_agent" yaml:"user_agent"`
	TargetType    string                 `json:"target_type" yaml:"target_type"`
	TargetID      string                 `json:"target_id" yaml:"target_id"`
	TargetName    string                 `json:"target_name" yaml:"target_name"`
	SessionID     string                 `json:"session_id" yaml:"session_id"`
	RequestID     string                 `json:"request_id" yaml:"request_id"`
	CorrelationID string                 `json:"correlation_id" yaml:"correlation_id"`
	Metadata      map[string]interface{} `json:"metadata" yaml:"metadata"`
	Tags          []string               `json:"tags" yaml:"tags"`
	StartedAt     *time.Time             `json:"started_at" yaml:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at" yaml:"completed_at"`
	DurationMs    int64                  `json:"duration_ms" yaml:"duration_ms"`
	Success       bool                   `json:"success" yaml:"success"`
	ErrorCode     string                 `json:"error_code" yaml:"error_code"`
	ErrorMessage  string                 `json:"error_message" yaml:"error_message"`
	StatusCode    int                    `json:"status_code" yaml:"status_code"`
	IsSensitive   bool                   `json:"is_sensitive" yaml:"is_sensitive"`
	RetentionDays int                    `json:"retention_days" yaml:"retention_days"`
}

// UpdateActivityRequest represents a request to update an existing activity
type UpdateActivityRequest struct {
	Description   string                 `json:"description" yaml:"description"`
	CompletedAt   *time.Time             `json:"completed_at" yaml:"completed_at"`
	DurationMs    int64                  `json:"duration_ms" yaml:"duration_ms"`
	Success       bool                   `json:"success" yaml:"success"`
	ErrorCode     string                 `json:"error_code" yaml:"error_code"`
	ErrorMessage  string                 `json:"error_message" yaml:"error_message"`
	StatusCode    int                    `json:"status_code" yaml:"status_code"`
	Metadata      map[string]interface{} `json:"metadata" yaml:"metadata"`
	Tags          []string               `json:"tags" yaml:"tags"`
	IsSensitive   bool                   `json:"is_sensitive" yaml:"is_sensitive"`
	RetentionDays int                    `json:"retention_days" yaml:"retention_days"`
}
