# Diagnostics Service

The Diagnostics Service provides comprehensive error tracking and execution path monitoring for complex nested operations. It allows you to track the execution flow, bubble up errors with full context, and debug issues more effectively.

## Features

- **Execution Path Tracking**: Track every step of complex operations with timestamps and metadata
- **Error Bubbling**: Errors bubble up through the call stack with full context and path information
- **Child Operations**: Support for nested operations with parent-child relationships
- **Concurrent Safety**: Thread-safe operations for concurrent access
- **Integration**: Seamless integration with existing notification system
- **Performance Monitoring**: Track operation durations and performance metrics
- **Debugging Support**: Rich debugging information with stack traces and metadata

## Core Concepts

### DiagnosticContext

A `DiagnosticContext` represents a single operation and tracks:
- Execution path with timestamps
- Errors and warnings with severity levels
- Metadata for debugging
- Parent-child relationships
- Performance metrics

### PathEntry

Each step in the execution path is recorded as a `PathEntry` containing:
- Operation name and component
- Timestamp and duration
- File name, line number, and function name
- Optional metadata

### DiagnosticError

Errors are tracked with:
- Error code and message
- Full execution path at time of error
- Severity level (low, medium, high, critical)
- Stack trace information
- Component and operation context

## Basic Usage

### Starting a Diagnostic Context

```go
import "github.com/cjlapao/locally-cli/internal/diagnostics"

// Start a new diagnostic context
svc := diagnostics.GetInstance()
ctx := svc.StartOperation("my-operation")
defer ctx.Complete()

// Add path entries
ctx.AddPathEntry("step1", "my-component", map[string]interface{}{
    "description": "First step",
    "data": "some data",
})

// Add errors
ctx.AddError("VALIDATION_FAILED", "Invalid input", "my-component", diagnostics.SeverityHigh)

// Add warnings
ctx.AddWarning("PERFORMANCE_ISSUE", "Operation is slow", "my-component")
```

### Using Convenience Functions

```go
// Execute a function with automatic diagnostic tracking
result, err := diagnostics.WithDiagnostics("my-operation", func(ctx *diagnostics.DiagnosticContext) (string, error) {
    ctx.AddPathEntry("processing", "my-component")
    // ... your logic here
    return "result", nil
})

// Execute a child operation
parent := svc.StartOperation("parent")
result, err := diagnostics.WithChildDiagnostics(parent, "child-operation", func(ctx *diagnostics.DiagnosticContext) (string, error) {
    // ... child logic
    return "child result", nil
})
```

### Using Diagnostic Results

```go
// Get a result with diagnostic information
result := diagnostics.WithDiagnosticResult("my-operation", func(ctx *diagnostics.DiagnosticContext) (string, error) {
    // ... your logic
    return "data", nil
})

if result.HasErrors {
    fmt.Printf("Operation failed with %d errors\n", result.Context.GetErrorCount())
}

if result.HasWarnings {
    fmt.Printf("Operation completed with %d warnings\n", result.Context.GetWarningCount())
}
```

## Integration with Existing Systems

### Dependency Tree Integration

The diagnostics service can be integrated with the dependency tree system to provide detailed error tracking:

```go
// Enhanced dependency tree building with diagnostics
func BuildDependencyTreeWithDiagnostics[T interfaces.LocallyService](values []T) ([]T, error) {
    return diagnostics.WithDiagnostics("build_dependency_tree", func(ctx *diagnostics.DiagnosticContext) ([]T, error) {
        ctx.AddPathEntry("start", "dependency_tree", map[string]interface{}{
            "service_count": len(values),
        })
        
        // Detect cycles
        if err := detectCyclesWithDiagnostics(ctx, values); err != nil {
            ctx.AddError(diagnostics.ErrorCodeDependencyError, "Cycle detection failed", "dependency_tree", diagnostics.SeverityCritical)
            return nil, err
        }
        
        // Perform topological sort
        result, err := performTopologicalSortWithDiagnostics(ctx, values)
        if err != nil {
            ctx.AddError(diagnostics.ErrorCodeDependencyError, "Topological sort failed", "dependency_tree", diagnostics.SeverityCritical)
            return nil, err
        }
        
        return result, nil
    })
}
```

### HTTP Handler Integration

```go
// Wrap HTTP handlers with diagnostic tracking
func (h *MyHandler) HandleRequest(w http.ResponseWriter, r *http.Request) {
    middleware := diagnostics.NewDiagnosticMiddleware()
    
    err := middleware.WrapHandler("http_request", func(ctx *diagnostics.DiagnosticContext) error {
        ctx.AddPathEntry("parse_request", "http_handler")
        
        // Parse request body
        var req MyRequest
        if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
            ctx.AddError(diagnostics.ErrorCodeValidationFailed, "Invalid request body", "http_handler", diagnostics.SeverityMedium)
            return err
        }
        
        ctx.AddPathEntry("process_request", "http_handler")
        
        // Process the request
        result, err := h.processRequest(ctx, req)
        if err != nil {
            ctx.AddError(diagnostics.ErrorCodeExecutionError, "Request processing failed", "http_handler", diagnostics.SeverityHigh)
            return err
        }
        
        // Return response
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(result)
        
        return nil
    })
    
    if err != nil {
        http.Error(w, "Internal server error", http.StatusInternalServerError)
    }
}
```

## Error Codes

The diagnostics service provides predefined error codes for common scenarios:

