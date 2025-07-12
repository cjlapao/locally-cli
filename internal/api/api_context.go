package api

import (
	"context"
	"time"

	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/types"

	"github.com/sirupsen/logrus"
)

// ApiContext wraps context.Context with additional application-specific data
type ApiContext struct {
	context.Context
	requestID string
	userID    string
	tenantID  string
	startTime time.Time
	metadata  map[string]interface{}
}

// NewContext creates a new CustomContext with the given parent context
func NewContext(parent context.Context) *ApiContext {
	return &ApiContext{
		Context:   parent,
		startTime: time.Now(),
		metadata:  make(map[string]interface{}),
	}
}

// WithRequestID creates a new context with the given request ID
func (c *ApiContext) WithRequestID(requestID string) *ApiContext {
	newCtx := c.clone()
	newCtx.requestID = requestID
	return newCtx
}

// WithUserID creates a new context with the given user ID
func (c *ApiContext) WithUserID(userID string) *ApiContext {
	newCtx := c.clone()
	newCtx.userID = userID
	return newCtx
}

// WithTenantID creates a new context with the given tenant ID
func (c *ApiContext) WithTenantID(tenantID string) *ApiContext {
	newCtx := c.clone()
	newCtx.tenantID = tenantID
	return newCtx
}

// WithMetadata adds metadata to the context
func (c *ApiContext) WithMetadata(key string, value interface{}) *ApiContext {
	newCtx := c.clone()
	newCtx.metadata[key] = value
	return newCtx
}

// WithMetadataMap adds multiple metadata entries to the context
func (c *ApiContext) WithMetadataMap(metadata map[string]interface{}) *ApiContext {
	newCtx := c.clone()
	for k, v := range metadata {
		newCtx.metadata[k] = v
	}
	return newCtx
}

// GetRequestID returns the request ID from the context
func (c *ApiContext) GetRequestID() string {
	return c.requestID
}

// GetUserID returns the user ID from the context
func (c *ApiContext) GetUserID() string {
	return c.userID
}

// GetTenantID returns the tenant ID from the context
func (c *ApiContext) GetTenantID() string {
	return c.tenantID
}

// GetStartTime returns the start time of the context
func (c *ApiContext) GetStartTime() time.Time {
	return c.startTime
}

// GetDuration returns the duration since the context was created
func (c *ApiContext) GetDuration() time.Duration {
	return time.Since(c.startTime)
}

// GetMetadata returns the metadata value for the given key
func (c *ApiContext) GetMetadata(key string) (interface{}, bool) {
	value, exists := c.metadata[key]
	return value, exists
}

