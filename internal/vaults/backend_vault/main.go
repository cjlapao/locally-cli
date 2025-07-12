package backend_vault

import (
	"fmt"
	"strings"

	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/notifications"
)

type CredentialsVault struct {
	name string
}

func New() *CredentialsVault {
	result := CredentialsVault{
		name: "backend",
	}

	return &result
}

func (c CredentialsVault) Name() string {
	return c.name
}

func (c CredentialsVault) Sync() (map[string]interface{}, error) {
	config := configuration.Get()
	ctx := config.GetCurrentContext()
	notify := notifications.Get()

	result := make(map[string]interface{})

	if ctx == nil {
		return result, nil
	}
	if !ctx.IsValid {
		return result, fmt.Errorf("invalid context selected")
	}

	if ctx.BackendConfig != nil {
		if ctx.BackendConfig.Azure != nil {
			if ctx.BackendConfig.Azure.Location != "" {
				result[strings.ToLower("azure.location")] = ctx.BackendConfig.Azure.Location
			}
			if ctx.BackendConfig.Azure.ResourceGroupName != "" {
				result[strings.ToLower("azure.resource_group_name")] = ctx.BackendConfig.Azure.ResourceGroupName
			}
			if ctx.BackendConfig.Azure.AccessKey != "" {
				result[strings.ToLower("azure.access_key")] = ctx.BackendConfig.Azure.AccessKey
			}
			if ctx.BackendConfig.Azure.StorageAccountName != "" {
				result[strings.ToLower("azure.storage_account_name")] = ctx.BackendConfig.Azure.StorageAccountName
			}
			if ctx.BackendConfig.Azure.ContainerName != "" {
				result[strings.ToLower("azure.container_name")] = ctx.BackendConfig.Azure.ContainerName
			}
			if ctx.BackendConfig.Azure.SubscriptionId != "" {
				result[strings.ToLower("azure.subscription_id")] = ctx.BackendConfig.Azure.SubscriptionId
			}
		}
	}

	for key, val := range result {
		notify.Debug("[%s] Synced %s key with value %s", c.name, key, val)
	}

	return result, nil
}
