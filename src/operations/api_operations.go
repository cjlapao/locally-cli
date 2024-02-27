package operations

import (
	restapi "github.com/cjlapao/common-go-restapi"
	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/controllers"
	"github.com/cjlapao/locally-cli/environment"
)

var globalApiOperations *ApiOperation

const (
	API_OPERATION_NAME = "api_operation"
)

type ApiOperation struct {
	listener *restapi.HttpListener
}

func NewApiOperation() *ApiOperation {
	if globalApiOperations == nil {
		listener := restapi.GetHttpListener()
		env := environment.Get()
		listener.Options.ApiPrefix = env.GetString("env", common.API_PREFIX_VAR)
		port := env.GetString("env", common.API_PORT_VAR)
		if port == "" {
			port = "7750"
		}
		listener.Options.HttpPort = port

		if listener.Options.ApiPrefix != "" && listener.Options.ApiPrefix[0] != '/' {
			listener.Options.ApiPrefix = "/" + listener.Options.ApiPrefix
		}

		listener.AddJsonContent().AddLogger().AddHealthCheck()
		controllers.RegisterControllers(listener)

		globalApiOperations = &ApiOperation{
			listener: restapi.GetHttpListener(),
		}
	}

	return globalApiOperations
}

func (api *ApiOperation) GetName() string {
	return API_OPERATION_NAME
}

func (api *ApiOperation) Run(arguments ...string) {
	api.listener.Start()
}
