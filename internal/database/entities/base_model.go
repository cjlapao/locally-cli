package entities

import (
	"time"

	"gorm.io/gorm"
)

type BaseModel struct {
	ID        string         `json:"id" gorm:"primarykey;type:text;not null;column:id"`
	Slug      string         `json:"slug" gorm:"not null;type:text"`
	CreatedAt time.Time      `json:"created_at" gorm:"column:created_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt time.Time      `json:"updated_at" gorm:"column:updated_at;type:timestamp;not null;default:CURRENT_TIMESTAMP"`
	DeletedAt gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`
}
