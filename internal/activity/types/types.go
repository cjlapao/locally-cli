// Package types contains the types for the activity service
package types

import "time"

type ActivityType string

const (
	ActivityTypeLogin     ActivityType = "login"
	ActivityTypeLogout    ActivityType = "logout"
	ActivityTypeCreate    ActivityType = "create"
	ActivityTypeUpdate    ActivityType = "update"
	ActivityTypeDelete    ActivityType = "delete"
	ActivityTypeBlock     ActivityType = "block"
	ActivityTypeUnblock   ActivityType = "unblock"
	ActivityTypeCall      ActivityType = "call"
	ActivityTypeStart     ActivityType = "start"
	ActivityTypeComplete  ActivityType = "complete"
	ActivityTypeFail      ActivityType = "fail"
	ActivityTypeSystem    ActivityType = "system"
	ActivityTypeSecurity  ActivityType = "security"
	ActivityTypeAudit     ActivityType = "audit"
	ActivityTypeError     ActivityType = "error"
	ActivityTypeWarning   ActivityType = "warning"
	ActivityTypeInfo      ActivityType = "info"
	ActivityTypeDebug     ActivityType = "debug"
	ActivityTypeTrace     ActivityType = "trace"
	ActivityTypeFatal     ActivityType = "fatal"
	ActivityTypePanic     ActivityType = "panic"
	ActivityTypeView      ActivityType = "view"
	ActivityTypeAvailable ActivityType = "available"
	ActivityTypeUnknown   ActivityType = "unknown"
	ActivityTypeOther     ActivityType = "other"
	ActivityTypeDisabled  ActivityType = "disabled"
	ActivityTypeEnabled   ActivityType = "enabled"
)

type ActivityLevel string

const (
	ActivityLevelInfo     ActivityLevel = "info"
	ActivityLevelWarning  ActivityLevel = "warning"
	ActivityLevelError    ActivityLevel = "error"
	ActivityLevelCritical ActivityLevel = "critical"
)

type ActorType string

const (
	ActorTypeUser    ActorType = "user"
	ActorTypeSystem  ActorType = "system"
	ActorTypeAPIKey  ActorType = "api_key"
	ActorTypeAuditor ActorType = "auditor"
	ActorTypeService ActorType = "service"
)

type ActivityRecord struct {
	TenantID      string             `json:"tenant_id" yaml:"tenant_id"`
	ActorID       string             `json:"actor_id" yaml:"actor_id"`
	ActorName     string             `json:"actor_name" yaml:"actor_name"`
	Module        string             `json:"module" yaml:"module"`
	Service       string             `json:"service" yaml:"service"`
	Message       string             `json:"message" yaml:"message"`
	Success       bool               `json:"success" yaml:"success"`
	ActorType     ActorType          `json:"actor_type" yaml:"actor_type"`
	ActivityType  ActivityType       `json:"activity_type" yaml:"activity_type"`
	ActivityLevel ActivityLevel      `json:"activity_level" yaml:"activity_level"`
	Data          *ActivityData      `json:"data" yaml:"data"`
	Error         *ActivityErrorData `json:"error" yaml:"error"`
}

type ActivityData struct {
	IsSensitive bool                   `json:"is_sensitive" yaml:"is_sensitive"`
	Metadata    map[string]interface{} `json:"metadata" yaml:"metadata"`
	Tags        []string               `json:"tags" yaml:"tags"`
	StartedAt   *time.Time             `json:"started_at" yaml:"started_at"`
	CompletedAt *time.Time             `json:"completed_at" yaml:"completed_at"`
}

// ActivityErrorData represents the error data for an activity
type ActivityErrorData struct {
	ErrorCode    string `json:"error_code" yaml:"error_code"`
	ErrorMessage string `json:"error_message" yaml:"error_message"`
	StatusCode   int    `json:"status_code" yaml:"status_code"`
}
