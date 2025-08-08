// Package models contains the models for the API keys service.
package models

import (
	"github.com/cjlapao/locally-cli/pkg/models"
)

type CreateApiKeyRequest struct {
	Name          string               `json:"name" yaml:"name" validate:"required"`
	ExpiresAt     string               `json:"expires_at" yaml:"expires_at"`
	Claims        []string             `json:"claims" yaml:"claims"`
	SecurityLevel models.SecurityLevel `json:"security_level" yaml:"security_level"`
}

type RevokeApiKeyRequest struct {
	RevocationReason string `json:"revocation_reason" yaml:"revocation_reason" validate:"required"`
}
