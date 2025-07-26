package models

import "time"

type Tenant struct {
	ID            string     `json:"id" yaml:"id"`
	Name          string     `json:"name" yaml:"name" validate:"required"`
	Slug          string     `json:"slug" yaml:"slug"`
	Description   string     `json:"description" yaml:"description"`
	Domain        string     `json:"domain" yaml:"domain" validate:"required"`
	OwnerID       string     `json:"owner_id" yaml:"owner_id" validate:"required"`
	ContactEmail  string     `json:"contact_email" yaml:"contact_email" validate:"required,email"`
	Status        string     `json:"status" yaml:"status"`
	ActivatedAt   *time.Time `json:"activated_at" yaml:"activated_at"`
	DeactivatedAt *time.Time `json:"deactivated_at" yaml:"deactivated_at"`
	Metadata      string     `json:"metadata" yaml:"metadata"`
	LogoURL       string     `json:"logo_url" yaml:"logo_url"`
	Require2FA    bool       `json:"require_2fa" yaml:"require_2fa"`
	CreatedAt     time.Time  `json:"created_at" yaml:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at" yaml:"updated_at"`
}
