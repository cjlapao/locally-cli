package azure_cli

// import (
// 	"context"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"net/http"
// 	"net/url"
// 	"os"
// 	"strings"
// 	"time"

// 	"github.com/cjlapao/locally-cli/internal/common"
// 	"github.com/cjlapao/locally-cli/internal/configuration"
// 	"github.com/cjlapao/locally-cli/internal/entities"
// 	"github.com/cjlapao/locally-cli/internal/environment"
// 	"github.com/cjlapao/locally-cli/internal/executer"
// 	"github.com/cjlapao/locally-cli/internal/helpers"
// 	"github.com/cjlapao/locally-cli/internal/icons"
// 	"github.com/cjlapao/locally-cli/internal/notifications"

// 	"github.com/pascaldekloe/jwt"
// )

// var globalLoginInformation *WrapperLoggedInformation

// const (
// 	ENV_SUBSCRIPTION_ID = "locally_AZURE_SUBSCRIPTION_ID"
// 	ENV_TENANT_ID       = "locally_AZURE_TENANT_ID"
// )

// type AzureCliWrapper struct {
// 	ToolPath      string
// 	CommandOutput string
// 	notifications *notifications.NotificationsService
// }

// func GetWrapper() *AzureCliWrapper {
// 	return &AzureCliWrapper{
// 		notifications: notifications.Get(),
// 	}
// }

// func (svc *AzureCliWrapper) Login(credentials *WrapperCredentials) error {
// 	envSvc := environment.GetInstance()
// 	notify.Rocket("Running Azure Cli Login...")

// 	if credentials != nil {
// 		credentials.Password = envSvc.Replace(credentials.Password)
// 		credentials.SubscriptionId = envSvc.Replace(credentials.SubscriptionId)
// 		credentials.TenantId = envSvc.Replace(credentials.TenantId)
// 		credentials.Username = envSvc.Replace(credentials.Username)
// 	}

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "login")
// 	if credentials != nil {
// 		if credentials.ServicePrincipal {
// 			runArgs = append(runArgs, "--service-principal")
// 		}
// 		if credentials.UseDeviceCode {
// 			runArgs = append(runArgs, "--use-device-code")
// 		}
// 		if credentials.Username != "" {
// 			runArgs = append(runArgs, "-u")
// 			runArgs = append(runArgs, credentials.Username)
// 		}
// 		if credentials.Password != "" {
// 			runArgs = append(runArgs, "-p")
// 			runArgs = append(runArgs, credentials.Username)
// 		}
// 		if credentials.TenantId != "" {
// 			runArgs = append(runArgs, "--tenant")
// 			runArgs = append(runArgs, credentials.TenantId)
// 		}
// 	}

// 	if globalLoginInformation != nil {
// 		if globalLoginInformation.LoggedIn {
// 			if credentials != nil && (credentials.SubscriptionId != globalLoginInformation.SubscriptionId || credentials.Username != globalLoginInformation.Username || credentials.TenantId != globalLoginInformation.TenantId) {
// 				globalLoginInformation.LoggedIn = false
// 			}
// 		}
// 	} else {
// 		globalLoginInformation = &WrapperLoggedInformation{
// 			LoggedIn:           false,
// 			IsServicePrincipal: false,
// 		}
// 	}

// 	if !globalLoginInformation.LoggedIn {
// 		if common.IsDebug() {
// 			notify.Debug("Login Run Arguments: %v", fmt.Sprintf("%v", runArgs))
// 		}

// 		output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), runArgs...)
// 		if err != nil {
// 			notify.FromError(err, "Something wrong running azure cli login")
// 			if output.GetAllOutput() != "" {
// 				notify.Error(output.GetAllOutput())
// 			}

// 			return err
// 		}

// 		var azResponse AzureLoginResponse
// 		if err := json.Unmarshal([]byte(output.GetAllOutput()), &azResponse); err != nil {
// 			return err
// 		}

