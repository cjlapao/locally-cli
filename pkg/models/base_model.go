package models

import (
	"time"
)

type BaseModel struct {
	ID        string    `json:"id" yaml:"id"`
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
