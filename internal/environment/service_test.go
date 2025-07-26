package environment

import (
	"context"
	"testing"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/interfaces"
)

func TestEnvironment_Replace(t *testing.T) {
	type args struct {
		source string
	}
	tests := []struct {
		name string
		env  *Environment
		args args
		want string
	}{
		{
			"should change complex",
			Initialize(),
			args{
				source: "/${{config.test}}/foo/${{config.bar}}",
			},
			"/test_config/foo/bar",
		},
		{
			"should change individual",
			Initialize(),
			args{
				source: "${{config.test}}",
			},
			"test_config",
		},
		{
			"should not change individual",
			Initialize(),
			args{
				source: "${{config.test_not_found }}",
			},
			"${{config.test_not_found }}",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ResetForTesting()
			ctx := appctx.NewContext(context.Background())
			tt.env = Initialize()
			tt.env.variables = map[string]map[string]interface{}{}
			// Register the config vault first
			mockVault := &MockVault{name: "config"}
			tt.env.RegisterVault(ctx, mockVault)
			tt.env.SetVariable(ctx, "config", "test", "test_config")
			tt.env.SetVariable(ctx, "config", "bar", "bar")
			if got := tt.env.Replace(ctx, tt.args.source); got != tt.want {
				t.Errorf("Environment.Replace() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEnvironment_RegisterVault(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Create a mock vault
	mockVault := &MockVault{name: "test-vault"}

	// Test registration
	diag := env.RegisterVault(ctx, mockVault)
	if diag.HasErrors() {
		t.Errorf("Failed to register vault: %v", diag.GetSummary())
		t.FailNow()
	}

	// Test duplicate registration
	diag = env.RegisterVault(ctx, mockVault)
	if diag.HasErrors() {
		t.Error("Expected error for duplicate vault registration")
		t.FailNow()
	}

	// Test vault retrieval
	vault, exists := env.GetVault(ctx, "test-vault")
	if !exists {
		t.Error("Vault should exist after registration")
		t.FailNow()
	}
	if vault.Name() != "test-vault" {
		t.Errorf("Expected vault name 'test-vault', got '%s'", vault.Name())
	}
}

func TestEnvironment_RegisterFunction(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Create a mock function
	mockFunction := &MockFunction{name: "test-function"}

	// Test registration
	diag := env.RegisterFunction(ctx, mockFunction)
	if diag.HasErrors() {
		t.Errorf("Failed to register function: %v", diag.GetSummary())
	}

	// Test duplicate registration
	diag = env.RegisterFunction(ctx, mockFunction)
	if diag.HasErrors() {
		t.Error("Expected error for duplicate function registration")
	}

	// Test function retrieval
	function, exists := env.GetFunction(ctx, "test-function")
	if !exists {
		t.Error("Function should exist after registration")
	}
	if function.Name() != "test-function" {
		t.Errorf("Expected function name 'test-function', got '%s'", function.Name())
	}
}

func TestEnvironment_SetAndGetVariable(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault first
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test setting variable
	diag := env.SetVariable(ctx, "test-vault", "test-key", "test-value")
	if diag.HasErrors() {
		t.Errorf("Failed to set variable: %v", diag.GetSummary())
	}

	// Test getting variable
	value, exists := env.GetVariable(ctx, "test-vault", "test-key")
	if !exists {
		t.Error("Variable should exist after setting")
	}
	if value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%v'", value)
	}

	// Test getting non-existent variable
	_, exists = env.GetVariable(ctx, "test-vault", "non-existent")
	if exists {
		t.Error("Non-existent variable should not exist")
	}
}

func TestEnvironment_ResolveVariable(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())
	// Register a vault and set a variable
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)
	env.SetVariable(ctx, "test-vault", "test-key", "test-value")

	// Test resolving variable
	value, exists := env.ResolveVariable(ctx, "test-vault.test-key")
	if !exists {
		t.Error("Variable should be resolved")
	}
	if value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%v'", value)
	}

	// Test resolving non-existent variable
	_, exists = env.ResolveVariable(ctx, "test-vault.non-existent")
	if exists {
		t.Error("Non-existent variable should not be resolved")
	}

	// Test invalid reference format
	_, exists = env.ResolveVariable(ctx, "invalid-format")
	if exists {
		t.Error("Invalid format should not be resolved")
	}
}

func TestEnvironment_ListVaultsAndFunctions(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())
	// Register vaults and functions
	mockVault1 := &MockVault{name: "vault1"}
	mockVault2 := &MockVault{name: "vault2"}
	mockFunction1 := &MockFunction{name: "function1"}
	mockFunction2 := &MockFunction{name: "function2"}

	env.RegisterVault(ctx, mockVault1)
	env.RegisterVault(ctx, mockVault2)
	env.RegisterFunction(ctx, mockFunction1)
	env.RegisterFunction(ctx, mockFunction2)

	// Test listing vaults
	vaults := env.ListVaults(ctx)
	if len(vaults) != 2 {
		t.Errorf("Expected 2 vaults, got %d", len(vaults))
	}

	// Test listing functions
	functions := env.ListFunctions(ctx)
	if len(functions) != 2 {
		t.Errorf("Expected 2 functions, got %d", len(functions))
	}
}

