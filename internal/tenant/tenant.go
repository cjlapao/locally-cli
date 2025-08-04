// Package tenant provides the tenant service and its dependencies
package tenant

import (
	claimsvc "github.com/cjlapao/locally-cli/internal/claim/interfaces"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	rolesvc "github.com/cjlapao/locally-cli/internal/role/interfaces"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
	"github.com/cjlapao/locally-cli/internal/tenant/handlers"
	"github.com/cjlapao/locally-cli/internal/tenant/interfaces"
	"github.com/cjlapao/locally-cli/internal/tenant/service"
	usersvc "github.com/cjlapao/locally-cli/internal/user/interfaces"
)

// Initialize initializes the tenant service
func Initialize(tenantStore stores.TenantDataStoreInterface, userService usersvc.UserServiceInterface, roleService rolesvc.RoleServiceInterface, systemService system_interfaces.SystemServiceInterface, claimService claimsvc.ClaimServiceInterface) interfaces.TenantServiceInterface {
	return service.Initialize(tenantStore, userService, roleService, systemService, claimService)
}

// GetInstance returns the tenant service instance
func GetInstance() interfaces.TenantServiceInterface {
	return service.GetInstance()
}

// Reset resets the tenant service for testing
func Reset() {
	service.Reset()
}

// NewApiHandler creates a new tenant API handler
func NewApiHandler(tenantService interfaces.TenantServiceInterface) *handlers.APIHandler {
	return handlers.NewApiHandler(tenantService)
}
