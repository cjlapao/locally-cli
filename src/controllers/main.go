package controllers

import (
	restapi "github.com/cjlapao/common-go-restapi"
	"github.com/cjlapao/common-go/helper/http_helper"
)

func RegisterControllers(listener *restapi.HttpListener) {
	// Context Controller
	listener.AddController(GetCurrentContext(), "context", "GET")

	// Environment Controller
	listener.AddController(IsEnvironmentInitialized(), http_helper.JoinUrl("environment", "initialized"), "GET")
}
