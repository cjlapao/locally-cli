// Package claim provides claim management functionality.
package claim

import (
	"github.com/cjlapao/locally-cli/internal/claim/handlers"
	"github.com/cjlapao/locally-cli/internal/claim/interfaces"
	"github.com/cjlapao/locally-cli/internal/claim/service"
	"github.com/cjlapao/locally-cli/internal/database/stores"
)

// Initialize initializes the claim service
func Initialize(claimStore stores.ClaimDataStoreInterface) interfaces.ClaimServiceInterface {
	return service.Initialize(claimStore)
}

// GetInstance returns the claim service instance
func GetInstance() interfaces.ClaimServiceInterface {
	return service.GetInstance()
}

// Reset resets the claim service for testing
func Reset() {
	service.Reset()
}

// NewApiHandler creates a new claims API handler
func NewApiHandler(claimService interfaces.ClaimServiceInterface) *handlers.ClaimsApiHandler {
	return handlers.NewClaimsApiHandler(claimService)
}
