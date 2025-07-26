//lint:file-ignore SA1012 Passing nil ctx is intentional in tests
package appctx

import (
	"context"
	"fmt"
	"io"
	"sync"
	"testing"
	"time"

	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/types"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestNewContext(t *testing.T) {
	// Test with nil parent
	ctx := NewContext(context.Background())
	assert.NotNil(t, ctx)
	assert.Equal(t, context.Background(), ctx.Context)
	assert.NotZero(t, ctx.GetStartTime())
	assert.Empty(t, ctx.GetRequestID())
	assert.Empty(t, ctx.GetUserID())
	assert.Empty(t, ctx.GetTenantID())

	// Test with background context
	parent := context.Background()
	ctx = NewContext(parent)
	assert.NotNil(t, ctx)
	assert.Equal(t, parent, ctx.Context)
	assert.NotNil(t, ctx.GetDiagnostics())
}

func TestAppContext_WithRequestID(t *testing.T) {
	ctx := NewContext(context.Background())
	requestID := "test-request-123"

	newCtx := ctx.WithRequestID(requestID)
	assert.NotEqual(t, ctx, newCtx) // Should be a new instance
	assert.Equal(t, requestID, newCtx.GetRequestID())

	// Test that it's accessible via context.Value
	value := newCtx.Value(types.RequestIDKey)
	assert.Equal(t, requestID, value)

	// Test that it's accessible via string key
	value = newCtx.Value(string(types.RequestIDKey))
	assert.Equal(t, requestID, value)

	// Test that diagnostics contains the request ID
	diag := newCtx.GetDiagnostics()
	metadata := diag.GetAllMetadata()
	assert.Equal(t, requestID, metadata["request_id"])
}

func TestAppContext_WithUserID(t *testing.T) {
	ctx := NewContext(context.Background())
	userID := "user-456"

	newCtx := ctx.WithUserID(userID)
	assert.NotEqual(t, ctx, newCtx)
	assert.Equal(t, userID, newCtx.GetUserID())

	// Test that it's accessible via context.Value
	value := newCtx.Value(types.UserIDKey)
	assert.Equal(t, userID, value)

	// Test that diagnostics contains the user ID
	diag := newCtx.GetDiagnostics()
	metadata := diag.GetAllMetadata()
	assert.Equal(t, userID, metadata["user_id"])
}

func TestAppContext_WithTenantID(t *testing.T) {
	ctx := NewContext(context.Background())
	tenantID := "tenant-789"

	newCtx := ctx.WithTenantID(tenantID)
	assert.NotEqual(t, ctx, newCtx)
	assert.Equal(t, tenantID, newCtx.GetTenantID())

	// Test that it's accessible via context.Value
	value := newCtx.Value(types.TenantIDKey)
	assert.Equal(t, tenantID, value)

	// Test that diagnostics contains the tenant ID
	diag := newCtx.GetDiagnostics()
	metadata := diag.GetAllMetadata()
	assert.Equal(t, tenantID, metadata["tenant_id"])
}

func TestAppContext_WithMetadata(t *testing.T) {
	ctx := NewContext(context.Background())
	key := "test_key"
	value := "test_value"

	newCtx := ctx.WithMetadata(key, value)
	assert.NotEqual(t, ctx, newCtx)

	// Test direct access
	retrievedValue, exists := newCtx.GetMetadata(key)
	assert.True(t, exists)
	assert.Equal(t, value, retrievedValue)

	// Test via context.Value
	ctxValue := newCtx.Value(key)
	assert.Equal(t, value, ctxValue)

	// Test that diagnostics contains the metadata
	diag := newCtx.GetDiagnostics()
	metadata := diag.GetAllMetadata()
	assert.Equal(t, value, metadata[key])
}

func TestAppContext_WithMetadataMap(t *testing.T) {
	ctx := NewContext(context.Background())
	metadata := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
		"key3": true,
	}

	newCtx := ctx.WithMetadataMap(metadata)
	assert.NotEqual(t, ctx, newCtx)

	// Test all metadata
	for k, v := range metadata {
		retrievedValue, exists := newCtx.GetMetadata(k)
		assert.True(t, exists)
		assert.Equal(t, v, retrievedValue)

		// Test via context.Value
		ctxValue := newCtx.Value(k)
		assert.Equal(t, v, ctxValue)
	}

	// Test that diagnostics contains all metadata
	diag := newCtx.GetDiagnostics()
	diagMetadata := diag.GetAllMetadata()
	for k, v := range metadata {
		assert.Equal(t, v, diagMetadata[k])
	}
}

