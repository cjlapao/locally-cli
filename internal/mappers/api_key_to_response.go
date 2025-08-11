package mappers

import (
	api_key_models "github.com/cjlapao/locally-cli/internal/api_keys/models"
	pkg_models "github.com/cjlapao/locally-cli/pkg/models"
)

// MapApiKeyDtoToCreateResponse constructs the create response with plaintext key
func MapApiKeyDtoToCreateResponse(dto *pkg_models.ApiKey, plaintextKey string) *api_key_models.CreateApiKeyResponse {
	if dto == nil {
		return nil
	}
	resp := &api_key_models.CreateApiKeyResponse{
		ID:     dto.ID,
		Name:   dto.Name,
		Key:    plaintextKey,
		Claims: make([]string, 0),
	}
	if dto.ExpiresAt != nil {
		resp.ExpiresAt = dto.ExpiresAt.Format("02-01-2006")
	}

	// Map the claims
	claims := []string{}
	for _, claim := range dto.Claims {
		claims = append(claims, claim.Slug)
	}
	resp.Claims = claims
	return resp
}
