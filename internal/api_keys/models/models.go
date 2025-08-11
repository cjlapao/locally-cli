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

type CreateApiKeyResponse struct {
	ID        string   `json:"id" yaml:"id"`
	Name      string   `json:"name" yaml:"name"`
	ExpiresAt string   `json:"expires_at" yaml:"expires_at"`
	Key       string   `json:"key" yaml:"key"`
	Claims    []string `json:"claims" yaml:"claims"`
}

type RevokeApiKeyRequest struct {
	RevocationReason string `json:"revocation_reason" yaml:"revocation_reason" validate:"required"`
}
