package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/cjlapao/locally-cli/pkg/utils"
	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// MiddlewareResult represents the result of middleware execution
type MiddlewareResult struct {
	Continue bool
	Error    error
}

// PreMiddleware represents a middleware that runs before the route handler
type PreMiddleware interface {
	Execute(w http.ResponseWriter, r *http.Request) MiddlewareResult
}

// PostMiddleware represents a middleware that runs after the route handler
type PostMiddleware interface {
	Execute(w http.ResponseWriter, r *http.Request, responseData *ResponseData) error
}

// PreMiddlewareFunc is a function-based pre-middleware
type PreMiddlewareFunc func(w http.ResponseWriter, r *http.Request) MiddlewareResult

func (f PreMiddlewareFunc) Execute(w http.ResponseWriter, r *http.Request) MiddlewareResult {
	return f(w, r)
}

// PostMiddlewareFunc is a function-based post-middleware
type PostMiddlewareFunc func(w http.ResponseWriter, r *http.Request, responseData *ResponseData) error

func (f PostMiddlewareFunc) Execute(w http.ResponseWriter, r *http.Request, responseData *ResponseData) error {
	return f(w, r, responseData)
}

// ResponseData captures response information for post-middlewares
type ResponseData struct {
	StatusCode   int
	ResponseSize int64
	ResponseBody *bytes.Buffer
	Duration     time.Duration
	PanicError   interface{}
}

// MiddlewareChain represents a chain of middlewares
type MiddlewareChain struct {
	preMiddlewares  []PreMiddleware
	postMiddlewares []PostMiddleware
}

// NewMiddlewareChain creates a new middleware chain
func NewMiddlewareChain() *MiddlewareChain {
	return &MiddlewareChain{
		preMiddlewares:  make([]PreMiddleware, 0),
		postMiddlewares: make([]PostMiddleware, 0),
	}
}

// AddPreMiddleware adds a pre-middleware to the chain
func (mc *MiddlewareChain) AddPreMiddleware(middleware PreMiddleware) {
	mc.preMiddlewares = append(mc.preMiddlewares, middleware)
}

// AddPostMiddleware adds a post-middleware to the chain
func (mc *MiddlewareChain) AddPostMiddleware(middleware PostMiddleware) {
	mc.postMiddlewares = append(mc.postMiddlewares, middleware)
}

// AddPreMiddlewareFunc adds a function-based pre-middleware
func (mc *MiddlewareChain) AddPreMiddlewareFunc(fn func(w http.ResponseWriter, r *http.Request) MiddlewareResult) {
	mc.AddPreMiddleware(PreMiddlewareFunc(fn))
}

// AddPostMiddlewareFunc adds a function-based post-middleware
func (mc *MiddlewareChain) AddPostMiddlewareFunc(fn func(w http.ResponseWriter, r *http.Request, responseData *ResponseData) error) {
	mc.AddPostMiddleware(PostMiddlewareFunc(fn))
}

// Execute executes the middleware chain with the given handler
func (mc *MiddlewareChain) Execute(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		// Execute pre-middlewares
		for _, middleware := range mc.preMiddlewares {
			result := middleware.Execute(w, r)
			if !result.Continue {
				// Middleware decided to short-circuit
				if result.Error != nil {
					// Only log if logging is initialized
					if logging.Logger != nil {
						logging.WithError(result.Error).Error("Pre-middleware short-circuited with error")
					}
				}
				return
			}
		}

		// Create response capture wrapper
		responseData := &ResponseData{
			ResponseBody: &bytes.Buffer{},
		}
		responseWriter := &responseCaptureWriter{
			ResponseWriter: w,
			responseData:   responseData,
		}

		// Execute the route handler with panic recovery
		func() {
			defer func() {
				if err := recover(); err != nil {
					responseData.PanicError = err
					responseWriter.WriteHeader(http.StatusInternalServerError)
					// Only log if logging is initialized
					if logging.Logger != nil {
						logging.WithError(fmt.Errorf("%v", err)).Error("Panic recovered in route handler")
					}
				}
			}()
			handler(responseWriter, r)
		}()

		// Calculate duration
		responseData.Duration = time.Since(start)

		// Execute post-middlewares
		for _, middleware := range mc.postMiddlewares {
			if err := middleware.Execute(w, r, responseData); err != nil {
				// Only log if logging is initialized
				if logging.Logger != nil {
					logging.WithError(err).Error("Post-middleware failed")
				}
				// Continue with other post-middlewares even if one fails
			}
		}
	}
}