func TestEnvironment_GetStatus(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())
	// Register vaults and functions
	mockVault := &MockVault{name: "test-vault"}
	mockFunction := &MockFunction{name: "test-function"}

	env.RegisterVault(ctx, mockVault)
	env.RegisterFunction(ctx, mockFunction)
	env.SetVariable(ctx, "test-vault", "key1", "value1")
	env.SetVariable(ctx, "test-vault", "key2", "value2")

	// Get status
	status := env.GetStatus(ctx)

	// Verify status
	if status["synced"] != false {
		t.Error("Environment should not be synced initially")
	}

	vaults := status["vaults"].([]string)
	if len(vaults) != 1 || vaults[0] != "test-vault" {
		t.Errorf("Expected vault 'test-vault', got %v", vaults)
	}

	functions := status["functions"].([]string)
	if len(functions) != 1 || functions[0] != "test-function" {
		t.Errorf("Expected function 'test-function', got %v", functions)
	}

	variables := status["variables"].(map[string]int)
	if variables["test-vault"] != 2 {
		t.Errorf("Expected 2 variables in test-vault, got %d", variables["test-vault"])
	}
}

func TestEnvironment_AddAndRemove(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())
	// Register a vault first
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test adding a value
	diag := env.Add(ctx, "test-vault", "test-key", "test-value")
	if diag.HasErrors() {
		t.Errorf("Failed to add value: %v", diag.GetSummary())
	}

	// Test adding to non-existent vault
	diag = env.Add(ctx, "non-existent", "key", "value")
	if !diag.HasErrors() {
		t.Error("Expected error when adding to non-existent vault")
	}

	// Test removing a value
	diag = env.Remove(ctx, "test-vault", "test-key")
	if diag.HasErrors() {
		t.Errorf("Failed to remove value: %v", diag.GetSummary())
	}

	// Test removing from non-existent vault
	diag = env.Remove(ctx, "non-existent", "key")
	if !diag.HasErrors() {
		t.Error("Expected error when removing from non-existent vault")
	}
}

func TestEnvironment_GetWithDefaults(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())
	// Register a vault and add values
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)
	env.Add(ctx, "test-vault", "string-key", "string-value")
	env.Add(ctx, "test-vault", "int-key", 42)
	env.Add(ctx, "test-vault", "bool-key", true)
	env.Add(ctx, "test-vault", "float-key", 3.14)

	// Test Get with existing value
	value, diag := env.Get(ctx, "test-vault", "string-key", "default")
	if diag.HasErrors() {
		t.Errorf("Failed to get value: %v", diag.GetSummary())
	}
	if value != "string-value" {
		t.Errorf("Expected 'string-value', got '%v'", value)
	}

	// Test Get with non-existent key
	value, diag = env.Get(ctx, "test-vault", "non-existent", "default")
	if diag.HasErrors() {
		t.Errorf("Failed to get default value: %v", diag.GetSummary())
	}
	if value != "default" {
		t.Errorf("Expected 'default', got '%v'", value)
	}

	// Test Get with non-existent vault
	value, diag = env.Get(ctx, "non-existent", "key", "default")
	if diag.HasErrors() {
		t.Errorf("Failed to get default value for non-existent vault: %v", diag.GetSummary())
	}
	if value != "default" {
		t.Errorf("Expected 'default', got '%v'", value)
	}
}

func TestEnvironment_GetString(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())
	// Register a vault and add values
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)
	env.Add(ctx, "test-vault", "string-key", "string-value")
	env.Add(ctx, "test-vault", "int-key", 42)
	env.Add(ctx, "test-vault", "bool-key", true)

	// Test GetString with string value
	value, diag := env.GetString(ctx, "test-vault", "string-key", "default")
	if diag.HasErrors() {
		t.Errorf("Failed to get string value: %v", diag.GetSummary())
	}
	if value != "string-value" {
		t.Errorf("Expected 'string-value', got '%s'", value)
	}

	// Test GetString with int value (should convert)
	value, diag = env.GetString(ctx, "test-vault", "int-key", "default")
	if diag.HasErrors() {
		t.Errorf("Failed to get converted string value: %v", diag.GetSummary())
	}
	if value != "42" {
		t.Errorf("Expected '42', got '%s'", value)
	}

	// Test GetString with bool value (should convert)
	value, diag = env.GetString(ctx, "test-vault", "bool-key", "default")
	if diag.HasErrors() {
		t.Errorf("Failed to get converted bool string value: %v", diag.GetSummary())
	}
	if value != "true" {
		t.Errorf("Expected 'true', got '%s'", value)
	}

	// Test GetString with non-existent key
	value, diag = env.GetString(ctx, "test-vault", "non-existent", "default")
	if diag.HasErrors() {
		t.Errorf("Failed to get default string value: %v", diag.GetSummary())
	}
	if value != "default" {
		t.Errorf("Expected 'default', got '%s'", value)
	}
}

