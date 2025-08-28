package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cjlapao/locally-cli/internal/api"
	"github.com/cjlapao/locally-cli/internal/api/models"
)

// TestResponse represents a generic API response for testing
type TestResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

func TestNewHandler(t *testing.T) {
	handler := NewHandler()
	if handler == nil {
		t.Error("NewHandler returned nil")
	}
}

// Example middleware for testing
func testMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Test-Middleware", "test-value")
		next.ServeHTTP(w, r)
	})
}

// Example handler method for testing
func (h *Handler) testHandler(w http.ResponseWriter, r *http.Request) {
	response := TestResponse{
		Message: "Test successful",
		Status:  "ok",
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func TestHandlerWithMiddleware(t *testing.T) {
	handler := NewHandler()

	// Create a test server with middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/test", handler.testHandler)

	server := httptest.NewServer(testMiddleware(mux))
	defer server.Close()

	// Test the endpoint
	resp, err := http.Get(server.URL + "/test")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status code %d, got %d", http.StatusOK, resp.StatusCode)
	}

	// Check middleware header
	if header := resp.Header.Get("X-Test-Middleware"); header != "test-value" {
		t.Errorf("Expected X-Test-Middleware header to be 'test-value', got '%s'", header)
	}

	// Check content type
	if contentType := resp.Header.Get("Content-Type"); contentType != "application/json" {
		t.Errorf("Expected Content-Type to be 'application/json', got '%s'", contentType)
	}

	// Parse response
	var response TestResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	// Check response content
	if response.Message != "Test successful" {
		t.Errorf("Expected message 'Test successful', got '%s'", response.Message)
	}
	if response.Status != "ok" {
		t.Errorf("Expected status 'ok', got '%s'", response.Status)
	}
}

func TestHandlerErrorHandling(t *testing.T) {
	// Create a test server
	mux := http.NewServeMux()
	mux.HandleFunc("/error", func(w http.ResponseWriter, r *http.Request) {
		api.WriteInternalError(w, r, "Test error message")
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Test the error endpoint
	resp, err := http.Get(server.URL + "/error")
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("Expected status code %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}

	// Parse error response
	var errorResponse models.APIError
	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Check error response content
	if errorResponse.Error.Code != api.ErrorCodeInternalError {
		t.Errorf("Expected error code '%s', got '%s'", api.ErrorCodeInternalError, errorResponse.Error.Code)
	}
	if errorResponse.Error.Message != "Test error message" {
		t.Errorf("Expected error message 'Test error message', got '%s'", errorResponse.Error.Message)
	}
}

func TestHandlerMethodNotAllowed(t *testing.T) {
	handler := NewHandler()

	// Create a test server that only allows GET
	mux := http.NewServeMux()
	mux.HandleFunc("/test", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			api.WriteMethodNotAllowed(w, r, "Only GET method is allowed")
			return
		}
		handler.testHandler(w, r)
	})

	server := httptest.NewServer(mux)
	defer server.Close()

	// Test POST request (should fail)
	resp, err := http.Post(server.URL+"/test", "application/json", nil)
	if err != nil {
		t.Fatalf("Failed to make request: %v", err)
	}
	defer resp.Body.Close()

	// Check status code
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("Expected status code %d, got %d", http.StatusMethodNotAllowed, resp.StatusCode)
	}

	// Parse error response
	var errorResponse models.APIError
	if err := json.NewDecoder(resp.Body).Decode(&errorResponse); err != nil {
		t.Fatalf("Failed to decode error response: %v", err)
	}

	// Check error response content
	if errorResponse.Error.Code != api.ErrorCodeMethodNotAllowed {
		t.Errorf("Expected error code '%s', got '%s'", api.ErrorCodeMethodNotAllowed, errorResponse.Error.Code)
	}
	if errorResponse.Error.Message != "Only GET method is allowed" {
		t.Errorf("Expected error message 'Only GET method is allowed', got '%s'", errorResponse.Error.Message)
	}
}
