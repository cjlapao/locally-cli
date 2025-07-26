package controllers

import (
	"encoding/json"
	"net/http"

	"github.com/cjlapao/common-go-restapi/controllers"
	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/environment"
)

func IsEnvironmentInitialized() controllers.Controller {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := appctx.FromContext(r.Context())
		env := environment.GetInstance()
		if env == nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(env.GetStatus(ctx))
	}
}