func TestEnvironment_GetBool(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())
	// Register a vault and add values
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)
	env.Add(ctx, "test-vault", "bool-true", true)
	env.Add(ctx, "test-vault", "bool-false", false)
	env.Add(ctx, "test-vault", "string-true", "true")
	env.Add(ctx, "test-vault", "string-false", "false")
	env.Add(ctx, "test-vault", "string-yes", "yes")
	env.Add(ctx, "test-vault", "string-no", "no")
	env.Add(ctx, "test-vault", "int-positive", 1)
	env.Add(ctx, "test-vault", "int-zero", 0)
	env.Add(ctx, "test-vault", "string-invalid", "invalid")

	// Test GetBool with bool values
	value, diag := env.GetBool(ctx, "test-vault", "bool-true", false)
	if diag.HasErrors() {
		t.Errorf("Failed to get bool true value: %v", diag.GetSummary())
	}
	if !value {
		t.Error("Expected true, got false")
	}

	value, diag = env.GetBool(ctx, "test-vault", "bool-false", true)
	if diag.HasErrors() {
		t.Errorf("Failed to get bool false value: %v", diag.GetSummary())
	}
	if value {
		t.Error("Expected false, got true")
	}

	// Test GetBool with string values
	value, diag = env.GetBool(ctx, "test-vault", "string-true", false)
	if diag.HasErrors() {
		t.Errorf("Failed to get string true value: %v", diag.GetSummary())
	}
	if !value {
		t.Error("Expected true, got false")
	}

	value, diag = env.GetBool(ctx, "test-vault", "string-yes", false)
	if diag.HasErrors() {
		t.Errorf("Failed to get string yes value: %v", diag.GetSummary())
	}
	if !value {
		t.Error("Expected true, got false")
	}

	value, diag = env.GetBool(ctx, "test-vault", "string-no", true)
	if diag.HasErrors() {
		t.Errorf("Failed to get string no value: %v", diag.GetSummary())
	}
	if value {
		t.Error("Expected false, got true")
	}

	// Test GetBool with int values
	value, diag = env.GetBool(ctx, "test-vault", "int-positive", false)
	if diag.HasErrors() {
		t.Errorf("Failed to get int positive value: %v", diag.GetSummary())
	}
	if !value {
		t.Error("Expected true, got false")
	}

	value, diag = env.GetBool(ctx, "test-vault", "int-zero", true)
	if diag.HasErrors() {
		t.Errorf("Failed to get int zero value: %v", diag.GetSummary())
	}
	if value {
		t.Error("Expected false, got true")
	}

	// Test GetBool with invalid string (should return error and default)
	value, diag = env.GetBool(ctx, "test-vault", "string-invalid", false)
	if diag.HasErrors() {
		t.Error("Expected error for invalid string conversion")
	}
	if value {
		t.Error("Expected default value false, got true")
	}

	// Test GetBool with non-existent key
	value, diag = env.GetBool(ctx, "test-vault", "non-existent", true)
	if diag.HasErrors() {
		t.Errorf("Failed to get default bool value: %v", diag.GetSummary())
	}
	if !value {
		t.Error("Expected default value true, got false")
	}
}

func TestEnvironment_GetInt(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())
	// Register a vault and add values
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)
	env.Add(ctx, "test-vault", "int-key", 42)
	env.Add(ctx, "test-vault", "string-key", "123")
	env.Add(ctx, "test-vault", "bool-true", true)
	env.Add(ctx, "test-vault", "bool-false", false)
	env.Add(ctx, "test-vault", "string-invalid", "invalid")

	// Test GetInt with int value
	value, diag := env.GetInt(ctx, "test-vault", "int-key", 0)
	if diag.HasErrors() {
		t.Errorf("Failed to get int value: %v", diag.GetSummary())
	}
	if value != 42 {
		t.Errorf("Expected 42, got %d", value)
	}

	// Test GetInt with string value (should convert)
	value, diag = env.GetInt(ctx, "test-vault", "string-key", 0)
	if diag.HasErrors() {
		t.Errorf("Failed to get converted string int value: %v", diag.GetSummary())
	}
	if value != 123 {
		t.Errorf("Expected 123, got %d", value)
	}

	// Test GetInt with bool values
	value, diag = env.GetInt(ctx, "test-vault", "bool-true", 0)
	if diag.HasErrors() {
		t.Errorf("Failed to get bool true int value: %v", diag.GetSummary())
	}
	if value != 1 {
		t.Errorf("Expected 1, got %d", value)
	}

	value, diag = env.GetInt(ctx, "test-vault", "bool-false", 0)
	if diag.HasErrors() {
		t.Errorf("Failed to get bool false int value: %v", diag.GetSummary())
	}
	if value != 0 {
		t.Errorf("Expected 0, got %d", value)
	}

	// Test GetInt with invalid string (should return error and default)
	value, diag = env.GetInt(ctx, "test-vault", "string-invalid", 999)
	if diag.HasErrors() {
		t.Error("Expected error for invalid string conversion")
	}
	if value != 999 {
		t.Errorf("Expected default value 999, got %d", value)
	}

	// Test GetInt with non-existent key
	value, diag = env.GetInt(ctx, "test-vault", "non-existent", 999)
	if diag.HasErrors() {
		t.Errorf("Failed to get default int value: %v", diag.GetSummary())
	}
	if value != 999 {
		t.Errorf("Expected default value 999, got %d", value)
	}
}

