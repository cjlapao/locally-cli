// Package api_keys provides API key management functionality.
package api_keys

import (
	"github.com/cjlapao/locally-cli/internal/api_keys/handlers"
	"github.com/cjlapao/locally-cli/internal/api_keys/interfaces"
	"github.com/cjlapao/locally-cli/internal/api_keys/service"
	"github.com/cjlapao/locally-cli/internal/database/stores"
)

// Initialize initializes the API keys service
func Initialize(apiKeyStore stores.ApiKeyStoreInterface) interfaces.ApiKeysServiceInterface {
	return service.Initialize(apiKeyStore)
}

// GetInstance returns the API keys service instance
func GetInstance() interfaces.ApiKeysServiceInterface {
	return service.GetInstance()
}

// Reset resets the API keys service for testing
func Reset() {
	service.Reset()
}

// NewApiHandler creates a new API keys API handler
func NewApiHandler(apiKeysService interfaces.ApiKeysServiceInterface) *handlers.ApiKeysApiHandler {
	return handlers.NewApiKeysApiHandler(apiKeysService)
}
