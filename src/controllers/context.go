package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/cjlapao/common-go-restapi/controllers"
	"github.com/cjlapao/locally-cli/configuration"
)

func GetCurrentContext() controllers.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		config := configuration.Get()
		context := config.GetCurrentContext()
		if context == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(context.Name)
	}
}