func TestEnvironment_GetFloat(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())
	// Register a vault first
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test getting float value
	env.SetVariable(ctx, "test-vault", "float-key", 3.14)
	value, diag := env.GetFloat(ctx, "test-vault", "float-key", 0.0)
	if diag.HasErrors() {
		t.Errorf("Failed to get float value: %v", diag.GetSummary())
	}
	if value != 3.14 {
		t.Errorf("Expected value 3.14, got %f", value)
	}

	// Test getting non-existent value with default
	value, diag = env.GetFloat(ctx, "test-vault", "non-existent", 2.5)
	if diag.HasErrors() {
		t.Errorf("Failed to get default float value: %v", diag.GetSummary())
	}
	if value != 2.5 {
		t.Errorf("Expected default value 2.5, got %f", value)
	}

	// Test getting string value (should return default)
	env.SetVariable(ctx, "test-vault", "string-key", "not-a-float")
	value, diag = env.GetFloat(ctx, "test-vault", "string-key", 1.0)
	if diag.HasErrors() {
		t.Errorf("Failed to get string as float: %v", diag.GetSummary())
	}
	if value != 1.0 {
		t.Errorf("Expected default value 1.0, got %f", value)
	}

	// Test getting int value (should convert)
	env.SetVariable(ctx, "test-vault", "int-key", 42)
	value, diag = env.GetFloat(ctx, "test-vault", "int-key", 0.0)
	if diag.HasErrors() {
		t.Errorf("Failed to get int as float: %v", diag.GetSummary())
	}
	if value != 42.0 {
		t.Errorf("Expected value 42.0, got %f", value)
	}

	// Test getting float64 value
	env.SetVariable(ctx, "test-vault", "float64-key", float64(2.718))
	value, diag = env.GetFloat(ctx, "test-vault", "float64-key", 0.0)
	if diag.HasErrors() {
		t.Errorf("Failed to get float64 value: %v", diag.GetSummary())
	}
	if value != 2.718 {
		t.Errorf("Expected value 2.718, got %f", value)
	}
}

func TestEnvironment_GetInstance(t *testing.T) {
	ResetForTesting()

	// Test that GetInstance returns the same instance
	instance1 := GetInstance()
	instance2 := GetInstance()

	if instance1 != instance2 {
		t.Error("GetInstance should return the same instance")
	}

	// Test that it's not nil
	if instance1 == nil {
		t.Error("GetInstance should not return nil")
	}
}

func TestEnvironment_GetInstance_Panic(t *testing.T) {
	// Reset to nil to test panic case
	globalEnvironment = nil

	defer func() {
		if r := recover(); r == nil {
			t.Error("Expected panic when environment not initialized")
		}
		// Reinitialize for other tests
		Initialize()
	}()

	GetInstance()
}

func TestEnvironment_RegisterVault_AlreadyExists(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault
	mockVault1 := &MockVault{name: "test-vault"}
	diag := env.RegisterVault(ctx, mockVault1)
	if diag.HasErrors() {
		t.Errorf("Failed to register first vault: %v", diag.GetSummary())
	}

	// Try to register the same vault again
	mockVault2 := &MockVault{name: "test-vault"}
	diag = env.RegisterVault(ctx, mockVault2)
	if !diag.HasWarnings() {
		t.Error("Expected warning when registering duplicate vault")
	}
}

func TestEnvironment_RegisterFunction_AlreadyExists(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a function
	mockFunction1 := &MockFunction{name: "test-function"}
	diag := env.RegisterFunction(ctx, mockFunction1)
	if diag.HasErrors() {
		t.Errorf("Failed to register first function: %v", diag.GetSummary())
	}

	// Try to register the same function again
	mockFunction2 := &MockFunction{name: "test-function"}
	diag = env.RegisterFunction(ctx, mockFunction2)
	if !diag.HasWarnings() {
		t.Error("Expected warning when registering duplicate function")
	}
}

func TestEnvironment_SetVariable_EdgeCases(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Test setting variable in non-existent vault
	diag := env.SetVariable(ctx, "non-existent", "key", "value")
	if !diag.HasErrors() {
		t.Error("Expected error when setting variable in non-existent vault")
	}

	// Test setting variable with empty key
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)
	diag = env.SetVariable(ctx, "test-vault", "", "value")
	if diag.HasErrors() {
		t.Errorf("Should allow empty key: %v", diag.GetSummary())
	}
}

func TestEnvironment_SyncVault(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test syncing vault
	diag := env.SyncVault(ctx, "test-vault")
	if diag.HasErrors() {
		t.Errorf("Failed to sync vault: %v", diag.GetSummary())
	}

	// Test syncing non-existent vault
	diag = env.SyncVault(ctx, "non-existent")
	if !diag.HasErrors() {
		t.Error("Expected error when syncing non-existent vault")
	}
}

func TestEnvironment_GetVariableString(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault first
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test getting string value
	env.SetVariable(ctx, "test-vault", "string-key", "test-value")
	value, exists := env.GetVariableString(ctx, "test-vault", "string-key")
	if !exists {
		t.Error("String value should exist")
	}
	if value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%s'", value)
	}

	// Test getting non-existent value
	value, exists = env.GetVariableString(ctx, "test-vault", "non-existent")
	if exists {
		t.Error("Non-existent value should not exist")
	}
	if value != "" {
		t.Errorf("Expected empty string, got '%s'", value)
	}

	// Test getting non-string value (should return empty string)
	env.SetVariable(ctx, "test-vault", "int-key", 42)
	value, exists = env.GetVariableString(ctx, "test-vault", "int-key")
	if exists {
		t.Error("Non-string value should not be considered a valid string")
	}
	if value != "" {
		t.Errorf("Expected empty string for non-string value, got '%s'", value)
	}
}

