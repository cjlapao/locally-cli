package ncontext

import (
	"sync"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
)

var (
	globalNContext *NContext
	once           sync.Once
)

// NContext represents the new context service that organizes services and loads configuration
type NContext struct {
	mu        sync.RWMutex
	contexts  map[string]*Context
	services  map[string]Service
	isLoaded  bool
	configDir string
}

// Context represents a single context with its configuration and services
type Context struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	IsDefault   bool                   `json:"is_default"`
	IsEnabled   bool                   `json:"is_enabled"`
	ConfigPath  string                 `json:"config_path"`
	Config      map[string]interface{} `json:"config"`
	Services    map[string]interface{} `json:"services"`
	CreatedAt   string                 `json:"created_at"`
	UpdatedAt   string                 `json:"updated_at"`
}

// Service represents a service interface that can be registered with the context
type Service interface {
	Name() string
	Initialize(ctx *appctx.AppContext, context *Context) *diagnostics.Diagnostics
	Load(ctx *appctx.AppContext, context *Context) *diagnostics.Diagnostics
	Save(ctx *appctx.AppContext, context *Context) *diagnostics.Diagnostics
	Validate(ctx *appctx.AppContext, context *Context) *diagnostics.Diagnostics
}

// Config represents the configuration for the NContext service
type Config struct {
	ConfigDirectory string
	AutoReload      bool
	WatchInterval   string
}

// new creates a new NContext instance
func new(config Config) *NContext {
	svc := &NContext{
		contexts:  make(map[string]*Context),
		services:  make(map[string]Service),
		isLoaded:  false,
		configDir: config.ConfigDirectory,
	}
	return svc
}

// Initialize initializes the NContext service singleton
func Initialize(config Config) *NContext {
	once.Do(func() {
		globalNContext = new(config)
	})
	return globalNContext
}

// GetInstance returns the global singleton NContext instance
func GetInstance() *NContext {
	if globalNContext == nil {
		panic("ncontext not initialized")
	}
	return globalNContext
}

// ResetForTesting resets the ncontext for testing purposes
// This should only be used in tests
func ResetForTesting() {
	globalNContext = new(Config{})
}

// RegisterService registers a service with the NContext
func (nc *NContext) RegisterService(ctx *appctx.AppContext, service Service) *diagnostics.Diagnostics {
	diag := diagnostics.New("register_service")
	defer diag.Complete()

	diag.AddPathEntry("start", "ncontext", map[string]interface{}{
		"service_name": service.Name(),
	})

	nc.mu.Lock()
	defer nc.mu.Unlock()

	// Check if service already exists
	if _, exists := nc.services[service.Name()]; exists {
		diag.AddWarning("SERVICE_ALREADY_EXISTS", "Service already registered", "ncontext", map[string]interface{}{
			"service_name": service.Name(),
		})
		return diag
	}

	// Register the service
	nc.services[service.Name()] = service

	diag.AddPathEntry("service_registered", "ncontext", map[string]interface{}{
		"service_name":   service.Name(),
		"total_services": len(nc.services),
	})

	logging.Infof("Registered service: %s", service.Name())
	return diag
}

// GetService returns a service by name
func (nc *NContext) GetService(ctx *appctx.AppContext, name string) (Service, bool) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	service, exists := nc.services[name]
	return service, exists
}

// ListServices returns a list of all registered service names
func (nc *NContext) ListServices(ctx *appctx.AppContext) []string {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	services := make([]string, 0, len(nc.services))
	for name := range nc.services {
		services = append(services, name)
	}
	return services
}

// AddContext adds a new context
func (nc *NContext) AddContext(ctx *appctx.AppContext, context *Context) *diagnostics.Diagnostics {
	diag := diagnostics.New("add_context")
	defer diag.Complete()

	diag.AddPathEntry("start", "ncontext", map[string]interface{}{
		"context_name": context.Name,
		"context_id":   context.ID,
	})

	nc.mu.Lock()
	defer nc.mu.Unlock()

	// Validate context
	if context.Name == "" {
		diag.AddError("INVALID_CONTEXT_NAME", "Context name cannot be empty", "ncontext", nil)
		return diag
	}

	if context.ID == "" {
		diag.AddError("INVALID_CONTEXT_ID", "Context ID cannot be empty", "ncontext", nil)
		return diag
	}

	// Check if context already exists
	if _, exists := nc.contexts[context.ID]; exists {
		diag.AddError("CONTEXT_ALREADY_EXISTS", "Context already exists", "ncontext", map[string]interface{}{
			"context_id": context.ID,
		})
		return diag
	}

	// Add the context
	nc.contexts[context.ID] = context

	diag.AddPathEntry("context_added", "ncontext", map[string]interface{}{
		"context_name":   context.Name,
		"context_id":     context.ID,
		"total_contexts": len(nc.contexts),
	})

	logging.Infof("Added context: %s (%s)", context.Name, context.ID)
	return diag
}