// 		if credentials != nil && credentials.SubscriptionId != "" {
// 			found := false
// 			for _, subscription := range azResponse {
// 				if strings.EqualFold(subscription.ID, credentials.SubscriptionId) {
// 					found = true
// 					notify.Success("Found subscription %s, setting it to be active", credentials.SubscriptionId)
// 					_, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), "account", "set", "--subscription", credentials.SubscriptionId)
// 					if err != nil {
// 						notify.FromError(err, "Something wrong running azureCli login")
// 						if output.GetAllOutput() != "" {
// 							notify.Error(output.GetAllOutput())
// 						}

// 						return err
// 					}
// 				}
// 			}

// 			if !found {
// 				err := fmt.Errorf("could not find the requested subscription %s in the list of available subscriptions", credentials.SubscriptionId)
// 				return err
// 			}
// 		}

// 		globalLoginInformation.LoggedIn = true
// 		if err := svc.SetSubscription(credentials.SubscriptionId); err != nil {
// 			return err
// 		}

// 		if err := os.Setenv(ENV_TENANT_ID, credentials.TenantId); err != nil {
// 			return err
// 		}

// 		if credentials != nil {
// 			globalLoginInformation.IsServicePrincipal = credentials.ServicePrincipal
// 			globalLoginInformation.Username = credentials.Username
// 			globalLoginInformation.TenantId = credentials.TenantId
// 			globalLoginInformation.SubscriptionId = credentials.SubscriptionId
// 		}

// 		svc.CommandOutput = output.GetAllOutput()

// 		notify.Success("Logged in successfully to Azure using azure cli")
// 	} else {
// 		notify.Info("Already logged in, skipping")
// 	}
// 	return nil
// }

// func (svc *AzureCliWrapper) UserLogin(subscriptionId, tenantId string) error {
// 	env := environment.GetInstance()
// 	msg := "Running Azure Cli User Login"
// 	if tenantId != "" {
// 		msg += fmt.Sprintf(" for tenant %s", tenantId)
// 	}
// 	notify.Rocket("%s...", msg)
// 	notify.InfoWithIcon(icons.IconBell, "This will be opening a browser, please complete the login there.")

// 	if subscriptionId != "" {
// 		subscriptionId = env.Replace(subscriptionId)
// 	}

// 	if tenantId != "" {
// 		tenantId = env.Replace(tenantId)
// 	}

// 	if globalLoginInformation != nil {
// 		if globalLoginInformation.LoggedIn {
// 			if (subscriptionId != "" &&
// 				subscriptionId != globalLoginInformation.SubscriptionId) ||
// 				(tenantId != "" && tenantId != globalLoginInformation.TenantId) {
// 				globalLoginInformation.LoggedIn = false
// 			}
// 		}
// 	} else {
// 		globalLoginInformation = &WrapperLoggedInformation{
// 			LoggedIn:           false,
// 			IsServicePrincipal: false,
// 		}
// 	}

// 	if globalLoginInformation.LoggedIn {
// 		notify.Info("User was already logged in to subscription %s in tenant %s", subscriptionId, tenantId)
// 		return nil
// 	}

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "login")
// 	if tenantId != "" {
// 		runArgs = append(runArgs, "--tenant")
// 		runArgs = append(runArgs, tenantId)
// 	}

// 	if common.IsDebug() {
// 		notify.Debug("User Login Run Arguments: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
// 	defer cancel()
// 	output, err := executer.ExecuteWithNoOutputContext(ctx, helpers.GetAzureCliPath(), runArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli user login")
// 		if output.GetAllOutput() != "" {
// 			notify.Error(output.GetAllOutput())
// 		}

// 		return err
// 	}

// 	globalLoginInformation.LoggedIn = true
// 	globalLoginInformation.IsServicePrincipal = false
// 	globalLoginInformation.SubscriptionId = subscriptionId
// 	globalLoginInformation.TenantId = tenantId

