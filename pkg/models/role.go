package models

import "time"

type Role struct {
	ID          string    `json:"id" yaml:"id"`
	Slug        string    `json:"slug" yaml:"slug"`
	Name        string    `json:"name" yaml:"name"`
	Description string    `json:"description" yaml:"description"`
	IsAdmin     bool      `json:"is_admin" yaml:"is_admin"`
	IsSuperUser bool      `json:"is_super_user" yaml:"is_super_user"`
	CreatedAt   time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt   time.Time `json:"updated_at" yaml:"updated_at"`
	Matched     bool      `json:"-" yaml:"-"`
}
