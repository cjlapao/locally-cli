// Package appctx provides a context for the application
package appctx

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/types"

	"github.com/sirupsen/logrus"
)

// AppContext wraps context.Context with additional application-specific data
type AppContext struct {
	context.Context
	requestID     string
	correlationID string
	userID        string
	username      string
	tenantID      string
	userIP        string
	userAgent     string
	startTime     time.Time
	metadata      map[string]interface{}
	diagnostics   *diagnostics.Diagnostics
	mu            sync.RWMutex
}

// NewContext creates a new AppContext with the given parent context
func NewContext(parent context.Context) *AppContext {
	if parent == nil {
		parent = context.Background()
	}

	return &AppContext{
		Context:     parent,
		startTime:   time.Now(),
		metadata:    make(map[string]interface{}),
		diagnostics: diagnostics.New("app_context"),
	}
}

// WithRequestID creates a new context with the given request ID
func (c *AppContext) WithRequestID(requestID string) *AppContext {
	newCtx := c.clone()
	newCtx.requestID = requestID

	// Update the underlying context
	newCtx.Context = context.WithValue(newCtx.Context, types.RequestIDKey, requestID)

	// Add to diagnostics
	newCtx.diagnostics.AddMetadata("request_id", requestID)

	return newCtx
}

// WithCorrelationID creates a new context with the given correlation ID

func (c *AppContext) WithCorrelationID(correlationID string) *AppContext {
	newCtx := c.clone()
	newCtx.correlationID = correlationID

	// Update the underlying context
	newCtx.Context = context.WithValue(newCtx.Context, types.CorrelationIDKey, correlationID)

	// Add to diagnostics
	newCtx.diagnostics.AddMetadata("correlation_id", correlationID)

	return newCtx
}

// WithUserID creates a new context with the given user ID
func (c *AppContext) WithUserID(userID string) *AppContext {
	newCtx := c.clone()
	newCtx.userID = userID

	// Update the underlying context
	newCtx.Context = context.WithValue(newCtx.Context, types.UserIDKey, userID)

	// Add to diagnostics
	newCtx.diagnostics.AddMetadata("user_id", userID)

	return newCtx
}

// WithUsername creates a new context with the given username

func (c *AppContext) WithUsername(username string) *AppContext {
	newCtx := c.clone()
	newCtx.username = username

	// Update the underlying context
	newCtx.Context = context.WithValue(newCtx.Context, types.UsernameKey, username)

	// Add to diagnostics
	newCtx.diagnostics.AddMetadata("username", username)

	return newCtx
}

// WithTenantID creates a new context with the given tenant ID
func (c *AppContext) WithTenantID(tenantID string) *AppContext {
	newCtx := c.clone()
	newCtx.tenantID = tenantID

	// Update the underlying context
	newCtx.Context = context.WithValue(newCtx.Context, types.TenantIDKey, tenantID)

	// Add to diagnostics
	newCtx.diagnostics.AddMetadata("tenant_id", tenantID)

	return newCtx
}

func (c *AppContext) WithUserIP(userIP string) *AppContext {
	newCtx := c.clone()
	newCtx.userIP = userIP

	// Update the underlying context
	newCtx.Context = context.WithValue(newCtx.Context, types.UserIPKey, userIP)

	// Add to diagnostics
	newCtx.diagnostics.AddMetadata("user_ip", userIP)

	return newCtx
}

func (c *AppContext) WithUserAgent(userAgent string) *AppContext {
	newCtx := c.clone()
	newCtx.userAgent = userAgent

	// Update the underlying context
	newCtx.Context = context.WithValue(newCtx.Context, types.UserAgentKey, userAgent)

	// Add to diagnostics
	newCtx.diagnostics.AddMetadata("user_agent", userAgent)

	return newCtx
}

// WithMetadata adds metadata to the context
func (c *AppContext) WithMetadata(key string, value interface{}) *AppContext {
	newCtx := c.clone()

	newCtx.mu.Lock()
	newCtx.metadata[key] = value
	newCtx.mu.Unlock()

	// Update the underlying context
	newCtx.Context = context.WithValue(newCtx.Context, types.AppContextKey(key), value)

	// Add to diagnostics
	newCtx.diagnostics.AddMetadata(key, value)

	return newCtx
}