// GetContext returns a context by ID
func (nc *NContext) GetContext(ctx *appctx.AppContext, id string) (*Context, bool) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	context, exists := nc.contexts[id]
	return context, exists
}

// ListContexts returns a list of all context IDs
func (nc *NContext) ListContexts(ctx *appctx.AppContext) []string {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	contexts := make([]string, 0, len(nc.contexts))
	for id := range nc.contexts {
		contexts = append(contexts, id)
	}
	return contexts
}

// GetDefaultContext returns the default context
func (nc *NContext) GetDefaultContext(ctx *appctx.AppContext) (*Context, bool) {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	for _, context := range nc.contexts {
		if context.IsDefault {
			return context, true
		}
	}
	return nil, false
}

// SetDefaultContext sets a context as the default
func (nc *NContext) SetDefaultContext(ctx *appctx.AppContext, id string) *diagnostics.Diagnostics {
	diag := diagnostics.New("set_default_context")
	defer diag.Complete()

	diag.AddPathEntry("start", "ncontext", map[string]interface{}{
		"context_id": id,
	})

	nc.mu.Lock()
	defer nc.mu.Unlock()

	// Check if context exists
	if _, exists := nc.contexts[id]; !exists {
		diag.AddError("CONTEXT_NOT_FOUND", "Context not found", "ncontext", map[string]interface{}{
			"context_id": id,
		})
		return diag
	}

	// Clear existing default
	for _, context := range nc.contexts {
		context.IsDefault = false
	}

	// Set new default
	nc.contexts[id].IsDefault = true

	diag.AddPathEntry("default_context_set", "ncontext", map[string]interface{}{
		"context_id": id,
	})

	logging.Infof("Set default context: %s", id)
	return diag
}

// Load loads all contexts and their configurations
func (nc *NContext) Load(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("load_ncontext")
	defer diag.Complete()

	diag.AddPathEntry("start", "ncontext", map[string]interface{}{
		"config_dir": nc.configDir,
	})

	nc.mu.Lock()
	defer nc.mu.Unlock()

	// TODO: Implement loading from files
	// For now, just mark as loaded
	nc.isLoaded = true

	diag.AddPathEntry("load_completed", "ncontext", map[string]interface{}{
		"contexts_count": len(nc.contexts),
		"services_count": len(nc.services),
	})

	logging.Info("NContext loaded successfully")
	return diag
}

// Save saves all contexts and their configurations
func (nc *NContext) Save(ctx *appctx.AppContext) *diagnostics.Diagnostics {
	diag := diagnostics.New("save_ncontext")
	defer diag.Complete()

	diag.AddPathEntry("start", "ncontext", map[string]interface{}{
		"contexts_count": len(nc.contexts),
	})

	nc.mu.RLock()
	defer nc.mu.RUnlock()

	// TODO: Implement saving to files
	// For now, just log the operation

	diag.AddPathEntry("save_completed", "ncontext", map[string]interface{}{
		"contexts_count": len(nc.contexts),
	})

	logging.Info("NContext saved successfully")
	return diag
}

// GetStatus returns the current status of the NContext
func (nc *NContext) GetStatus(ctx *appctx.AppContext) map[string]interface{} {
	nc.mu.RLock()
	defer nc.mu.RUnlock()

	status := map[string]interface{}{
		"loaded":     nc.isLoaded,
		"contexts":   make([]string, 0),
		"services":   make([]string, 0),
		"config_dir": nc.configDir,
	}

	// Add context IDs
	for id := range nc.contexts {
		status["contexts"] = append(status["contexts"].([]string), id)
	}

	// Add service names
	for name := range nc.services {
		status["services"] = append(status["services"].([]string), name)
	}

	return status
}

// IsLoaded returns whether the NContext has been loaded
func (nc *NContext) IsLoaded() bool {
	nc.mu.RLock()
	defer nc.mu.RUnlock()
	return nc.isLoaded
}