### Error Codes
- `VALIDATION_FAILED`: Input validation errors
- `DEPENDENCY_ERROR`: Dependency-related errors
- `SERVICE_UNAVAILABLE`: Service availability issues
- `CONFIGURATION_ERROR`: Configuration problems
- `AUTHENTICATION_ERROR`: Authentication failures
- `AUTHORIZATION_ERROR`: Authorization failures
- `RESOURCE_NOT_FOUND`: Missing resources
- `TIMEOUT_ERROR`: Operation timeouts
- `NETWORK_ERROR`: Network-related issues
- `DATABASE_ERROR`: Database operation failures
- `FILESYSTEM_ERROR`: File system operations
- `EXECUTION_ERROR`: General execution errors

### Warning Codes
- `DEPRECATED_FEATURE`: Use of deprecated features
- `PERFORMANCE_ISSUE`: Performance concerns
- `RESOURCE_USAGE`: Resource usage warnings
- `CONFIGURATION_WARNING`: Configuration warnings
- `SECURITY_WARNING`: Security-related warnings

## Debugging and Monitoring

### Getting Diagnostic Information

```go
// Get a summary of the diagnostic context
summary := ctx.GetSummary()
fmt.Println(summary)

// Get the full execution path
path := ctx.GetFullPath()
fmt.Println(path)

// Get specific errors
errors := ctx.GetErrors()
for _, err := range errors {
    fmt.Printf("Error [%s]: %s\n", err.Code, err.Message)
    fmt.Printf("  Path: %s\n", err.StackTrace)
}

// Get specific warnings
warnings := ctx.GetWarnings()
for _, warning := range warnings {
    fmt.Printf("Warning [%s]: %s\n", warning.Code, warning.Message)
}
```

### Global Diagnostic Summary

```go
// Get a summary of all diagnostic contexts
summary := diagnostics.GetDiagnosticSummary()
fmt.Printf("Total Operations: %d\n", summary.TotalOperations)
fmt.Printf("Successful: %d\n", summary.SuccessfulOperations)
fmt.Printf("Failed: %d\n", summary.FailedOperations)
fmt.Printf("Total Errors: %d\n", summary.TotalErrors)
fmt.Printf("Total Warnings: %d\n", summary.TotalWarnings)

// Print formatted summary
formatted := diagnostics.PrintDiagnosticSummary()
fmt.Println(formatted)
```

### Diagnostic Logger

```go
// Create a diagnostic logger
logger := diagnostics.NewDiagnosticLogger(ctx)

// Use the logger for various operations
logger.Info("Starting operation")
logger.Warning("This is a warning")
logger.Error("ERROR_CODE", "This is an error")
logger.Debug("Debug information")
```

## Best Practices

### 1. Always Complete Contexts

```go
ctx := svc.StartOperation("my-operation")
defer ctx.Complete() // Always complete the context
```

### 2. Use Descriptive Operation Names

```go
// Good
ctx := svc.StartOperation("build_dependency_tree")

// Bad
ctx := svc.StartOperation("op")
```

### 3. Add Meaningful Metadata

```go
ctx.AddPathEntry("process_service", "service_processor", map[string]interface{}{
    "service_name": service.GetName(),
    "service_type": service.GetType(),
    "dependencies_count": len(service.GetDependencies()),
})
```

### 4. Use Appropriate Error Severity

```go
// Critical: System cannot continue
ctx.AddError("CRITICAL_ERROR", "Database connection lost", "db", diagnostics.SeverityCritical)

// High: Operation failed but system can continue
ctx.AddError("VALIDATION_ERROR", "Invalid input", "validator", diagnostics.SeverityHigh)

// Medium: Warning that might indicate a problem
ctx.AddError("PERFORMANCE_WARNING", "Slow operation", "processor", diagnostics.SeverityMedium)

// Low: Minor issue
ctx.AddError("DEPRECATION_WARNING", "Using deprecated API", "api", diagnostics.SeverityLow)
```

### 5. Leverage Child Operations for Complex Workflows

```go
parent := svc.StartOperation("complex_workflow")
defer parent.Complete()

// Process each step as a child operation
for i, step := range steps {
    child := svc.StartChildOperation(parent, fmt.Sprintf("step_%d", i))
    defer child.Complete()
    
    // Process the step
    if err := processStep(child, step); err != nil {
        child.AddError("STEP_FAILED", "Step processing failed", "workflow", diagnostics.SeverityHigh)
        return err
    }
}
```

## Performance Considerations

- Diagnostic contexts are lightweight and designed for minimal overhead
- Path entries are stored in memory and should be cleaned up when no longer needed
- Use `ClearAllDiagnostics()` for testing or when you need to reset the system
- Consider the level of detail needed - too many path entries can impact performance

## Thread Safety

All diagnostic operations are thread-safe and can be used in concurrent environments. The service uses read-write mutexes to ensure data consistency while maintaining good performance for read operations.

## Integration with Existing Error Handling

The diagnostics service integrates seamlessly with existing error handling patterns:

```go
// Traditional error handling
if err != nil {
    return fmt.Errorf("operation failed: %w", err)
}

// With diagnostics
if err != nil {
    ctx.AddError("OPERATION_FAILED", err.Error(), "my-component", diagnostics.SeverityHigh)
    return fmt.Errorf("operation failed: %w", err)
}
```

This approach provides both the traditional error handling and rich diagnostic information for debugging and monitoring. 