// WithMetadataMap adds multiple metadata entries to the context
func (c *AppContext) WithMetadataMap(metadata map[string]interface{}) *AppContext {
	newCtx := c.clone()

	newCtx.mu.Lock()
	for k, v := range metadata {
		newCtx.metadata[k] = v
	}
	newCtx.mu.Unlock()

	// Update the underlying context for each key
	for k, v := range metadata {
		newCtx.Context = context.WithValue(newCtx.Context, types.AppContextKey(k), v)
	}

	// Add to diagnostics
	for k, v := range metadata {
		newCtx.diagnostics.AddMetadata(k, v)
	}

	return newCtx
}

// WithDiagnostics adds diagnostics to the context
func (c *AppContext) WithDiagnostics(diag *diagnostics.Diagnostics) *AppContext {
	newCtx := c.clone()

	if diag != nil {
		// Create a new diagnostics instance to avoid sharing
		newDiag := diagnostics.New("app_context_with_diagnostics")
		newDiag.Append(c.diagnostics)
		newDiag.Append(diag)
		newCtx.diagnostics = newDiag
	}

	return newCtx
}

// GetRequestID returns the request ID from the context
func (c *AppContext) GetRequestID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.requestID
}

// GetCorrelationID returns the correlation ID from the context

func (c *AppContext) GetCorrelationID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.correlationID
}

// GetUserID returns the user ID from the context
func (c *AppContext) GetUserID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userID
}

// GetUsername returns the username from the context

func (c *AppContext) GetUsername() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.username
}

// GetTenantID returns the tenant ID from the context
func (c *AppContext) GetTenantID() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.tenantID
}

// GetStartTime returns the start time of the context
func (c *AppContext) GetStartTime() time.Time {
	return c.startTime
}

// GetDuration returns the duration since the context was created
func (c *AppContext) GetDuration() time.Duration {
	return time.Since(c.startTime)
}

// GetMetadata returns the metadata value for the given key
func (c *AppContext) GetMetadata(key string) (interface{}, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	value, exists := c.metadata[key]
	return value, exists
}

// GetMetadataString returns the metadata value as a string
func (c *AppContext) GetMetadataString(key string) string {
	if value, exists := c.GetMetadata(key); exists {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return ""
}

// GetMetadataInt returns the metadata value as an int
func (c *AppContext) GetMetadataInt(key string) int {
	if value, exists := c.GetMetadata(key); exists {
		switch v := value.(type) {
		case int:
			return v
		case int64:
			return int(v)
		case float64:
			return int(v)
		}
	}
	return 0
}

// GetMetadataBool returns the metadata value as a bool
func (c *AppContext) GetMetadataBool(key string) bool {
	if value, exists := c.GetMetadata(key); exists {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return false
}

// GetAllMetadata returns all metadata as a map
func (c *AppContext) GetAllMetadata() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make(map[string]interface{})
	for k, v := range c.metadata {
		result[k] = v
	}
	return result
}

// GetUserIP returns the user IP from the context
func (c *AppContext) GetUserIP() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userIP
}

// GetUserAgent returns the user agent from the context
func (c *AppContext) GetUserAgent() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.userAgent
}

// GetDiagnostics returns the diagnostics from the context
func (c *AppContext) GetDiagnostics() *diagnostics.Diagnostics {
	return c.diagnostics
}

// Log creates a logrus.Entry with context information
func (c *AppContext) Log() *logrus.Entry {
	if c == nil {
		if logging.Logger != nil {
			return logging.Logger.WithFields(logrus.Fields{
				"context": "nil",
			})
		}
		logger := logrus.New()
		return logger.WithFields(logrus.Fields{
			"context": "nil",
		})
	}

	tenantID := c.GetTenantID()
	userID := c.GetUserID()
	requestID := c.GetRequestID()

	fields := logrus.Fields{
		"duration": c.GetDuration().String(),
	}

	// Add tenant, user and request ID fields if they are not empty
	if tenantID != "" {
		fields["tenant_id"] = tenantID
	}

	if userID != "" {
		fields["user_id"] = userID
	}

	if requestID != "" {
		fields["request_id"] = requestID
	}

	// Add metadata fields
	c.mu.RLock()
	for k, v := range c.metadata {
		fields[k] = v
	}
	c.mu.RUnlock()

	if logging.Logger != nil {
		return logging.Logger.WithFields(fields)
	}
	logger := logrus.New()
	return logger.WithFields(fields)
}

// LogWithLocation creates a logrus.Entry with context information and file/line location
func (c *AppContext) LogWithLocation() *logrus.Entry {
	entry := c.Log()

	// Add file/line context
	_, file, line, _ := runtime.Caller(1)
	filename := filepath.Base(file)
	return entry.WithField("location", fmt.Sprintf("%s:%d", filename, line))
}

