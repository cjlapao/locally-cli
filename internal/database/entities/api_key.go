package entities

import (
	"time"

	"github.com/cjlapao/locally-cli/internal/database/types"
)

// APIKey represents an API key for authentication
type APIKey struct {
	BaseModel
	UserID           string     `json:"user_id" gorm:"not null;type:text;index"`
	Name             string     `json:"name" gorm:"not null;type:text"`
	KeyHash          string     `json:"-" gorm:"not null;type:text;uniqueIndex"`     // Never expose the actual key
	KeyPrefix        string     `json:"key_prefix" gorm:"not null;type:text;index"`  // First 8 chars for identification
	Permissions      string     `json:"permissions" gorm:"type:text;default:'read'"` // JSON array of permissions
	TenantID         string     `json:"tenant_id" gorm:"type:text;index"`
	ExpiresAt        *time.Time `json:"expires_at" gorm:"type:timestamp"`
	LastUsedAt       *time.Time `json:"last_used_at" gorm:"type:timestamp"`
	IsActive         bool       `json:"is_active" gorm:"type:boolean;not null;default:true"`
	CreatedBy        string     `json:"created_by" gorm:"type:text"`
	RevokedAt        *time.Time `json:"revoked_at" gorm:"type:timestamp"`
	RevokedBy        string     `json:"revoked_by" gorm:"type:text"`
	RevocationReason string     `json:"revocation_reason" gorm:"type:text"`
}

// APIKeyUsage represents usage tracking for API keys
type APIKeyUsage struct {
	BaseModel
	APIKeyID     string `json:"api_key_id" gorm:"not null;type:text;index"`
	UserID       string `json:"user_id" gorm:"not null;type:text;index"`
	IPAddress    string `json:"ip_address" gorm:"type:text"`
	UserAgent    string `json:"user_agent" gorm:"type:text"`
	Endpoint     string `json:"endpoint" gorm:"type:text"`
	Method       string `json:"method" gorm:"type:text"`
	StatusCode   int    `json:"status_code" gorm:"type:int"`
	ResponseTime int64  `json:"response_time_ms" gorm:"type:bigint"`
	TenantID     string `json:"tenant_id" gorm:"type:text;index"`
}

// APIKeyPermissions defines the structure for API key permissions
type APIKeyPermissions struct {
	Read   bool              `json:"read"`
	Write  bool              `json:"write"`
	Delete bool              `json:"delete"`
	Admin  bool              `json:"admin"`
	Scopes types.StringSlice `json:"scopes"` // Specific resource scopes
}

func (APIKey) TableName() string {
	return "api_keys"
}

func (APIKeyUsage) TableName() string {
	return "api_key_usage"
}
