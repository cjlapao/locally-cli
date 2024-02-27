package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cjlapao/common-go-restapi/controllers"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/context"
	context_entities "github.com/cjlapao/locally-cli/context/entities"
	"github.com/cjlapao/locally-cli/controllers/entities"
	"github.com/cjlapao/locally-cli/environment"
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

func GetAllContexts() controllers.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		env := environment.Get()
		if env == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		config := configuration.Get()
		result := make([]entities.EnvironmentApiResponse, 0)

		if len(config.GlobalConfiguration.Contexts) == 0 {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		for _, context := range config.GlobalConfiguration.Contexts {
			environment := entities.EnvironmentApiResponse{
				Name:    context.Name,
				Id:      context.ID,
				Enabled: context.IsEnabled,
				Valid:   context.IsValid,
			}
			if context.IsValid {
				if context.Configuration.Location != nil {
					environment.Location = context.Configuration.Location.Path
					environment.Type = context.Configuration.Location.Type
				}
			}
			result = append(result, environment)
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func AddNewContext() controllers.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		var request entities.NewContextRequest
		if err := MapBodyFromRequest(r, &request); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(entities.NewApiErrorResponseFromError("bad_body", err))
			return
		}

		config := configuration.Get()
		if config.ContextExists(request.Name) {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(entities.NewApiErrorResponse("context_exist", "context already exists", 400))
			return
		}

		newContext := context.Context{
			Name: request.Name,
			Configuration: &context_entities.ContextConfiguration{
				Domain:    request.DomainName,
				Subdomain: request.SubDomainName,
			},
		}

		if err := config.AddContext(&newContext); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(entities.NewApiErrorResponseFromError("bad_context", err))
			return
		}

		fmt.Println(request.Name)

		w.WriteHeader(http.StatusOK)
	}
}
