// Package role provides role management functionality.
package role

import (
	claim_interfaces "github.com/cjlapao/locally-cli/internal/claim/interfaces"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/role/handlers"
	"github.com/cjlapao/locally-cli/internal/role/interfaces"
	"github.com/cjlapao/locally-cli/internal/role/service"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
)

// Initialize initializes the role service
func Initialize(roleStore stores.RoleDataStoreInterface, systemService system_interfaces.SystemServiceInterface, claimService claim_interfaces.ClaimServiceInterface) interfaces.RoleServiceInterface {
	return service.Initialize(roleStore, systemService, claimService)
}

// GetInstance returns the role service instance
func GetInstance() interfaces.RoleServiceInterface {
	return service.GetInstance()
}

// Reset resets the role service for testing
func Reset() {
	service.Reset()
}

// NewApiHandler creates a new role API handler
func NewApiHandler(roleService interfaces.RoleServiceInterface) *handlers.RolesApiHandler {
	return handlers.NewRolesApiHandler(roleService)
}