// responseCaptureWriter wraps http.ResponseWriter to capture response data
type responseCaptureWriter struct {
	http.ResponseWriter
	responseData *ResponseData
	wroteHeader  bool
}

func (rw *responseCaptureWriter) WriteHeader(code int) {
	if !rw.wroteHeader {
		rw.responseData.StatusCode = code
		rw.wroteHeader = true
		rw.ResponseWriter.WriteHeader(code)
	}
}

func (rw *responseCaptureWriter) Write(b []byte) (int, error) {
	if !rw.wroteHeader {
		rw.WriteHeader(http.StatusOK)
	}
	size, err := rw.ResponseWriter.Write(b)
	rw.responseData.ResponseSize += int64(size)
	rw.responseData.ResponseBody.Write(b)
	return size, err
}

// Flush implements http.Flusher for SSE support
func (rw *responseCaptureWriter) Flush() {
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
			"Accept-Encoding",
			"Content-Type",
			"Content-Language",
			"Origin",
			"Authorization",
			"X-Requested-With",
			"X-Request-ID",
			"X-HTTP-Method-Override",
			"Cache-Control",
			"Pragma",
			"Expires",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"Content-Type",
			"X-Request-ID",
			"X-Total-Count",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
	}
}

// CORSMiddleware creates a CORS pre-middleware
func CORSMiddleware(config CORSConfig) PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		origin := r.Header.Get("Origin")

		// Determine if origin is allowed
		allowOrigin := ""
		if len(config.AllowOrigins) == 0 || contains(config.AllowOrigins, "*") {
			allowOrigin = "*"
		} else if origin != "" && contains(config.AllowOrigins, origin) {
			allowOrigin = origin
		}

		// Handle preflight OPTIONS request
		if r.Method == http.MethodOptions {
			return handlePreflightRequest(w, r, config, allowOrigin)
		}

		// Handle actual requests (non-preflight)
		return handleActualRequest(w, r, config, allowOrigin)
	})
}

// handlePreflightRequest handles CORS preflight OPTIONS requests
func handlePreflightRequest(w http.ResponseWriter, r *http.Request, config CORSConfig, allowOrigin string) MiddlewareResult {
	// Validate requested method
	requestMethod := r.Header.Get("Access-Control-Request-Method")
	if requestMethod != "" && !contains(config.AllowMethods, requestMethod) {
		logging.Warnf("CORS preflight rejected: method %s not allowed", requestMethod)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return MiddlewareResult{Continue: false}
	}

	// Validate requested headers
	requestHeaders := r.Header.Get("Access-Control-Request-Headers")
	if requestHeaders != "" {
		headers := strings.Split(requestHeaders, ",")
		for _, header := range headers {
			header = strings.TrimSpace(header)
			if header != "" && !contains(config.AllowHeaders, header) {
				logging.Warnf("CORS preflight rejected: header %s not allowed", header)
				w.WriteHeader(http.StatusForbidden)
				return MiddlewareResult{Continue: false}
			}
		}
	}

	// Set CORS headers for preflight response
	setCORSHeaders(w, config, allowOrigin)
	w.WriteHeader(http.StatusOK)
	return MiddlewareResult{Continue: false}
}

// handleActualRequest handles actual CORS requests (non-preflight)
func handleActualRequest(w http.ResponseWriter, r *http.Request, config CORSConfig, allowOrigin string) MiddlewareResult {
	// Validate origin for actual requests
	if allowOrigin == "" {
		logging.Warnf("CORS request rejected: origin %s not allowed", r.Header.Get("Origin"))
		w.WriteHeader(http.StatusForbidden)
		return MiddlewareResult{Continue: false}
	}

	// Validate HTTP method
	if !contains(config.AllowMethods, r.Method) {
		logging.Warnf("CORS request rejected: method %s not allowed", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return MiddlewareResult{Continue: false}
	}

	// Validate non-CORS headers against AllowHeaders list
	corsHeaders := []string{
		"Origin",
		"Access-Control-Request-Method",
		"Access-Control-Request-Headers",
		"Access-Control-Allow-Origin",
		"Access-Control-Allow-Methods",
		"Access-Control-Allow-Headers",
		"Access-Control-Allow-Credentials",
		"Access-Control-Max-Age",
		"Access-Control-Expose-Headers",
	}

	for headerName := range r.Header {
		// Skip CORS headers - they're always allowed
		if contains(corsHeaders, headerName) {
			continue
		}

		// Check if non-CORS header is allowed
		if !contains(config.AllowHeaders, headerName) {
			logging.Warnf("CORS request rejected: header %s not allowed", headerName)
			w.WriteHeader(http.StatusForbidden)
			return MiddlewareResult{Continue: false}
		}
	}

	// For actual requests, we don't set CORS headers back
	// CORS headers are only set in preflight responses
	return MiddlewareResult{Continue: true}
}

// setCORSHeaders sets the appropriate CORS headers on the response
func setCORSHeaders(w http.ResponseWriter, config CORSConfig, allowOrigin string) {
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
}

// RequestIDMiddleware creates a request ID pre-middleware
func RequestIDMiddleware() PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
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

		// Get or create AppContext from the current context
		appCtx := appctx.FromContext(r.Context())

		// Update the request ID in the AppContext
		appCtx = appCtx.WithRequestID(requestID)

		// Continue with the updated AppContext
		*r = *r.WithContext(appCtx)

		return MiddlewareResult{Continue: true}
	})
}