// 	svc.CommandOutput = output.GetAllOutput()
// 	if subscriptionId != "" {
// 		if err := svc.SetSubscription(subscriptionId); err != nil {
// 			return err
// 		}
// 	}
// 	if tenantId != "" {
// 		if err := os.Setenv(ENV_TENANT_ID, tenantId); err != nil {
// 			return err
// 		}
// 	}

// 	notify.Success("User Logged in successfully to Azure using azure cli")
// 	return nil
// }

// func (svc *AzureCliWrapper) Logout() error {
// 	notify.Rocket("Running Azure Cli Logout...")

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "logout")

// 	if globalLoginInformation == nil || !globalLoginInformation.LoggedIn {
// 		notify.Info("Not logged in, skipping")
// 		return nil
// 	}

// 	if common.IsDebug() {
// 		notify.Debug("Logout Run Arguments: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), runArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli logout")
// 		if output.GetAllOutput() != "" {
// 			notify.Error(output.GetAllOutput())
// 		}

// 		return err
// 	}

// 	svc.CommandOutput = output.GetAllOutput()

// 	globalLoginInformation = nil

// 	notify.Success("Logged out successfully to Azure using azure cli")
// 	return nil
// }

// func (svc *AzureCliWrapper) ListApps(displayName string) (*AzureAdAppList, error) {
// 	notify.Rocket("Running Azure Cli List AD Apps...")

// 	envSvc := environment.GetInstance()
// 	displayName = envSvc.Replace(displayName)

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "ad")
// 	runArgs = append(runArgs, "app")
// 	runArgs = append(runArgs, "list")
// 	if displayName != "" {
// 		runArgs = append(runArgs, "--query")
// 		runArgs = append(runArgs, fmt.Sprintf("[?displayName == '%s'].appId", displayName))
// 		runArgs = append(runArgs, "--all")
// 	}

// 	if common.IsDebug() {
// 		notify.Debug("Logout Run Arguments: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), runArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli list apps")
// 		if output.GetAllOutput() != "" {
// 			notify.Error(output.GetAllOutput())
// 		}
// 		return nil, err
// 	}

// 	var appList AzureAdAppList
// 	if err := json.Unmarshal([]byte(output.StdOut), &appList); err != nil {
// 		return nil, err
// 	}

// 	svc.CommandOutput = output.GetAllOutput()

// 	return &appList, nil
// }

// func (svc *AzureCliWrapper) CreateServicePrincipal(displayName string, subscriptionId string) (*ServicePrincipalResponse, error) {
// 	config := configuration.Get()
// 	notify.Rocket("Running Azure Cli Create Service Principal...")
// 	envSvc := environment.GetInstance()

// 	displayName = envSvc.Replace(displayName)
// 	subscriptionId = envSvc.Replace(subscriptionId)

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "ad")
// 	runArgs = append(runArgs, "sp")
// 	runArgs = append(runArgs, "create-for-rbac")
// 	runArgs = append(runArgs, "--name")
// 	runArgs = append(runArgs, displayName)
// 	if subscriptionId != "" {
// 		runArgs = append(runArgs, "--role")
// 		runArgs = append(runArgs, "contributor")
// 		runArgs = append(runArgs, "--scopes")
// 		runArgs = append(runArgs, fmt.Sprintf("/subscriptions/%s", subscriptionId))
// 	}

// 	if common.IsDebug() {
// 		notify.Debug("Run Arguments: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), runArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli ad service principal")
// 		if output.GetAllOutput() != "" {
// 			notify.Error(output.GetAllOutput())
// 		}
// 		return nil, err
// 	}

// 	var spResponse ServicePrincipalResponse
// 	if err := json.Unmarshal([]byte(output.StdOut), &spResponse); err != nil {
// 		return nil, err
// 	}

