package models

import (
	"time"
)

type BaseModel struct {
	ID        string    `json:"id" yaml:"id"`
	Slug      string    `json:"slug" yaml:"slug"`
	CreatedBy string    `json:"created_by" yaml:"created_by"`
	UpdatedBy string    `json:"updated_by" yaml:"updated_by"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
}

func (b *BaseModel) GetID() string {
	return b.ID
}

func (b *BaseModel) GetCreatedAt() time.Time {
	return b.CreatedAt
}

func (b *BaseModel) GetUpdatedAt() time.Time {
	return b.UpdatedAt
}

type BaseModelWithTenant struct {
	ID        string    `json:"id" yaml:"id"`
	TenantID  string    `json:"tenant_id" yaml:"tenant_id"`
	Slug      string    `json:"slug" yaml:"slug"`
	CreatedBy string    `json:"created_by" yaml:"created_by"`
	UpdatedBy string    `json:"updated_by" yaml:"updated_by"`
	CreatedAt time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt time.Time `json:"updated_at" yaml:"updated_at"`
}

func (b *BaseModelWithTenant) GetTenantID() string {
	return b.TenantID
}