// LogWithFields creates a logrus.Entry with context information and additional fields
func (c *AppContext) LogWithFields(fields logrus.Fields) *logrus.Entry {
	entry := c.Log()
	if len(fields) > 0 {
		entry = entry.WithFields(fields)
	}
	return entry
}

// LogWithField creates a logrus.Entry with context information and a single additional field
func (c *AppContext) LogWithField(key string, value interface{}) *logrus.Entry {
	return c.LogWithFields(logrus.Fields{key: value})
}

// LogWithError creates a logrus.Entry with context information and an error
func (c *AppContext) LogWithError(err error) *logrus.Entry {
	return c.Log().WithError(err)
}

// LogWithLevel logs a message with the given level and context information
func (c *AppContext) LogWithLevel(level logrus.Level, message string) {
	entry := c.LogWithLocation()
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

// LogWithLevelf logs a formatted message with the given level and context information
func (c *AppContext) LogWithLevelf(level logrus.Level, format string, args ...interface{}) {
	entry := c.LogWithLocation()
	switch level {
	case logrus.TraceLevel:
		entry.Tracef(format, args...)
	case logrus.DebugLevel:
		entry.Debugf(format, args...)
	case logrus.InfoLevel:
		entry.Infof(format, args...)
	case logrus.WarnLevel:
		entry.Warnf(format, args...)
	case logrus.ErrorLevel:
		entry.Errorf(format, args...)
	case logrus.FatalLevel:
		entry.Fatalf(format, args...)
	case logrus.PanicLevel:
		entry.Panicf(format, args...)
	}
}

// LogDebug logs a debug message with context information
func (c *AppContext) LogDebug(message string) {
	c.LogWithLevel(logrus.DebugLevel, message)
}

// LogDebugf logs a formatted debug message with context information
func (c *AppContext) LogDebugf(format string, args ...interface{}) {
	c.LogWithLevelf(logrus.DebugLevel, format, args...)
}

// LogInfo logs an info message with context information
func (c *AppContext) LogInfo(message string) {
	c.LogWithLevel(logrus.InfoLevel, message)
}

// LogInfof logs a formatted info message with context information
func (c *AppContext) LogInfof(format string, args ...interface{}) {
	c.LogWithLevelf(logrus.InfoLevel, format, args...)
}

// LogWarn logs a warning message with context information
func (c *AppContext) LogWarn(message string) {
	c.LogWithLevel(logrus.WarnLevel, message)
}

// LogWarnf logs a formatted warning message with context information
func (c *AppContext) LogWarnf(format string, args ...interface{}) {
	c.LogWithLevelf(logrus.WarnLevel, format, args...)
}

// LogError logs an error message with context information
func (c *AppContext) LogError(message string) {
	c.LogWithLevel(logrus.ErrorLevel, message)
}

// LogErrorf logs a formatted error message with context information
func (c *AppContext) LogErrorf(format string, args ...interface{}) {
	c.LogWithLevelf(logrus.ErrorLevel, format, args...)
}

// LogFatal logs a fatal message with context information
func (c *AppContext) LogFatal(message string) {
	c.LogWithLevel(logrus.FatalLevel, message)
}

// LogFatalf logs a formatted fatal message with context information
func (c *AppContext) LogFatalf(format string, args ...interface{}) {
	c.LogWithLevelf(logrus.FatalLevel, format, args...)
}

// LogPanic logs a panic message with context information
func (c *AppContext) LogPanic(message string) {
	c.LogWithLevel(logrus.PanicLevel, message)
}

// LogPanicf logs a formatted panic message with context information
func (c *AppContext) LogPanicf(format string, args ...interface{}) {
	c.LogWithLevelf(logrus.PanicLevel, format, args...)
}

// LogTrace logs a trace message with context information
func (c *AppContext) LogTrace(message string) {
	c.LogWithLevel(logrus.TraceLevel, message)
}

// LogTracef logs a formatted trace message with context information
func (c *AppContext) LogTracef(format string, args ...interface{}) {
	c.LogWithLevelf(logrus.TraceLevel, format, args...)
}

// LogErrorWithError logs an error message with an error and context information
func (c *AppContext) LogErrorWithError(message string, err error) {
	c.LogWithLocation().WithError(err).Error(message)
}

// LogErrorWithErrorf logs a formatted error message with an error and context information
func (c *AppContext) LogErrorWithErrorf(format string, err error, args ...interface{}) {
	c.LogWithLocation().WithError(err).Errorf(format, args...)
}

// LogWithFieldsAndLevel logs a message with additional fields and the given level
func (c *AppContext) LogWithFieldsAndLevel(level logrus.Level, fields logrus.Fields, message string) {
	entry := c.LogWithLocation().WithFields(fields)
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

// LogWithFieldsAndLevelf logs a formatted message with additional fields and the given level
func (c *AppContext) LogWithFieldsAndLevelf(level logrus.Level, fields logrus.Fields, format string, args ...interface{}) {
	entry := c.LogWithLocation().WithFields(fields)
	switch level {
	case logrus.TraceLevel:
		entry.Tracef(format, args...)
	case logrus.DebugLevel:
		entry.Debugf(format, args...)
	case logrus.InfoLevel:
		entry.Infof(format, args...)
	case logrus.WarnLevel:
		entry.Warnf(format, args...)
	case logrus.ErrorLevel:
		entry.Errorf(format, args...)
	case logrus.FatalLevel:
		entry.Fatalf(format, args...)
	case logrus.PanicLevel:
		entry.Panicf(format, args...)
	}
}

// clone creates a deep copy of the context
func (c *AppContext) clone() *AppContext {
	c.mu.RLock()
	defer c.mu.RUnlock()

	newCtx := &AppContext{
		Context:       c.Context,
		requestID:     c.requestID,
		correlationID: c.correlationID,
		userID:        c.userID,
		username:      c.username,
		tenantID:      c.tenantID,
		userIP:        c.userIP,
		userAgent:     c.userAgent,
		startTime:     c.startTime,
		metadata:      make(map[string]interface{}),
		diagnostics:   diagnostics.New("app_context_clone"),
	}

	// Copy metadata
	for k, v := range c.metadata {
		newCtx.metadata[k] = v
	}

	// Copy diagnostics
	if c.diagnostics != nil {
		newCtx.diagnostics.Append(c.diagnostics)
	}

	return newCtx
}

// Value implements context.Context interface
func (c *AppContext) Value(key interface{}) interface{} {
	// Check our custom keys first
	switch k := key.(type) {
	case types.AppContextKey:
		switch k {
		case types.RequestIDKey:
			return c.GetRequestID()
		case types.CorrelationIDKey:
			return c.GetCorrelationID()
		case types.UserIDKey:
			return c.GetUserID()
		case types.UsernameKey:
			return c.GetUsername()
		case types.TenantIDKey:
			return c.GetTenantID()
		case types.StartTimeKey:
			return c.GetStartTime()
		case types.MetadataKey:
			return c.GetAllMetadata()
		}
	case string:
		// Check if it's one of our context keys
		switch types.AppContextKey(k) {
		case types.RequestIDKey:
			return c.GetRequestID()
		case types.CorrelationIDKey:
			return c.GetCorrelationID()
		case types.UserIDKey:
			return c.GetUserID()
		case types.TenantIDKey:
			return c.GetTenantID()
		case types.StartTimeKey:
			return c.GetStartTime()
		case types.MetadataKey:
			return c.GetAllMetadata()
		}

		// Check if it's a metadata key
		if value, exists := c.GetMetadata(k); exists {
			return value
		}
	}

	// Fall back to the parent context
	return c.Context.Value(key)
}

// Deadline implements context.Context interface
func (c *AppContext) Deadline() (deadline time.Time, ok bool) {
	return c.Context.Deadline()
}

// Done implements context.Context interface
func (c *AppContext) Done() <-chan struct{} {
	return c.Context.Done()
}

// Err implements context.Context interface
func (c *AppContext) Err() error {
	return c.Context.Err()
}

// String returns a string representation of the context
func (c *AppContext) String() string {
	return fmt.Sprintf("AppContext{requestID=%s, userID=%s, tenantID=%s, duration=%v}",
		c.GetRequestID(), c.GetUserID(), c.GetTenantID(), c.GetDuration())
}

// FromContext extracts an AppContext from a standard context.Context
func FromContext(ctx context.Context) *AppContext {
	if ctx == nil {
		return NewContext(context.Background())
	}

	// If it's already an AppContext, return it
	if appCtx, ok := ctx.(*AppContext); ok {
		return appCtx
	}

	// Create a new AppContext and populate it from the standard context
	appCtx := NewContext(ctx)

	// Extract values from the standard context
	if requestID, ok := ctx.Value(types.RequestIDKey).(string); ok {
		appCtx.requestID = requestID
	}

	if correlationID, ok := ctx.Value(types.CorrelationIDKey).(string); ok {
		appCtx.correlationID = correlationID
	}

	if userID, ok := ctx.Value(types.UserIDKey).(string); ok {
		appCtx.userID = userID
	}

	if tenantID, ok := ctx.Value(types.TenantIDKey).(string); ok {
		appCtx.tenantID = tenantID
	}

	if userIP, ok := ctx.Value(types.UserIPKey).(string); ok {
		appCtx.userIP = userIP
	}

	if userAgent, ok := ctx.Value(types.UserAgentKey).(string); ok {
		appCtx.userAgent = userAgent
	}

	if startTime, ok := ctx.Value(types.StartTimeKey).(time.Time); ok {
		appCtx.startTime = startTime
	}

	return appCtx
}

// WithRequestID creates a new AppContext with the given request ID
func WithRequestID(ctx context.Context, requestID string) *AppContext {
	appCtx := FromContext(ctx)
	return appCtx.WithRequestID(requestID)
}

// WithUserID creates a new AppContext with the given user ID
func WithUserID(ctx context.Context, userID string) *AppContext {
	appCtx := FromContext(ctx)
	return appCtx.WithUserID(userID)
}

// WithTenantID creates a new AppContext with the given tenant ID
func WithTenantID(ctx context.Context, tenantID string) *AppContext {
	appCtx := FromContext(ctx)
	return appCtx.WithTenantID(tenantID)
}

// WithMetadata creates a new AppContext with the given metadata
func WithMetadata(ctx context.Context, key string, value interface{}) *AppContext {
	appCtx := FromContext(ctx)
	return appCtx.WithMetadata(key, value)
}

// GetRequestID extracts the request ID from a context
func GetRequestID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try to get from AppContext first
	if appCtx, ok := ctx.(*AppContext); ok {
		return appCtx.GetRequestID()
	}

	// Try to get from standard context
	if requestID, ok := ctx.Value(types.RequestIDKey).(string); ok {
		return requestID
	}

	return ""
}

// GetCorrelationID extracts the correlation ID from a context
func GetCorrelationID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try to get from AppContext first
	if appCtx, ok := ctx.(*AppContext); ok {
		return appCtx.GetCorrelationID()
	}

	// Try to get from standard context
	if correlationID, ok := ctx.Value(types.CorrelationIDKey).(string); ok {
		return correlationID
	}

	return ""
}

