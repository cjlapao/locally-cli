package models

import (
	"strings"
	"time"

	"github.com/cjlapao/locally-cli/internal/config"
)

type User struct {
	ID                    string    `json:"id" yaml:"id"`
	Slug                  string    `json:"slug" yaml:"slug"`
	Name                  string    `json:"name" yaml:"name"`
	Username              string    `json:"username" yaml:"username"`
	Password              string    `json:"password" yaml:"password"`
	Email                 string    `json:"email" yaml:"email"`
	Roles                 []Role    `json:"roles" yaml:"roles"`
	Claims                []Claim   `json:"claims" yaml:"claims"`
	Status                string    `json:"status" yaml:"status"`
	TenantID              string    `json:"tenant_id" yaml:"tenant_id"`
	TwoFactorEnabled      bool      `json:"two_factor_enabled" yaml:"two_factor_enabled"`
	TwoFactorSecret       string    `json:"two_factor_secret" yaml:"two_factor_secret"`
	TwoFactorVerified     bool      `json:"two_factor_verified" yaml:"two_factor_verified"`
	Blocked               bool      `json:"blocked" yaml:"blocked"`
	RefreshToken          string    `json:"refresh_token" yaml:"refresh_token"`
	RefreshTokenExpiresAt time.Time `json:"refresh_token_expires_at" yaml:"refresh_token_expires_at"`
	CreatedAt             time.Time `json:"created_at" yaml:"created_at"`
	UpdatedAt             time.Time `json:"updated_at" yaml:"updated_at"`
}

func (u *User) IsSuperUser() bool {
	for _, role := range u.Roles {
		if strings.EqualFold(role.Name, config.SuperUserRole) || strings.EqualFold(role.Slug, config.SuperUserRole) {
			return true
		}
	}
	return false
}

func (u *User) HasRole(roleIdOrSlug string) bool {
	for _, role := range u.Roles {
		if strings.EqualFold(role.Name, roleIdOrSlug) || strings.EqualFold(role.Slug, roleIdOrSlug) || role.ID == roleIdOrSlug {
			return true
		}
	}
	return false
}
