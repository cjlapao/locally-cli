package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/validation"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/gorilla/mux"
)

func GetTenantIDFromRequest(r *http.Request) (string, error) {
	tenantID := ""
	contextTenantID := r.Context().Value(config.TenantIDContextKey)
	if contextTenantID == nil || contextTenantID == "" {
		vars := mux.Vars(r)
		tenantID = vars["tenant_id"]
		if tenantID == "" {
			// attempt to get tenant id from header
			tenantID = r.Header.Get("X-Tenant-ID")
		}
	} else {
		// Safely convert to string, fallback to URL if conversion fails
		if strTenantID, ok := contextTenantID.(string); ok {
			tenantID = strTenantID
		} else {
			// If context value is not a string, fallback to URL parameter
			vars := mux.Vars(r)
			tenantID = vars["tenant_id"]
		}
	}

	if tenantID == "" {
		return "", fmt.Errorf("tenant_id is required")
	}

	return tenantID, nil
}

func ParseAndValidateBody[T any](r *http.Request) (T, *diagnostics.Diagnostics) {
	diag := diagnostics.New("validate_request_body")
	defer diag.Complete()

	var obj T
	err := json.NewDecoder(r.Body).Decode(&obj)
	if err != nil {
		diag.AddError("invalid_request_body", "invalid request body", "request_body", map[string]interface{}{
			"error": err.Error(),
		})
		return obj, diag
	}
	if errors := validation.Validate(obj); errors != nil {
		for _, err := range errors {
			diag.AddError("invalid_request_body", err.Message, "request_body")
		}
		return obj, diag
	}

	return obj, nil
}
