# Executer Service

The Executer service provides a reliable wrapper for executing commands with enhanced debugging and diagnostics capabilities. It replaces the conventional error-based approach with a more robust diagnostics system and integrates seamlessly with the application's context system.

## Features

- **AppContext Integration**: All functions accept `*appctx.AppContext` for enhanced debugging and logging
- **Diagnostics System**: Returns `*diagnostics.Diagnostics` instead of conventional errors for better error handling
- **Configurable Execution**: Flexible configuration options for timeout, retry logic, working directory, environment variables, and more
- **Retry Logic**: Built-in retry mechanism with configurable retry count and delay
- **Exit Code Validation**: Optional validation of command exit codes
- **Output Control**: Configurable output capture and display options
- **Command Validation**: Built-in command existence validation
- **Comprehensive Logging**: Detailed logging with structured fields for debugging
- **Backward Compatibility**: Legacy functions maintained for existing code

## Core Types

### ExecuteOutput

Represents the output of a command execution with comprehensive metadata:

```go
type ExecuteOutput struct {
    StdOut      string    `json:"stdout"`
    StdErr      string    `json:"stderr"`
    ErrorCode   string    `json:"error_code,omitempty"`
    ExitCode    int       `json:"exit_code"`
    Duration    string    `json:"duration"`
    Command     string    `json:"command"`
    Args        []string  `json:"args"`
    StartTime   time.Time `json:"start_time"`
    EndTime     time.Time `json:"end_time"`
    Success     bool      `json:"success"`
}
```

### ExecuteConfig

Configuration for command execution:

```go
type ExecuteConfig struct {
    Timeout           time.Duration
    WorkingDirectory  string
    Environment       []string
    CaptureOutput     bool
    ShowOutput        bool
    RetryCount        int
    RetryDelay        time.Duration
    ValidateExitCode  bool
    ExpectedExitCodes []int
}
```

## Usage Examples

### Basic Command Execution

```go
ctx := appctx.NewContext(nil)
output, diag := executer.ExecuteSimple(ctx, "echo", "hello world")

if diag.HasErrors() {
    ctx.LogError("Command failed: " + diag.GetSummary())
    return
}

ctx.LogInfo("Command output: " + output.StdOut)
```

### Command with Custom Configuration

```go
ctx := appctx.NewContext(nil)
config := executer.ExecuteConfig{
    Timeout:           10 * time.Second,
    WorkingDirectory:  "/tmp",
    Environment:       []string{"CUSTOM_VAR=value"},
    RetryCount:        3,
    RetryDelay:        1 * time.Second,
    ValidateExitCode:  true,
    ExpectedExitCodes: []int{0, 1},
}

output, diag := executer.Execute(ctx, config, "my-script.sh", "arg1", "arg2")
```

### Command with Timeout

```go
ctx := appctx.NewContext(nil)
output, diag := executer.ExecuteWithTimeout(ctx, 5*time.Second, "long-running-command")
```

### Command with Retry Logic

```go
ctx := appctx.NewContext(nil)
output, diag := executer.ExecuteWithRetry(ctx, 3, 2*time.Second, "unreliable-command")
```

### Command in Specific Directory

```go
ctx := appctx.NewContext(nil)
output, diag := executer.ExecuteInDirectory(ctx, "/path/to/working/dir", "git", "status")
```

### Command with Custom Environment

```go
ctx := appctx.NewContext(nil)
env := []string{"DEBUG=true", "LOG_LEVEL=verbose"}
output, diag := executer.ExecuteWithEnvironment(ctx, env, "my-app")
```

### Silent Command Execution

```go
ctx := appctx.NewContext(nil)
output, diag := executer.ExecuteSilent(ctx, "secret-command")
// Output is captured but not displayed
```

### Command with Real-time Output

```go
ctx := appctx.NewContext(nil)
output, diag := executer.ExecuteAndWatch(ctx, "build-script.sh")
// Output is displayed in real-time and also captured
```

### Command Validation

```go
ctx := appctx.NewContext(nil)
diag := executer.ValidateCommand(ctx, "docker")
if diag.HasErrors() {
    ctx.LogError("Docker is not available")
    return
}
```

## Configuration Options

### Default Configuration

