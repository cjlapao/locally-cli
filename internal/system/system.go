// Package system provides the system service and its dependencies
package system

import (
	"github.com/cjlapao/locally-cli/internal/system/interfaces"
	"github.com/cjlapao/locally-cli/internal/system/service"
)

// Initialize initializes the system service
func Initialize() interfaces.SystemServiceInterface {
	return service.Initialize()
}

// GetInstance returns the system service instance
func GetInstance() interfaces.SystemServiceInterface {
	return service.GetInstance()
}
