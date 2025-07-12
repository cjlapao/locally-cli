package global_vault

import (
	"fmt"
	"strings"

	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/notifications"
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

	if context == nil {
		return result, nil
	}
	if !context.IsValid {
		return result, fmt.Errorf("invalid context selected")
	}

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