func TestAppContext_WithDiagnostics(t *testing.T) {
	ctx := NewContext(context.Background())
	diag := diagnostics.New("test_diagnostics")
	diag.AddError("TEST_ERROR", "Test error", "test")

	newCtx := ctx.WithDiagnostics(diag)
	assert.NotEqual(t, ctx, newCtx)

	// Test that diagnostics were merged
	resultDiag := newCtx.GetDiagnostics()
	assert.Equal(t, 1, resultDiag.GetErrorCount())

	// Test that the original diagnostics is not modified
	originalDiag := ctx.GetDiagnostics()
	assert.Equal(t, 0, originalDiag.GetErrorCount())
}

func TestAppContext_GetMetadataTypes(t *testing.T) {
	ctx := NewContext(context.Background())
	ctx = ctx.WithMetadata("string_key", "string_value")
	ctx = ctx.WithMetadata("int_key", 42)
	ctx = ctx.WithMetadata("bool_key", true)

	// Test string retrieval
	assert.Equal(t, "string_value", ctx.GetMetadataString("string_key"))
	assert.Equal(t, "", ctx.GetMetadataString("int_key")) // Should return empty string for non-string

	// Test int retrieval
	assert.Equal(t, 42, ctx.GetMetadataInt("int_key"))
	assert.Equal(t, 0, ctx.GetMetadataInt("string_key")) // Should return 0 for non-int

	// Test bool retrieval
	assert.Equal(t, true, ctx.GetMetadataBool("bool_key"))
	assert.Equal(t, false, ctx.GetMetadataBool("string_key")) // Should return false for non-bool
}

func TestAppContext_GetAllMetadata(t *testing.T) {
	ctx := NewContext(context.Background())
	ctx = ctx.WithMetadata("key1", "value1")
	ctx = ctx.WithMetadata("key2", 42)

	metadata := ctx.GetAllMetadata()
	assert.Equal(t, "value1", metadata["key1"])
	assert.Equal(t, 42, metadata["key2"])
	assert.Len(t, metadata, 2)
}

func TestAppContext_GetDuration(t *testing.T) {
	ctx := NewContext(context.Background())

	// Wait a bit
	time.Sleep(1 * time.Millisecond)

	duration := ctx.GetDuration()
	assert.Greater(t, duration, time.Duration(0))
	assert.GreaterOrEqual(t, duration, time.Millisecond)
}

func TestAppContext_ContextInterface(t *testing.T) {
	ctx := NewContext(context.Background())

	// Test Deadline
	deadline, ok := ctx.Deadline()
	assert.False(t, ok) // Background context has no deadline
	assert.True(t, deadline.IsZero())

	// Test Done
	done := ctx.Done()
	assert.Nil(t, done) // Background context never done

	// Test Err
	err := ctx.Err()
	assert.Nil(t, err) // Background context has no error
}

func TestAppContext_Value(t *testing.T) {
	ctx := NewContext(context.Background())
	ctx = ctx.WithRequestID("req-123")
	ctx = ctx.WithUserID("user-456")
	ctx = ctx.WithTenantID("tenant-789")
	ctx = ctx.WithMetadata("custom_key", "custom_value")

	// Test custom context keys
	assert.Equal(t, "req-123", ctx.Value(types.RequestIDKey))
	assert.Equal(t, "user-456", ctx.Value(types.UserIDKey))
	assert.Equal(t, "tenant-789", ctx.Value(types.TenantIDKey))
	assert.Equal(t, "custom_value", ctx.Value("custom_key"))

	// Test string keys
	assert.Equal(t, "req-123", ctx.Value(string(types.RequestIDKey)))
	assert.Equal(t, "user-456", ctx.Value(string(types.UserIDKey)))
	assert.Equal(t, "tenant-789", ctx.Value(string(types.TenantIDKey)))

	// Test metadata key
	metadata := ctx.Value(types.MetadataKey)
	assert.NotNil(t, metadata)
	metadataMap, ok := metadata.(map[string]interface{})
	assert.True(t, ok)
	assert.Equal(t, "custom_value", metadataMap["custom_key"])

	// Test unknown key
	assert.Nil(t, ctx.Value("unknown_key"))
}

