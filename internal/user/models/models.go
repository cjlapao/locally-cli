// Package models contains the models for the user service.
package models

import (
	"fmt"

	"github.com/cjlapao/locally-cli/pkg/models"
)

type CreateUserRequest struct {
	ID               string `json:"-" yaml:"-"`
	Name             string `json:"name" yaml:"name" validate:"required"`
	Username         string `json:"username" yaml:"username" validate:"required"`
	Password         string `json:"password" yaml:"password" validate:"required,password_complexity"`
	Email            string `json:"email" yaml:"email" validate:"required,email"`
	TwoFactorEnabled bool   `json:"two_factor_enabled" yaml:"two_factor_enabled"`
	Role             string `json:"role" yaml:"role" validate:"required"`
}

type CreateUserResponse struct {
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
}

type UpdateUserRequest struct {
	Name     string `json:"name" yaml:"name"`
	Username string `json:"username" yaml:"username"`
	Password string `json:"password" yaml:"password"`
	Email    string `json:"email" yaml:"email"`
	Role     string `json:"role" yaml:"role"`
}

type UpdateUserResponse struct {
	ID     string `json:"id" yaml:"id"`
	Name   string `json:"name" yaml:"name"`
	Status string `json:"status" yaml:"status"`
}

type UpdateUserPasswordRequest struct {
	Password string `json:"password" yaml:"password" validate:"required,password_complexity"`
}

type CreateRoleRequest struct {
	Name          string               `json:"name" yaml:"name" validate:"required"`
	Description   string               `json:"description" yaml:"description"`
	SecurityLevel models.SecurityLevel `json:"security_level" yaml:"security_level" validate:"required"`
}

type CreateRoleResponse struct {
	ID            string               `json:"id" yaml:"id"`
	Name          string               `json:"name" yaml:"name"`
	Status        string               `json:"status" yaml:"status"`
	SecurityLevel models.SecurityLevel `json:"security_level" yaml:"security_level"`
}

type UpdateRoleRequest struct {
	ID            string               `json:"-" yaml:"-"`
	Slug          string               `json:"slug" yaml:"slug"`
	Name          string               `json:"name" yaml:"name"`
	Description   string               `json:"description" yaml:"description"`
	SecurityLevel models.SecurityLevel `json:"security_level" yaml:"security_level"`
}

type CreateClaimRequest struct {
	Module        string               `json:"module" yaml:"module" validate:"required"`
	Service       string               `json:"service" yaml:"service" validate:"required"`
	Action        models.AccessLevel   `json:"action" yaml:"action" validate:"required"`
	SecurityLevel models.SecurityLevel `json:"security_level" yaml:"security_level" validate:"required"`
}

func (c *CreateClaimRequest) GetSlug() string {
	return fmt.Sprintf("%s::%s::%s", c.Service, c.Module, c.Action)
}

type UpdateClaimRequest struct {
	ID            string               `json:"-" yaml:"-"`
	SecurityLevel models.SecurityLevel `json:"security_level" yaml:"security_level" validate:"required"`
}
