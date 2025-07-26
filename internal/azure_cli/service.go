package azure_cli

import (
	"encoding/json"
	"errors"
	"os"

	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/context/infrastructure_component"
	"github.com/cjlapao/locally-cli/internal/executer"
	"github.com/cjlapao/locally-cli/internal/helpers"
	"github.com/cjlapao/locally-cli/internal/icons"
	"github.com/cjlapao/locally-cli/internal/notifications"
)

var globalAzureCliService *AzureCliService

type AzureCliService struct {
	notify  *notifications.NotificationsService
	wrapper *AzureCliWrapper
}

func New() *AzureCliService {
	svc := AzureCliService{
		wrapper: GetWrapper(),
		notify:  notifications.New(ServiceName),
	}

	return &svc
}

func Get() *AzureCliService {
	if globalAzureCliService != nil {
		return globalAzureCliService
	}

	return New()
}

func (svc *AzureCliService) CheckForAzureCli(softFail bool) {
	config := configuration.Get()
	if !config.GlobalConfiguration.Tools.Checked.AzureCliChecked {
		notify.InfoWithIcon(icons.IconFlag, "Checking for AzureCli tool in the system")
		if output, err := executer.ExecuteWithNoOutput(helpers.GetAzureCliPath(), "version"); err != nil {
			if !softFail {
				notify.Error("AzureCli tool not found in system, this is required for the selected function")
				os.Exit(1)
			} else {
				notify.Warning("AzureCli tool not found in system, this might generate an error in the future")
			}
		} else {
			var jOutput AzureCLIVersion
			if err := json.Unmarshal([]byte(output.StdOut), &jOutput); err != nil {
				if !softFail {
					notify.Error("AzureCli tool not found in system, this is required for the selected function")
					os.Exit(1)
				} else {
					notify.Warning("AzureCli tool not found in system, this might generate an error in the future")
				}
			}

			notify.Success("AzureCli found with version %s, core version %s, telemetry version %s", jOutput.AzureCLI, jOutput.AzureCLICore, jOutput.AzureCLITelemetry)
			if common.IsVerbose() {
				for extension, version := range jOutput.Extensions {
					notify.Success("Found AzureCli extension %s with version %s", extension, version)
				}
			}
		}

		config.GlobalConfiguration.Tools.Checked.AzureCliChecked = true
	}
}

func (svc *AzureCliService) InitBackendResources(config *infrastructure_component.InfrastructureAzureBackendConfig) error {
	configSvc := configuration.Get()
	context := configSvc.GetCurrentContext()
	if context.Credentials == nil {
		return errors.New("credentials cannot be empty")
	}
	if context.Credentials.Azure == nil {
		return errors.New("azure Credentials cannot be empty")
	}

	if !svc.wrapper.IsLoggedIn(context.Credentials.Azure.SubscriptionId, context.Credentials.Azure.TenantId) {
		svc.wrapper.Login(&WrapperCredentials{
			ServicePrincipal: true,
			Username:         context.Credentials.Azure.ClientId,
			Password:         context.Credentials.Azure.ClientSecret,
			SubscriptionId:   context.Credentials.Azure.SubscriptionId,
			TenantId:         context.Credentials.Azure.TenantId,
		})
	}

	var err error
	if err := svc.wrapper.UpsertResourceGroup(config.ResourceGroupName, config.SubscriptionId, config.Location); err != nil {
		return err
	}

	if err := svc.wrapper.UpsertStorageAccount(config.StorageAccountName, config.ResourceGroupName, config.SubscriptionId); err != nil {
		return err
	}

	config.AccessKey, err = svc.wrapper.GetStorageAccountKey(config.StorageAccountName, config.ResourceGroupName, config.SubscriptionId)
	if err != nil {
		return err
	}

	if err := svc.wrapper.UpsertStorageAccountContainer(config.ContainerName, config.StorageAccountName, config.AccessKey); err != nil {
		return err
	}

	return nil
}

func (svc *AzureCliService) IsLoggedIn(subscriptionId, tenantId string) bool {
	return svc.wrapper.IsLoggedIn(subscriptionId, tenantId)
}

func (svc *AzureCliService) Login() error {
	config := configuration.Get()
	context := config.GetCurrentContext()
	if context.Credentials == nil {
		return errors.New("credentials cannot be empty")
	}
	if context.Credentials.Azure == nil {
		return errors.New("azure Credentials cannot be empty")
	}

	err := svc.wrapper.Login(&WrapperCredentials{
		ServicePrincipal: true,
		Username:         context.Credentials.Azure.ClientId,
		Password:         context.Credentials.Azure.ClientSecret,
		SubscriptionId:   context.Credentials.Azure.SubscriptionId,
		TenantId:         context.Credentials.Azure.TenantId,
	})

	return err
}

func (svc *AzureCliService) UserLogin(subscriptionId, tenantId string) error {
	config := configuration.Get()
	context := config.GetCurrentContext()
	if context.Credentials == nil {
		return errors.New("credentials cannot be empty")
	}
	if context.Credentials.Azure == nil {
		return errors.New("azure Credentials cannot be empty")
	}

	err := svc.wrapper.UserLogin(subscriptionId, tenantId)

	return err
}

func (svc *AzureCliService) CreateServicePrincipal(displayName string, subscriptionId string) (*ServicePrincipalResponse, error) {
	return svc.wrapper.CreateServicePrincipal(displayName, subscriptionId)
}

func (svc *AzureCliService) GetAcrRefreshToken(acr, subscriptionId string, tenantId string) (string, error) {
	return svc.wrapper.GetAcrRefreshToken(acr, subscriptionId, tenantId)
}

func (svc *AzureCliService) ExchangeRefreshTokenForAccessToken(acr, scope, subscriptionId, tenantId string) (string, error) {
	return svc.wrapper.ExchangeRefreshTokenForAccessToken(acr, scope, subscriptionId, tenantId)
}
