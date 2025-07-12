package api

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// responseWriter wraps http.ResponseWriter to capture the status code and response size
type responseWriter struct {
	http.ResponseWriter
	statusCode   int
	responseSize int64
	responseBody *bytes.Buffer
	wroteHeader  bool
	captureBody  bool // New field to control body capture
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK, // Default status code
		responseBody:   &bytes.Buffer{},
		captureBody:    true, // Capture body for normal requests
	}
}

// newResponseWriterNoCapture creates a response writer that doesn't capture body (for SSE)
func newResponseWriterNoCapture(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
		responseBody:   &bytes.Buffer{},
		captureBody:    false, // Don't capture body for SSE
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.statusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.responseSize += int64(size)

	// Only capture body if enabled (not for SSE)
	if rw.captureBody {
		rw.responseBody.Write(b)
	}

	return size, err
}

// Flush implements http.Flusher for SSE support
func (rw *responseWriter) Flush() {
	if flusher, ok := rw.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

// CORSConfig represents CORS configuration
type CORSConfig struct {
	AllowOrigins     []string
	AllowMethods     []string
	AllowHeaders     []string
	ExposeHeaders    []string
	AllowCredentials bool
	MaxAge           int
}

// DefaultCORSConfig returns a default CORS configuration
func DefaultCORSConfig() CORSConfig {
	return CORSConfig{
		AllowOrigins: []string{"http://localhost:3000", "http://127.0.0.1:3000"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodPost,
			http.MethodPut,
			http.MethodDelete,
			http.MethodPatch,
			http.MethodOptions,
			http.MethodHead,
		},
		AllowHeaders: []string{
			"Accept",
			"Accept-Language",
			"Content-Type",
			"Content-Language",
			"Origin",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"X-HTTP-Method-Override",
			"Cache-Control",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Request-ID",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware creates a CORS middleware with the given configuration
func CORSMiddleware(config CORSConfig) Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			origin := r.Header.Get("Origin")

			// Check if origin is allowed
			allowOrigin := ""
			if len(config.AllowOrigins) == 0 || contains(config.AllowOrigins, "*") {
				allowOrigin = "*"
			} else if contains(config.AllowOrigins, origin) {
				allowOrigin = origin
			}

			// Set CORS headers
			if allowOrigin != "" {
				w.Header().Set("Access-Control-Allow-Origin", allowOrigin)
			}

			if config.AllowCredentials {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			if len(config.AllowMethods) > 0 {
				w.Header().Set("Access-Control-Allow-Methods", strings.Join(config.AllowMethods, ", "))
			}

			if len(config.AllowHeaders) > 0 {
				w.Header().Set("Access-Control-Allow-Headers", strings.Join(config.AllowHeaders, ", "))
			}

			if len(config.ExposeHeaders) > 0 {
				w.Header().Set("Access-Control-Expose-Headers", strings.Join(config.ExposeHeaders, ", "))
			}

			if config.MaxAge > 0 {
				w.Header().Set("Access-Control-Max-Age", fmt.Sprintf("%d", config.MaxAge))
			}

			// Handle preflight OPTIONS request
			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusOK)
				return
			}

			// Continue with next handler
			next(w, r)
		}
	}
}

func RequestIDMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get("X-Request-ID")
		if requestID == "" {
			// trying to get it from the current context
			contextRequestID := r.Context().Value(types.RequestIDKey)
			if contextRequestID != nil {
				requestID = contextRequestID.(string)
			}
			if requestID == "" {
				requestID = uuid.New().String()
			}
		}

		// copy the current context with the new request id
		ctx := context.WithValue(r.Context(), types.RequestIDKey, requestID)
		next(w, r.WithContext(ctx))
	}
}

// contains checks if a slice contains a string
func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// LoggingMiddleware logs HTTP requests and responses
func LoggingMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		isSSE := r.Header.Get("Accept") == "text/event-stream" ||
			strings.HasSuffix(r.URL.Path, "/events/stream")
		var rw *responseWriter
		if isSSE {
			rw = newResponseWriterNoCapture(w)
		} else {
			rw = newResponseWriter(w)
		}

		if !isSSE {
			logRequest(r)
		} else {
			logSSERequest(r)
		}

		panicOccurred := false
		defer func() {
			if err := recover(); err != nil {
				panicOccurred = true
				rw.WriteHeader(http.StatusInternalServerError)
				logging.WithError(fmt.Errorf("%v", err)).Error("Panic recovered in logging middleware")
				if !isSSE {
					logResponse(r, rw, time.Since(start))
				}
				panic(err)
			}
		}()

		if next == nil {
			notFound := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				http.NotFound(w, r)
			})
			notFound.ServeHTTP(rw, r)
		} else {
			next.ServeHTTP(rw, r)
		}

		duration := time.Since(start)
		if !isSSE && !panicOccurred {
			logResponse(r, rw, duration)
		} else if isSSE {
			logSSEResponse(r, rw, duration)
		}
	}
}