func TestAppContext_ConcurrentAccess(t *testing.T) {
	ctx := NewContext(context.Background())

	// Test concurrent metadata access
	done := make(chan bool, 10)
	for i := 0; i < 10; i++ {
		go func(index int) {
			ctx.WithMetadata("concurrent_key", index)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}

	// Should not panic
	assert.NotNil(t, ctx)
}

func TestFromContext(t *testing.T) {
	// Test with nil context
	ctx := FromContext(context.Background())
	assert.NotNil(t, ctx)
	assert.Equal(t, context.Background(), ctx.Context)

	// Test with AppContext
	originalCtx := NewContext(context.Background())
	originalCtx = originalCtx.WithRequestID("req-123")

	resultCtx := FromContext(originalCtx)
	assert.Equal(t, originalCtx, resultCtx)

	// Test with standard context containing our keys
	standardCtx := context.Background()
	standardCtx = context.WithValue(standardCtx, types.RequestIDKey, "req-456")
	standardCtx = context.WithValue(standardCtx, types.UserIDKey, "user-789")

	resultCtx = FromContext(standardCtx)
	assert.Equal(t, "req-456", resultCtx.GetRequestID())
	assert.Equal(t, "user-789", resultCtx.GetUserID())
}

func TestWithRequestID_Helper(t *testing.T) {
	ctx := context.Background()
	requestID := "req-123"

	resultCtx := WithRequestID(ctx, requestID)
	assert.Equal(t, requestID, resultCtx.GetRequestID())
	assert.Equal(t, requestID, resultCtx.Value(types.RequestIDKey))
}

func TestWithUserID_Helper(t *testing.T) {
	ctx := context.Background()
	userID := "user-456"

	resultCtx := WithUserID(ctx, userID)
	assert.Equal(t, userID, resultCtx.GetUserID())
	assert.Equal(t, userID, resultCtx.Value(types.UserIDKey))
}

func TestWithTenantID_Helper(t *testing.T) {
	ctx := context.Background()
	tenantID := "tenant-789"

	resultCtx := WithTenantID(ctx, tenantID)
	assert.Equal(t, tenantID, resultCtx.GetTenantID())
	assert.Equal(t, tenantID, resultCtx.Value(types.TenantIDKey))
}

func TestWithMetadata_Helper(t *testing.T) {
	ctx := context.Background()
	key := "test_key"
	value := "test_value"

	resultCtx := WithMetadata(ctx, key, value)
	retrievedValue, exists := resultCtx.GetMetadata(key)
	assert.True(t, exists)
	assert.Equal(t, value, retrievedValue)
	assert.Equal(t, value, resultCtx.Value(key))
}

func TestGetRequestID_Helper(t *testing.T) {
	// Test with nil context
	assert.Equal(t, "", GetRequestID(context.Background()))

	// Test with AppContext
	appCtx := NewContext(context.Background()).WithRequestID("req-123")
	assert.Equal(t, "req-123", GetRequestID(appCtx))

	// Test with standard context
	standardCtx := context.WithValue(context.Background(), types.RequestIDKey, "req-456")
	assert.Equal(t, "req-456", GetRequestID(standardCtx))

	// Test with context without request ID
	assert.Equal(t, "", GetRequestID(context.Background()))
}

func TestGetUserID_Helper(t *testing.T) {
	// Test with nil context
	assert.Equal(t, "", GetUserID(context.Background()))

	// Test with AppContext
	appCtx := NewContext(context.Background()).WithUserID("user-123")
	assert.Equal(t, "user-123", GetUserID(appCtx))

	// Test with standard context
	standardCtx := context.WithValue(context.Background(), types.UserIDKey, "user-456")
	assert.Equal(t, "user-456", GetUserID(standardCtx))
}

func TestGetTenantID_Helper(t *testing.T) {
	// Test with nil context
	assert.Equal(t, "", GetTenantID(context.Background()))

	// Test with AppContext
	appCtx := NewContext(context.Background()).WithTenantID("tenant-123")
	assert.Equal(t, "tenant-123", GetTenantID(appCtx))

	// Test with standard context
	standardCtx := context.WithValue(context.Background(), types.TenantIDKey, "tenant-456")
	assert.Equal(t, "tenant-456", GetTenantID(standardCtx))
}

func TestGetMetadata_Helper(t *testing.T) {
	// Test with nil context
	value, exists := GetMetadata(context.Background(), "key")
	assert.Nil(t, value)
	assert.False(t, exists)

	// Test with AppContext
	appCtx := NewContext(context.Background()).WithMetadata("key", "value")
	value, exists = GetMetadata(appCtx, "key")
	assert.Equal(t, "value", value)
	assert.True(t, exists)

	// Test with standard context
	standardCtx := context.WithValue(context.Background(), "key", "value")
	value, exists = GetMetadata(standardCtx, "key")
	assert.Equal(t, "value", value)
	assert.True(t, exists)
}

func TestGetDiagnostics_Helper(t *testing.T) {
	// Test with nil context
	assert.Nil(t, GetDiagnostics(context.Background()))

	// Test with AppContext
	appCtx := NewContext(context.Background())
	diag := appCtx.GetDiagnostics()
	assert.NotNil(t, diag)
	assert.Equal(t, diag, GetDiagnostics(appCtx))

	// Test with standard context
	standardCtx := context.Background()
	assert.Nil(t, GetDiagnostics(standardCtx))
}

func TestAppContext_String(t *testing.T) {
	ctx := NewContext(context.Background())
	ctx = ctx.WithRequestID("req-123")
	ctx = ctx.WithUserID("user-456")
	ctx = ctx.WithTenantID("tenant-789")

	str := ctx.String()
	assert.Contains(t, str, "req-123")
	assert.Contains(t, str, "user-456")
	assert.Contains(t, str, "tenant-789")
	assert.Contains(t, str, "AppContext")
}

func TestAppContext_Logging(t *testing.T) {
	ctx := NewContext(context.Background())
	ctx = ctx.WithRequestID("req-123")
	ctx = ctx.WithUserID("user-456")
	ctx = ctx.WithTenantID("tenant-789")
	ctx = ctx.WithMetadata("log_key", "log_value")

	// Test Log() method
	logEntry := ctx.Log()
	assert.NotNil(t, logEntry)

	// Test logging methods (should not panic)
	ctx.LogDebug("Debug message")
	ctx.LogInfo("Info message")
	ctx.LogWarn("Warning message")
	ctx.LogError("Error message")
	ctx.LogErrorWithError("Error with err", assert.AnError)
}

func TestAppContext_Chaining(t *testing.T) {
	ctx := NewContext(context.Background())

	// Test method chaining
	resultCtx := ctx.
		WithRequestID("req-123").
		WithUserID("user-456").
		WithTenantID("tenant-789").
		WithMetadata("key1", "value1").
		WithMetadata("key2", "value2")

	assert.Equal(t, "req-123", resultCtx.GetRequestID())
	assert.Equal(t, "user-456", resultCtx.GetUserID())
	assert.Equal(t, "tenant-789", resultCtx.GetTenantID())

	value1, exists := resultCtx.GetMetadata("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value1)

	value2, exists := resultCtx.GetMetadata("key2")
	assert.True(t, exists)
	assert.Equal(t, "value2", value2)
}

func TestAppContext_ContextWithTimeout(t *testing.T) {
	// Test that AppContext works with context.WithTimeout
	ctx := NewContext(context.Background())
	ctx = ctx.WithRequestID("req-123")

	timeoutCtx, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()

	// Extract AppContext from timeout context
	appCtx := FromContext(timeoutCtx)
	assert.Equal(t, "req-123", appCtx.GetRequestID())

	// Test deadline
	deadline, ok := timeoutCtx.Deadline()
	assert.True(t, ok)
	assert.True(t, deadline.After(time.Now()))

	// Test done channel
	select {
	case <-timeoutCtx.Done():
		// Should not be done immediately
		t.Fatal("Context should not be done immediately")
	default:
		// Expected
	}

	// Wait for timeout
	time.Sleep(150 * time.Millisecond)

	select {
	case <-timeoutCtx.Done():
		// Should be done after timeout
	default:
		t.Fatal("Context should be done after timeout")
	}
}

func TestAppContext_ContextWithCancel(t *testing.T) {
	// Test that AppContext works with context.WithCancel
	ctx := NewContext(context.Background())
	ctx = ctx.WithRequestID("req-123")

	cancelCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Extract AppContext from cancel context
	appCtx := FromContext(cancelCtx)
	assert.Equal(t, "req-123", appCtx.GetRequestID())

	// Test that it's not done initially
	select {
	case <-cancelCtx.Done():
		t.Fatal("Context should not be done initially")
	default:
		// Expected
	}

	// Cancel the context
	cancel()

	// Test that it's done after cancel
	select {
	case <-cancelCtx.Done():
		// Expected
	default:
		t.Fatal("Context should be done after cancel")
	}

	// Test error
	err := cancelCtx.Err()
	assert.Equal(t, context.Canceled, err)
}

func TestAppContext_IntegrationWithDiagnostics(t *testing.T) {
	ctx := NewContext(context.Background())
	ctx = ctx.WithRequestID("req-123")

	// Create some diagnostics
	diag := diagnostics.New("test_operation")
	diag.AddError("TEST_ERROR", "Test error", "test")
	diag.AddWarning("TEST_WARNING", "Test warning", "test")

	// Add diagnostics to context
	ctx = ctx.WithDiagnostics(diag)

	// Verify diagnostics are accessible
	resultDiag := ctx.GetDiagnostics()
	assert.Equal(t, 1, resultDiag.GetErrorCount())
	assert.Equal(t, 1, resultDiag.GetWarningCount())

	// Verify context metadata is in diagnostics
	metadata := resultDiag.GetAllMetadata()
	assert.Equal(t, "req-123", metadata["request_id"])
}

func TestAppContext_NilHandling(t *testing.T) {
	// Test that all methods handle nil gracefully

	// Test that we can create a new context from nil
	newCtx := FromContext(nil)
	assert.NotNil(t, newCtx)
	assert.Equal(t, "", newCtx.GetRequestID())
	assert.Equal(t, "", newCtx.GetUserID())
	assert.Equal(t, "", newCtx.GetTenantID())

	// Test that helper functions handle nil gracefully
	assert.Equal(t, "", GetRequestID(nil))
	assert.Equal(t, "", GetUserID(nil))
	assert.Equal(t, "", GetTenantID(nil))

	value, exists := GetMetadata(nil, "key")
	assert.Nil(t, value)
	assert.False(t, exists)

	assert.Nil(t, GetDiagnostics(nil))
}

func TestNewContext_WithNilParent(t *testing.T) {
	// Test NewContext with nil parent
	ctx := NewContext(nil)
	assert.NotNil(t, ctx)
	assert.Equal(t, context.Background(), ctx.Context)
	assert.NotZero(t, ctx.GetStartTime())
	assert.Empty(t, ctx.GetRequestID())
	assert.Empty(t, ctx.GetUserID())
	assert.Empty(t, ctx.GetTenantID())
	assert.NotNil(t, ctx.GetDiagnostics())
}

func TestGetMetadataInt_WithDifferentTypes(t *testing.T) {
	ctx := NewContext(nil)
	ctx = ctx.WithMetadata("int_key", 42)
	ctx = ctx.WithMetadata("int64_key", int64(123))
	ctx = ctx.WithMetadata("float64_key", float64(456.7))
	ctx = ctx.WithMetadata("string_key", "not_a_number")

	// Test int
	assert.Equal(t, 42, ctx.GetMetadataInt("int_key"))

	// Test int64
	assert.Equal(t, 123, ctx.GetMetadataInt("int64_key"))

	// Test float64
	assert.Equal(t, 456, ctx.GetMetadataInt("float64_key"))

	// Test string (should return 0)
	assert.Equal(t, 0, ctx.GetMetadataInt("string_key"))

	// Test non-existent key
	assert.Equal(t, 0, ctx.GetMetadataInt("non_existent"))
}

func TestLogWithLevel_AllLevels(t *testing.T) {
	ctx := NewContext(nil)
	ctx = ctx.WithRequestID("req-123")

	// Test all log levels except FatalLevel and PanicLevel
	levels := []logrus.Level{
		logrus.TraceLevel,
		logrus.DebugLevel,
		logrus.InfoLevel,
		logrus.WarnLevel,
		logrus.ErrorLevel,
		// logrus.FatalLevel, // Don't test this directly!
		// logrus.PanicLevel, // Don't test this directly!
	}

	for _, level := range levels {
		// These should not panic
		ctx.LogWithLevel(level, "Test message")
	}
}

func TestValue_EdgeCases(t *testing.T) {
	ctx := NewContext(nil)
	ctx = ctx.WithRequestID("req-123")
	ctx = ctx.WithUserID("user-456")
	ctx = ctx.WithTenantID("tenant-789")
	ctx = ctx.WithMetadata("custom_key", "custom_value")

	// Test with unknown type
	unknownKey := 42
	result := ctx.Value(unknownKey)
	assert.Nil(t, result)

	// Test with nil key
	result = ctx.Value(nil)
	assert.Nil(t, result)

	// Test with empty string key
	result = ctx.Value("")
	assert.Nil(t, result)

	// Test with unknown string key
	result = ctx.Value("unknown_string_key")
	assert.Nil(t, result)
}

func TestFromContext_EdgeCases(t *testing.T) {
	// Test with context containing invalid types
	standardCtx := context.Background()
	standardCtx = context.WithValue(standardCtx, types.RequestIDKey, 123) // Wrong type
	standardCtx = context.WithValue(standardCtx, types.UserIDKey, true)   // Wrong type
	standardCtx = context.WithValue(standardCtx, types.TenantIDKey, 456)  // Wrong type

	resultCtx := FromContext(standardCtx)
	assert.Equal(t, "", resultCtx.GetRequestID()) // Should be empty due to wrong type
	assert.Equal(t, "", resultCtx.GetUserID())    // Should be empty due to wrong type
	assert.Equal(t, "", resultCtx.GetTenantID())  // Should be empty due to wrong type
}

func TestAppContext_ConcurrentMetadataAccess(t *testing.T) {
	ctx := NewContext(nil)

	// Test concurrent metadata reads and writes
	var wg sync.WaitGroup
	numGoroutines := 10

	// Start goroutines that write metadata
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ctx.WithMetadata(fmt.Sprintf("key_%d", index), fmt.Sprintf("value_%d", index))
		}(i)
	}

	// Start goroutines that read metadata
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			ctx.GetMetadata(fmt.Sprintf("key_%d", index))
			ctx.GetAllMetadata()
		}(i)
	}

	wg.Wait()
	// Should not panic
	assert.NotNil(t, ctx)
}

