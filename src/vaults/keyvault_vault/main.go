package keyvault_vault

import (
	"fmt"
	"strings"

	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/notifications"
)

type KeyvaultVault struct {
	name string
}

func New() *KeyvaultVault {
	result := KeyvaultVault{
		name: "keyvault",
	}

	return &result
}

func (c KeyvaultVault) Name() string {
	return c.name
}

func (c KeyvaultVault) Sync() (map[string]interface{}, error) {
	config := configuration.Get()
	context := config.GetCurrentContext()
	notify := notifications.Get()
	result := make(map[string]interface{})

	// Adding Global Variables
	if context.EnvironmentVariables != nil && context.EnvironmentVariables.KeyVault != nil && len(context.EnvironmentVariables.KeyVault) > 0 {
		for key, value := range context.EnvironmentVariables.KeyVault {
			formattedKey := fmt.Sprintf("%s", strings.ToLower(key))
			notify.Debug("Synced %s key with value %s", formattedKey, value)
			result[formattedKey] = value
		}
	}

	return result, nil
}
