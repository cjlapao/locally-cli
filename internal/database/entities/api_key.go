package entities

import (
	"time"
)

// ApiKey represents an API key for authentication
type ApiKey struct {
	BaseModelWithTenant
	Name             string     `json:"name" gorm:"not null;type:text"`
	KeyHash          string     `json:"-" gorm:"not null;type:text;uniqueIndex"`    // Never expose the actual key
	KeyPrefix        string     `json:"key_prefix" gorm:"not null;type:text;index"` // First 8 chars for identification
	Claims           []Claim    `json:"claims" gorm:"many2many:api_key_claims"`
	ExpiresAt        *time.Time `json:"expires_at" gorm:"type:timestamp"`
	LastUsedAt       *time.Time `json:"last_used_at" gorm:"type:timestamp"`
	IsActive         bool       `json:"is_active" gorm:"type:boolean;not null;default:true"`
	RevokedAt        *time.Time `json:"revoked_at" gorm:"type:timestamp"`
	RevokedBy        string     `json:"revoked_by" gorm:"type:text"`
	RevocationReason string     `json:"revocation_reason" gorm:"type:text"`
}

func (ApiKey) TableName() string {
	return "api_keys"
}