func TestAppContext_ConcurrentDiagnosticsAccess(t *testing.T) {
	ctx := NewContext(nil)

	// Test concurrent diagnostics access
	var wg sync.WaitGroup
	numGoroutines := 5

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			diag := diagnostics.New(fmt.Sprintf("test_%d", index))
			diag.AddError("TEST_ERROR", "Test error", "test")
			ctx.WithDiagnostics(diag)
		}(i)
	}

	wg.Wait()
	// Should not panic
	assert.NotNil(t, ctx)
}

func TestAppContext_WithDiagnostics_NilDiagnostics(t *testing.T) {
	ctx := NewContext(nil)
	ctx = ctx.WithRequestID("req-123")

	// Test with nil diagnostics
	newCtx := ctx.WithDiagnostics(nil)
	assert.NotEqual(t, ctx, newCtx)

	// Should still have the original diagnostics
	resultDiag := newCtx.GetDiagnostics()
	assert.NotNil(t, resultDiag)
	assert.Equal(t, 0, resultDiag.GetErrorCount())
}

func TestAppContext_Value_WithParentContext(t *testing.T) {
	// Create a parent context with some values
	parentCtx := context.Background()
	parentCtx = context.WithValue(parentCtx, "parent_key", "parent_value")

	// Create AppContext with parent
	ctx := NewContext(parentCtx)
	ctx = ctx.WithRequestID("req-123")

	// Test that we can access parent context values
	result := ctx.Value("parent_key")
	assert.Equal(t, "parent_value", result)

	// Test that we can access our own values
	result = ctx.Value(types.RequestIDKey)
	assert.Equal(t, "req-123", result)
}