```go
config := executer.DefaultConfig()
// Returns:
// - Timeout: 30 seconds
// - WorkingDirectory: ""
// - Environment: nil
// - CaptureOutput: true
// - ShowOutput: false
// - RetryCount: 0
// - RetryDelay: 1 second
// - ValidateExitCode: false
// - ExpectedExitCodes: [0]
```

### Custom Configuration Examples

#### High-Reliability Configuration

```go
config := executer.ExecuteConfig{
    Timeout:           60 * time.Second,
    RetryCount:        5,
    RetryDelay:        2 * time.Second,
    ValidateExitCode:  true,
    ExpectedExitCodes: []int{0},
    CaptureOutput:     true,
    ShowOutput:        false,
}
```

#### Development Configuration

```go
config := executer.ExecuteConfig{
    Timeout:       10 * time.Second,
    ShowOutput:    true,
    CaptureOutput: true,
    RetryCount:    0,
}
```

#### Production Configuration

```go
config := executer.ExecuteConfig{
    Timeout:           30 * time.Second,
    RetryCount:        3,
    RetryDelay:        5 * time.Second,
    ValidateExitCode:  true,
    ExpectedExitCodes: []int{0},
    CaptureOutput:     true,
    ShowOutput:        false,
}
```

## Error Handling

The executer uses the diagnostics system for comprehensive error handling:

```go
ctx := appctx.NewContext(nil)
output, diag := executer.ExecuteSimple(ctx, "failing-command")

if diag.HasErrors() {
    // Log all errors
    for _, err := range diag.GetErrors() {
        ctx.LogErrorf("Error: %s - %s", err.Code, err.Message)
    }
    
    // Log all warnings
    for _, warn := range diag.GetWarnings() {
        ctx.LogWarnf("Warning: %s - %s", warn.Code, warn.Message)
    }
    
    // Get summary
    ctx.LogError("Summary: " + diag.GetSummary())
    return
}
```

## Logging and Debugging

The executer provides comprehensive logging with structured fields:

```go
ctx := appctx.NewContext(nil)
output, diag := executer.ExecuteSimple(ctx, "my-command")

// Logs include:
// - Command and arguments
// - Execution duration
// - Exit code
// - Success/failure status
// - Retry attempts (if applicable)
// - Working directory (if specified)
// - Environment variables count (if specified)
```

## Legacy Functions

For backward compatibility, the following legacy functions are available (deprecated):

- `ExecuteLegacy(command string, args ...string) (ExecuteOutput, error)`
- `ExecuteWithNoOutput(command string, args ...string) (ExecuteOutput, error)`
- `ExecuteWithNoOutputContext(ctx context.Context, command string, args ...string) (ExecuteOutput, error)`
- `ExecuteAndWatchLegacy(command string, args ...string) (ExecuteOutput, error)`

**Note**: These functions are deprecated and should be replaced with the new AppContext-based functions.

## Best Practices

1. **Always use AppContext**: Pass a proper AppContext for better debugging and logging
2. **Handle Diagnostics**: Always check `diag.HasErrors()` before proceeding
3. **Configure Timeouts**: Set appropriate timeouts for your use case
4. **Use Retry Logic**: For unreliable commands, configure retry logic
5. **Validate Commands**: Use `ValidateCommand` to check command availability
6. **Validate Exit Codes**: For critical commands, validate expected exit codes
7. **Capture Output**: Always capture output for debugging purposes
8. **Use Structured Logging**: Leverage the built-in structured logging for better debugging

## Migration from Legacy Functions

### Before (Legacy)
```go
output, err := executer.Execute("echo", "hello")
if err != nil {
    log.Printf("Error: %v", err)
    return
}
```

### After (New)
```go
ctx := appctx.NewContext(nil)
output, diag := executer.ExecuteSimple(ctx, "echo", "hello")
if diag.HasErrors() {
    ctx.LogError("Command failed: " + diag.GetSummary())
    return
}
```

## Testing

The package includes comprehensive tests covering:

- Basic command execution
- Timeout handling
- Retry logic
- Exit code validation
- Environment variable handling
- Working directory configuration
- Output capture and display
- Command validation
- Legacy function compatibility

Run tests with:
```bash
go test ./internal/services/executer -v
``` 