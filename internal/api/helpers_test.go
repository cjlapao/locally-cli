package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/gorilla/mux"
)

func TestGetTenantIDFromRequest(t *testing.T) {
	tests := []struct {
		name          string
		contextValue  interface{}
		urlPattern    string
		urlPath       string
		expectedID    string
		expectedError bool
		errorMessage  string
	}{
		{
			name:          "Tenant ID from context",
			contextValue:  "tenant-123",
			urlPattern:    "/api/v1/tenants/{tenant_id}",
			urlPath:       "/api/v1/tenants/tenant-456",
			expectedID:    "tenant-123",
			expectedError: false,
		},
		{
			name:          "Tenant ID from URL",
			contextValue:  nil,
			urlPattern:    "/api/v1/tenants/{tenant_id}",
			urlPath:       "/api/v1/tenants/tenant-456",
			expectedID:    "tenant-456",
			expectedError: false,
		},
		{
			name:          "Empty tenant ID in context",
			contextValue:  "",
			urlPattern:    "/api/v1/tenants/{tenant_id}",
			urlPath:       "/api/v1/tenants/tenant-456",
			expectedID:    "tenant-456",
			expectedError: false,
		},
		{
			name:          "No tenant ID in context or URL",
			contextValue:  nil,
			urlPattern:    "/api/v1/tenants/{tenant_id}",
			urlPath:       "/api/v1/tenants/",
			expectedID:    "",
			expectedError: true,
			errorMessage:  "tenant_id is required",
		},
		{
			name:          "Invalid context value type",
			contextValue:  123, // non-string value
			urlPattern:    "/api/v1/tenants/{tenant_id}",
			urlPath:       "/api/v1/tenants/tenant-456",
			expectedID:    "tenant-456",
			expectedError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new router
			router := mux.NewRouter()
			router.HandleFunc(tt.urlPattern, func(w http.ResponseWriter, r *http.Request) {
				// Test the function
				tenantID, err := GetTenantIDFromRequest(r)

				// Check error
				if tt.expectedError {
					if err == nil {
						t.Error("Expected error but got none")
					} else if err.Error() != tt.errorMessage {
						t.Errorf("Expected error message '%s', got '%s'", tt.errorMessage, err.Error())
					}
					return
				}

				// Check tenant ID
				if tenantID != tt.expectedID {
					t.Errorf("Expected tenant ID '%s', got '%s'", tt.expectedID, tenantID)
				}
			})

			// Create a test request
			req := httptest.NewRequest("GET", tt.urlPath, nil)

			// Add context value if provided
			if tt.contextValue != nil {
				ctx := context.WithValue(req.Context(), config.TenantIDContextKey, tt.contextValue)
				req = req.WithContext(ctx)
			}

			// Create a response recorder
			rr := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(rr, req)
		})
	}
}

func TestGetTenantIDFromRequestWithInvalidURL(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/tenants/{tenant_id}", func(w http.ResponseWriter, r *http.Request) {
		// Test the function
		tenantID, err := GetTenantIDFromRequest(r)

		// Should return error
		if err == nil {
			t.Error("Expected error but got none")
		}

		// Should return empty tenant ID
		if tenantID != "" {
			t.Errorf("Expected empty tenant ID, got '%s'", tenantID)
		}
	})

	// Create a test request with invalid URL (no tenant_id parameter)
	req := httptest.NewRequest("GET", "/api/v1/tenants", nil)
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)
}

func TestGetTenantIDFromRequestWithEmptyContext(t *testing.T) {
	// Create a new router
	router := mux.NewRouter()
	router.HandleFunc("/api/v1/tenants/{tenant_id}", func(w http.ResponseWriter, r *http.Request) {
		// Test the function
		tenantID, err := GetTenantIDFromRequest(r)

		// Should return error
		if err == nil {
			t.Error("Expected error but got none")
		}

		// Should return empty tenant ID
		if tenantID != "" {
			t.Errorf("Expected empty tenant ID, got '%s'", tenantID)
		}
	})

	// Create a test request with empty context
	req := httptest.NewRequest("GET", "/api/v1/tenants", nil)
	ctx := context.WithValue(req.Context(), config.TenantIDContextKey, "")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(rr, req)
}
