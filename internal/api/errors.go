package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/cjlapao/locally-cli/internal/api/models"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
)

// Predefined error codes
const (
	// Authentication errors
	ErrorCodeUnauthorized = "UNAUTHORIZED"
	ErrorCodeForbidden    = "FORBIDDEN"
	ErrorCodeInvalidToken = "INVALID_TOKEN"
	ErrorCodeTokenExpired = "TOKEN_EXPIRED"

	// Validation errors
	ErrorCodeBadRequest   = "BAD_REQUEST"
	ErrorCodeInvalidInput = "INVALID_INPUT"
	ErrorCodeMissingField = "MISSING_FIELD"

	// Resource errors
	ErrorCodeNotFound      = "NOT_FOUND"
	ErrorCodeConflict      = "CONFLICT"
	ErrorCodeAlreadyExists = "ALREADY_EXISTS"

	// Server errors
	ErrorCodeInternalError      = "INTERNAL_ERROR"
	ErrorCodeServiceUnavailable = "SERVICE_UNAVAILABLE"
	ErrorCodeDatabaseError      = "DATABASE_ERROR"

	// Method errors
	ErrorCodeMethodNotAllowed = "METHOD_NOT_ALLOWED"
)

// Error code to HTTP status mapping
var errorCodeToStatus = map[string]int{
	ErrorCodeUnauthorized:       http.StatusUnauthorized,
	ErrorCodeForbidden:          http.StatusForbidden,
	ErrorCodeInvalidToken:       http.StatusUnauthorized,
	ErrorCodeTokenExpired:       http.StatusUnauthorized,
	ErrorCodeBadRequest:         http.StatusBadRequest,
	ErrorCodeInvalidInput:       http.StatusBadRequest,
	ErrorCodeMissingField:       http.StatusBadRequest,
	ErrorCodeNotFound:           http.StatusNotFound,
	ErrorCodeConflict:           http.StatusConflict,
	ErrorCodeAlreadyExists:      http.StatusConflict,
	ErrorCodeInternalError:      http.StatusInternalServerError,
	ErrorCodeServiceUnavailable: http.StatusServiceUnavailable,
	ErrorCodeDatabaseError:      http.StatusInternalServerError,
	ErrorCodeMethodNotAllowed:   http.StatusMethodNotAllowed,
}

// WriteError writes a standardized error response
func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, errorCode, message string, details ...string) {
	// Use default error code if not provided
	if errorCode == "" {
		errorCode = getDefaultErrorCode(statusCode)
	}

	// Use default status code if error code is provided but status is 0
	if statusCode == 0 && errorCode != "" {
		if defaultStatus, exists := errorCodeToStatus[errorCode]; exists {
			statusCode = defaultStatus
		} else {
			statusCode = http.StatusInternalServerError
		}
	}

	// Default status code if nothing is provided
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	errorDetails := models.ErrorDetails{
		Code:    errorCode,
		Message: message,
	}

	// Add details if provided
	if len(details) > 0 {
		errorDetails.Details = details[0]
	}

	apiError := models.APIError{
		Error:     errorDetails,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(apiError)
	if err != nil {
		logging.WithError(err).Error("Failed to encode error")
	}
}

// WriteErrorWithDiagnostics writes a standardized error response
func WriteErrorWithDiagnostics(w http.ResponseWriter, r *http.Request, statusCode int, errorCode, message string, diag *diagnostics.Diagnostics) {
	// Use default error code if not provided
	cfg := config.GetInstance().Get()
	if errorCode == "" {
		errorCode = getDefaultErrorCode(statusCode)
	}

	// Use default status code if error code is provided but status is 0
	if statusCode == 0 && errorCode != "" {
		if defaultStatus, exists := errorCodeToStatus[errorCode]; exists {
			statusCode = defaultStatus
		} else {
			statusCode = http.StatusInternalServerError
		}
	}

	// Default status code if nothing is provided
	if statusCode == 0 {
		statusCode = http.StatusInternalServerError
	}

	errorDetails := models.ErrorDetails{
		Code:    errorCode,
		Message: message,
	}

	// Add details if provided
	if diag != nil {
		if cfg.IsDebug() {
			errorDetails.Diagnostics = diag
		} else {
			errorDetails.Errors = make([]models.ErrorDetailsError, 0)
			for _, err := range diag.Errors {
				errorDetails.Errors = append(errorDetails.Errors, models.ErrorDetailsError{
					Code:    err.Code,
					Message: err.Message,
				})
			}
		}
	}

	apiError := models.APIError{
		Error:     errorDetails,
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Path:      r.URL.Path,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(apiError)
	if err != nil {
		logging.WithError(err).Error("Failed to encode error")
	}
}

// Helper functions for common errors

// WriteUnauthorized writes an unauthorized error
func WriteUnauthorized(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	if message == "" {
		message = "Authentication required"
	}
	WriteError(w, r, http.StatusUnauthorized, ErrorCodeUnauthorized, message, details...)
}

// WriteForbidden writes a forbidden error
func WriteForbidden(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	if message == "" {
		message = "Access denied"
	}
	WriteError(w, r, http.StatusForbidden, ErrorCodeForbidden, message, details...)
}

// WriteBadRequest writes a bad request error
func WriteBadRequest(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	if message == "" {
		message = "Invalid request"
	}
	WriteError(w, r, http.StatusBadRequest, ErrorCodeBadRequest, message, details...)
}

// WriteNotFound writes a not found error
func WriteNotFound(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	if message == "" {
		message = "Resource not found"
	}
	WriteError(w, r, http.StatusNotFound, ErrorCodeNotFound, message, details...)
}

// WriteConflict writes a conflict error
func WriteConflict(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	if message == "" {
		message = "Resource conflict"
	}
	WriteError(w, r, http.StatusConflict, ErrorCodeConflict, message, details...)
}

// WriteInternalError writes an internal server error
func WriteInternalError(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	if message == "" {
		message = "Internal server error"
	}
	WriteError(w, r, http.StatusInternalServerError, ErrorCodeInternalError, message, details...)
}

// WriteMethodNotAllowed writes a method not allowed error
func WriteMethodNotAllowed(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	if message == "" {
		message = "Method not allowed"
	}
	WriteError(w, r, http.StatusMethodNotAllowed, ErrorCodeMethodNotAllowed, message, details...)
}

// WriteInvalidToken writes an invalid token error
func WriteInvalidToken(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	if message == "" {
		message = "Invalid or expired token"
	}
	WriteError(w, r, http.StatusUnauthorized, ErrorCodeInvalidToken, message, details...)
}

// WriteValidationError writes a validation error
func WriteValidationError(w http.ResponseWriter, r *http.Request, message string, details ...string) {
	if message == "" {
		message = "Validation failed"
	}
	WriteError(w, r, http.StatusBadRequest, ErrorCodeInvalidInput, message, details...)
}

// getDefaultErrorCode returns a default error code based on HTTP status
func getDefaultErrorCode(statusCode int) string {
	switch statusCode {
	case http.StatusBadRequest:
		return ErrorCodeBadRequest
	case http.StatusUnauthorized:
		return ErrorCodeUnauthorized
	case http.StatusForbidden:
		return ErrorCodeForbidden
	case http.StatusNotFound:
		return ErrorCodeNotFound
	case http.StatusMethodNotAllowed:
		return ErrorCodeMethodNotAllowed
	case http.StatusConflict:
		return ErrorCodeConflict
	case http.StatusInternalServerError:
		return ErrorCodeInternalError
	case http.StatusServiceUnavailable:
		return ErrorCodeServiceUnavailable
	default:
		return ErrorCodeInternalError
	}
}
