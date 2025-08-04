// Package models contains the models for the claim service.
package models

import (
	"fmt"

	"github.com/cjlapao/locally-cli/pkg/models"
)

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