// 	credential := entities.AzureCredentials{
// 		AppName:        displayName,
// 		ClientId:       spResponse.AppID,
// 		ClientSecret:   spResponse.Password,
// 		SubscriptionId: subscriptionId,
// 		TenantId:       spResponse.TenantId,
// 	}

// 	ctx := config.GetCurrentContext()
// 	ctx.AddAzureCredential(&credential)
// 	cc := config.GetCurrentContext()
// 	if err := cc.SaveCredentials(); err != nil {
// 		return nil, err
// 	}

// 	env := environment.GetInstance()
// 	env.SyncVault("credentials")

// 	svc.CommandOutput = output.GetAllOutput()

// 	return &spResponse, nil
// }

// func (svc *AzureCliWrapper) UpsertResourceGroup(resourceGroupName string, subscriptionId string, location string) error {
// 	envSvc := environment.GetInstance()
// 	resourceGroupName = envSvc.Replace(resourceGroupName)
// 	subscriptionId = envSvc.Replace(subscriptionId)
// 	location = envSvc.Replace(location)

// 	notify.Rocket("Running Azure Cli Upsert Resource Group...")

// 	if resourceGroupName == "" {
// 		err := errors.New("resource group name cannot be null or empty")
// 		return err
// 	}

// 	if subscriptionId == "" {
// 		err := errors.New("subscription id cannot be null or empty")
// 		return err
// 	}

// 	if location == "" {
// 		err := errors.New("location cannot be null or empty")
// 		return err
// 	}

// 	if os.Getenv(ENV_SUBSCRIPTION_ID) != subscriptionId {
// 		if err := svc.SetSubscription(subscriptionId); err != nil {
// 			return err
// 		}
// 	}

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "group")
// 	runArgs = append(runArgs, "list")
// 	runArgs = append(runArgs, "--query")
// 	runArgs = append(runArgs, fmt.Sprintf("[?name=='%s'].{name: name}", resourceGroupName))

// 	if common.IsDebug() {
// 		notify.Debug("Run Parameters: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), runArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli querying resource")
// 		if output.GetAllOutput() != "" {
// 			notify.Error(output.GetAllOutput())
// 		}
// 		return err
// 	}

// 	var response ResourceGroupResponse
// 	if err := json.Unmarshal([]byte(output.StdOut), &response); err != nil {
// 		return err
// 	}

// 	if len(response) > 0 {
// 		notify.Info("Resource group %s already exists, skipping...", resourceGroupName)
// 		return nil
// 	}

// 	createRunArgs := make([]string, 0)
// 	createRunArgs = append(createRunArgs, "group")
// 	createRunArgs = append(createRunArgs, "create")
// 	createRunArgs = append(createRunArgs, "--subscription")
// 	createRunArgs = append(createRunArgs, subscriptionId)
// 	createRunArgs = append(createRunArgs, "--location")
// 	createRunArgs = append(createRunArgs, location)
// 	createRunArgs = append(createRunArgs, "--name")
// 	createRunArgs = append(createRunArgs, resourceGroupName)

// 	if common.IsDebug() {
// 		notify.Debug("Run Parameters: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	createOutput, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), createRunArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli creating resource")
// 		if createOutput.GetAllOutput() != "" {
// 			notify.Error(createOutput.GetAllOutput())
// 		}
// 		return err
// 	}

// 	notify.Success("Azure Resource group %s was created successfully", resourceGroupName)
// 	return nil
// }

// func (svc *AzureCliWrapper) UpsertStorageAccount(name, resourceGroupName, subscriptionId string) error {
// 	envSvc := environment.GetInstance()
// 	name = envSvc.Replace(name)
// 	resourceGroupName = envSvc.Replace(resourceGroupName)
// 	subscriptionId = envSvc.Replace(subscriptionId)

// 	notify.Rocket("Running Azure Cli Upsert Storage Account...")

// 	if !ValidateStorageAccountName(name) {
// 		err := fmt.Errorf("%s is not a valid storage account name. Storage account name must be between 3 and 24 characters in length and use numbers and lower-case letters only", name)
// 		return err
// 	}

