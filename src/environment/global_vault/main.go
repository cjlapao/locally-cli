package global_vault

import (
	"fmt"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/notifications"
	"strings"
)

type GlobalVault struct {
	name string
}

func New() *GlobalVault {
	result := GlobalVault{
		name: "global",
	}

	return &result
}

func (c GlobalVault) Name() string {
	return c.name
}

func (c GlobalVault) Sync() (map[string]interface{}, error) {
	config := configuration.Get()
	context := config.GetCurrentContext()
	notify := notifications.Get()
	result := make(map[string]interface{})

	// Adding Global Variables
	if context.EnvironmentVariables != nil && context.EnvironmentVariables.Global != nil && len(context.EnvironmentVariables.Global) > 0 {
		for key, value := range context.EnvironmentVariables.Global {
			formattedKey := fmt.Sprintf("%s", strings.ToLower(key))
			notify.Debug("Synced %s key with value %s", formattedKey, value)
			result[formattedKey] = value
		}
	}

	return result, nil
}