func TestAppContext_Clone_WithFullData(t *testing.T) {
	ctx := NewContext(nil)
	ctx = ctx.WithRequestID("req-123")
	ctx = ctx.WithUserID("user-456")
	ctx = ctx.WithTenantID("tenant-789")
	ctx = ctx.WithMetadata("key1", "value1")
	ctx = ctx.WithMetadata("key2", 42)

	// Add some diagnostics
	diag := diagnostics.New("test_operation")
	diag.AddError("TEST_ERROR", "Test error", "test")
	diag.AddWarning("TEST_WARNING", "Test warning", "test")
	ctx = ctx.WithDiagnostics(diag)

	// Clone the context
	clonedCtx := ctx.clone()

	// Verify all data is copied
	assert.Equal(t, ctx.GetRequestID(), clonedCtx.GetRequestID())
	assert.Equal(t, ctx.GetUserID(), clonedCtx.GetUserID())
	assert.Equal(t, ctx.GetTenantID(), clonedCtx.GetTenantID())
	assert.Equal(t, ctx.GetStartTime(), clonedCtx.GetStartTime())

	// Verify metadata is copied
	value1, exists := clonedCtx.GetMetadata("key1")
	assert.True(t, exists)
	assert.Equal(t, "value1", value1)

	value2, exists := clonedCtx.GetMetadata("key2")
	assert.True(t, exists)
	assert.Equal(t, 42, value2)

	// Verify diagnostics are copied
	clonedDiag := clonedCtx.GetDiagnostics()
	assert.Equal(t, 1, clonedDiag.GetErrorCount())
	assert.Equal(t, 1, clonedDiag.GetWarningCount())

	// Verify they are separate instances
	assert.NotEqual(t, ctx, clonedCtx)
}