// 	if resourceGroupName == "" {
// 		err := errors.New("resource group name cannot be null or empty")
// 		return err
// 	}

// 	if subscriptionId == "" {
// 		err := errors.New("subscription id cannot be null or empty")
// 		return err
// 	}

// 	if os.Getenv(ENV_SUBSCRIPTION_ID) != subscriptionId {
// 		if err := svc.SetSubscription(subscriptionId); err != nil {
// 			return err
// 		}
// 	}

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "storage")
// 	runArgs = append(runArgs, "account")
// 	runArgs = append(runArgs, "list")
// 	runArgs = append(runArgs, "--subscription")
// 	runArgs = append(runArgs, subscriptionId)
// 	runArgs = append(runArgs, "--resource-group")
// 	runArgs = append(runArgs, resourceGroupName)
// 	runArgs = append(runArgs, "--query")
// 	runArgs = append(runArgs, fmt.Sprintf("[?name=='%s'].{name: name}", name))

// 	if common.IsDebug() {
// 		notify.Debug("Run Parameters: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), runArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli querying resource")
// 		if output.GetAllOutput() != "" {
// 			notify.Error(output.GetAllOutput())
// 		}
// 		return err
// 	}

// 	var response StorageAccountResponse
// 	if err := json.Unmarshal([]byte(output.StdOut), &response); err != nil {
// 		return err
// 	}

// 	notify.Debug("List response %s", fmt.Sprintf("%v", response))

// 	if len(response) > 0 {
// 		notify.Info("Storage Account %s already exists in resource group %s for subscription %s, skipping...", name, resourceGroupName, subscriptionId)
// 		return nil
// 	}

// 	createRunArgs := make([]string, 0)
// 	createRunArgs = append(createRunArgs, "storage")
// 	createRunArgs = append(createRunArgs, "account")
// 	createRunArgs = append(createRunArgs, "create")
// 	createRunArgs = append(createRunArgs, "--subscription")
// 	createRunArgs = append(createRunArgs, subscriptionId)
// 	createRunArgs = append(createRunArgs, "--resource-group")
// 	createRunArgs = append(createRunArgs, resourceGroupName)
// 	createRunArgs = append(createRunArgs, "--name")
// 	createRunArgs = append(createRunArgs, name)
// 	createRunArgs = append(createRunArgs, "--sku")
// 	createRunArgs = append(createRunArgs, "Standard_LRS")
// 	createRunArgs = append(createRunArgs, "--encryption-services")
// 	createRunArgs = append(createRunArgs, "blob")

// 	if common.IsDebug() {
// 		notify.Debug("Run Parameters: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	createOutput, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), createRunArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli creating resource")
// 		if createOutput.GetAllOutput() != "" {
// 			notify.Error(createOutput.GetAllOutput())
// 		}
// 		return err
// 	}

// 	notify.Success("Azure Storage Account %s was created successfully", name)
// 	return nil
// }

// func (svc *AzureCliWrapper) GetStorageAccountKey(name, resourceGroupName, subscriptionId string) (string, error) {
// 	envSvc := environment.GetInstance()
// 	name = envSvc.Replace(name)
// 	resourceGroupName = envSvc.Replace(resourceGroupName)
// 	subscriptionId = envSvc.Replace(subscriptionId)

// 	notify.Rocket("Running Azure Cli Upsert Storage Account...")

// 	if !ValidateStorageAccountName(name) {
// 		err := fmt.Errorf("%s is not a valid storage account name. Storage account name must be between 3 and 24 characters in length and use numbers and lower-case letters only", name)
// 		return "", err
// 	}

// 	if resourceGroupName == "" {
// 		err := errors.New("resource group name cannot be null or empty")
// 		return "", err
// 	}