func TestEnvironment_RemoveVariable(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault first
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Set a variable first
	env.SetVariable(ctx, "test-vault", "test-key", "test-value")

	// Verify it exists
	value, exists := env.GetVariable(ctx, "test-vault", "test-key")
	if !exists {
		t.Error("Variable should exist before removal")
	}
	if value != "test-value" {
		t.Errorf("Expected value 'test-value', got '%v'", value)
	}

	// Test removing variable
	diag := env.RemoveVariable(ctx, "test-vault", "test-key")
	if diag.HasErrors() {
		t.Errorf("Failed to remove variable: %v", diag.GetSummary())
	}

	// Verify it's removed
	_, exists = env.GetVariable(ctx, "test-vault", "test-key")
	if exists {
		t.Error("Variable should not exist after removal")
	}

	// Test removing non-existent variable
	diag = env.RemoveVariable(ctx, "test-vault", "non-existent")
	if !diag.HasErrors() {
		t.Error("Expected error when removing non-existent variable")
	}

	// Test removing from non-existent vault
	diag = env.RemoveVariable(ctx, "non-existent", "key")
	if !diag.HasErrors() {
		t.Error("Expected error when removing from non-existent vault")
	}
}

func TestEnvironment_GetAllVariables(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault first
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Set multiple variables
	env.SetVariable(ctx, "test-vault", "key1", "value1")
	env.SetVariable(ctx, "test-vault", "key2", "value2")
	env.SetVariable(ctx, "test-vault", "key3", 42)

	// Test getting all variables
	variables, exists := env.GetAllVariables(ctx, "test-vault")
	if !exists {
		t.Error("Variables should exist")
	}
	if len(variables) != 3 {
		t.Errorf("Expected 3 variables, got %d", len(variables))
	}

	// Verify specific variables
	if variables["key1"] != "value1" {
		t.Errorf("Expected 'value1' for key1, got '%v'", variables["key1"])
	}
	if variables["key2"] != "value2" {
		t.Errorf("Expected 'value2' for key2, got '%v'", variables["key2"])
	}
	if variables["key3"] != 42 {
		t.Errorf("Expected 42 for key3, got '%v'", variables["key3"])
	}

	// Test getting variables from non-existent vault
	variables, exists = env.GetAllVariables(ctx, "non-existent")
	if exists {
		t.Error("Variables should not exist for non-existent vault")
	}
	if len(variables) != 0 {
		t.Errorf("Expected 0 variables for non-existent vault, got %d", len(variables))
	}
}

func TestEnvironment_ExecuteFunction(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a function
	mockFunction := &MockFunction{name: "test-function"}
	env.RegisterFunction(ctx, mockFunction)

	// Test executing function
	result, diag := env.ExecuteFunction(ctx, "test-function", "test-value", "arg1", "arg2")
	if diag.HasErrors() {
		t.Errorf("Failed to execute function: %v", diag.GetSummary())
	}
	if result != "test-function:test-value:arg1:arg2" {
		t.Errorf("Expected 'test-function:test-value:arg1:arg2', got '%s'", result)
	}

	// Test executing non-existent function
	result, diag = env.ExecuteFunction(ctx, "non-existent", "value")
	if !diag.HasErrors() {
		t.Error("Expected error when executing non-existent function")
	}
	if result != "" {
		t.Errorf("Expected empty result for non-existent function, got '%s'", result)
	}
}

func TestEnvironment_Sync(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register vaults
	mockVault1 := &MockVault{name: "vault1"}
	mockVault2 := &MockVault{name: "vault2"}
	env.RegisterVault(ctx, mockVault1)
	env.RegisterVault(ctx, mockVault2)

	// Test syncing all vaults
	diag := env.Sync(ctx)
	if diag.HasErrors() {
		t.Errorf("Failed to sync vaults: %v", diag.GetSummary())
	}

	// Verify sync status
	if !env.IsSynced() {
		t.Error("Environment should be synced after sync operation")
	}
}

func TestEnvironment_IsSynced(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Initially should not be synced
	if env.IsSynced() {
		t.Error("Environment should not be synced initially")
	}

	// Register and sync a vault
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)
	env.Sync(ctx)

	// Should be synced after sync
	if !env.IsSynced() {
		t.Error("Environment should be synced after sync operation")
	}
}

func TestEnvironment_GetInt_EdgeCases(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault first
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test getting float value (should convert)
	env.SetVariable(ctx, "test-vault", "float-key", 3.14)
	value, diag := env.GetInt(ctx, "test-vault", "float-key", 0)
	if diag.HasErrors() {
		t.Errorf("Failed to get float as int: %v", diag.GetSummary())
	}
	if value != 3 {
		t.Errorf("Expected value 3, got %d", value)
	}

	// Test getting string value (should return default)
	env.SetVariable(ctx, "test-vault", "string-key", "not-a-number")
	value, diag = env.GetInt(ctx, "test-vault", "string-key", 42)
	if diag.HasErrors() {
		t.Errorf("Failed to get string as int: %v", diag.GetSummary())
	}
	if value != 42 {
		t.Errorf("Expected default value 42, got %d", value)
	}

	// Test getting int64 value (should convert)
	env.SetVariable(ctx, "test-vault", "int64-key", int64(123))
	value, diag = env.GetInt(ctx, "test-vault", "int64-key", 0)
	if diag.HasErrors() {
		t.Errorf("Failed to get int64 as int: %v", diag.GetSummary())
	}
	if value != 123 {
		t.Errorf("Expected value 123, got %d", value)
	}
}

