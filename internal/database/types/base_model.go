// Package types contains the types for the database
package types

import (
	"time"
)

type BaseModel struct {
	ID        string    `gorm:"primarykey;type:uuid;not null;column:id" json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
