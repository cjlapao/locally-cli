// Package models contains the models for the role service.
package models

import (
	"github.com/cjlapao/locally-cli/pkg/models"
)

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
