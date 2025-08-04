package models

import "time"

type Role struct {
	ID            string        `json:"id" yaml:"id"`
	TenantID      string        `json:"tenant_id" yaml:"tenant_id"`
	Slug          string        `json:"slug" yaml:"slug"`
	Name          string        `json:"name" yaml:"name"`
	Description   string        `json:"description" yaml:"description"`
	CreatedAt     time.Time     `json:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at" yaml:"updated_at"`
	SecurityLevel SecurityLevel `json:"security_level" yaml:"security_level"`
	Claims        []Claim       `json:"default_claims,omitempty" yaml:"default_claims,omitempty"`
	Matched       bool          `json:"-" yaml:"-"`
}
