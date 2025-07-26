package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestRequestIDMiddleware_PreservesAppContext(t *testing.T) {
	// Create an AppContext with tenant_id and user_id
	ctx := context.Background()
	appCtx := appctx.NewContext(ctx)
	appCtx = appCtx.WithTenantID("test-tenant-123")
	appCtx = appCtx.WithUserID("test-user-456")

	// Create a request with the AppContext
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(appCtx)

	// Add a request ID header
	req.Header.Set("X-Request-ID", "test-request-789")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Track what context is passed to the handler
	var handlerContext context.Context

	// Create a test handler that captures the context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerContext = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware chain with RequestID middleware
	chain := NewMiddlewareChain()
	chain.AddPreMiddleware(RequestIDMiddleware())

	// Apply the middleware chain
	middleware := chain.Execute(handler)
	middleware.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the context was passed to the handler
	assert.NotNil(t, handlerContext)

	// Extract AppContext from the handler context
	resultAppCtx := appctx.FromContext(handlerContext)

	// Verify all fields are preserved
	assert.Equal(t, "test-tenant-123", resultAppCtx.GetTenantID())
	assert.Equal(t, "test-user-456", resultAppCtx.GetUserID())
	assert.Equal(t, "test-request-789", resultAppCtx.GetRequestID())

	// Verify the underlying context.Context also has the request ID
	contextRequestID := handlerContext.Value(types.RequestIDKey)
	assert.Equal(t, "test-request-789", contextRequestID)
}

func TestRequestIDMiddleware_GeneratesRequestID(t *testing.T) {
	// Create an AppContext with tenant_id and user_id
	ctx := context.Background()
	appCtx := appctx.NewContext(ctx)
	appCtx = appCtx.WithTenantID("test-tenant-123")
	appCtx = appCtx.WithUserID("test-user-456")

	// Create a request with the AppContext but no request ID header
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(appCtx)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Track what context is passed to the handler
	var handlerContext context.Context

	// Create a test handler that captures the context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerContext = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware chain with RequestID middleware
	chain := NewMiddlewareChain()
	chain.AddPreMiddleware(RequestIDMiddleware())

	// Apply the middleware chain
	middleware := chain.Execute(handler)
	middleware.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the context was passed to the handler
	assert.NotNil(t, handlerContext)

	// Extract AppContext from the handler context
	resultAppCtx := appctx.FromContext(handlerContext)

	// Verify tenant and user IDs are preserved
	assert.Equal(t, "test-tenant-123", resultAppCtx.GetTenantID())
	assert.Equal(t, "test-user-456", resultAppCtx.GetUserID())

	// Verify a request ID was generated (should not be empty)
	assert.NotEmpty(t, resultAppCtx.GetRequestID())

	// Verify the underlying context.Context also has the request ID
	contextRequestID := handlerContext.Value(types.RequestIDKey)
	assert.NotEmpty(t, contextRequestID)
	assert.Equal(t, resultAppCtx.GetRequestID(), contextRequestID)
}

func TestRequestIDMiddleware_ExistingContextRequestID(t *testing.T) {
	// Create an AppContext with tenant_id and user_id
	ctx := context.Background()
	appCtx := appctx.NewContext(ctx)
	appCtx = appCtx.WithTenantID("test-tenant-123")
	appCtx = appCtx.WithUserID("test-user-456")
	appCtx = appCtx.WithRequestID("existing-request-123")

	// Create a request with the AppContext but no request ID header
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(appCtx)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Track what context is passed to the handler
	var handlerContext context.Context

	// Create a test handler that captures the context
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerContext = r.Context()
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware chain with RequestID middleware
	chain := NewMiddlewareChain()
	chain.AddPreMiddleware(RequestIDMiddleware())

	// Apply the middleware chain
	middleware := chain.Execute(handler)
	middleware.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify the context was passed to the handler
	assert.NotNil(t, handlerContext)

	// Extract AppContext from the handler context
	resultAppCtx := appctx.FromContext(handlerContext)

	// Verify tenant and user IDs are preserved
	assert.Equal(t, "test-tenant-123", resultAppCtx.GetTenantID())
	assert.Equal(t, "test-user-456", resultAppCtx.GetUserID())

	// Verify the existing request ID is preserved
	assert.Equal(t, "existing-request-123", resultAppCtx.GetRequestID())

	// Verify the underlying context.Context also has the request ID
	contextRequestID := handlerContext.Value(types.RequestIDKey)
	assert.Equal(t, "existing-request-123", contextRequestID)
}

