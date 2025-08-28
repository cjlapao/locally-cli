package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/cjlapao/locally-cli/internal/api/models"
)

func TestWriteError(t *testing.T) {
	tests := []struct {
		name           string
		statusCode     int
		errorCode      string
		message        string
		details        []string
		expectedStatus int
		expectedError  models.APIError
	}{
		{
			name:           "Basic error",
			statusCode:     http.StatusBadRequest,
			errorCode:      ErrorCodeBadRequest,
			message:        "Test error",
			expectedStatus: http.StatusBadRequest,
			expectedError: models.APIError{
				Error: models.ErrorDetails{
					Code:    ErrorCodeBadRequest,
					Message: "Test error",
				},
			},
		},
		{
			name:           "Error with details",
			statusCode:     http.StatusNotFound,
			errorCode:      ErrorCodeNotFound,
			message:        "Not found",
			details:        []string{"Resource ID: 123"},
			expectedStatus: http.StatusNotFound,
			expectedError: models.APIError{
				Error: models.ErrorDetails{
					Code:    ErrorCodeNotFound,
					Message: "Not found",
					Details: "Resource ID: 123",
				},
			},
		},
		{
			name:           "Default error code",
			statusCode:     http.StatusInternalServerError,
			errorCode:      "",
			message:        "Server error",
			expectedStatus: http.StatusInternalServerError,
			expectedError: models.APIError{
				Error: models.ErrorDetails{
					Code:    ErrorCodeInternalError,
					Message: "Server error",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Call the function
			WriteError(w, req, tt.statusCode, tt.errorCode, tt.message, tt.details...)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// Parse response
			var response models.APIError
			err := json.NewDecoder(w.Body).Decode(&response)
			if err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			// Check error code
			if response.Error.Code != tt.expectedError.Error.Code {
				t.Errorf("expected error code %s, got %s", tt.expectedError.Error.Code, response.Error.Code)
			}

			// Check message
			if response.Error.Message != tt.expectedError.Error.Message {
				t.Errorf("expected message %s, got %s", tt.expectedError.Error.Message, response.Error.Message)
			}

			// Check details if provided
			if len(tt.details) > 0 && response.Error.Details != tt.expectedError.Error.Details {
				t.Errorf("expected details %s, got %s", tt.expectedError.Error.Details, response.Error.Details)
			}

			// Check timestamp format
			_, err = time.Parse(time.RFC3339, response.Timestamp)
			if err != nil {
				t.Errorf("invalid timestamp format: %v", err)
			}

			// Check path
			if response.Path != "/test" {
				t.Errorf("expected path /test, got %s", response.Path)
			}
		})
	}
}

func TestHelperErrorFunctions(t *testing.T) {
	tests := []struct {
		name           string
		handler        func(w http.ResponseWriter, r *http.Request, message string, details ...string)
		message        string
		details        []string
		expectedStatus int
		expectedCode   string
	}{
		{
			name:           "WriteUnauthorized",
			handler:        WriteUnauthorized,
			message:        "Custom unauthorized message",
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   ErrorCodeUnauthorized,
		},
		{
			name:           "WriteForbidden",
			handler:        WriteForbidden,
			message:        "Custom forbidden message",
			expectedStatus: http.StatusForbidden,
			expectedCode:   ErrorCodeForbidden,
		},
		{
			name:           "WriteBadRequest",
			handler:        WriteBadRequest,
			message:        "Custom bad request message",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeBadRequest,
		},
		{
			name:           "WriteNotFound",
			handler:        WriteNotFound,
			message:        "Custom not found message",
			expectedStatus: http.StatusNotFound,
			expectedCode:   ErrorCodeNotFound,
		},
		{
			name:           "WriteConflict",
			handler:        WriteConflict,
			message:        "Custom conflict message",
			expectedStatus: http.StatusConflict,
			expectedCode:   ErrorCodeConflict,
		},
		{
			name:           "WriteInternalError",
			handler:        WriteInternalError,
			message:        "Custom internal error message",
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   ErrorCodeInternalError,
		},
		{
			name:           "WriteMethodNotAllowed",
			handler:        WriteMethodNotAllowed,
			message:        "Custom method not allowed message",
			expectedStatus: http.StatusMethodNotAllowed,
			expectedCode:   ErrorCodeMethodNotAllowed,
		},
		{
			name:           "WriteInvalidToken",
			handler:        WriteInvalidToken,
			message:        "Custom invalid token message",
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   ErrorCodeInvalidToken,
		},
		{
			name:           "WriteValidationError",
			handler:        WriteValidationError,
			message:        "Custom validation error message",
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeInvalidInput,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Call the handler
			tt.handler(w, req, tt.message, tt.details...)

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// Parse response
			var response models.APIError
			err := json.NewDecoder(w.Body).Decode(&response)
			if err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			// Check error code
			if response.Error.Code != tt.expectedCode {
				t.Errorf("expected error code %s, got %s", tt.expectedCode, response.Error.Code)
			}

			// Check message
			if response.Error.Message != tt.message {
				t.Errorf("expected message %s, got %s", tt.message, response.Error.Message)
			}
		})
	}
}