func logRequest(r *http.Request) {
	cfg := config.GetInstance().Get()

	fields := logrus.Fields{
		"timestamp": time.Now().Format(time.RFC3339),
		"method":    r.Method,
		"path":      r.URL.Path,
		"remote":    r.RemoteAddr,
		"headers":   formatHeaders(r.Header),
	}

	// Log request body for POST/PUT/PATCH methods
	var requestBody string
	if isBodyMethod(r.Method) {
		body, err := io.ReadAll(r.Body)
		if err == nil {
			r.Body = io.NopCloser(bytes.NewBuffer(body))

			if cfg.Get(config.DebugKey).GetBool() {
				// Try to parse as JSON for pretty printing
				var prettyJSON bytes.Buffer
				if json.Indent(&prettyJSON, body, "", "") == nil {
					requestBody = prettyJSON.String()
				} else {
					requestBody = string(body)
				}
			}
		}
	}
	if !cfg.IsDebug() {
		requestBody = "[REDACTED]"
	}

	fields["request_body"] = requestBody

	logging.WithFields(fields).Info("Api Request")
}

func logResponse(r *http.Request, rw *responseWriter, duration time.Duration) {
	cfg := config.GetInstance().Get()

	// Try to parse response body as JSON for pretty printing
	var responseBody string
	if rw.responseBody.Len() > 0 && cfg.Get(config.DebugKey).GetBool() {
		var prettyJSON bytes.Buffer
		body := rw.responseBody.Bytes()
		if json.Indent(&prettyJSON, body, "", "  ") == nil {
			responseBody = prettyJSON.String()
		} else {
			responseBody = string(body)
		}
	}
	if !cfg.IsDebug() {
		responseBody = "[REDACTED]"
	}

	fields := logrus.Fields{
		"timestamp":     time.Now().Format(time.RFC3339),
		"method":        r.Method,
		"path":          r.URL.Path,
		"status_code":   rw.statusCode,
		"response_size": rw.responseSize,
		"duration":      utils.FormatDuration(duration),
		"response_body": responseBody,
	}

	logging.WithFields(fields).Info("Api Response")
}

func formatHeaders(headers http.Header) map[string]string {
	result := make(map[string]string)
	for key, values := range headers {
		// Skip sensitive headers
		if isSensitiveHeader(key) {
			result[key] = "[REDACTED]"
		} else {
			result[key] = strings.Join(values, ", ")
		}
	}
	return result
}

func isSensitiveHeader(header string) bool {
	sensitiveHeaders := map[string]bool{
		"Authorization":  true,
		"Cookie":         true,
		"Set-Cookie":     true,
		"X-Api-Key":      true,
		"X-Access-Token": true,
		"Password":       true,
	}
	return sensitiveHeaders[header]
}

func isBodyMethod(method string) bool {
	return method == http.MethodPost ||
		method == http.MethodPut ||
		method == http.MethodPatch ||
		method == http.MethodDelete
}

// logSSERequest logs SSE connection requests without body capture
func logSSERequest(r *http.Request) {
	reqDetails := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"method":    r.Method,
		"path":      r.URL.Path,
		"remote":    r.RemoteAddr,
		"type":      "SSE_REQUEST",
	}

	if details, err := json.Marshal(reqDetails); err == nil {
		logging.WithField("sse_request", string(details)).Info("SSE Request")
	}
}

// logSSEResponse logs SSE connection responses without body capture
func logSSEResponse(r *http.Request, rw *responseWriter, duration time.Duration) {
	respDetails := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"method":      r.Method,
		"path":        r.URL.Path,
		"status_code": rw.statusCode,
		"duration_ms": duration.Milliseconds(),
		"type":        "SSE_RESPONSE",
		"note":        "SSE stream established",
	}

	if details, err := json.Marshal(respDetails); err == nil {
		logging.WithField("sse_response", string(details)).Info("SSE Response")
	}
}