func TestAppContext_Log_WithNilLogger(t *testing.T) {
	// Save original logger
	originalLogger := logging.Logger

	// Set logger to nil
	logging.Logger = nil

	// Create context and test logging
	ctx := NewContext(nil)
	ctx = ctx.WithRequestID("req-123")

	// This should not panic
	logEntry := ctx.Log()
	assert.NotNil(t, logEntry)

	// Test logging methods
	ctx.LogDebug("Debug message")
	ctx.LogInfo("Info message")
	ctx.LogWarn("Warning message")
	ctx.LogError("Error message")

	// Restore original logger
	logging.Logger = originalLogger
}

func TestAppContext_GetMetadataTypes_EdgeCases(t *testing.T) {
	ctx := NewContext(nil)
	ctx = ctx.WithMetadata("string_key", "string_value")
	ctx = ctx.WithMetadata("int_key", 42)
	ctx = ctx.WithMetadata("bool_key", true)
	ctx = ctx.WithMetadata("nil_key", nil)
	ctx = ctx.WithMetadata("map_key", map[string]string{"nested": "value"})

	// Test string retrieval with different types
	assert.Equal(t, "string_value", ctx.GetMetadataString("string_key"))
	assert.Equal(t, "", ctx.GetMetadataString("int_key"))
	assert.Equal(t, "", ctx.GetMetadataString("bool_key"))
	assert.Equal(t, "", ctx.GetMetadataString("nil_key"))
	assert.Equal(t, "", ctx.GetMetadataString("map_key"))
	assert.Equal(t, "", ctx.GetMetadataString("non_existent"))

	// Test bool retrieval with different types
	assert.Equal(t, false, ctx.GetMetadataBool("string_key"))
	assert.Equal(t, false, ctx.GetMetadataBool("int_key"))
	assert.Equal(t, true, ctx.GetMetadataBool("bool_key"))
	assert.Equal(t, false, ctx.GetMetadataBool("nil_key"))
	assert.Equal(t, false, ctx.GetMetadataBool("map_key"))
	assert.Equal(t, false, ctx.GetMetadataBool("non_existent"))
}

