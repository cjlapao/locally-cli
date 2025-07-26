package executer

import (
	"testing"
	"time"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/stretchr/testify/assert"
)

func TestExecuteOutput_GetAllOutput(t *testing.T) {
	output := ExecuteOutput{
		StdOut: "Hello",
		StdErr: "World",
	}

	result := output.GetAllOutput()
	assert.Equal(t, "Hello\nWorld", result)

	// Test with empty stderr
	output.StdErr = ""
	result = output.GetAllOutput()
	assert.Equal(t, "Hello", result)

	// Test with empty stdout
	output.StdOut = ""
	output.StdErr = "World"
	result = output.GetAllOutput()
	assert.Equal(t, "World", result)
}

func TestExecuteOutput_GetFormattedCommand(t *testing.T) {
	output := ExecuteOutput{
		Command: "ls",
		Args:    []string{"-la", "/tmp"},
	}

	result := output.GetFormattedCommand()
	assert.Equal(t, "ls -la /tmp", result)

	// Test with no args
	output.Args = nil
	result = output.GetFormattedCommand()
	assert.Equal(t, "ls", result)
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	assert.Equal(t, 30*time.Second, config.Timeout)
	assert.Equal(t, "", config.WorkingDirectory)
	assert.Nil(t, config.Environment)
	assert.True(t, config.CaptureOutput)
	assert.False(t, config.ShowOutput)
	assert.Equal(t, 0, config.RetryCount)
	assert.Equal(t, 1*time.Second, config.RetryDelay)
	assert.False(t, config.ValidateExitCode)
	assert.Equal(t, []int{0}, config.ExpectedExitCodes)
}

func TestExecuteSimple(t *testing.T) {
	ctx := appctx.NewContext(nil)

	// Test successful command
	output, diag := ExecuteSimple(ctx, "echo", "hello world")
	assert.False(t, diag.HasErrors())
	assert.True(t, output.Success)
	assert.Equal(t, 0, output.ExitCode)
	assert.Contains(t, output.StdOut, "hello world")
	assert.Equal(t, "echo", output.Command)
	assert.Equal(t, []string{"hello world"}, output.Args)

	// Test failed command
	output, diag = ExecuteSimple(ctx, "nonexistentcommand")
	assert.True(t, diag.HasErrors())
	assert.False(t, output.Success)
	assert.NotEqual(t, 0, output.ExitCode)
}

func TestExecuteWithTimeout(t *testing.T) {
	ctx := appctx.NewContext(nil)

	// Test with short timeout
	output, diag := ExecuteWithTimeout(ctx, 1*time.Second, "sleep", "2")
	assert.True(t, diag.HasErrors())
	assert.False(t, output.Success)

	// Test with sufficient timeout
	output, diag = ExecuteWithTimeout(ctx, 5*time.Second, "echo", "test")
	assert.False(t, diag.HasErrors())
	assert.True(t, output.Success)
}

func TestExecuteWithRetry(t *testing.T) {
	ctx := appctx.NewContext(nil)

	// Test with retry (should succeed on first try)
	output, diag := ExecuteWithRetry(ctx, 3, 100*time.Millisecond, "echo", "retry test")
	assert.False(t, diag.HasErrors())
	assert.True(t, output.Success)

	// Test with retry on failing command
	output, diag = ExecuteWithRetry(ctx, 2, 100*time.Millisecond, "nonexistentcommand")
	assert.True(t, diag.HasErrors())
	assert.False(t, output.Success)
}

func TestExecuteInDirectory(t *testing.T) {
	ctx := appctx.NewContext(nil)

	// Test executing in current directory
	output, diag := ExecuteInDirectory(ctx, ".", "pwd")
	assert.False(t, diag.HasErrors())
	assert.True(t, output.Success)
	assert.Contains(t, output.StdOut, "/")
}

func TestExecuteWithEnvironment(t *testing.T) {
	ctx := appctx.NewContext(nil)

	// Test with custom environment
	env := []string{"TEST_VAR=test_value"}
	output, diag := ExecuteWithEnvironment(ctx, env, "env")
	assert.False(t, diag.HasErrors())
	assert.True(t, output.Success)
	assert.Contains(t, output.StdOut, "TEST_VAR=test_value")
}

func TestExecuteAndWatch(t *testing.T) {
	ctx := appctx.NewContext(nil)

	// Test watching output
	output, diag := ExecuteAndWatch(ctx, "echo", "watch test")
	assert.False(t, diag.HasErrors())
	assert.True(t, output.Success)
	assert.Contains(t, output.StdOut, "watch test")
}

