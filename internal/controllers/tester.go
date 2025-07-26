package controllers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/cjlapao/common-go-restapi/controllers"
	controller_entities "github.com/cjlapao/locally-cli/internal/controllers/entities"
	"github.com/cjlapao/locally-cli/internal/entities"
	aws_service "github.com/cjlapao/locally-cli/internal/services/aws"
	azure_service "github.com/cjlapao/locally-cli/internal/services/azure"
)

func TestAzureConnectionController() controllers.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := azure_service.Get()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(controller_entities.NewApiErrorResponseFromError("bad_body", err))
			return
		}

		var credentials entities.AzureCredentials
		if err := json.Unmarshal(body, &credentials); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(controller_entities.NewApiErrorResponseFromError("bad_body", err))
			return
		}

		if err := svc.TestConnection(credentials); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(controller_entities.NewApiErrorResponseFromError("unauthorized", err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

func TestAwsConnectionController() controllers.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		svc := aws_service.Get()
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(controller_entities.NewApiErrorResponseFromError("bad_body", err))
			return
		}

		var credentials entities.AwsCredentials
		if err := json.Unmarshal(body, &credentials); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(controller_entities.NewApiErrorResponseFromError("bad_body", err))
			return
		}

		if err := svc.TestConnection(credentials); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(controller_entities.NewApiErrorResponseFromError("unauthorized", err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
