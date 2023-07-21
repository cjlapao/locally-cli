package azure_keyvault

import (
	"github.com/cjlapao/locally-cli/entities"
	"github.com/cjlapao/locally-cli/notifications"
)

// TODO: Add the ability to create storage accounts and containers
var notify = notifications.Get()

type AzureStorageService struct {
	name        string
	credentials entities.AzureCredentials
}

func New(name string, credentials entities.AzureCredentials) *AzureStorageService {
	result := AzureStorageService{
		name:        name,
		credentials: credentials,
	}

	return &result
}

func (c *AzureStorageService) Name() string {
	return c.name
}

// func (c AzureStorageService) Exists() (bool, error) {
// 	config := configuration.Get()
// 	configContext := config.GetCurrentContext()

// 	// setting the timeout to 5 minutes
// 	context, cancel := context.WithTimeout(context.Background(), time.Second*300)

// 	defer cancel()

// 	c.setAuthorization(c.credentials)
// 	cred, err := azidentity.NewDefaultAzureCredential(nil)
// 	if err != nil {
// 		notify.FromError(err, "failed to obtain a credential: %v")
// 		return false, err
// 	}

//   azblob.
// 	serviceClient, err := azblob.NewServiceClient(url, cred, nil)
// 	if err != nil {
// 		return false, err
// 	}

// 	pager := client.NewListSecretsPager(nil)
// 	for pager.More() {
// 		page, err := pager.NextPage(context)
// 		if err != nil {
// 			notify.Error("There was an error getting a page from the keyvault, the vault %s was only partially sync", c.name)
// 			return nil, err
// 		}

// 		for _, secret := range page.Value {
// 			value := ""
// 			secretResp, err := client.GetSecret(context, secret.ID.Name(), secret.ID.Version(), nil)
// 			if err != nil {
// 				notify.Error("There was an error getting secret %s from the keyvault, it will not sync", secret.ID.Name())
// 				return nil, err
// 			}
// 			if secretResp.Value == nil {
// 				notify.Warning("Secret %s is empty, will not sync", secretResp.ID.Name())
// 			}

// 			if c.options.DecodeBase64 {
// 				b, err := base64.StdEncoding.DecodeString(*secretResp.Value)
// 				if err != nil {
// 					notify.Warning("Secret %s is not a valid base64 encoded, will sync the raw value", secretResp.ID.Name())
// 					value = *secretResp.Value
// 				} else {
// 					value = string(b)
// 				}
// 			} else {
// 				value = string(*secretResp.Value)
// 			}

// 			if value != "" {
// 				formattedKey := strings.ToLower(secret.ID.Name())
// 				if c.partition != "" {
// 					formattedKey = fmt.Sprintf("%s.%s", strings.ToLower(c.partition), strings.ToLower(secret.ID.Name()))
// 				}
// 				result[formattedKey] = value

// 				if configContext.EnvironmentVariables != nil {
// 					if configContext.EnvironmentVariables.KeyVault == nil {
// 						configContext.EnvironmentVariables.KeyVault = make(map[string]interface{})
// 					}

// 					notify.Debug("Adding %s to the %s keyvault environment variable", formattedKey, c.partition)
// 					configContext.EnvironmentVariables.KeyVault[formattedKey] = value
// 				}

// 				notify.Debug("%s: %s", fmt.Sprintf("%s.%s", c.partition, secret.ID.Name()), value)

// 			} else {
// 				notify.Debug("Secret %s is empty, will not sync", secretResp.ID.Name())
// 			}
// 		}

// 		if page.NextLink == nil {
// 			break
// 		}
// 	}

// 	notify.Debug("Saving the sync of the keyvault %s to the keyvault %s environment variables", c.options.KeyVaultUri, c.partition)
// 	configContext.SaveEnvironmentVariables()
// 	return result, nil
// }

// func (c AzureStorageService) setAuthorization(credentials entities.AzureCredentials) error {
// 	if credentials.ClientId == "" {
// 		return errors.New("cannot find infrastructure authorization client id for authorization")
// 	}
// 	if credentials.ClientSecret == "" {
// 		return errors.New("cannot find infrastructure authorization client secret for authorization")
// 	}
// 	if credentials.TenantId == "" {
// 		return errors.New("cannot find infrastructure authorization tenant id for authorization")
// 	}

// 	os.Setenv("AZURE_TENANT_ID", credentials.TenantId)
// 	os.Setenv("AZURE_CLIENT_ID", credentials.ClientId)
// 	os.Setenv("AZURE_CLIENT_SECRET", credentials.ClientSecret)
// 	return nil
// }