func TestDefaultMessages(t *testing.T) {
	tests := []struct {
		name           string
		handler        func(w http.ResponseWriter, r *http.Request, message string, details ...string)
		expectedStatus int
		expectedCode   string
		expectedMsg    string
	}{
		{
			name:           "WriteUnauthorized default",
			handler:        WriteUnauthorized,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   ErrorCodeUnauthorized,
			expectedMsg:    "Authentication required",
		},
		{
			name:           "WriteForbidden default",
			handler:        WriteForbidden,
			expectedStatus: http.StatusForbidden,
			expectedCode:   ErrorCodeForbidden,
			expectedMsg:    "Access denied",
		},
		{
			name:           "WriteBadRequest default",
			handler:        WriteBadRequest,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeBadRequest,
			expectedMsg:    "Invalid request",
		},
		{
			name:           "WriteNotFound default",
			handler:        WriteNotFound,
			expectedStatus: http.StatusNotFound,
			expectedCode:   ErrorCodeNotFound,
			expectedMsg:    "Resource not found",
		},
		{
			name:           "WriteConflict default",
			handler:        WriteConflict,
			expectedStatus: http.StatusConflict,
			expectedCode:   ErrorCodeConflict,
			expectedMsg:    "Resource conflict",
		},
		{
			name:           "WriteInternalError default",
			handler:        WriteInternalError,
			expectedStatus: http.StatusInternalServerError,
			expectedCode:   ErrorCodeInternalError,
			expectedMsg:    "Internal server error",
		},
		{
			name:           "WriteMethodNotAllowed default",
			handler:        WriteMethodNotAllowed,
			expectedStatus: http.StatusMethodNotAllowed,
			expectedCode:   ErrorCodeMethodNotAllowed,
			expectedMsg:    "Method not allowed",
		},
		{
			name:           "WriteInvalidToken default",
			handler:        WriteInvalidToken,
			expectedStatus: http.StatusUnauthorized,
			expectedCode:   ErrorCodeInvalidToken,
			expectedMsg:    "Invalid or expired token",
		},
		{
			name:           "WriteValidationError default",
			handler:        WriteValidationError,
			expectedStatus: http.StatusBadRequest,
			expectedCode:   ErrorCodeInvalidInput,
			expectedMsg:    "Validation failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test request
			req := httptest.NewRequest("GET", "/test", nil)
			w := httptest.NewRecorder()

			// Call the handler with empty message
			tt.handler(w, req, "")

			// Check status code
			if w.Code != tt.expectedStatus {
				t.Errorf("expected status code %d, got %d", tt.expectedStatus, w.Code)
			}

			// Parse response
			var response models.APIError
			err := json.NewDecoder(w.Body).Decode(&response)
			if err != nil {
				t.Fatalf("failed to decode response: %v", err)
			}

			// Check error code
			if response.Error.Code != tt.expectedCode {
				t.Errorf("expected error code %s, got %s", tt.expectedCode, response.Error.Code)
			}

			// Check default message
			if response.Error.Message != tt.expectedMsg {
				t.Errorf("expected message %s, got %s", tt.expectedMsg, response.Error.Message)
			}
		})
	}
}
