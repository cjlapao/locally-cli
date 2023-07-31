package terraform_vault

import (
	"fmt"
	"strings"

	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/notifications"
)

type TerraformVault struct {
	name string
}

func New() *TerraformVault {
	result := TerraformVault{
		name: "terraform",
	}

	return &result
}

func (c TerraformVault) Name() string {
	return c.name
}

func (c TerraformVault) Sync() (map[string]interface{}, error) {
	config := configuration.Get()
	context := config.GetCurrentContext()
	notify := notifications.Get()
	result := make(map[string]interface{})

	// Adding Global Variables
	if context.EnvironmentVariables != nil && context.EnvironmentVariables.Terraform != nil && len(context.EnvironmentVariables.Terraform) > 0 {
		for key, value := range context.EnvironmentVariables.Terraform {
			formattedKey := fmt.Sprintf("%s", strings.ToLower(key))
			notify.Debug("Synced %s key with value %s", formattedKey, value)
			result[formattedKey] = value
		}
	}

	return result, nil
}