// RequestLoggingMiddleware creates a request logging pre-middleware
func RequestLoggingMiddleware() PreMiddleware {
	return PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		isSSE := r.Header.Get("Accept") == "text/event-stream" ||
			strings.HasSuffix(r.URL.Path, "/events/stream")

		if !isSSE {
			logRequest(r)
		} else {
			logSSERequest(r)
		}

		return MiddlewareResult{Continue: true}
	})
}

// ResponseLoggingMiddleware creates a response logging post-middleware
func ResponseLoggingMiddleware() PostMiddleware {
	return PostMiddlewareFunc(func(w http.ResponseWriter, r *http.Request, responseData *ResponseData) error {
		isSSE := r.Header.Get("Accept") == "text/event-stream" ||
			strings.HasSuffix(r.URL.Path, "/events/stream")

		if !isSSE {
			logResponse(r, responseData)
		} else {
			logSSEResponse(r, responseData)
		}

		return nil
	})
}

// contains checks if a slice contains a string (case-insensitive)
func contains(slice []string, item string) bool {
	// If the slice contains "*", everything is allowed
	for _, s := range slice {
		if s == "*" {
			return true
		}
	}

	// Otherwise, check for exact match (case-insensitive)
	for _, s := range slice {
		if strings.EqualFold(s, item) {
			return true
		}
	}
	return false
}

func logRequest(r *http.Request) {
	cfg := config.GetInstance().Get()
	ctx := appctx.FromContext(r.Context())

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

	ctx.LogWithFields(fields).Info("Api Request")
}

func logResponse(r *http.Request, responseData *ResponseData) {
	cfg := config.GetInstance().Get()
	ctx := appctx.FromContext(r.Context())

	// Try to parse response body as JSON for pretty printing
	var responseBody string
	if responseData.ResponseBody.Len() > 0 && cfg.Get(config.DebugKey).GetBool() {
		var prettyJSON bytes.Buffer
		body := responseData.ResponseBody.Bytes()
		if json.Indent(&prettyJSON, body, "", "  ") == nil {
			responseBody = prettyJSON.String()
		} else {
			responseBody = string(body)
		}
	}
	if !cfg.IsDebug() {
		responseBody = "[REDACTED]"
	}

	ctx.LogWithField("duration", utils.FormatDuration(responseData.Duration)).
		WithField("method", r.Method).
		WithField("path", r.URL.Path).
		WithField("status_code", responseData.StatusCode).
		WithField("response_size", responseData.ResponseSize).
		WithField("response_body", responseBody).
		Info("Api Response")
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
	ctx := appctx.FromContext(r.Context())

	reqDetails := map[string]interface{}{
		"timestamp": time.Now().Format(time.RFC3339),
		"method":    r.Method,
		"path":      r.URL.Path,
		"remote":    r.RemoteAddr,
		"type":      "SSE_REQUEST",
	}

	if details, err := json.Marshal(reqDetails); err == nil {
		ctx.LogWithField("sse_request", string(details)).Info("SSE Request")
	}
}

// logSSEResponse logs SSE connection responses without body capture
func logSSEResponse(r *http.Request, responseData *ResponseData) {
	ctx := appctx.FromContext(r.Context())

	respDetails := map[string]interface{}{
		"timestamp":   time.Now().Format(time.RFC3339),
		"method":      r.Method,
		"path":        r.URL.Path,
		"status_code": responseData.StatusCode,
		"duration_ms": responseData.Duration.Milliseconds(),
		"type":        "SSE_RESPONSE",
		"note":        "SSE stream established",
	}

	if details, err := json.Marshal(respDetails); err == nil {
		ctx.LogWithField("sse_response", string(details)).Info("SSE Response")
	}
}
