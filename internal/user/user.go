// Package user provides the user service.
package user

import (
	claimsvc "github.com/cjlapao/locally-cli/internal/claim/interfaces"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	rolesvc "github.com/cjlapao/locally-cli/internal/role/interfaces"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
	"github.com/cjlapao/locally-cli/internal/user/handlers"
	"github.com/cjlapao/locally-cli/internal/user/interfaces"
	"github.com/cjlapao/locally-cli/internal/user/service"
)

// Initialize initializes the claim service
func Initialize(userStore stores.UserDataStoreInterface, roleService rolesvc.RoleServiceInterface, claimService claimsvc.ClaimServiceInterface, systemService system_interfaces.SystemServiceInterface) interfaces.UserServiceInterface {
	return service.Initialize(userStore, roleService, claimService, systemService)
}

// GetInstance returns the claim service instance
func GetInstance() interfaces.UserServiceInterface {
	return service.GetInstance()
}

// Reset resets the claim service for testing
func Reset() {
	service.Reset()
}

// NewClaimsApiHandler creates a new claims API handler
func NewApiHandler(userService interfaces.UserServiceInterface) *handlers.ApiHandler {
	return handlers.NewUsersApiHandler(userService)
}
