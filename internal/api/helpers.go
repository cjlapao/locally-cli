package api

import (
	"fmt"
	"net/http"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/gorilla/mux"
)

func GetTenantIDFromRequest(r *http.Request) (string, error) {
	tenantID := ""
	contextTenantID := r.Context().Value(config.TenantIDContextKey)
	if contextTenantID == nil || contextTenantID == "" {
		vars := mux.Vars(r)
		tenantID = vars["tenant_id"]
	} else {
		tenantID = contextTenantID.(string)
	}

	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}

	return tenantID, nil
}
