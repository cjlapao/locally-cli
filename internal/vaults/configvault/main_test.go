package configvault

import (
	"testing"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/environment"
	"github.com/stretchr/testify/assert"
)

func TestConfigVault_New(t *testing.T) {
	vault := New()
	assert.NotNil(t, vault)
	assert.Equal(t, "config", vault.Name())
}

func TestConfigVault_Sync(t *testing.T) {
	// Initialize config service
	configSvc, err := config.Initialize()
	assert.NoError(t, err)
	assert.NotNil(t, configSvc)

	// Create vault
	vault := New()
	assert.NotNil(t, vault)

	// Sync vault
	variables, diag := vault.Sync()
	assert.False(t, diag.HasErrors())
	assert.NotNil(t, variables)

	// Verify that we have configuration variables
	assert.Greater(t, len(variables), 0)

	// Check for some expected configuration keys
	expectedKeys := []string{
		"debug",
		"log_level",
		"server.api_port",
		"server.bind_to",
		"database.type",
		"database.storage_path",
	}

	for _, expectedKey := range expectedKeys {
		value, exists := variables[expectedKey]
		assert.True(t, exists, "Expected key %s to exist", expectedKey)
		assert.NotEmpty(t, value, "Expected key %s to have a value", expectedKey)
	}
}

func TestConfigVault_Get(t *testing.T) {
	// Initialize config service
	configSvc, err := config.Initialize()
	assert.NoError(t, err)
	assert.NotNil(t, configSvc)

	// Create vault
	vault := New()
	assert.NotNil(t, vault)

	// Test getting a known configuration value
	value, exists := vault.Get("debug")
	assert.True(t, exists)
	assert.NotNil(t, value)

	// Test getting a non-existent value
	value, exists = vault.Get("non_existent_key")
	assert.False(t, exists)
	assert.Nil(t, value)
}

func TestConfigVault_Set(t *testing.T) {
	// Initialize config service
	configSvc, err := config.Initialize()
	assert.NoError(t, err)
	assert.NotNil(t, configSvc)

	// Create vault
	vault := New()
	assert.NotNil(t, vault)

	// Test setting a string value
	diag := vault.Set("test_key", "test_value")
	assert.False(t, diag.HasErrors())

	// Verify the value was set
	value, exists := vault.Get("test_key")
	assert.True(t, exists)
	assert.Equal(t, "test_value", value)

	// Test setting a boolean value
	diag = vault.Set("test_bool", true)
	assert.False(t, diag.HasErrors())

	// Verify the boolean value was set
	value, exists = vault.Get("test_bool")
	assert.True(t, exists)
	assert.Equal(t, "true", value)

	// Test setting an integer value
	diag = vault.Set("test_int", 42)
	assert.False(t, diag.HasErrors())

	// Verify the integer value was set
	value, exists = vault.Get("test_int")
	assert.True(t, exists)
	assert.Equal(t, "42", value)
}

func TestConfigVault_Remove(t *testing.T) {
	// Initialize config service
	configSvc, err := config.Initialize()
	assert.NoError(t, err)
	assert.NotNil(t, configSvc)

	// Create vault
	vault := New()
	assert.NotNil(t, vault)

	// Set a value first
	diag := vault.Set("test_remove_key", "test_value")
	assert.False(t, diag.HasErrors())

	// Verify the value exists
	value, exists := vault.Get("test_remove_key")
	assert.True(t, exists)
	assert.Equal(t, "test_value", value)

	// Remove the value
	diag = vault.Remove("test_remove_key")
	assert.False(t, diag.HasErrors())

	// Verify the value was removed (should not exist anymore since it's empty)
	value, exists = vault.Get("test_remove_key")
	assert.False(t, exists) // Key doesn't exist since value is empty
	assert.Nil(t, value)
}

func TestConfigVault_IntegrationWithEnvironment(t *testing.T) {
	// Initialize config service
	configSvc, err := config.Initialize()
	assert.NoError(t, err)
	assert.NotNil(t, configSvc)

	// Initialize environment service
	env := environment.Initialize()
	ctx := appctx.NewContext(nil)

	// Create and register config vault
	vault := New()
	diag := env.RegisterVault(ctx, vault)
	assert.False(t, diag.HasErrors())

	// Sync the vault
	diag = env.SyncVault(ctx, "config")
	assert.False(t, diag.HasErrors())

	// Test getting a value through the environment service
	value, diag := env.Get(ctx, "config", "debug", nil)
	assert.False(t, diag.HasErrors())
	assert.NotNil(t, value)

	// Test setting a value through the environment service
	diag = env.SetVariable(ctx, "config", "test_env_key", "test_env_value")
	assert.False(t, diag.HasErrors())

	// Verify the value was set
	value, diag = env.Get(ctx, "config", "test_env_key", nil)
	assert.False(t, diag.HasErrors())
	assert.Equal(t, "test_env_value", value)
}
