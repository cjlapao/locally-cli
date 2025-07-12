package controllers

import (
	"net/http"

	restapi "github.com/cjlapao/common-go-restapi"
	"github.com/cjlapao/common-go/helper/http_helper"
	"github.com/gorilla/mux"
)

func RegisterControllers(listener *restapi.HttpListener) {

	// Context Controller
	listener.AddController(GetCurrentContext(), http_helper.JoinUrl("contexts", "current"), "GET")
	listener.AddController(GetAllContexts(), http_helper.JoinUrl("contexts"), "GET")
	listener.AddController(AddNewContext(), http_helper.JoinUrl("contexts"), "POST")

	listener.AddController(TestAzureConnectionController(), http_helper.JoinUrl("test", "azure", "credentials"), "GET")
	listener.AddController(TestAwsConnectionController(), http_helper.JoinUrl("test", "aws", "credentials"), "GET")

	// Environment Controller
	listener.AddController(IsEnvironmentInitialized(), http_helper.JoinUrl("environments", "initialized"), "GET")
}

func MapBodyFromRequest(request *http.Request, target interface{}) error {
	err := http_helper.MapRequestBody(request, target)

	if err != nil {
		return err
	}

	return nil
}

func GetVariableName(request *http.Request, name string) string {
	vars := mux.Vars(request)
	value := vars[name]

	return value
}
