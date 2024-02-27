package azure_keyvault

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/environment"
	"github.com/cjlapao/locally-cli/notifications"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/keyvault/azsecrets"
)

var notify = notifications.Get()

type AzureKeyVault struct {
	name      string
	partition string
	options   *AzureKeyVaultOptions
}

type AzureKeyVaultOptions struct {
	KeyVaultUri  string
	DecodeBase64 bool
}

func New(id string, options *AzureKeyVaultOptions) *AzureKeyVault {
	result := AzureKeyVault{
		name:      "keyvault",
		partition: id,
		options:   options,
	}

	return &result
}

func (c AzureKeyVault) Name() string {
	return c.name
}

func (c AzureKeyVault) Sync() (map[string]interface{}, error) {
	env := environment.Get()
	config := configuration.Get()
	configContext := config.GetCurrentContext()
	result := make(map[string]interface{})

	if configContext == nil {
		return result, nil
	}
	if !configContext.IsValid {
		return result, fmt.Errorf("invalid context selected")
	}

	// setting the timeout to 5 minutes
	context, cancel := context.WithTimeout(context.Background(), time.Second*300)

	defer cancel()

	c.setAuthorization()
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		notify.FromError(err, "failed to obtain a credential: %v")
		return nil, err
	}

	client, err := azsecrets.NewClient(c.options.KeyVaultUri, cred, nil)
	if err != nil {
		notify.Error("There was an error getting the keyvault client, the vault %s will not be sync", c.name)
		return nil, err
	}

	pager := client.NewListSecretsPager(nil)
	for pager.More() {
		page, err := pager.NextPage(context)
		if err != nil {
			notify.Error("There was an error getting a page from the keyvault, the vault %s was only partially sync", c.name)
			return nil, err
		}

		for _, secret := range page.Value {
			value := ""
			secretResp, err := client.GetSecret(context, secret.ID.Name(), secret.ID.Version(), nil)
			if err != nil {
				notify.Error("There was an error getting secret %s from the keyvault, it will not sync", secret.ID.Name())
				return nil, err
			}
			if secretResp.Value == nil {
				notify.Warning("Secret %s is empty, will not sync", secretResp.ID.Name())
			}

			if c.options.DecodeBase64 {
				b, err := base64.StdEncoding.DecodeString(*secretResp.Value)
				if err != nil {
					notify.Warning("Secret %s is not a valid base64 encoded, will sync the raw value", secretResp.ID.Name())
					value = *secretResp.Value
				} else {
					value = string(b)
				}
			} else {
				value = string(*secretResp.Value)
			}

			if value != "" {
				formattedKey := strings.ToLower(secret.ID.Name())
				if c.partition != "" {
					formattedKey = fmt.Sprintf("%s.%s", strings.ToLower(c.partition), strings.ToLower(secret.ID.Name()))
				}
				result[formattedKey] = value

				if configContext.EnvironmentVariables != nil {
					if configContext.EnvironmentVariables.KeyVault == nil {
						configContext.EnvironmentVariables.KeyVault = make(map[string]interface{})
					}

					notify.Debug("Adding %s to the %s keyvault environment variable", formattedKey, c.partition)
					configContext.EnvironmentVariables.KeyVault[formattedKey] = value

					// forcing adding to existing environment vault
					env.Add(c.name, formattedKey, value)
				}

				notify.Debug("%s: %s", fmt.Sprintf("%s.%s", c.partition, secret.ID.Name()), value)

			} else {
				notify.Debug("Secret %s is empty, will not sync", secretResp.ID.Name())
			}
		}

		if page.NextLink == nil {
			break
		}
	}

	notify.Debug("Saving the sync of the keyvault %s to the keyvault %s environment variables", c.options.KeyVaultUri, c.partition)
	configContext.SaveEnvironmentVariables()
	return result, nil
}

func (c AzureKeyVault) setAuthorization() error {
	config := configuration.Get()
	context := config.GetCurrentContext()

	if context.Infrastructure == nil {
		return errors.New("cannot find infrastructure for authorization")
	}

	if context.Infrastructure.Authorization == nil {
		return errors.New("cannot find infrastructure authorization for authorization")
	}

	if context.Infrastructure.Authorization.ClientId == "" {
		return errors.New("cannot find infrastructure authorization client id for authorization")
	}
	if context.Infrastructure.Authorization.ClientSecret == "" {
		return errors.New("cannot find infrastructure authorization client secret for authorization")
	}
	if context.Infrastructure.Authorization.TenantId == "" {
		return errors.New("cannot find infrastructure authorization tenant id for authorization")
	}

	os.Setenv("AZURE_TENANT_ID", context.Infrastructure.Authorization.TenantId)
	os.Setenv("AZURE_CLIENT_ID", context.Infrastructure.Authorization.ClientId)
	os.Setenv("AZURE_CLIENT_SECRET", context.Infrastructure.Authorization.ClientSecret)
	return nil
}
