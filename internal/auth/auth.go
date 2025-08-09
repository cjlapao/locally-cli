// Package auth provides the public facade for the authentication service
package auth

import (
	activity_interfaces "github.com/cjlapao/locally-cli/internal/activity/interfaces"
	"github.com/cjlapao/locally-cli/internal/auth/handlers"
	auth_interfaces "github.com/cjlapao/locally-cli/internal/auth/interfaces"
	"github.com/cjlapao/locally-cli/internal/auth/service"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
)

// AuthServiceConfig represents the authentication service configuration
type AuthServiceConfig = service.AuthServiceConfig

// Initialize initializes the auth service
func Initialize(cfg AuthServiceConfig, authDataStore stores.ApiKeyStoreInterface, userStore stores.UserDataStoreInterface, tenantStore stores.TenantDataStoreInterface) (auth_interfaces.AuthServiceInterface, *diagnostics.Diagnostics) {
	return service.Initialize(cfg, authDataStore, userStore, tenantStore)
}

// GetInstance returns the auth service instance
func GetInstance() auth_interfaces.AuthServiceInterface {
	return service.GetInstance()
}

// Reset resets the auth service singleton (useful for tests)
func Reset() {
	service.Reset()
}

// NewApiHandler creates a new auth API handler
func NewApiHandler(authService auth_interfaces.AuthServiceInterface, store stores.ApiKeyStoreInterface, activityService activity_interfaces.ActivityServiceInterface) *handlers.AuthApiHandler {
	return handlers.NewApiHandler(authService, store, activityService)
}
