// Package configvault provides a vault for storing configuration items
package configvault

import (
	"fmt"
	"strings"

	"github.com/cjlapao/locally-cli/internal/config"
	"github.com/cjlapao/locally-cli/internal/environment/interfaces"
	"github.com/cjlapao/locally-cli/internal/logging"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"
	"github.com/cjlapao/locally-cli/pkg/models"
	"github.com/cjlapao/locally-cli/pkg/types"
)

type ConfigVault struct {
	name    string
	enabled bool
	synced  bool
	items   []interfaces.EnvironmentVaultItem
}

func New() *ConfigVault {
	result := ConfigVault{
		name:    "config",
		enabled: true,
		synced:  false,
		items:   make([]interfaces.EnvironmentVaultItem, 0),
	}

	return &result
}

func (c ConfigVault) Name() string {
	return c.name
}

func (c ConfigVault) Sync() ([]interfaces.EnvironmentVaultItem, *diagnostics.Diagnostics) {
	diag := diagnostics.New("config_vault_sync")
	defer diag.Complete()

	diag.AddPathEntry("start", "config_vault", map[string]interface{}{
		"vault_name": c.name,
	})

	// Get the config service instance
	configInstance := config.GetInstance()
	if configInstance == nil {
		diag.AddError("CONFIG_SERVICE_NOT_INITIALIZED", "Config service not initialized", "config_vault", map[string]interface{}{
			"vault_name": c.name,
		})
		return c.items, diag
	}

	// Get the current configuration
	cfg := configInstance.Get()
	if cfg == nil {
		diag.AddError("CONFIG_NOT_AVAILABLE", "Configuration not available", "config_vault", map[string]interface{}{
			"vault_name": c.name,
		})
		return c.items, diag
	}

	// Automatically import all configuration items
	for _, item := range cfg.Items {
		if item.IsSet() {
			// Convert the key to lowercase for consistency with other vaults
			key := strings.ToLower(item.Key)
			envItem := models.EnvironmentVaultItem{
				Key:       key,
				Value:     item.Value,
				Encrypted: false,
				Secret:    false,
				ValueType: types.EnvironmentVaultItemTypeString,
			}
			c.items = append(c.items, envItem)

			logging.Debugf("Config vault: synced %s = %s", key, item.Value)
		}
	}

	diag.AddPathEntry("sync_completed", "config_vault", map[string]interface{}{
		"vault_name":      c.name,
		"variables_count": len(c.items),
	})

	logging.Infof("Config vault synced with %d variables", len(c.items))
	c.synced = true
	return c.items, diag
}

func (c ConfigVault) Get(key string) (interface{}, bool) {
	// Get the config service instance
	configInstance := config.GetInstance()
	if configInstance == nil {
		return nil, false
	}

	// Get the current configuration
	cfg := configInstance.Get()
	if cfg == nil {
		return nil, false
	}

	// Look for the item by key (case-insensitive)
	item := cfg.Get(key)
	if item == nil || !item.IsSet() {
		return nil, false
	}

	return item.Value, true
}

func (c ConfigVault) Set(key string, value interface{}) *diagnostics.Diagnostics {
	diag := diagnostics.New("config_vault_set")
	defer diag.Complete()

	diag.AddPathEntry("start", "config_vault", map[string]interface{}{
		"vault_name": c.name,
		"key":        key,
		"value_type": fmt.Sprintf("%T", value),
	})

	// Get the config service instance
	configInstance := config.GetInstance()
	if configInstance == nil {
		diag.AddError("CONFIG_SERVICE_NOT_INITIALIZED", "Config service not initialized", "config_vault", map[string]interface{}{
			"vault_name": c.name,
		})
		return diag
	}

	// Get the current configuration
	cfg := configInstance.Get()
	if cfg == nil {
		diag.AddError("CONFIG_NOT_AVAILABLE", "Configuration not available", "config_vault", map[string]interface{}{
			"vault_name": c.name,
		})
		return diag
	}

	// Convert value to string
	var stringValue string
	switch v := value.(type) {
	case string:
		stringValue = v
	case bool:
		if v {
			stringValue = "true"
		} else {
			stringValue = "false"
		}
	case int, int32, int64:
		stringValue = fmt.Sprintf("%d", v)
	case float32, float64:
		stringValue = fmt.Sprintf("%f", v)
	default:
		stringValue = fmt.Sprintf("%v", v)
	}

	// Set the configuration value
	cfg.Set(key, stringValue)

	diag.AddPathEntry("value_set", "config_vault", map[string]interface{}{
		"vault_name": c.name,
		"key":        key,
		"value":      stringValue,
	})

	logging.Infof("Config vault: set %s = %s", key, stringValue)
	return diag
}

func (c ConfigVault) Remove(key string) *diagnostics.Diagnostics {
	diag := diagnostics.New("config_vault_remove")
	defer diag.Complete()

	diag.AddPathEntry("start", "config_vault", map[string]interface{}{
		"vault_name": c.name,
		"key":        key,
	})

	// Get the config service instance
	configInstance := config.GetInstance()
	if configInstance == nil {
		diag.AddError("CONFIG_SERVICE_NOT_INITIALIZED", "Config service not initialized", "config_vault", map[string]interface{}{
			"vault_name": c.name,
		})
		return diag
	}

	// Get the current configuration
	cfg := configInstance.Get()
	if cfg == nil {
		diag.AddError("CONFIG_NOT_AVAILABLE", "Configuration not available", "config_vault", map[string]interface{}{
			"vault_name": c.name,
		})
		return diag
	}

	// Set the configuration value to empty string (effectively removing it)
	cfg.Set(key, "")

	diag.AddPathEntry("value_removed", "config_vault", map[string]interface{}{
		"vault_name": c.name,
		"key":        key,
	})

	logging.Infof("Config vault: removed %s", key)
	return diag
}

func (c ConfigVault) GetItems() []interfaces.EnvironmentVaultItem {
	return c.items
}

func (c ConfigVault) IsEnabled() bool {
	return c.enabled
}

func (c ConfigVault) IsSynced() bool {
	return c.synced
}

func (c ConfigVault) GetMetadata() map[string]interface{} {
	return make(map[string]interface{})
}