// 	if subscriptionId == "" {
// 		err := errors.New("subscription id cannot be null or empty")
// 		return "", err
// 	}

// 	if os.Getenv(ENV_SUBSCRIPTION_ID) != subscriptionId {
// 		if err := svc.SetSubscription(subscriptionId); err != nil {
// 			return "", err
// 		}
// 	}

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "storage")
// 	runArgs = append(runArgs, "account")
// 	runArgs = append(runArgs, "keys")
// 	runArgs = append(runArgs, "list")
// 	runArgs = append(runArgs, "--subscription")
// 	runArgs = append(runArgs, subscriptionId)
// 	runArgs = append(runArgs, "--resource-group")
// 	runArgs = append(runArgs, resourceGroupName)
// 	runArgs = append(runArgs, "--account-name")
// 	runArgs = append(runArgs, name)
// 	runArgs = append(runArgs, "--query")
// 	runArgs = append(runArgs, "[0].value")

// 	if common.IsDebug() {
// 		notify.Debug("Run Parameters: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), runArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli querying resource")
// 		if output.GetAllOutput() != "" {
// 			notify.Error(output.GetAllOutput())
// 		}
// 		return "", err
// 	}

// 	notify.Debug("Out: %s", output.GetAllOutput())

// 	var key string
// 	if err := json.Unmarshal([]byte(output.StdOut), &key); err != nil {
// 		return "", err
// 	}

// 	notify.Debug("Key: %s", key)

// 	notify.Success("Got storage account %s key successfully", name)
// 	return key, nil
// }

// func (svc *AzureCliWrapper) UpsertStorageAccountContainer(name, storageAccountName, storageAccountKey string) error {
// 	envSvc := environment.GetInstance()
// 	name = envSvc.Replace(name)
// 	storageAccountName = envSvc.Replace(storageAccountName)
// 	storageAccountKey = envSvc.Replace(storageAccountKey)
// 	notify.Rocket("Running Azure Cli Upsert Storage Account Container...")

// 	if !ValidateStorageAccountName(storageAccountName) {
// 		err := fmt.Errorf("%s is not a valid storage account name. Storage account name must be between 3 and 24 characters in length and use numbers and lower-case letters only", storageAccountName)
// 		return err
// 	}

// 	if storageAccountKey == "" {
// 		err := errors.New("storage account key cannot be null or empty")
// 		return err
// 	}

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "storage")
// 	runArgs = append(runArgs, "container")
// 	runArgs = append(runArgs, "list")
// 	runArgs = append(runArgs, "--account-name")
// 	runArgs = append(runArgs, storageAccountName)
// 	runArgs = append(runArgs, "--account-key")
// 	runArgs = append(runArgs, storageAccountKey)
// 	runArgs = append(runArgs, "--query")
// 	runArgs = append(runArgs, fmt.Sprintf("[?name=='%s'].{name: name}", name))

// 	if common.IsDebug() {
// 		notify.Debug("Run Parameters: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), runArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli querying resource")
// 		if output.GetAllOutput() != "" {
// 			notify.Error(output.GetAllOutput())
// 		}
// 		return err
// 	}

// 	var response StorageAccountContainerResponse
// 	if err := json.Unmarshal([]byte(output.StdOut), &response); err != nil {
// 		return err
// 	}

// 	notify.Debug("List response %s", fmt.Sprintf("%v", response))

// 	if len(response) > 0 {
// 		notify.Info("Storage Account Container %s already exists in storage account %s, skipping...", name, storageAccountName)
// 		return nil
// 	}

// 	createRunArgs := make([]string, 0)
// 	createRunArgs = append(createRunArgs, "storage")
// 	createRunArgs = append(createRunArgs, "container")
// 	createRunArgs = append(createRunArgs, "create")
// 	createRunArgs = append(createRunArgs, "--name")
// 	createRunArgs = append(createRunArgs, name)
// 	createRunArgs = append(createRunArgs, "--account-name")
// 	createRunArgs = append(createRunArgs, storageAccountName)
// 	createRunArgs = append(createRunArgs, "--account-key")
// 	createRunArgs = append(createRunArgs, storageAccountKey)

