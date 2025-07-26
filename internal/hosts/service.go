package hosts

import (
	"fmt"

	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/notifications"
)

var globalHostsService *HostsService

type HostsService struct {
	notify  *notifications.NotificationsService
	wrapper *HostsCommandWrapper
}

func New() *HostsService {
	svc := HostsService{
		wrapper: GetWrapper(),
		notify:  notifications.New(ServiceName),
	}

	return &svc
}

func Get() *HostsService {
	if globalHostsService != nil {
		return globalHostsService
	}

	return New()
}

func (svc *HostsService) Clean() error {
	notify.Info("Cleaning tenant Host Entries")

	wrapper := GetWrapper()

	return wrapper.Clean()
}

func (svc *HostsService) GenerateHostsEntries() error {
	notify.Wrench("Generating tenant Host Entries")
	wrapper := GetWrapper()
	config := configuration.Get()

	if err := wrapper.Read(); err != nil {
		return err
	}

	if _, err := wrapper.Add(config.GlobalConfiguration.Network.LocalIP, fmt.Sprintf("%v.%v", config.GetCurrentContext().Configuration.RootURI, config.GlobalConfiguration.Network.DomainName), "Cluster BaseUrl"); err != nil {
		notify.Warning("Unable to add the base url to the host file, err: %v", err.Error())
	}

	for _, webService := range config.GetCurrentContext().SpaServices {
		if webService.URI != "." {
			if _, err := wrapper.Add(config.GlobalConfiguration.Network.LocalIP, fmt.Sprintf("%v.%v", webService.URI, config.GlobalConfiguration.Network.DomainName), fmt.Sprintf("for web client %v", webService.Name)); err != nil {
				notify.Warning("%v was not able to be added to the host file, err: %v", webService.Name, err.Error())
			}
		}
	}

	for _, backendService := range config.GetCurrentContext().BackendServices {
		if backendService.URI != "." && backendService.URI != "" {
			if _, err := wrapper.Add(config.GlobalConfiguration.Network.LocalIP, fmt.Sprintf("%v.%v", backendService.URI, config.GlobalConfiguration.Network.DomainName), fmt.Sprintf("for backend client %v", backendService.Name)); err != nil {
				notify.Warning("%v was not able to be added to the host file, err: %v", backendService.Name, err.Error())
			}
		}
	}

	for _, tenant := range config.GetCurrentContext().Tenants {
		if _, err := wrapper.Add(config.GlobalConfiguration.Network.LocalIP, fmt.Sprintf("%v.%v", tenant.URI, config.GlobalConfiguration.Network.DomainName), fmt.Sprintf("for tenant %v", tenant.Name)); err != nil {
			notify.Warning("%v was not able to be added to the host file, err: %v", tenant.Name, err.Error())
		}
	}

	// Adding the internal config service url for locally to do its magic for non ICF services
	if _, err := wrapper.Add(config.GlobalConfiguration.Network.LocalIP, "config-service.dev-ops.svc.cluster.local", "config service internal traffic"); err != nil {
		notify.Warning("config-service was not able to be added to the host file, err: %v", err.Error())
	}

	if err := wrapper.Save(); err != nil {
		return err
	}

	return nil
}