// GetMetadataString returns the metadata value as a string
func (c *ApiContext) GetMetadataString(key string) string {
	if value, exists := c.metadata[key]; exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetMetadataInt returns the metadata value as an int
func (c *ApiContext) GetMetadataInt(key string) int {
	if value, exists := c.metadata[key]; exists {
		if i, ok := value.(int); ok {
			return i
		}
	}
	return 0
}

// GetMetadataBool returns the metadata value as a bool
func (c *ApiContext) GetMetadataBool(key string) bool {
	if value, exists := c.metadata[key]; exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// GetAllMetadata returns all metadata as a map
func (c *ApiContext) GetAllMetadata() map[string]interface{} {
	result := make(map[string]interface{})
	for k, v := range c.metadata {
		result[k] = v
	}
	return result
}

// Log creates a logrus.Entry with context information
func (c *ApiContext) Log() *logrus.Entry {
	fields := logrus.Fields{
		"request_id": c.GetRequestID(),
		"user_id":    c.GetUserID(),
		"tenant_id":  c.GetTenantID(),
		"duration":   c.GetDuration().String(),
	}

	// Add metadata fields
	for k, v := range c.metadata {
		fields[k] = v
	}

	return logging.Logger.WithFields(fields)
}

// LogWithLevel logs a message with the given level and context information
func (c *ApiContext) LogWithLevel(level logrus.Level, message string) {
	entry := c.Log()
	switch level {
	case logrus.TraceLevel:
		entry.Trace(message)
	case logrus.DebugLevel:
		entry.Debug(message)
	case logrus.InfoLevel:
		entry.Info(message)
	case logrus.WarnLevel:
		entry.Warn(message)
	case logrus.ErrorLevel:
		entry.Error(message)
	case logrus.FatalLevel:
		entry.Fatal(message)
	case logrus.PanicLevel:
		entry.Panic(message)
	}
}

// LogDebug logs a debug message with context information
func (c *ApiContext) LogDebug(message string) {
	c.LogWithLevel(logrus.DebugLevel, message)
}

// LogInfo logs an info message with context information
func (c *ApiContext) LogInfo(message string) {
	c.LogWithLevel(logrus.InfoLevel, message)
}

// LogWarn logs a warning message with context information
func (c *ApiContext) LogWarn(message string) {
	c.LogWithLevel(logrus.WarnLevel, message)
}

// LogError logs an error message with context information
func (c *ApiContext) LogError(message string) {
	c.LogWithLevel(logrus.ErrorLevel, message)
}

// LogErrorWithError logs an error message with an error and context information
func (c *ApiContext) LogErrorWithError(message string, err error) {
	c.Log().WithError(err).Error(message)
}

// clone creates a deep copy of the context
func (c *ApiContext) clone() *ApiContext {
	newCtx := &ApiContext{
		Context:   c.Context,
		requestID: c.requestID,
		userID:    c.userID,
		tenantID:  c.tenantID,
		startTime: c.startTime,
		metadata:  make(map[string]interface{}),
	}

	// Copy metadata
	for k, v := range c.metadata {
		newCtx.metadata[k] = v
	}

	return newCtx
}

// Value implements context.Context interface
func (c ApiContext) Value(key interface{}) interface{} {
	// Check our custom keys first
	switch k := key.(type) {
	case types.ContextKey:
		switch k {
		case types.RequestIDKey:
			return c.requestID
		case types.UserIDKey:
			return c.userID
		case types.TenantIDKey:
			return c.tenantID
		case types.StartTimeKey:
			return c.startTime
		case types.MetadataKey:
			return c.metadata
		}
	case string:
		// Check if it's one of our context keys
		switch types.ContextKey(k) {
		case types.RequestIDKey:
			return c.requestID
		case types.UserIDKey:
			return c.userID
		case types.TenantIDKey:
			return c.tenantID
		case types.StartTimeKey:
			return c.startTime
		case types.MetadataKey:
			return c.metadata
		}
	}

	// Fall back to parent context
	return c.Context.Value(key)
}

// Deadline implements context.Context interface
func (c ApiContext) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

// Done implements context.Context interface
func (c ApiContext) Done() <-chan struct{} {
	return c.Context.Done()
}

// Err implements context.Context interface
func (c ApiContext) Err() error {
	return c.Context.Err()
}

// Helper functions for working with CustomContext

// FromContext extracts a CustomContext from a standard context
func FromContext(ctx context.Context) *ApiContext {
	if customCtx, ok := ctx.(*ApiContext); ok {
		return customCtx
	}
	requestID := GetRequestID(ctx)

	// Create a new CustomContext from the standard context
	return &ApiContext{
		Context:   ctx,
		startTime: time.Now(),
		requestID: requestID,
		metadata:  make(map[string]interface{}),
	}
}

// WithRequestID creates a new context with request ID
func WithRequestID(ctx context.Context, requestID string) *ApiContext {
	return FromContext(ctx).WithRequestID(requestID)
}

// WithUserID creates a new context with user ID
func WithUserID(ctx context.Context, userID string) *ApiContext {
	return FromContext(ctx).WithUserID(userID)
}

// WithTenantID creates a new context with tenant ID
func WithTenantID(ctx context.Context, tenantID string) *ApiContext {
	return FromContext(ctx).WithTenantID(tenantID)
}

// WithMetadata adds metadata to a context
func WithMetadata(ctx context.Context, key string, value interface{}) *ApiContext {
	return FromContext(ctx).WithMetadata(key, value)
}

// GetRequestID extracts request ID from context
func GetRequestID(ctx context.Context) string {
	if customCtx, ok := ctx.(*ApiContext); ok {
		return customCtx.GetRequestID()
	}

	requestIDValue := ctx.Value(types.RequestIDKey)
	if requestIDValue == nil {
		return ""
	}
	requestID, ok := requestIDValue.(string)
	if !ok {
		return ""
	}
	return requestID
}

// GetUserID extracts user ID from context
func GetUserID(ctx context.Context) string {
	if customCtx, ok := ctx.(*ApiContext); ok {
		return customCtx.GetUserID()
	}
	return ""
}

// GetTenantID extracts tenant ID from context
func GetTenantID(ctx context.Context) string {
	if customCtx, ok := ctx.(*ApiContext); ok {
		return customCtx.GetTenantID()
	}
	return ""
}

// GetMetadata extracts metadata from context
func GetMetadata(ctx context.Context, key string) (interface{}, bool) {
	if customCtx, ok := ctx.(*ApiContext); ok {
		return customCtx.GetMetadata(key)
	}
	return nil, false
}
