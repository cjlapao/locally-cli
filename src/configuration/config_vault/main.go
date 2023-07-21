package config_vault

import (
	"fmt"
	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/notifications"
	"strings"

	"github.com/cjlapao/common-go/helper"
)

type ConfigVault struct {
	name string
}

func New() *ConfigVault {
	result := ConfigVault{
		name: "config",
	}

	return &result
}

func (c ConfigVault) Name() string {
	return c.name
}

func (c ConfigVault) Sync() (map[string]interface{}, error) {
	config := configuration.Get()
	context := config.GetCurrentContext()
	notify := notifications.Get()

	result := make(map[string]interface{})

	baseContextConfigPath := context.RootConfigFilePath
	defaultPath := helper.JoinPath(baseContextConfigPath, configuration.DEFAULT_CONTEXT_SERVICE_FOLDER)
	infrastructurePath := helper.JoinPath(defaultPath, configuration.DEFAULT_CONTEXT_INFRASTRUCTURE_FOLDER)
	backendPath := helper.JoinPath(defaultPath, configuration.DEFAULT_CONTEXT_SERVICE_BACKEND_FOLDER)
	mocksPath := helper.JoinPath(defaultPath, configuration.DEFAULT_CONTEXT_SERVICE_MOCKS_FOLDER)
	webclientsPath := helper.JoinPath(defaultPath, configuration.DEFAULT_CONTEXT_SERVICE_WEBCLIENTS_FOLDER)

	basePath := config.GetCurrentContext().Configuration.OutputPath
	caddyFolder := helper.JoinPath(basePath, common.CADDY_PATH)
	webClientsFolder := helper.JoinPath(basePath, common.SPA_PATH)
	infrastructure := helper.JoinPath(basePath, common.INFRASTRUCTURE_PATH)
	sources := helper.JoinPath(basePath, common.SOURCES_PATH)
	pipelines := helper.JoinPath(basePath, common.PIPELINES_PATH)
	configServiceDataPath := helper.JoinPath(basePath, common.DEFAULT_CONFIG_SERVICE_PATH)

	// Adding context
	result[strings.ToLower("context.path.config")] = baseContextConfigPath
	result[strings.ToLower("context.path.services")] = defaultPath
	result[strings.ToLower("context.path.infrastructure")] = infrastructurePath
	result[strings.ToLower("context.path.backend")] = backendPath
	result[strings.ToLower("context.path.mocks")] = mocksPath
	result[strings.ToLower("context.path.webclients")] = webclientsPath
	result[strings.ToLower("context.domain")] = config.GlobalConfiguration.Network.DomainName
	result[strings.ToLower("context.baseUrl")] = fmt.Sprintf("https://%s.%s", context.Configuration.RootURI, config.GlobalConfiguration.Network.DomainName)
	result[strings.ToLower("context.url_prefix")] = context.Configuration.RootURI

	// Adding Path variables
	result[strings.ToLower("path.base")] = basePath
	result[strings.ToLower("path.caddy")] = caddyFolder
	result[strings.ToLower("path.webclients")] = webClientsFolder
	result[strings.ToLower("path.infrastructure")] = infrastructure
	result[strings.ToLower("path.sources")] = sources
	result[strings.ToLower("path.pipelines")] = pipelines
	result[strings.ToLower("config_service.data.path")] = configServiceDataPath

	// Adding Infrastructure authentication
	if context.Infrastructure != nil && context.Infrastructure.Authorization != nil {
		result[strings.ToLower("infrastructure.authorization.client_id")] = context.Infrastructure.Authorization.ClientId
		result[strings.ToLower("infrastructure.authorization.client_secret")] = context.Infrastructure.Authorization.ClientSecret
		result[strings.ToLower("infrastructure.authorization.subscription_id")] = context.Infrastructure.Authorization.SubscriptionId
		result[strings.ToLower("infrastructure.authorization.tenant_id")] = context.Infrastructure.Authorization.TenantId
	}

	// Adding Path variables
	if config.GlobalConfiguration != nil && config.GlobalConfiguration.Tools != nil {
		if config.GlobalConfiguration.Tools.Git != nil {
			if config.GlobalConfiguration.Tools.Git.Path != "" {
				result["tools.git.path"] = config.GlobalConfiguration.Tools.Git.Path
			}
		}
	}

	for key, val := range result {
		notify.Debug("Synced %s key with value %s", key, val)
	}

	return result, nil
}