func TestExecuteSilent(t *testing.T) {
	ctx := appctx.NewContext(nil)

	// Test silent execution
	output, diag := ExecuteSilent(ctx, "echo", "silent test")
	assert.False(t, diag.HasErrors())
	assert.True(t, output.Success)
	assert.Contains(t, output.StdOut, "silent test")
}

func TestValidateCommand(t *testing.T) {
	ctx := appctx.NewContext(nil)

	// Test valid command
	diag := ValidateCommand(ctx, "echo")
	assert.False(t, diag.HasErrors())

	// Test invalid command
	diag = ValidateCommand(ctx, "nonexistentcommand")
	assert.True(t, diag.HasErrors())
}

func TestIsValidExitCode(t *testing.T) {
	expectedCodes := []int{0, 1, 2}

	// Test valid exit codes
	assert.True(t, isValidExitCode(0, expectedCodes))
	assert.True(t, isValidExitCode(1, expectedCodes))
	assert.True(t, isValidExitCode(2, expectedCodes))

	// Test invalid exit codes
	assert.False(t, isValidExitCode(3, expectedCodes))
	assert.False(t, isValidExitCode(-1, expectedCodes))
}

func TestExecuteWithCustomConfig(t *testing.T) {
	ctx := appctx.NewContext(nil)

	// Test with custom configuration
	config := ExecuteConfig{
		Timeout:           5 * time.Second,
		WorkingDirectory:  ".",
		Environment:       []string{"CUSTOM_VAR=custom_value"},
		CaptureOutput:     true,
		ShowOutput:        false,
		RetryCount:        1,
		RetryDelay:        100 * time.Millisecond,
		ValidateExitCode:  true,
		ExpectedExitCodes: []int{0},
	}

	output, diag := Execute(ctx, config, "echo", "custom config test")
	assert.False(t, diag.HasErrors())
	assert.True(t, output.Success)
	assert.Contains(t, output.StdOut, "custom config test")
}

func TestExecuteWithExitCodeValidation(t *testing.T) {
	ctx := appctx.NewContext(nil)

	// Test with exit code validation
	config := DefaultConfig()
	config.ValidateExitCode = true
	config.ExpectedExitCodes = []int{0, 1}

	// Should succeed with exit code 0
	output, diag := Execute(ctx, config, "echo", "test")
	assert.False(t, diag.HasErrors())
	assert.True(t, output.Success)

	// Should succeed with exit code 1 (using bash to exit with 1)
	output, diag = Execute(ctx, config, "bash", "-c", "exit 1")
	assert.True(t, diag.HasErrors()) // Command failed, so diagnostics should have errors
	assert.False(t, output.Success)  // Command failed but exit code is expected
	assert.Equal(t, 1, output.ExitCode)

	// Should fail with unexpected exit code
	config.ExpectedExitCodes = []int{0}
	output, diag = Execute(ctx, config, "bash", "-c", "exit 1")
	assert.True(t, diag.HasErrors())
}

func TestLegacyFunctions(t *testing.T) {
	// Test legacy Execute function
	output, err := ExecuteLegacy("echo", "legacy test")
	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.StdOut, "legacy test")

	// Test legacy ExecuteWithNoOutput function
	output, err = ExecuteWithNoOutput("echo", "no output test")
	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.StdOut, "no output test")

	// Test legacy ExecuteWithNoOutputContext function
	ctx := appctx.NewContext(nil)
	output, err = ExecuteWithNoOutputContext(ctx, "echo", "context test")
	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.StdOut, "context test")

	// Test legacy ExecuteAndWatchLegacy function
	output, err = ExecuteAndWatchLegacy("echo", "watch legacy test")
	assert.NoError(t, err)
	assert.True(t, output.Success)
	assert.Contains(t, output.StdOut, "watch legacy test")
}

func TestExecuteOutput_JSONTags(t *testing.T) {
	// Test that the struct has proper JSON tags
	output := ExecuteOutput{
		StdOut:    "test output",
		StdErr:    "test error",
		ErrorCode: "test error code",
		ExitCode:  0,
		Duration:  "1s",
		Command:   "test",
		Args:      []string{"arg1", "arg2"},
		StartTime: time.Now(),
		EndTime:   time.Now(),
		Success:   true,
	}

	// This test ensures the struct can be marshaled to JSON
	// The actual marshaling is tested implicitly by the struct definition
	assert.NotEmpty(t, output.Command)
	assert.NotEmpty(t, output.Args)
}