// 	if common.IsDebug() {
// 		notify.Debug("Run Parameters: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	createOutput, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), createRunArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running azure cli creating resource")
// 		if createOutput.GetAllOutput() != "" {
// 			notify.Error(createOutput.GetAllOutput())
// 		}
// 		return err
// 	}

// 	notify.Success("Azure Storage Account Container %s was created successfully", name)
// 	return nil
// }

// func (svc *AzureCliWrapper) SetSubscription(subscriptionId string) error {
// 	envSvc := environment.GetInstance()
// 	subscriptionId = envSvc.Replace(subscriptionId)

// 	notify.Rocket("Running Azure Cli Set Subscription to %s...", subscriptionId)

// 	if subscriptionId == "" {
// 		err := errors.New("subscription id cannot be null or empty")
// 		return err
// 	}

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "account")
// 	runArgs = append(runArgs, "set")
// 	runArgs = append(runArgs, "--subscription")
// 	runArgs = append(runArgs, subscriptionId)

// 	if common.IsDebug() {
// 		notify.Debug("run parameters: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), runArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running setting the subscription")
// 		if output.GetAllOutput() != "" {
// 			notify.Error(output.GetAllOutput())
// 		}
// 		return err
// 	}

// 	os.Setenv(ENV_SUBSCRIPTION_ID, subscriptionId)

// 	return nil
// }

// func (svc *AzureCliWrapper) IsLoggedIn(subscriptionId, tenantId string) bool {
// 	env := environment.GetInstance()
// 	if subscriptionId != "" {
// 		subscriptionId = env.Replace(subscriptionId)
// 	}

// 	if tenantId != "" {
// 		tenantId = env.Replace(tenantId)
// 	}

// 	envSubscriptionId := os.Getenv(ENV_SUBSCRIPTION_ID)
// 	envTenantId := os.Getenv(ENV_TENANT_ID)
// 	if envSubscriptionId != subscriptionId || envTenantId != tenantId {
// 		return false
// 	}

// 	return globalLoginInformation.LoggedIn
// }

// func (svc *AzureCliWrapper) GetAcrRefreshToken(acr, subscriptionId, tenantId string) (string, error) {
// 	env := environment.GetInstance()

// 	if subscriptionId != "" {
// 		subscriptionId = env.Replace(subscriptionId)
// 	} else {
// 		subscriptionId = "empty"
// 	}

// 	if tenantId != "" {
// 		tenantId = env.Replace(tenantId)
// 	} else {
// 		tenantId = "empty"
// 	}

// 	if acr == "" {
// 		err := errors.New("acr cannot be null or empty")
// 		return "", err
// 	}

// 	acr = strings.TrimPrefix(acr, "https://")
// 	acr = strings.TrimPrefix(acr, "http://")

// 	encodedAcrName := common.EncodeName(acr)
// 	token := os.Getenv(fmt.Sprintf("locally_AZURE_%s_%s_%s_ACR_TOKEN", subscriptionId, tenantId, encodedAcrName))
// 	notify.Debug("Token value: %s", token)

// 	if token != "" {
// 		tokenBytes := []byte(token)
// 		rawToken, _ := jwt.ParseWithoutCheck(tokenBytes)
// 		notify.Debug("Raw token: %v", rawToken)
// 		minus1Minute := time.Now().Add((time.Minute * 1) * -1)
// 		isExpired := rawToken.Expires.Time().Before(minus1Minute)
// 		if !isExpired {
// 			notify.Debug("Using same token as it was not expired")
// 			return token, nil
// 		}
// 	}

// 	// checking if we need to login to that subscription/tenant
// 	if !svc.IsLoggedIn(subscriptionId, tenantId) {
// 		err := svc.UserLogin(subscriptionId, tenantId)
// 		if err != nil {
// 			return "", err
// 		}
// 	}