// GetUserID extracts the user ID from a context
func GetUserID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try to get from AppContext first
	if appCtx, ok := ctx.(*AppContext); ok {
		return appCtx.GetUserID()
	}

	// Try to get from standard context
	if userID, ok := ctx.Value(types.UserIDKey).(string); ok {
		return userID
	}

	return ""
}

// GetTenantID extracts the tenant ID from a context
func GetTenantID(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try to get from AppContext first
	if appCtx, ok := ctx.(*AppContext); ok {
		return appCtx.GetTenantID()
	}

	// Try to get from standard context
	if tenantID, ok := ctx.Value(types.TenantIDKey).(string); ok {
		return tenantID
	}

	return ""
}

// GetUserAgent extracts the user agent from a context
func GetUserAgent(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try to get from AppContext first
	if appCtx, ok := ctx.(*AppContext); ok {
		return appCtx.GetUserAgent()
	}

	// Try to get from standard context
	if userAgent, ok := ctx.Value(types.UserAgentKey).(string); ok {
		return userAgent
	}

	return ""
}

// GetUserIP extracts the user IP from a context
func GetUserIP(ctx context.Context) string {
	if ctx == nil {
		return ""
	}

	// Try to get from standard context
	if userIP, ok := ctx.Value(types.UserIPKey).(string); ok {
		return userIP
	}

	return ""
}

// GetMetadata extracts metadata from a context
func GetMetadata(ctx context.Context, key string) (interface{}, bool) {
	if ctx == nil {
		return nil, false
	}

	// Try to get from AppContext first
	if appCtx, ok := ctx.(*AppContext); ok {
		return appCtx.GetMetadata(key)
	}

	// Try to get from standard context
	if value := ctx.Value(key); value != nil {
		return value, true
	}

	return nil, false
}

// GetDiagnostics extracts diagnostics from a context
func GetDiagnostics(ctx context.Context) *diagnostics.Diagnostics {
	if ctx == nil {
		return nil
	}

	// Try to get from AppContext first
	if appCtx, ok := ctx.(*AppContext); ok {
		return appCtx.GetDiagnostics()
	}

	return nil
}
