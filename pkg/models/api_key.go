package models

import "time"

type ApiKey struct {
	BaseModelWithTenant
	Name             string     `json:"name" yaml:"name"`
	KeyHash          string     `json:"-" yaml:"-"`
	KeyPrefix        string     `json:"key_prefix" yaml:"key_prefix"`
	Claims           []Claim    `json:"claims" yaml:"claims"`
	ExpiresAt        *time.Time `json:"expires_at" yaml:"expires_at"`
	LastUsedAt       *time.Time `json:"last_used_at" yaml:"last_used_at"`
	IsActive         bool       `json:"is_active" yaml:"is_active"`
	RevokedAt        *time.Time `json:"revoked_at" yaml:"revoked_at"`
	RevokedBy        string     `json:"revoked_by" yaml:"revoked_by"`
	RevocationReason string     `json:"revocation_reason" yaml:"revocation_reason"`
}
