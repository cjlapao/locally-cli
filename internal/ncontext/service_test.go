package ncontext

import (
	"testing"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/stretchr/testify/assert"
)

// MockService is a mock implementation of the Service interface for testing
type MockService struct {
	name string
}

func (m *MockService) Name() string {
	return m.name
}

func (m *MockService) Initialize(ctx *appctx.AppContext, context *Context) *diagnostics.Diagnostics {
	return diagnostics.New("mock_initialize")
}

func (m *MockService) Load(ctx *appctx.AppContext, context *Context) *diagnostics.Diagnostics {
	return diagnostics.New("mock_load")
}

func (m *MockService) Save(ctx *appctx.AppContext, context *Context) *diagnostics.Diagnostics {
	return diagnostics.New("mock_save")
}

func (m *MockService) Validate(ctx *appctx.AppContext, context *Context) *diagnostics.Diagnostics {
	return diagnostics.New("mock_validate")
}

func TestNContext_Initialize(t *testing.T) {
	// Reset for testing
	ResetForTesting()

	// Test initialization
	config := Config{
		ConfigDirectory: "/test/config",
		AutoReload:      true,
		WatchInterval:   "30s",
	}

	svc := Initialize(config)
	assert.NotNil(t, svc)
	assert.Equal(t, "/test/config", svc.configDir)

	// Test singleton behavior
	svc2 := Initialize(config)
	assert.Equal(t, svc, svc2)
}

func TestNContext_GetInstance(t *testing.T) {
	// Reset for testing
	ResetForTesting()

	// Test GetInstance before initialization
	assert.Panics(t, func() {
		GetInstance()
	})

	// Initialize and test GetInstance
	config := Config{ConfigDirectory: "/test"}
	Initialize(config)

	svc := GetInstance()
	assert.NotNil(t, svc)
}

func TestNContext_RegisterService(t *testing.T) {
	// Reset for testing
	svc := Initialize(Config{ConfigDirectory: "/test"})
	ctx := appctx.NewContext(nil)

	// Create mock service
	mockService := &MockService{name: "test-service"}

	// Register service
	diag := svc.RegisterService(ctx, mockService)
	assert.False(t, diag.HasErrors())

	// Verify service is registered
	service, exists := svc.GetService(ctx, "test-service")
	assert.True(t, exists)
	assert.Equal(t, mockService, service)

	// Test registering the same service again (should give warning)
	diag = svc.RegisterService(ctx, mockService)
	assert.True(t, diag.HasWarnings())
}

func TestNContext_ListServices(t *testing.T) {
	// Reset for testing
	svc := Initialize(Config{ConfigDirectory: "/test"})
	ctx := appctx.NewContext(nil)

	// Register multiple services
	mockService1 := &MockService{name: "service1"}
	mockService2 := &MockService{name: "service2"}

	svc.RegisterService(ctx, mockService1)
	svc.RegisterService(ctx, mockService2)

	// List services
	services := svc.ListServices(ctx)
	assert.Len(t, services, 2)
	assert.Contains(t, services, "service1")
	assert.Contains(t, services, "service2")
}

func TestNContext_AddContext(t *testing.T) {
	// Reset for testing
	svc := Initialize(Config{ConfigDirectory: "/test"})
	ctx := appctx.NewContext(nil)

	// Create context
	context := &Context{
		ID:          "test-context-1",
		Name:        "Test Context",
		Description: "A test context",
		IsEnabled:   true,
		ConfigPath:  "/test/config.yml",
		Config:      make(map[string]interface{}),
		Services:    make(map[string]interface{}),
	}

	// Add context
	diag := svc.AddContext(ctx, context)
	assert.False(t, diag.HasErrors())

	// Verify context is added
	retrievedContext, exists := svc.GetContext(ctx, "test-context-1")
	assert.True(t, exists)
	assert.Equal(t, context, retrievedContext)

	// Test adding context with empty name
	invalidContext := &Context{
		ID:   "invalid-context",
		Name: "",
	}
	diag = svc.AddContext(ctx, invalidContext)
	assert.True(t, diag.HasErrors())

	// Test adding context with empty ID
	invalidContext2 := &Context{
		ID:   "",
		Name: "Invalid Context",
	}
	diag = svc.AddContext(ctx, invalidContext2)
	assert.True(t, diag.HasErrors())

	// Test adding duplicate context
	diag = svc.AddContext(ctx, context)
	assert.True(t, diag.HasErrors())
}

