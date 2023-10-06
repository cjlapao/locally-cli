package credentials_vault

import (
	"fmt"
	"strings"

	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/notifications"
)

type CredentialsVault struct {
	name string
}

func New() *CredentialsVault {
	result := CredentialsVault{
		name: "credentials",
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

	if ctx.Credentials != nil {
		if ctx.Credentials.Azure != nil {
			if ctx.Credentials.Azure.AppName != "" {
				result[strings.ToLower("azure.app_name")] = ctx.Credentials.Azure.AppName
			}
			if ctx.Credentials.Azure.ClientId != "" {
				result[strings.ToLower("azure.client_id")] = ctx.Credentials.Azure.ClientId
			}
			if ctx.Credentials.Azure.ClientSecret != "" {
				result[strings.ToLower("azure.client_secret")] = ctx.Credentials.Azure.ClientSecret
			}
			if ctx.Credentials.Azure.SubscriptionId != "" {
				result[strings.ToLower("azure.subscription_id")] = ctx.Credentials.Azure.SubscriptionId
			}
			if ctx.Credentials.Azure.TenantId != "" {
				result[strings.ToLower("azure.tenant_id")] = ctx.Credentials.Azure.TenantId
			}
		}
	}

	for key, val := range result {
		notify.Debug("[%s] Synced %s key with value %s", c.name, key, val)
	}

	return result, nil
}