func TestAppContext_ContextWithDeadline(t *testing.T) {
	ctx := NewContext(nil)
	ctx = ctx.WithRequestID("req-123")

	// Test with deadline
	deadline := time.Now().Add(100 * time.Millisecond)
	deadlineCtx, cancel := context.WithDeadline(ctx, deadline)
	defer cancel()

	// Extract AppContext from deadline context
	appCtx := FromContext(deadlineCtx)
	assert.Equal(t, "req-123", appCtx.GetRequestID())

	// Test deadline
	resultDeadline, ok := deadlineCtx.Deadline()
	assert.True(t, ok)
	assert.True(t, resultDeadline.Before(deadline.Add(time.Millisecond)))
	assert.True(t, resultDeadline.After(deadline.Add(-time.Millisecond)))

	// Test that it's not done initially
	select {
	case <-deadlineCtx.Done():
		t.Fatal("Context should not be done initially")
	default:
		// Expected
	}

	// Wait for deadline
	time.Sleep(150 * time.Millisecond)

	select {
	case <-deadlineCtx.Done():
		// Expected
	default:
		t.Fatal("Context should be done after deadline")
	}

	// Test error
	err := deadlineCtx.Err()
	assert.Equal(t, context.DeadlineExceeded, err)
}

func TestAppContext_EnhancedLogging(t *testing.T) {
	// Create a simple logger for testing without full initialization
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard output for testing
	logger.SetLevel(logrus.InfoLevel)

	// Store original logger and restore it after test
	originalLogger := logging.Logger
	logging.Logger = logger
	defer func() {
		logging.Logger = originalLogger
	}()

	ctx := NewContext(context.Background())
	ctx = ctx.WithRequestID("req-123")
	ctx = ctx.WithUserID("user-456")
	ctx = ctx.WithTenantID("tenant-789")
	ctx = ctx.WithMetadata("operation", "test-operation")

	// Test basic logging methods
	t.Run("BasicLogging", func(t *testing.T) {
		// These should log with context information
		ctx.LogInfo("This is an info message")
		ctx.LogDebug("This is a debug message")
		ctx.LogWarn("This is a warning message")
		ctx.LogError("This is an error message")
	})

	t.Run("FormattedLogging", func(t *testing.T) {
		// Test formatted logging
		ctx.LogInfof("Processing request %s for user %s", "req-123", "user-456")
		ctx.LogDebugf("Debug info: %s", "some debug data")
		ctx.LogWarnf("Warning: %s", "something to watch")
		ctx.LogErrorf("Error occurred: %s", "something went wrong")
	})

	t.Run("LoggingWithFields", func(t *testing.T) {
		// Test logging with additional fields
		ctx.LogWithFieldsAndLevel(logrus.InfoLevel, logrus.Fields{
			"custom_field": "custom_value",
			"count":        42,
		}, "Message with custom fields")

		ctx.LogWithFieldsAndLevelf(logrus.InfoLevel, logrus.Fields{
			"operation": "test",
			"status":    "success",
		}, "Formatted message with fields: %s", "test data")
	})

	t.Run("ErrorLogging", func(t *testing.T) {
		// Test error logging
		err := fmt.Errorf("test error")
		ctx.LogErrorWithError("An error occurred", err)
		ctx.LogErrorWithErrorf("Error processing request: %v", err)
	})

	t.Run("LogEntryCreation", func(t *testing.T) {
		// Test creating log entries with different methods
		entry := ctx.Log()
		if entry == nil {
			t.Error("Log() should return a valid entry")
		}

		entryWithLocation := ctx.LogWithLocation()
		if entryWithLocation == nil {
			t.Error("LogWithLocation() should return a valid entry")
		}

		entryWithFields := ctx.LogWithFields(logrus.Fields{"test": "value"})
		if entryWithFields == nil {
			t.Error("LogWithFields() should return a valid entry")
		}

		entryWithField := ctx.LogWithField("single_field", "single_value")
		if entryWithField == nil {
			t.Error("LogWithField() should return a valid entry")
		}

		entryWithError := ctx.LogWithError(fmt.Errorf("test error"))
		if entryWithError == nil {
			t.Error("LogWithError() should return a valid entry")
		}
	})
}

func TestAppContext_LoggingWithNilContext(t *testing.T) {
	// Create a simple logger for testing without full initialization
	logger := logrus.New()
	logger.SetOutput(io.Discard) // Discard output for testing
	logger.SetLevel(logrus.InfoLevel)

	// Store original logger and restore it after test
	originalLogger := logging.Logger
	logging.Logger = logger
	defer func() {
		logging.Logger = originalLogger
	}()

	// Test logging with nil context
	var ctx *AppContext = nil

	// These should not panic
	ctx.LogInfo("This should not panic")
	ctx.LogDebug("This should not panic")
	ctx.LogWarn("This should not panic")
	ctx.LogError("This should not panic")

	// Test entry creation with nil context
	entry := ctx.Log()
	if entry == nil {
		t.Error("Log() should return a valid entry even with nil context")
	}
}