func TestNContext_ListContexts(t *testing.T) {
	// Reset for testing
	svc := Initialize(Config{ConfigDirectory: "/test"})
	ctx := appctx.NewContext(nil)

	// Add multiple contexts
	context1 := &Context{ID: "context1", Name: "Context 1"}
	context2 := &Context{ID: "context2", Name: "Context 2"}

	svc.AddContext(ctx, context1)
	svc.AddContext(ctx, context2)

	// List contexts
	contexts := svc.ListContexts(ctx)
	assert.Len(t, contexts, 2)
	assert.Contains(t, contexts, "context1")
	assert.Contains(t, contexts, "context2")
}

func TestNContext_GetDefaultContext(t *testing.T) {
	// Reset for testing
	svc := Initialize(Config{ConfigDirectory: "/test"})
	ctx := appctx.NewContext(nil)

	// Add contexts
	context1 := &Context{ID: "context1", Name: "Context 1", IsDefault: false}
	context2 := &Context{ID: "context2", Name: "Context 2", IsDefault: true}

	svc.AddContext(ctx, context1)
	svc.AddContext(ctx, context2)

	// Get default context
	defaultContext, exists := svc.GetDefaultContext(ctx)
	assert.True(t, exists)
	assert.Equal(t, "context2", defaultContext.ID)
}

func TestNContext_SetDefaultContext(t *testing.T) {
	// Reset for testing
	svc := Initialize(Config{ConfigDirectory: "/test"})
	ctx := appctx.NewContext(nil)

	// Add contexts
	context1 := &Context{ID: "context1", Name: "Context 1", IsDefault: true}
	context2 := &Context{ID: "context2", Name: "Context 2", IsDefault: false}

	svc.AddContext(ctx, context1)
	svc.AddContext(ctx, context2)

	// Set new default
	diag := svc.SetDefaultContext(ctx, "context2")
	assert.False(t, diag.HasErrors())

	// Verify default changed
	defaultContext, exists := svc.GetDefaultContext(ctx)
	assert.True(t, exists)
	assert.Equal(t, "context2", defaultContext.ID)

	// Test setting non-existent context as default
	diag = svc.SetDefaultContext(ctx, "non-existent")
	assert.True(t, diag.HasErrors())
}

func TestNContext_Load(t *testing.T) {
	// Reset for testing
	svc := Initialize(Config{ConfigDirectory: "/test"})
	ctx := appctx.NewContext(nil)

	// Test load
	diag := svc.Load(ctx)
	assert.False(t, diag.HasErrors())
	assert.True(t, svc.IsLoaded())
}

func TestNContext_Save(t *testing.T) {
	// Reset for testing
	svc := Initialize(Config{ConfigDirectory: "/test"})
	ctx := appctx.NewContext(nil)

	// Add some contexts
	context := &Context{ID: "test-context", Name: "Test Context"}
	svc.AddContext(ctx, context)

	// Test save
	diag := svc.Save(ctx)
	assert.False(t, diag.HasErrors())
}

func TestNContext_GetStatus(t *testing.T) {
	// Reset for testing
	svc := Initialize(Config{ConfigDirectory: "/test"})
	ctx := appctx.NewContext(nil)

	// Add some data
	mockService := &MockService{name: "test-service"}
	context := &Context{ID: "test-context", Name: "Test Context"}

	svc.RegisterService(ctx, mockService)
	svc.AddContext(ctx, context)
	svc.Load(ctx)

	// Get status
	status := svc.GetStatus(ctx)

	// Verify status
	assert.True(t, status["loaded"].(bool))
	assert.Equal(t, "/test", status["config_dir"].(string))

	contexts := status["contexts"].([]string)
	assert.Len(t, contexts, 1)
	assert.Contains(t, contexts, "test-context")

	services := status["services"].([]string)
	assert.Len(t, services, 1)
	assert.Contains(t, contexts, "test-service")
}