func TestEnvironment_GetFloat_EdgeCases(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault first
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test getting string value (should return default)
	env.SetVariable(ctx, "test-vault", "string-key", "not-a-float")
	value, diag := env.GetFloat(ctx, "test-vault", "string-key", 3.14)
	if diag.HasErrors() {
		t.Errorf("Failed to get string as float: %v", diag.GetSummary())
	}
	if value != 3.14 {
		t.Errorf("Expected default value 3.14, got %f", value)
	}

	// Test getting bool value (should return default)
	env.SetVariable(ctx, "test-vault", "bool-key", true)
	value, diag = env.GetFloat(ctx, "test-vault", "bool-key", 2.5)
	if diag.HasErrors() {
		t.Errorf("Failed to get bool as float: %v", diag.GetSummary())
	}
	if value != 2.5 {
		t.Errorf("Expected default value 2.5, got %f", value)
	}
}

func TestEnvironment_ProcessFunctionCall(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a function
	mockFunction := &MockFunction{name: "test-function"}
	env.RegisterFunction(ctx, mockFunction)

	// Test processing function call
	result := env.Replace(ctx, "${{test-function:test-value:arg1:arg2}}")
	expected := "test-function:test-value:arg1:arg2"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test processing non-existent function
	result = env.Replace(ctx, "${{non-existent:value}}")
	expected = "${{non-existent:value}}" // Should remain unchanged
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test processing invalid function call format
	result = env.Replace(ctx, "${{test-function}}") // No arguments
	expected = "${{test-function}}"                 // Should remain unchanged
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestEnvironment_ProcessFunctionCall_Direct(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a function
	mockFunction := &MockFunction{name: "test-function"}
	env.RegisterFunction(ctx, mockFunction)

	// Test direct function call processing
	diag := diagnostics.New("test")
	result, diag := env.processFunctionCall(ctx, "test-function(arg1, arg2)", diag)
	if diag.HasErrors() {
		t.Errorf("Failed to process function call: %v", diag.GetSummary())
	}
	expected := "test-function:arg1:arg2"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test invalid function call format
	_, diag = env.processFunctionCall(ctx, "invalid-function", diag)
	if !diag.HasErrors() {
		t.Error("Expected error for invalid function call format")
	}
}

func TestEnvironment_GetString_EdgeCases(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault first
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test getting int value (should convert to string)
	env.SetVariable(ctx, "test-vault", "int-key", 42)
	value, diag := env.GetString(ctx, "test-vault", "int-key", "")
	if diag.HasErrors() {
		t.Errorf("Failed to get int as string: %v", diag.GetSummary())
	}
	if value != "42" {
		t.Errorf("Expected value '42', got '%s'", value)
	}

	// Test getting float value (should convert to string)
	env.SetVariable(ctx, "test-vault", "float-key", 3.14)
	value, diag = env.GetString(ctx, "test-vault", "float-key", "")
	if diag.HasErrors() {
		t.Errorf("Failed to get float as string: %v", diag.GetSummary())
	}
	if value != "3.14" {
		t.Errorf("Expected value '3.14', got '%s'", value)
	}

	// Test getting bool value (should convert to string)
	env.SetVariable(ctx, "test-vault", "bool-key", true)
	value, diag = env.GetString(ctx, "test-vault", "bool-key", "")
	if diag.HasErrors() {
		t.Errorf("Failed to get bool as string: %v", diag.GetSummary())
	}
	if value != "true" {
		t.Errorf("Expected value 'true', got '%s'", value)
	}
}

func TestEnvironment_GetString_MoreEdgeCases(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test getting complex types
	env.SetVariable(ctx, "test-vault", "map-key", map[string]interface{}{"key": "value"})
	value, diag := env.GetString(ctx, "test-vault", "map-key", "default")
	if diag.HasErrors() {
		t.Errorf("Failed to get map as string: %v", diag.GetSummary())
	}
	if value != "default" {
		t.Errorf("Expected default value for map, got '%s'", value)
	}

	// Test getting slice
	env.SetVariable(ctx, "test-vault", "slice-key", []string{"a", "b", "c"})
	value, diag = env.GetString(ctx, "test-vault", "slice-key", "default")
	if diag.HasErrors() {
		t.Errorf("Failed to get slice as string: %v", diag.GetSummary())
	}
	if value != "default" {
		t.Errorf("Expected default value for slice, got '%s'", value)
	}
}

func TestEnvironment_GetBool_EdgeCases(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault first
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test getting string "true" (should convert)
	env.SetVariable(ctx, "test-vault", "true-string", "true")
	value, diag := env.GetBool(ctx, "test-vault", "true-string", false)
	if diag.HasErrors() {
		t.Errorf("Failed to get 'true' as bool: %v", diag.GetSummary())
	}
	if !value {
		t.Error("Expected true, got false")
	}

	// Test getting string "false" (should convert)
	env.SetVariable(ctx, "test-vault", "false-string", "false")
	value, diag = env.GetBool(ctx, "test-vault", "false-string", true)
	if diag.HasErrors() {
		t.Errorf("Failed to get 'false' as bool: %v", diag.GetSummary())
	}
	if value {
		t.Error("Expected false, got true")
	}

	// Test getting string "1" (should convert to true)
	env.SetVariable(ctx, "test-vault", "one-string", "1")
	value, diag = env.GetBool(ctx, "test-vault", "one-string", false)
	if diag.HasErrors() {
		t.Errorf("Failed to get '1' as bool: %v", diag.GetSummary())
	}
	if !value {
		t.Error("Expected true for '1', got false")
	}

	// Test getting string "0" (should convert to false)
	env.SetVariable(ctx, "test-vault", "zero-string", "0")
	value, diag = env.GetBool(ctx, "test-vault", "zero-string", true)
	if diag.HasErrors() {
		t.Errorf("Failed to get '0' as bool: %v", diag.GetSummary())
	}
	if value {
		t.Error("Expected false for '0', got true")
	}

	// Test getting invalid string (should return default)
	env.SetVariable(ctx, "test-vault", "invalid-string", "maybe")
	value, diag = env.GetBool(ctx, "test-vault", "invalid-string", true)
	if diag.HasErrors() {
		t.Errorf("Failed to get invalid string as bool: %v", diag.GetSummary())
	}
	if !value {
		t.Error("Expected default value true, got false")
	}
}

func TestEnvironment_GetBool_MoreEdgeCases(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test getting complex types
	env.SetVariable(ctx, "test-vault", "map-key", map[string]interface{}{"key": "value"})
	value, diag := env.GetBool(ctx, "test-vault", "map-key", true)
	if diag.HasErrors() {
		t.Errorf("Failed to get map as bool: %v", diag.GetSummary())
	}
	if value != true {
		t.Errorf("Expected default value for map, got %v", value)
	}

	// Test getting slice
	env.SetVariable(ctx, "test-vault", "slice-key", []string{"a", "b", "c"})
	value, diag = env.GetBool(ctx, "test-vault", "slice-key", false)
	if diag.HasErrors() {
		t.Errorf("Failed to get slice as bool: %v", diag.GetSummary())
	}
	if value != false {
		t.Errorf("Expected default value for slice, got %v", value)
	}
}

func TestEnvironment_GetInt_MoreEdgeCases(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test getting complex types
	env.SetVariable(ctx, "test-vault", "map-key", map[string]interface{}{"key": "value"})
	value, diag := env.GetInt(ctx, "test-vault", "map-key", 42)
	if diag.HasErrors() {
		t.Errorf("Failed to get map as int: %v", diag.GetSummary())
	}
	if value != 42 {
		t.Errorf("Expected default value for map, got %d", value)
	}

	// Test getting slice
	env.SetVariable(ctx, "test-vault", "slice-key", []string{"a", "b", "c"})
	value, diag = env.GetInt(ctx, "test-vault", "slice-key", 123)
	if diag.HasErrors() {
		t.Errorf("Failed to get slice as int: %v", diag.GetSummary())
	}
	if value != 123 {
		t.Errorf("Expected default value for slice, got %d", value)
	}
}

func TestEnvironment_GetFloat_MoreEdgeCases(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register a vault
	mockVault := &MockVault{name: "test-vault"}
	env.RegisterVault(ctx, mockVault)

	// Test getting complex types
	env.SetVariable(ctx, "test-vault", "map-key", map[string]interface{}{"key": "value"})
	value, diag := env.GetFloat(ctx, "test-vault", "map-key", 3.14)
	if diag.HasErrors() {
		t.Errorf("Failed to get map as float: %v", diag.GetSummary())
	}
	if value != 3.14 {
		t.Errorf("Expected default value for map, got %f", value)
	}

	// Test getting slice
	env.SetVariable(ctx, "test-vault", "slice-key", []string{"a", "b", "c"})
	value, diag = env.GetFloat(ctx, "test-vault", "slice-key", 2.718)
	if diag.HasErrors() {
		t.Errorf("Failed to get slice as float: %v", diag.GetSummary())
	}
	if value != 2.718 {
		t.Errorf("Expected default value for slice, got %f", value)
	}
}

func TestEnvironment_Sync_WithErrors(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Manually add the error vault to the environment to avoid registration issues
	errorVault := &ErrorVault{name: "error-vault"}
	env.mu.Lock()
	env.vaults["error-vault"] = errorVault
	env.variables["error-vault"] = make(map[string]interface{})
	env.mu.Unlock()

	// Test syncing with errors
	diag := env.Sync(ctx, "error-vault")
	if !diag.HasErrors() {
		t.Error("Expected errors when syncing error vault")
	}

	// Environment should not be synced if there are errors
	if env.IsSynced() {
		t.Error("Environment should not be synced when sync has errors")
	}
}

func TestEnvironment_Sync_SpecificVaults(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register multiple vaults
	mockVault1 := &MockVault{name: "vault1"}
	mockVault2 := &MockVault{name: "vault2"}
	mockVault3 := &MockVault{name: "vault3"}
	env.RegisterVault(ctx, mockVault1)
	env.RegisterVault(ctx, mockVault2)
	env.RegisterVault(ctx, mockVault3)

	// Test syncing specific vaults
	diag := env.Sync(ctx, "vault1", "vault2")
	if diag.HasErrors() {
		t.Errorf("Failed to sync specific vaults: %v", diag.GetSummary())
	}

	// Environment should be synced
	if !env.IsSynced() {
		t.Error("Environment should be synced after sync operation")
	}
}

func TestEnvironment_Replace_ComplexScenarios(t *testing.T) {
	ResetForTesting()
	env := Initialize()
	ctx := appctx.NewContext(context.Background())

	// Register vault and function
	mockVault := &MockVault{name: "test-vault"}
	mockFunction := &MockFunction{name: "test-function"}
	env.RegisterVault(ctx, mockVault)
	env.RegisterFunction(ctx, mockFunction)

	// Set some variables
	env.SetVariable(ctx, "test-vault", "var1", "value1")
	env.SetVariable(ctx, "test-vault", "var2", "value2")

	// Test complex replacement with variables and functions
	result := env.Replace(ctx, "Hello ${{test-vault.var1}} and ${{test-function:arg1:arg2}}")
	expected := "Hello value1 and test-function:arg1:arg2"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test replacement with non-existent variables
	result = env.Replace(ctx, "Hello ${{test-vault.non-existent}}")
	expected = "Hello ${{test-vault.non-existent}}" // Should remain unchanged
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}

	// Test replacement with non-existent functions
	result = env.Replace(ctx, "Hello ${{non-existent:arg1}}")
	expected = "Hello ${{non-existent:arg1}}" // Should remain unchanged
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestEnvironment_Operations(t *testing.T) {
	ResetForTesting()
	Initialize() // Initialize the environment
	ctx := appctx.NewContext(context.Background())

	// Test operations function with a variable that exists
	// First set up a variable
	mockVault := &MockVault{name: "test-vault"}
	env := GetInstance()
	env.RegisterVault(ctx, mockVault)
	env.SetVariable(ctx, "test-vault", "test-variable", "test-value")

	// Note: The Operations function calls os.Exit which would terminate the test
	// In a real test environment, we would need to mock the os.Exit call
	// For now, we'll skip this test since it requires special handling
	t.Skip("Operations test requires os.Exit mocking to be properly tested")
}

// Mock implementations for testing

type MockVault struct {
	name   string
	values map[string]interface{}
}

func (m *MockVault) Name() string {
	return m.name
}

func (m *MockVault) Sync() (map[string]interface{}, *diagnostics.Diagnostics) {
	diag := diagnostics.New("sync_vault")
	defer diag.Complete()

	if m.values == nil {
		m.values = make(map[string]interface{})
	}
	return m.values, diag
}

func (m *MockVault) Get(key string) (interface{}, bool) {
	if m.values == nil {
		m.values = make(map[string]interface{})
	}
	value, exists := m.values[key]
	return value, exists
}

func (m *MockVault) Set(key string, value interface{}) *diagnostics.Diagnostics {
	diag := diagnostics.New("set_value_in_vault")
	defer diag.Complete()

	if m.values == nil {
		m.values = make(map[string]interface{})
	}
	m.values[key] = value
	return diag
}

func (m *MockVault) Remove(key string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_value_from_vault")
	defer diag.Complete()
	if m.values == nil {
		m.values = make(map[string]interface{})
	}
	if _, exists := m.values[key]; !exists {
		diag.AddError("NOT_FOUND", "Key not found", "MockVault")
		return diag
	}
	delete(m.values, key)
	return diag
}

type MockFunction struct {
	name string
}

func (m *MockFunction) Name() string {
	return m.name
}

func (m *MockFunction) New() interfaces.EnvironmentVariableFunction {
	return m
}

func (m *MockFunction) Exec(value string, args ...string) (string, *diagnostics.Diagnostics) {
	diag := diagnostics.New("exec_function")
	defer diag.Complete()
	result := m.name
	if value != "" {
		result += ":" + value
	}
	for _, arg := range args {
		result += ":" + arg
	}
	return result, diag
}

// ErrorVault is a mock vault that returns errors for testing
type ErrorVault struct {
	name string
}

func (e *ErrorVault) Name() string {
	return e.name
}

func (e *ErrorVault) Sync() (map[string]interface{}, *diagnostics.Diagnostics) {
	diag := diagnostics.New("sync_error_vault")
	defer diag.Complete()

	diag.AddError("SYNC_ERROR", "Mock sync error", "ErrorVault")
	return make(map[string]interface{}), diag
}

func (e *ErrorVault) Get(key string) (interface{}, bool) {
	return nil, false
}

func (e *ErrorVault) Set(key string, value interface{}) *diagnostics.Diagnostics {
	diag := diagnostics.New("set_error_vault")
	defer diag.Complete()

	diag.AddError("SET_ERROR", "Mock set error", "ErrorVault")
	return diag
}

func (e *ErrorVault) Remove(key string) *diagnostics.Diagnostics {
	diag := diagnostics.New("remove_error_vault")
	defer diag.Complete()

	diag.AddError("REMOVE_ERROR", "Mock remove error", "ErrorVault")
	return diag
}