// 	if strings.HasSuffix(acr, ".azurecr.io") {
// 		acr = strings.TrimSuffix(acr, ".azurecr.io")
// 	}

// 	runArgs := make([]string, 0)
// 	runArgs = append(runArgs, "acr")
// 	runArgs = append(runArgs, "login")
// 	runArgs = append(runArgs, "-n")
// 	runArgs = append(runArgs, acr)
// 	if subscriptionId != "" {
// 		runArgs = append(runArgs, "--subscription")
// 		runArgs = append(runArgs, subscriptionId)
// 	}
// 	runArgs = append(runArgs, "--expose-token")

// 	if common.IsDebug() {
// 		notify.Debug("run parameters: %v", fmt.Sprintf("%v", runArgs))
// 	}

// 	output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), runArgs...)
// 	if err != nil {
// 		notify.FromError(err, "Something wrong running setting the subscription")
// 		if output.GetAllOutput() != "" {
// 			notify.Error(output.GetAllOutput())
// 		}
// 		return "", err
// 	}

// 	var response AcrLoginResponse
// 	if err := json.Unmarshal([]byte(output.StdOut), &response); err != nil {
// 		notify.Error("failed to unmarshal the response")
// 	}

// 	os.Setenv(fmt.Sprintf("locally_AZURE_%s_ACR_TOKEN", encodedAcrName), response.AccessToken)

// 	return response.AccessToken, nil
// }

// func (svc *AzureCliWrapper) ExchangeRefreshTokenForAccessToken(acr, scope, subscriptionId, tenantId string) (string, error) {
// 	notify.Rocket("Running Azure Refresh token exchange")

// 	if acr == "" {
// 		err := errors.New("subscription id cannot be null or empty")
// 		return "", err
// 	}

// 	token, _ := svc.GetAcrRefreshToken(acr, subscriptionId, tenantId)

// 	notify.Debug(token)
// 	acr = strings.TrimPrefix(acr, "http://")
// 	acr = strings.TrimPrefix(acr, "https://")

// 	if !strings.Contains(acr, ".azurecr.io") {
// 		acr = fmt.Sprintf("%s.azurecr.io", acr)
// 		notify.Debug("not found the domain, added it", acr)
// 	}

// 	if scope == "" {
// 		scope = "repository:*:metadata_read"
// 	}
// 	oauth2Endpoint := fmt.Sprintf("https://%s/oauth2/token", acr)
// 	body := url.Values{}
// 	body.Add("grant_type", "refresh_token")
// 	body.Add("service", acr)
// 	body.Add("scope", scope)
// 	body.Add("refresh_token", token)

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

// 	defer cancel()

// 	notify.Debug("Body: %v", body.Encode())
// 	notify.Debug("Host: %v", oauth2Endpoint)
// 	req, err := http.NewRequestWithContext(ctx, "POST", oauth2Endpoint, strings.NewReader(body.Encode()))
// 	if err != nil {
// 		return "", err
// 	}

// 	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
// 	client := &http.Client{}

// 	resp, err := client.Do(req)
// 	if err != nil {
// 		return "", err
// 	}

// 	if resp.StatusCode != 200 {
// 		return "", fmt.Errorf("invalid http response, got %s", fmt.Sprintf("%v", resp.StatusCode))
// 	}

// 	if resp.Body == nil {
// 		return "", fmt.Errorf("body cannot be nil")
// 	}

// 	defer resp.Body.Close()

// 	respBody, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", err
// 	}

// 	notify.Debug("Parsing the response body %s", fmt.Sprintf("%v", string(respBody)))
// 	var response Oauth2ACRTokenExchangeResponse
// 	if err := json.Unmarshal(respBody, &response); err != nil {
// 		return "", err
// 	}

// 	return response.AccessToken, nil
// }
