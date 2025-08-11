package models

import "time"

type ApiKey struct {
	BaseModelWithTenant
	Name             string     `json:"name" yaml:"name"`
	KeyHash          string     `json:"-" yaml:"-"`
	KeyPrefix        string     `json:"key_prefix" yaml:"key_prefix"`
	PlaintextKey     string     `json:"-" yaml:"-"`
	Claims           []Claim    `json:"claims" yaml:"claims"`
	ExpiresAt        *time.Time `json:"expires_at" yaml:"expires_at"`
	LastUsedAt       *time.Time `json:"last_used_at,omitempty" yaml:"last_used_at,omitempty"`
	IsActive         bool       `json:"is_active" yaml:"is_active"`
	RevokedAt        *time.Time `json:"revoked_at,omitempty" yaml:"revoked_at,omitempty"`
	RevokedBy        string     `json:"revoked_by,omitempty" yaml:"revoked_by,omitempty"`
	RevocationReason string     `json:"revocation_reason,omitempty" yaml:"revocation_reason,omitempty"`
}