func TestMiddlewareChain_WithAppContext(t *testing.T) {
	// Initialize config service for the test
	config.Initialize()

	// Create an AppContext with tenant_id and user_id
	ctx := context.Background()
	appCtx := appctx.NewContext(ctx)
	appCtx = appCtx.WithTenantID("test-tenant-123")
	appCtx = appCtx.WithUserID("test-user-456")
	appCtx = appCtx.WithRequestID("test-request-789")

	// Create a request with the AppContext
	req := httptest.NewRequest("GET", "/test", nil)
	req = req.WithContext(appCtx)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test response"))
	})

	// Create middleware chain with request and response logging
	chain := NewMiddlewareChain()
	chain.AddPreMiddleware(RequestLoggingMiddleware())
	chain.AddPostMiddleware(ResponseLoggingMiddleware())

	// Apply the middleware chain
	middleware := chain.Execute(handler)
	middleware.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "test response", w.Body.String())

	// Note: We can't easily capture the log output in tests without mocking the logger
	// But we can verify that the middleware doesn't panic and processes the request correctly
	// The actual log output verification would require integration tests with a real server
}

func TestMiddlewareChain_ShortCircuit(t *testing.T) {
	// Initialize logging for the test
	config.Initialize()

	// Create a short-circuiting middleware
	shortCircuitMiddleware := PreMiddlewareFunc(func(w http.ResponseWriter, r *http.Request) MiddlewareResult {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte("access denied"))
		return MiddlewareResult{Continue: false, Error: fmt.Errorf("access denied")}
	})

	// Create a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Create a test handler (should not be called)
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware chain with short-circuiting middleware
	chain := NewMiddlewareChain()
	chain.AddPreMiddleware(shortCircuitMiddleware)

	// Apply the middleware chain
	middleware := chain.Execute(handler)
	middleware.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Equal(t, "access denied", w.Body.String())

	// Verify the handler was not called
	assert.False(t, handlerCalled)
}

func TestMiddlewareChain_PanicRecovery(t *testing.T) {
	// Initialize logging for the test
	config.Initialize()

	// Create a request
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	// Create a handler that panics
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	// Create middleware chain
	chain := NewMiddlewareChain()

	// Apply the middleware chain
	middleware := chain.Execute(handler)
	middleware.ServeHTTP(w, req)

	// Verify the response (should be 500 due to panic recovery)
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestCORSMiddleware_OptionsRequest(t *testing.T) {
	// Create an OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	// Create a test handler (should not be called for OPTIONS)
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
	})

	// Create middleware chain with CORS middleware
	chain := NewMiddlewareChain()
	chain.AddPreMiddleware(CORSMiddleware(DefaultCORSConfig()))

	// Apply the middleware chain
	middleware := chain.Execute(handler)
	middleware.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Verify CORS headers are set
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

	// Verify the handler was not called (short-circuited by CORS)
	assert.False(t, handlerCalled)
}

func TestCORSMiddleware_RegularRequest(t *testing.T) {
	// Create a regular GET request
	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("Origin", "http://localhost:3000")
	w := httptest.NewRecorder()

	// Create a test handler
	handlerCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("success"))
	})

	// Create middleware chain with CORS middleware
	chain := NewMiddlewareChain()
	chain.AddPreMiddleware(CORSMiddleware(DefaultCORSConfig()))

	// Apply the middleware chain
	middleware := chain.Execute(handler)
	middleware.ServeHTTP(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "success", w.Body.String())

	// Verify CORS headers are set
	assert.Equal(t, "http://localhost:3000", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))

	// Verify the handler was called
	assert.True(t, handlerCalled)
}
