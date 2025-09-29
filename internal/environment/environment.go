package environment

import (
	"github.com/cjlapao/locally-cli/internal/database/stores"
	"github.com/cjlapao/locally-cli/internal/environment/handlers"
	"github.com/cjlapao/locally-cli/internal/environment/interfaces"
	"github.com/cjlapao/locally-cli/internal/environment/service"
)

func Initialize(environmentStore stores.EnvironmentDataStoreInterface) interfaces.EnvironmentServiceInterface {
	return service.Initialize(environmentStore)
}

func GetInstance() interfaces.EnvironmentServiceInterface {
	return service.GetInstance()
}

func Reset() {
	service.Reset()
}

func NewApiHandler(environmentService interfaces.EnvironmentServiceInterface) *handlers.EnvironmentApiHandler {
	return handlers.NewEnvironmentApiHandler(environmentService)
}
