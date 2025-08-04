// Package models contains the models for the tenant service
package models

type TenantCreateRequest struct {
	ID                string                 `json:"-" yaml:"-"`
	Name              string                 `json:"name" yaml:"name" validate:"required"`
	Description       string                 `json:"description" yaml:"description"`
	Domain            string                 `json:"domain" yaml:"domain" validate:"required"`
	ContactEmail      string                 `json:"contact_email" yaml:"contact_email" validate:"required,email"`
	CreateAdminUser   bool                   `json:"-" yaml:"-"`
	AdminUser         string                 `json:"admin_user" yaml:"admin_user" validate:"required"`
	AdminPassword     string                 `json:"admin_password" yaml:"admin_password" validate:"required,password_complexity"`
	AdminName         string                 `json:"admin_name" yaml:"admin_name" validate:"required"`
	AdminContactEmail string                 `json:"admin_contact_email" yaml:"admin_contact_email" validate:"required,email"`
	Metadata          map[string]interface{} `json:"metadata" yaml:"metadata"`
}

type TenantUpdateRequest struct {
	ID           string `json:"id" yaml:"id"`
	Name         string `json:"name" yaml:"name"`
	Description  string `json:"description" yaml:"description"`
	Domain       string `json:"domain" yaml:"domain"`
	OwnerID      string `json:"owner_id" yaml:"owner_id"`
	ContactEmail string `json:"contact_email" yaml:"contact_email"`
}
