// Package activity
package activity

import (
	"github.com/cjlapao/locally-cli/internal/activity/handlers"
	"github.com/cjlapao/locally-cli/internal/activity/interfaces"
	"github.com/cjlapao/locally-cli/internal/activity/service"
	"github.com/cjlapao/locally-cli/internal/database/stores"
	system_interfaces "github.com/cjlapao/locally-cli/internal/system/interfaces"
)

// Initialize initializes the claim service
func Initialize(activityStore stores.ActivityDataStoreInterface) interfaces.ActivityServiceInterface {
	return service.Initialize(activityStore)
}

// GetInstance returns the claim service instance
func GetInstance() interfaces.ActivityServiceInterface {
	return service.GetInstance()
}

// Reset resets the claim service for testing
func Reset() {
	service.Reset()
}

// NewApiHandler creates a new claims API handler
func NewApiHandler(activityService interfaces.ActivityServiceInterface, systemService system_interfaces.SystemServiceInterface) *handlers.ActivityApiHandler {
	return handlers.NewActivityApiHandler(activityService, systemService)
}
