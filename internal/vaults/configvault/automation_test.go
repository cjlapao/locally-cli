package configvault

import (
	"testing"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestConfigVault_AutomatedImport(t *testing.T) {
	// Initialize config service
	configSvc, err := config.Initialize()
	assert.NoError(t, err)
	assert.NotNil(t, configSvc)

	// Create vault
	vault := New()
	assert.NotNil(t, vault)

	// Sync vault to automatically import all configuration items
	variables, diag := vault.Sync()
	assert.False(t, diag.HasErrors())
	assert.NotNil(t, variables)

	// Verify that all default configuration items are automatically imported
	// This demonstrates the automation - we don't need to manually specify each item

	// Server configuration
	assert.Contains(t, variables, "server.api_port")
	assert.Contains(t, variables, "server.bind_to")
	assert.Contains(t, variables, "server.base_url")
	assert.Contains(t, variables, "server.api_prefix")

	// Auth configuration
	assert.Contains(t, variables, "auth.root_password")
	assert.Contains(t, variables, "jwt.auth_secret")

	// Database configuration
	assert.Contains(t, variables, "database.type")
	assert.Contains(t, variables, "database.storage_path")
	assert.Contains(t, variables, "database.host")
	assert.Contains(t, variables, "database.port")
	assert.Contains(t, variables, "database.database")
	assert.Contains(t, variables, "database.username")
	assert.Contains(t, variables, "database.password")
	assert.Contains(t, variables, "database.ssl_mode")
	assert.Contains(t, variables, "database.migrate")

	// Message processor configuration
	assert.Contains(t, variables, "message_processor.poll_interval")
	assert.Contains(t, variables, "message_processor.processing_timeout")
	assert.Contains(t, variables, "message_processor.default_max_retries")
	assert.Contains(t, variables, "message_processor.recovery_enabled")
	assert.Contains(t, variables, "message_processor.max_processing_age")
	assert.Contains(t, variables, "message_processor.cleanup_enabled")
	assert.Contains(t, variables, "message_processor.cleanup_max_age")
	assert.Contains(t, variables, "message_processor.cleanup_interval")
	assert.Contains(t, variables, "message_processor.keep_complete_messages")
	assert.Contains(t, variables, "message_processor.debug")

	// CORS configuration
	assert.Contains(t, variables, "cors.allow_origins")
	assert.Contains(t, variables, "cors.allow_methods")
	assert.Contains(t, variables, "cors.allow_headers")
	assert.Contains(t, variables, "cors.expose_headers")

	// Root user configuration
	assert.Contains(t, variables, "root_user.username")
	assert.Contains(t, variables, "root_user.password")

	// Other configuration
	assert.Contains(t, variables, "debug")
	assert.Contains(t, variables, "log_level")
	assert.Contains(t, variables, "api.key")
	assert.Contains(t, variables, "seeding.demo_data")

	t.Logf("Successfully imported %d configuration items automatically", len(variables))

	// Print all imported keys for verification
	for key, value := range variables {
		t.Logf("  %s = %s", key, value)
	}
}

func TestConfigVault_NewConfigurationItems(t *testing.T) {
	// Initialize config service
	configSvc, err := config.Initialize()
	assert.NoError(t, err)
	assert.NotNil(t, configSvc)

	// Get the current configuration
	cfg := configSvc.Get()
	assert.NotNil(t, cfg)

	// Add a new configuration item (simulating adding a new config item to the system)
	newKey := "new.feature.enabled"
	newValue := "true"
	cfg.Set(newKey, newValue)

	// Create vault
	vault := New()
	assert.NotNil(t, vault)

	// Sync vault - this should automatically pick up the new configuration item
	variables, diag := vault.Sync()
	assert.False(t, diag.HasErrors())
	assert.NotNil(t, variables)

	// Verify that the new configuration item was automatically imported
	assert.Contains(t, variables, newKey)
	assert.Equal(t, newValue, variables[newKey])

	t.Logf("New configuration item '%s' was automatically imported with value '%s'", newKey, newValue)
}

func TestConfigVault_ConfigurationUpdates(t *testing.T) {
	// Initialize config service
	configSvc, err := config.Initialize()
	assert.NoError(t, err)
	assert.NotNil(t, configSvc)

	// Create vault
	vault := New()
	assert.NotNil(t, vault)

	// Initial sync
	variables1, diag := vault.Sync()
	assert.False(t, diag.HasErrors())
	assert.NotNil(t, variables1)

	// Update a configuration value
	updatedKey := "debug"
	updatedValue := "true"
	diag = vault.Set(updatedKey, updatedValue)
	assert.False(t, diag.HasErrors())

	// Sync again to get updated values
	variables2, diag := vault.Sync()
	assert.False(t, diag.HasErrors())
	assert.NotNil(t, variables2)

	// Verify that the updated value is reflected
	assert.Equal(t, updatedValue, variables2[updatedKey])

	t.Logf("Configuration update for '%s' was automatically reflected: %s", updatedKey, variables2[updatedKey])
}
