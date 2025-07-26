// Package executer provides a reliable wrapper for executing commands with enhanced debugging and diagnostics
package executer

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
)

// ExecuteOutput represents the output of a command execution
type ExecuteOutput struct {
	StdOut    string    `json:"stdout"`
	StdErr    string    `json:"stderr"`
	ErrorCode string    `json:"error_code,omitempty"`
	ExitCode  int       `json:"exit_code"`
	Duration  string    `json:"duration"`
	Command   string    `json:"command"`
	Args      []string  `json:"args"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Success   bool      `json:"success"`
}

// GetAllOutput returns combined stdout and stderr output
func (exe ExecuteOutput) GetAllOutput() string {
	output := ""
	if exe.StdOut != "" {
		output += exe.StdOut
	}
	if exe.StdErr != "" {
		if output != "" {
			output += "\n"
		}
		output += exe.StdErr
	}
	return output
}

// GetFormattedCommand returns the command as a formatted string
func (exe ExecuteOutput) GetFormattedCommand() string {
	if len(exe.Args) == 0 {
		return exe.Command
	}
	return fmt.Sprintf("%s %s", exe.Command, strings.Join(exe.Args, " "))
}

// ExecuteConfig represents configuration for command execution
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

// DefaultConfig returns a default execution configuration
func DefaultConfig() ExecuteConfig {
	return ExecuteConfig{
		Timeout:           30 * time.Second,
		WorkingDirectory:  "",
		Environment:       nil,
		CaptureOutput:     true,
		ShowOutput:        false,
		RetryCount:        0,
		RetryDelay:        1 * time.Second,
		ValidateExitCode:  false,
		ExpectedExitCodes: []int{0},
	}
}

// Execute executes a command with the given configuration
func Execute(ctx *appctx.AppContext, config ExecuteConfig, command string, args ...string) (ExecuteOutput, *diagnostics.Diagnostics) {
	diag := diagnostics.New("execute_command")
	defer diag.Complete()

	startTime := time.Now()

	diag.AddPathEntry("start", "executer", map[string]interface{}{
		"command":     command,
		"args":        args,
		"timeout":     config.Timeout.String(),
		"retry_count": config.RetryCount,
	})

	ctx.LogWithFields(map[string]interface{}{
		"command":     command,
		"args":        args,
		"timeout":     config.Timeout.String(),
		"retry_count": config.RetryCount,
	}).Info("Executing command")

	var lastOutput ExecuteOutput
	var lastError error

	// Retry logic
	for attempt := 0; attempt <= config.RetryCount; attempt++ {
		if attempt > 0 {
			ctx.LogWithFields(map[string]interface{}{
				"attempt": attempt,
				"command": command,
			}).Info("Retrying command execution")

			diag.AddPathEntry("retry_attempt", "executer", map[string]interface{}{
				"attempt": attempt,
				"command": command,
			})

			// Wait before retry
			time.Sleep(config.RetryDelay)
		}

		// Create context with timeout
		execCtx, cancel := context.WithTimeout(ctx, config.Timeout)
		defer cancel()

		// Create command
		cmd := exec.CommandContext(execCtx, command, args...)

		// Set working directory if specified
		if config.WorkingDirectory != "" {
			cmd.Dir = config.WorkingDirectory
			ctx.LogWithField("working_directory", config.WorkingDirectory).Debug("Set working directory")
		}

		// Set environment variables if specified
		if config.Environment != nil {
			cmd.Env = append(os.Environ(), config.Environment...)
			ctx.LogWithField("env_count", len(config.Environment)).Debug("Set environment variables")
		}

		// Prepare output buffers
		var stdout, stderr bytes.Buffer
		var writers []io.Writer

		// Configure output capture
		if config.CaptureOutput {
			writers = append(writers, &stdout)
		}
		if config.ShowOutput {
			writers = append(writers, os.Stdout)
		}

		if len(writers) > 0 {
			cmd.Stdout = io.MultiWriter(writers...)
		}

		// Configure error output
		var errWriters []io.Writer
		if config.CaptureOutput {
			errWriters = append(errWriters, &stderr)
		}
		if config.ShowOutput {
			errWriters = append(errWriters, os.Stderr)
		}

		if len(errWriters) > 0 {
			cmd.Stderr = io.MultiWriter(errWriters...)
		}

		// Execute command
		err := cmd.Run()
		endTime := time.Now()
		duration := endTime.Sub(startTime)

		// Prepare output
		output := ExecuteOutput{
			StdOut:    stdout.String(),
			StdErr:    stderr.String(),
			ExitCode:  cmd.ProcessState.ExitCode(),
			Duration:  duration.String(),
			Command:   command,
			Args:      args,
			StartTime: startTime,
			EndTime:   endTime,
			Success:   err == nil,
		}

		if err != nil {
			output.ErrorCode = err.Error()
			lastError = err
			lastOutput = output

			ctx.LogWithFields(map[string]interface{}{
				"command":   command,
				"exit_code": output.ExitCode,
				"error":     err.Error(),
				"duration":  duration.String(),
			}).Error("Command execution failed")

			diag.AddPathEntry("command_failed", "executer", map[string]interface{}{
				"command":   command,
				"exit_code": output.ExitCode,
				"error":     err.Error(),
				"attempt":   attempt,
			})

			// Check if we should retry
			if attempt < config.RetryCount {
				continue
			}
		} else {
			// Command succeeded
			ctx.LogWithFields(map[string]interface{}{
				"command":   command,
				"exit_code": output.ExitCode,
				"duration":  duration.String(),
			}).Info("Command executed successfully")

			diag.AddPathEntry("command_succeeded", "executer", map[string]interface{}{
				"command":   command,
				"exit_code": output.ExitCode,
				"duration":  duration.String(),
			})

			// Validate exit code if required
			if config.ValidateExitCode {
				if !isValidExitCode(output.ExitCode, config.ExpectedExitCodes) {
					diag.AddError("INVALID_EXIT_CODE", "Command returned unexpected exit code", "executer", map[string]interface{}{
						"command":        command,
						"exit_code":      output.ExitCode,
						"expected_codes": config.ExpectedExitCodes,
					})
					return output, diag
				}
			}

			return output, diag
		}
	}

	// All retries failed
	diag.AddError("COMMAND_EXECUTION_FAILED", "Command execution failed after all retries", "executer", map[string]interface{}{
		"command":     command,
		"retry_count": config.RetryCount,
		"last_error":  lastError.Error(),
	})

	return lastOutput, diag
}

// ExecuteSimple executes a command with default configuration
func ExecuteSimple(ctx *appctx.AppContext, command string, args ...string) (ExecuteOutput, *diagnostics.Diagnostics) {
	config := DefaultConfig()
	return Execute(ctx, config, command, args...)
}

// ExecuteWithTimeout executes a command with a specific timeout
func ExecuteWithTimeout(ctx *appctx.AppContext, timeout time.Duration, command string, args ...string) (ExecuteOutput, *diagnostics.Diagnostics) {
	config := DefaultConfig()
	config.Timeout = timeout
	return Execute(ctx, config, command, args...)
}

// ExecuteWithRetry executes a command with retry logic
func ExecuteWithRetry(ctx *appctx.AppContext, retryCount int, retryDelay time.Duration, command string, args ...string) (ExecuteOutput, *diagnostics.Diagnostics) {
	config := DefaultConfig()
	config.RetryCount = retryCount
	config.RetryDelay = retryDelay
	return Execute(ctx, config, command, args...)
}

// ExecuteInDirectory executes a command in a specific working directory
func ExecuteInDirectory(ctx *appctx.AppContext, workingDirectory string, command string, args ...string) (ExecuteOutput, *diagnostics.Diagnostics) {
	config := DefaultConfig()
	config.WorkingDirectory = workingDirectory
	return Execute(ctx, config, command, args...)
}

// ExecuteWithEnvironment executes a command with custom environment variables
func ExecuteWithEnvironment(ctx *appctx.AppContext, environment []string, command string, args ...string) (ExecuteOutput, *diagnostics.Diagnostics) {
	config := DefaultConfig()
	config.Environment = environment
	return Execute(ctx, config, command, args...)
}

// ExecuteAndWatch executes a command and shows real-time output
func ExecuteAndWatch(ctx *appctx.AppContext, command string, args ...string) (ExecuteOutput, *diagnostics.Diagnostics) {
	config := DefaultConfig()
	config.ShowOutput = true
	config.CaptureOutput = true
	return Execute(ctx, config, command, args...)
}

// ExecuteSilent executes a command without showing output
func ExecuteSilent(ctx *appctx.AppContext, command string, args ...string) (ExecuteOutput, *diagnostics.Diagnostics) {
	config := DefaultConfig()
	config.ShowOutput = false
	config.CaptureOutput = true
	return Execute(ctx, config, command, args...)
}

// ValidateCommand checks if a command exists and is executable
func ValidateCommand(ctx *appctx.AppContext, command string) *diagnostics.Diagnostics {
	diag := diagnostics.New("validate_command")
	defer diag.Complete()

	diag.AddPathEntry("start", "executer", map[string]interface{}{
		"command": command,
	})

	ctx.LogWithField("command", command).Debug("Validating command")

	// Check if command exists in PATH
	path, err := exec.LookPath(command)
	if err != nil {
		diag.AddError("COMMAND_NOT_FOUND", "Command not found in PATH", "executer", map[string]interface{}{
			"command": command,
			"error":   err.Error(),
		})
		return diag
	}

	ctx.LogWithField("command_path", path).Debug("Command found")

	diag.AddPathEntry("command_validated", "executer", map[string]interface{}{
		"command":      command,
		"command_path": path,
	})

	return diag
}

// isValidExitCode checks if the exit code is in the list of expected codes
func isValidExitCode(exitCode int, expectedCodes []int) bool {
	for _, expected := range expectedCodes {
		if exitCode == expected {
			return true
		}
	}
	return false
}

// Legacy functions for backward compatibility (deprecated)

// ExecuteLegacy executes a command (legacy function - deprecated)
// Deprecated: Use ExecuteSimple instead
func ExecuteLegacy(command string, args ...string) (ExecuteOutput, error) {
	ctx := appctx.NewContext(context.Background())
	output, diag := ExecuteSimple(ctx, command, args...)

	if diag.HasErrors() {
		return output, fmt.Errorf("command execution failed: %s", diag.GetSummary())
	}

	return output, nil
}

// ExecuteWithNoOutput executes a command without showing output (legacy function - deprecated)
// Deprecated: Use ExecuteSilent instead
func ExecuteWithNoOutput(command string, args ...string) (ExecuteOutput, error) {
	ctx := appctx.NewContext(context.Background())
	output, diag := ExecuteSilent(ctx, command, args...)

	if diag.HasErrors() {
		return output, fmt.Errorf("command execution failed: %s", diag.GetSummary())
	}

	return output, nil
}

// ExecuteWithNoOutputContext executes a command with context (legacy function - deprecated)
// Deprecated: Use ExecuteWithTimeout instead
func ExecuteWithNoOutputContext(ctx context.Context, command string, args ...string) (ExecuteOutput, error) {
	appCtx := appctx.FromContext(ctx)
	if appCtx == nil {
		appCtx = appctx.NewContext(ctx)
	}

	output, diag := ExecuteSilent(appCtx, command, args...)

	if diag.HasErrors() {
		return output, fmt.Errorf("command execution failed: %s", diag.GetSummary())
	}

	return output, nil
}

// ExecuteAndWatchLegacy executes a command and watches output (legacy function - deprecated)
// Deprecated: Use ExecuteAndWatch instead
func ExecuteAndWatchLegacy(command string, args ...string) (ExecuteOutput, error) {
	ctx := appctx.NewContext(context.Background())
	output, diag := ExecuteAndWatch(ctx, command, args...)

	if diag.HasErrors() {
		return output, fmt.Errorf("command execution failed: %s", diag.GetSummary())
	}

	return output, nil
}
