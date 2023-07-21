package caddy

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/docker"
	"github.com/cjlapao/locally-cli/executer"
	"github.com/cjlapao/locally-cli/hosts"
	"github.com/cjlapao/locally-cli/icons"
	"github.com/cjlapao/locally-cli/notifications"

	"github.com/cjlapao/common-go/helper"
)

var globalCaddyService *CaddyService

type CaddyService struct {
	notify  *notifications.NotificationsService
	wrapper *CaddyCommandWrapper
}

func New() *CaddyService {
	svc := CaddyService{
		wrapper: GetWrapper(),
		notify:  notifications.New(ServiceName),
	}

	return &svc
}

func Get() *CaddyService {
	if globalCaddyService != nil {
		return globalCaddyService
	}

	return New()
}

func (svc *CaddyService) CheckForCaddy(softFail bool) {
	config = configuration.Get()
	if !config.GlobalConfiguration.Tools.Checked.CaddyChecked {
		notify.InfoWithIcon(icons.IconFlag, "Checking for Caddy tool in the system")
		if output, err := executer.ExecuteWithNoOutput(configuration.GetCaddyPath(), "version"); err != nil {
			if !softFail {
				notify.Error("Caddy tool not found in system, this is required for the selected function")
				os.Exit(1)
			} else {
				notify.Warning("Caddy tool not found in system, this might generate an error in the future")
			}
		} else {
			notify.Success("Caddy tool found with version %s", output.StdOut)
		}
		config.GlobalConfiguration.Tools.Checked.CaddyChecked = true
	}
}

func (svc *CaddyService) GenerateDockerFiles() {
	notify.Wrench("Generating locally Docker Files")

	svc.generateDockerfile()

	for _, client := range config.GetCurrentContext().SpaServices {
		// We only copy if the service is not a reverse proxy
		if !client.UseReverseProxy {
			if err := svc.copyWebClient(client); err != nil {
				notify.Error(err.Error())
			}
		}
	}

	svc.generateDockerComposeFile()
}

func (svc *CaddyService) GenerateCaddyFiles() error {
	hostSvc := hosts.Get()
	if !helper.FileExists(helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH)) {
		notify.Hammer("Creating Caddy folder")
		if !helper.CreateDirectory(helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH), fs.ModePerm) {
			return errors.New("error creating the caddy folder")
		}
	}

	if config.GlobalConfiguration.Network != nil && config.GlobalConfiguration.Network.CERTPath != "" && config.GlobalConfiguration.Network.PrivateKeyPath != "" {
		if !helper.FileExists(helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH, common.TLS_PATH)) {
			notify.Hammer("Creating Caddy SSL Folder folder")
			if !helper.CreateDirectory(helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH, common.TLS_PATH), fs.ModePerm) {
				return errors.New("error creating the caddy SSL folder")
			}
		}

		svc.copyCertificates()
	}

	if err := svc.generateMainCaddyFile(); err != nil {
		notify.AddError(err.Error())
	}
	if err := svc.generateBackendRootServicesAndRoutesCaddyFile(); err != nil {
		notify.AddError(err.Error())
	}
	if err := svc.generateHostedBackendServicesCaddyFile(); err != nil {
		notify.AddError(err.Error())
	}
	if err := svc.generateSpaServicesCaddyFile(); err != nil {
		notify.AddError(err.Error())
	}
	if err := svc.generateTenantsCaddyFile(); err != nil {
		notify.AddError(err.Error())
	}
	if err := svc.generateCaddyMockServicesRoutesFile(); err != nil {
		notify.AddError(err.Error())
	}
	if helper.GetFlagSwitch("update-dns", false) {
		if err := hostSvc.GenerateHostsEntries(); err != nil {
			notify.AddError(err.Error())
		}
	}

	return nil
}

func (svc *CaddyService) BuildContainer() error {
	notify.Rocket("Running Docker Compose Build for locally service")
	dockerCmdWrapper := docker.GetWrapper()
	if err := dockerCmdWrapper.Build(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, ""); err != nil {
		return err
	}

	return nil
}

func (svc *CaddyService) RebuildContainer() error {
	notify.Rocket("Running Docker Compose Rebuild for locally service")
	dockerCmdWrapper := docker.GetWrapper()

	if err := svc.ContainerDown(); err != nil {
		return err
	}

	if err := svc.GenerateCaddyFiles(); err != nil {
		return err
	}

	if err := dockerCmdWrapper.Build(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, ""); err != nil {
		return err
	}

	if err := svc.ContainerUp(); err != nil {
		return err
	}

	return nil
}

func (svc *CaddyService) ContainerUp() error {
	dockerCmdWrapper := docker.GetWrapper()
	if err := dockerCmdWrapper.Up(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, ""); err != nil {
		return err
	}
	return nil
}

func (svc *CaddyService) ContainerDown() error {
	dockerCmdWrapper := docker.GetWrapper()
	images, err := dockerCmdWrapper.GetServiceImages(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, "")
	if err != nil {
		return err
	}

	if err := dockerCmdWrapper.Down(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, ""); err != nil {
		return err
	}

	if len(images) == 0 {
		notify.Warning("No image was found to delete, there should be at least one")
	}

	for _, image := range images {
		if err := dockerCmdWrapper.RemoveImage(image.Repository, image.Tag); err != nil {
			return err
		}
	}

	return nil
}

func (svc *CaddyService) StartContainer() error {
	dockerCmdWrapper := docker.GetWrapper()
	if err := dockerCmdWrapper.Start(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, ""); err != nil {
		return err
	}
	return nil
}

func (svc *CaddyService) StoplocallyContainer() error {
	dockerCmdWrapper := docker.GetWrapper()
	if err := dockerCmdWrapper.Stop(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, ""); err != nil {
		return err
	}
	return nil
}

func (svc *CaddyService) PauseContainer() error {
	dockerCmdWrapper := docker.GetWrapper()
	if err := dockerCmdWrapper.Pause(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, ""); err != nil {
		return err
	}
	return nil
}

func (svc *CaddyService) ResumeContainer() error {
	dockerCmdWrapper := docker.GetWrapper()
	if err := dockerCmdWrapper.Resume(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, ""); err != nil {
		return err
	}
	return nil
}

func (svc *CaddyService) ContainerStatus() error {
	dockerCmdWrapper := docker.GetWrapper()
	if err := dockerCmdWrapper.Status(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, ""); err != nil {
		return err
	}

	return nil
}

func (svc *CaddyService) ContainerLogs() error {
	dockerCmdWrapper := docker.GetWrapper()
	if err := dockerCmdWrapper.Logs(config.GetCurrentContext().Configuration.OutputPath, common.SERVICE_NAME, ""); err != nil {
		return err
	}

	return nil
}

func (svc *CaddyService) generateSpaServicesCaddyFile() error {
	folderPath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH, common.CADDY_UI_PATH)
	if !helper.FileExists(folderPath) {
		notify.Hammer("Folder %v was not found, creating...", common.CADDY_UI_PATH)
		if !helper.CreateDirectory(folderPath, fs.ModePerm) {
			err := fmt.Errorf("there was an error creating the folder %v", folderPath)
			notify.Error(err.Error())
			return err
		} else {
			if config.Verbose() {
				notify.InfoWithIcon(icons.IconCheckMark, "Folder %v was created successfully", common.CADDY_UI_PATH)
			}
		}
	}

	for _, client := range config.GetCurrentContext().SpaServices {
		if client.Default {
			continue
		}

		if client.UseReverseProxy && client.ReverseProxyURI != "" {
			notify.Wrench("Generating %v as a reverse proxy caddy file", client.Name)
			clientFolderName := configuration.EncodeName(client.Name)
			filePath := helper.JoinPath(folderPath, fmt.Sprintf("%v.caddyfile", clientFolderName))
			caddyFile := ""
			if client.URI == "." || client.URI == "" {
				caddyFile += fmt.Sprintf("%v {\n", config.GlobalConfiguration.Network.DomainName)
			} else {
				caddyFile += fmt.Sprintf("%v.%v {\n", client.URI, config.GlobalConfiguration.Network.DomainName)
			}

			if client.RouteReplace != nil {
				if client.RouteReplace.Old != "" {
					if client.RouteReplace.Type == "" {
						client.RouteReplace.Type = "replace"
					}

					caddyFile += "\n"
					switch client.RouteReplace.Type {
					case "replace":
						caddyFile += fmt.Sprintf("  uri replace %v %v\n", client.RouteReplace.Old, client.RouteReplace.New)
					case "strip_prefix":
						caddyFile += fmt.Sprintf("  uri strip_prefix %v\n", client.RouteReplace.Old)
					}
					caddyFile += "\n"
				}
			}

			caddyFile += fmt.Sprintf("  reverse_proxy %v\n", client.ReverseProxyURI)

			if fragment, err := svc.generateServiceMockRouteFragment(clientFolderName, client.MockRoutes); err == nil {
				caddyFile += fragment
			}

			caddyFile += "}\n"

			notify.Debug(caddyFile)
			if err := helper.WriteToFile(caddyFile, filePath); err != nil {
				return err
			}

			if config.Verbose() {
				notify.InfoWithIcon(icons.IconCheckMark, "Finished generating %v as a reverse proxy caddy file", client.Name)
			}
		} else {
			notify.Wrench("Generating %v caddy file", client.Name)
			clientFolderName := configuration.EncodeName(client.Name)
			filePath := helper.JoinPath(folderPath, fmt.Sprintf("%v.caddyfile", clientFolderName))
			caddyFile := ""
			if client.URI == "" {
				caddyFile += fmt.Sprintf("root * {$%v_path}\n", clientFolderName)
				caddyFile += "\n"
				caddyFile += "encode gzip\n"
				caddyFile += "try_files {path} /index.html\n"
				caddyFile += "file_server\n"
				caddyFile += "\n"
				caddyFile += "@javascript {\n"
				caddyFile += "  path_regexp ^.*\\.js\n"
				caddyFile += "}\n"
				caddyFile += "\n"
				caddyFile += "route @javascript {\n"
				caddyFile += "  header Content-Type application/javascript\n"
				caddyFile += "}\n"
				caddyFile += "route * {\n"
				caddyFile += "\n"
				caddyFile += fmt.Sprintf("  import cors_headers https://%v.%v\n", config.GetCurrentContext().Configuration.RootURI, config.GlobalConfiguration.Network.DomainName)
				for _, tenant := range config.GetCurrentContext().Tenants {
					caddyFile += fmt.Sprintf("  import cors_headers https://%v.%v\n", tenant.URI, config.GlobalConfiguration.Network.DomainName)
				}
				for _, webClient := range config.GetCurrentContext().SpaServices {
					if webClient.URI != "" && webClient.URI != "." {
						println(webClient.URI)
						caddyFile += fmt.Sprintf("  import cors_headers https://%v.%v\n", webClient.URI, config.GlobalConfiguration.Network.DomainName)
					}
				}

				caddyFile += "}\n"

				if fragment, err := svc.generateServiceMockRouteFragment(clientFolderName, client.MockRoutes); err == nil {
					caddyFile += fragment
				}

			} else {
				if client.URI == "." {
					caddyFile += fmt.Sprintf("%v {\n", config.GlobalConfiguration.Network.DomainName)
				} else {
					caddyFile += fmt.Sprintf("%v.%v {\n", client.URI, config.GlobalConfiguration.Network.DomainName)
				}
				caddyFile += fmt.Sprintf("  root * {$%v_path}\n", clientFolderName)
				caddyFile += "\n"
				caddyFile += "  encode gzip\n"
				caddyFile += "  try_files {path} /index.html\n"
				caddyFile += "  file_server\n"
				caddyFile += "\n"
				caddyFile += "  @javascript {\n"
				caddyFile += "    path_regexp ^.*\\.js\n"
				caddyFile += "  }\n"
				caddyFile += "\n"
				caddyFile += "  route @javascript {\n"
				caddyFile += "    header Content-Type application/javascript\n"
				caddyFile += "  }\n"
				caddyFile += "  route * {\n"
				caddyFile += "\n"
				caddyFile += fmt.Sprintf("    import cors_headers https://%v.%v\n", config.GetCurrentContext().Configuration.RootURI, config.GlobalConfiguration.Network.DomainName)
				for _, tenant := range config.GetCurrentContext().Tenants {
					caddyFile += fmt.Sprintf("    import cors_headers https://%v.%v\n", tenant.URI, config.GlobalConfiguration.Network.DomainName)
				}
				for _, webClient := range config.GetCurrentContext().SpaServices {
					if webClient.URI != "" && webClient.URI != "." {
						caddyFile += fmt.Sprintf("    import cors_headers https://%v.%v\n", webClient.URI, config.GlobalConfiguration.Network.DomainName)
					}
				}
				caddyFile += "  }\n"

				if fragment, err := svc.generateServiceMockRouteFragment(clientFolderName, client.MockRoutes); err == nil {
					caddyFile += fragment
				}

				caddyFile += "}\n"
			}

			if err := helper.WriteToFile(caddyFile, filePath); err != nil {
				return err
			}

			if config.Verbose() {
				notify.InfoWithIcon(icons.IconCheckMark, "Finished generating %v caddy file", client.Name)
			}
		}
	}

	return nil
}

func (svc *CaddyService) generateBackendRootServicesAndRoutesCaddyFile() error {
	// Generating the services, routes and mock routes folder
	rootServicesFolderPath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH, common.CADDY_ROOT_SERVICES_PATH)
	if !helper.FileExists(rootServicesFolderPath) {
		notify.Hammer("Folder %v was not found, creating", common.CADDY_ROOT_SERVICES_PATH)
		if !helper.CreateDirectory(rootServicesFolderPath, fs.ModePerm) {
			err := fmt.Errorf("there was an error creating the folder %v", rootServicesFolderPath)
			notify.Error(err.Error())
			return err
		} else {
			if config.Verbose() {
				notify.Hammer("Folder %v was created successfully", common.CADDY_ROOT_SERVICES_PATH)
			}
		}
	}

	servicesHostFolderPath := helper.JoinPath(rootServicesFolderPath, common.CADDY_ROOT_SERVICES_HOSTS_PATH)
	if !helper.FileExists(servicesHostFolderPath) {
		notify.Hammer("Folder %v was not found, creating", common.CADDY_ROOT_SERVICES_HOSTS_PATH)
		if !helper.CreateDirectory(servicesHostFolderPath, fs.ModePerm) {
			err := fmt.Errorf("there was an error creating the folder %v", servicesHostFolderPath)
			notify.Error(err.Error())
			return err
		} else {
			if config.Verbose() {
				notify.InfoWithIcon(icons.IconCheckMark, "Folder %v was created successfully", common.CADDY_ROOT_SERVICES_HOSTS_PATH)
			}
		}
	}

	servicesRouteFolderPath := helper.JoinPath(rootServicesFolderPath, common.CADDY_ROOT_SERVICES_ROUTES_PATH)
	if !helper.FileExists(servicesRouteFolderPath) {
		notify.Hammer("Folder %v was not found, creating", common.CADDY_ROOT_SERVICES_ROUTES_PATH)
		if !helper.CreateDirectory(servicesRouteFolderPath, fs.ModePerm) {
			err := fmt.Errorf("there was an error creating the folder %v", servicesRouteFolderPath)
			notify.Error(err.Error())
			return err
		} else {
			if config.Verbose() {
				notify.InfoWithIcon(icons.IconCheckMark, "Folder %v was created successfully", common.CADDY_ROOT_SERVICES_ROUTES_PATH)
			}
		}
	}

	for _, service := range config.GetCurrentContext().BackendServices {
		if service.URI == "" {
			for _, component := range service.Components {
				if component.ReverseProxyURI != "" && len(component.Routes) > 0 {
					notify.Wrench("Generating %v service %v component caddy file", service.Name, component.Name)
					componentFolderName := fmt.Sprintf("%v_%v", configuration.EncodeName(service.Name), configuration.EncodeName(component.Name))
					serviceFilePath := helper.JoinPath(servicesHostFolderPath, fmt.Sprintf("%v.caddyfile", componentFolderName))
					componentCaddyFile := fmt.Sprintf("(%v) {\n", componentFolderName)
					componentCaddyFile += fmt.Sprintf("  reverse_proxy %v\n", component.ReverseProxyURI)
					componentCaddyFile += "}\n"

					if err := helper.WriteToFile(componentCaddyFile, serviceFilePath); err != nil {
						return err
					}

					if config.Verbose() {
						notify.InfoWithIcon(icons.IconCheckMark, "Finished generating %v service %v component caddy file", service.Name, component.Name)
					}

					notify.Wrench("Generating %v service %v component routes caddy file", service.Name, component.Name)
					componentRoutesFolderName := fmt.Sprintf("%v_%v_routes", configuration.EncodeName(service.Name), configuration.EncodeName(component.Name))
					serviceRoutesFilePath := helper.JoinPath(servicesRouteFolderPath, fmt.Sprintf("%v.caddyfile", componentRoutesFolderName))
					componentRoutesCaddyFile := ""
					for _, route := range component.Routes {
						routeEndpoint := configuration.EncodeName(route.Name)
						componentRoutesCaddyFile += fmt.Sprintf("@%v_%v_endpoint {\n", componentRoutesFolderName, routeEndpoint)
						componentRoutesCaddyFile += fmt.Sprintf("  path_regexp %v\n", route.Regex)
						componentRoutesCaddyFile += "}\n"
						componentRoutesCaddyFile += "\n"
						componentRoutesCaddyFile += fmt.Sprintf("handle @%v_%v_endpoint {\n", componentRoutesFolderName, routeEndpoint)
						for _, headerMap := range route.Headers {
							for header, value := range headerMap {
								componentRoutesCaddyFile += fmt.Sprintf("  header %v \"%v\"\n", header, value)
							}
						}
						if route.Replace.Old != "" {
							if route.Replace.Type == "" {
								route.Replace.Type = "replace"
							}
							componentRoutesCaddyFile += "\n"
							switch route.Replace.Type {
							case "replace":
								componentRoutesCaddyFile += fmt.Sprintf("  uri replace %v %v\n", route.Replace.Old, route.Replace.New)
							case "strip_prefix":
								componentRoutesCaddyFile += fmt.Sprintf("  uri strip_prefix %v\n", route.Replace.Old)
							}
						}
						componentRoutesCaddyFile += "\n"
						componentRoutesCaddyFile += fmt.Sprintf("  import %v\n", componentFolderName)

						componentRoutesCaddyFile += "}\n"
						componentRoutesCaddyFile += "\n"
					}

					if fragment, err := svc.generateServiceMockRouteFragment(componentRoutesFolderName, component.MockRoutes); err == nil {
						componentRoutesCaddyFile += fragment
					}

					if err := helper.WriteToFile(componentRoutesCaddyFile, serviceRoutesFilePath); err != nil {
						return err
					}

					if config.Verbose() {
						notify.InfoWithIcon(icons.IconCheckMark, "Finished generating %v service %v component routes caddy file", service.Name, component.Name)
					}
				}
			}
		}
	}

	return nil
}

func (svc *CaddyService) generateTenantsCaddyFile() error {
	folderPath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH, common.CADDY_TENANTS_PATH)
	if !helper.FileExists(folderPath) {
		notify.Hammer("Folder %v was not found, creating", common.CADDY_TENANTS_PATH)
		if !helper.CreateDirectory(folderPath, fs.ModePerm) {
			err := fmt.Errorf("there was an error creating the folder %v", folderPath)
			notify.Error(err.Error())
			return err
		} else {
			if config.Verbose() {
				notify.InfoWithIcon(icons.IconCheckMark, "Folder %v was created successfully", common.CADDY_TENANTS_PATH)
			}
		}
	}

	defaultSpa := svc.getDefaultSpaService()
	for _, tenant := range config.GetCurrentContext().Tenants {
		notify.Wrench("Generating %v tenant caddy file", tenant.Name)
		tenantFolderName := configuration.EncodeName(tenant.Name)
		filePath := helper.JoinPath(folderPath, fmt.Sprintf("%v.caddyfile", tenantFolderName))
		webClientShellName := configuration.EncodeName(common.WEB_CLIENT_SHELL_NAME)
		caddyFile := fmt.Sprintf("%v.%v {\n", tenant.URI, config.GlobalConfiguration.Network.DomainName)
		caddyFile += "\n"
		caddyFile += fmt.Sprintf("  import cors_headers https://%v.%v\n", config.GetCurrentContext().Configuration.RootURI, config.GlobalConfiguration.Network.DomainName)
		for _, tenant := range config.GetCurrentContext().Tenants {
			caddyFile += fmt.Sprintf("  import cors_headers https://%v.%v\n", tenant.URI, config.GlobalConfiguration.Network.DomainName)
		}
		for _, webClient := range config.GetCurrentContext().SpaServices {
			if webClient.URI != "" && webClient.URI != "." {
				caddyFile += fmt.Sprintf("  import cors_headers https://%v.%v\n", webClient.URI, config.GlobalConfiguration.Network.DomainName)
			}
		}
		caddyFile += "\n"
		if defaultSpa != nil {
			if defaultSpa.UseReverseProxy {
				caddyFile += fmt.Sprintf("  reverse_proxy %s\n", defaultSpa.ReverseProxyURI)
			} else {
				caddyFile += fmt.Sprintf("  import {$root_path}/%v/%v.caddyfile\n", defaultSpa.Name, webClientShellName)
			}
		}

		caddyFile += "}\n"

		if err := helper.WriteToFile(caddyFile, filePath); err != nil {
			return err
		}

		if config.Verbose() {
			notify.InfoWithIcon(icons.IconCheckMark, "Finished generating %v tenant caddy file", tenant.Name)
		}

	}

	return nil
}

func (svc *CaddyService) generateHostedBackendServicesCaddyFile() error {
	folderPath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH, common.CADDY_HOSTED_SERVICES_PATH)
	if !helper.FileExists(folderPath) {
		notify.Hammer("Folder %v was not found, creating", common.CADDY_HOSTED_SERVICES_PATH)
		if !helper.CreateDirectory(folderPath, fs.ModePerm) {
			err := fmt.Errorf("there was an error creating the folder %v", folderPath)
			notify.Error(err.Error())
			return err
		} else {
			if config.Verbose() {
				notify.InfoWithIcon(icons.IconCheckMark, "Folder %v was created successfully", common.CADDY_HOSTED_SERVICES_PATH)
			}
		}
	}

	for _, backend := range config.GetCurrentContext().BackendServices {
		if backend.URI != "" {
			notify.Wrench("Generating %v hosted backend service caddy file", backend.Name)
			backendFolderName := configuration.EncodeName(backend.Name)
			filePath := helper.JoinPath(folderPath, fmt.Sprintf("%v.caddyfile", backendFolderName))
			caddyFile := fmt.Sprintf("%v.%v {\n", backend.URI, config.GlobalConfiguration.Network.DomainName)
			caddyFile += "\n"
			if config.GlobalConfiguration.Network != nil && config.GlobalConfiguration.Network.CERTPath != "" && config.GlobalConfiguration.Network.PrivateKeyPath != "" {
				certPath := filepath.Base(config.GlobalConfiguration.Network.CERTPath)
				privKeyPath := filepath.Base(config.GlobalConfiguration.Network.PrivateKeyPath)
				caddyFile += fmt.Sprintf("  tls {$root_path}/%v/%v {$root_path}/%v/%v\n", common.TLS_PATH, certPath, common.TLS_PATH, privKeyPath)
			}
			caddyFile += "\n"
			caddyFile += fmt.Sprintf("  import cors_headers https://%v.%v\n", backend.URI, config.GlobalConfiguration.Network.DomainName)

			if len(backend.AllowedOrigins) > 0 {
				for _, origin := range backend.AllowedOrigins {
					if !strings.HasPrefix(origin, "http://") || !strings.HasPrefix(origin, "https://") {
						origin = fmt.Sprintf("https://%s", origin)
					}

					caddyFile += fmt.Sprintf("  import cors_headers %s\n", origin)
				}
			}

			caddyFile += "\n"

			for _, component := range backend.Components {
				for _, route := range component.Routes {
					componentName := fmt.Sprintf("%v_%v", configuration.EncodeName(backend.Name), configuration.EncodeName(component.Name))
					routeEndpoint := configuration.EncodeName(route.Name)
					caddyFile += fmt.Sprintf("  @%v_%v_endpoint {\n", componentName, routeEndpoint)

					caddyFile += fmt.Sprintf("    path_regexp %v\n", route.Regex)
					caddyFile += "  }\n"
					caddyFile += "  \n"
					caddyFile += fmt.Sprintf("  handle @%v_%v_endpoint {\n", componentName, routeEndpoint)
					for _, headerMap := range route.Headers {
						for header, value := range headerMap {
							caddyFile += fmt.Sprintf("    header %v \"%v\"\n", header, value)
						}
					}
					if route.Replace.Old != "" {
						if route.Replace.Type == "" {
							route.Replace.Type = "replace"
						}
						caddyFile += "\n"
						switch route.Replace.Type {
						case "replace":
							caddyFile += fmt.Sprintf("  uri replace %v %v\n", route.Replace.Old, route.Replace.New)
						case "strip_prefix":
							caddyFile += fmt.Sprintf("  uri strip_prefix %v\n", route.Replace.Old)
						}
					}
					caddyFile += "  \n"

					if fragment, err := svc.generateServiceMockRouteFragment(componentName, component.MockRoutes); err == nil {
						caddyFile += fragment
					}

					caddyFile += "  \n"

					caddyFile += fmt.Sprintf("    reverse_proxy %s\n", component.ReverseProxyURI)
					caddyFile += "  }\n"
					caddyFile += "  \n"
				}
			}

			caddyFile += "}\n"

			if err := helper.WriteToFile(caddyFile, filePath); err != nil {
				return err
			}

			if config.Verbose() {
				notify.InfoWithIcon(icons.IconCheckMark, "%s Finished generating %v hosted backend service caddy file", backend.Name)
			}
		}
	}

	return nil
}

func (svc *CaddyService) getDefaultSpaService() *configuration.SpaService {
	config := configuration.Get()
	for _, spaService := range config.GetCurrentContext().SpaServices {
		if spaService.Default {
			return spaService
		}
	}

	return nil
}

func (svc *CaddyService) generateCaddyMockServicesRoutesFile() error {
	servicesMockRouteFolderPath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH, common.CADDY_MOCK_ROUTES_PATH)
	notify.Debug(servicesMockRouteFolderPath)
	if !helper.FileExists(servicesMockRouteFolderPath) {
		notify.Hammer("Folder %v was not found, creating", common.CADDY_MOCK_ROUTES_PATH)
		if !helper.CreateDirectory(servicesMockRouteFolderPath, fs.ModePerm) {
			err := fmt.Errorf("there was an error creating the folder %v", servicesMockRouteFolderPath)
			notify.Error(err.Error())
			return err
		} else {
			if config.Verbose() {
				notify.InfoWithIcon(icons.IconCheckMark, "Folder %v was created", common.CADDY_MOCK_ROUTES_PATH)
			}
		}
	}

	for _, mockService := range config.GetCurrentContext().MockServices {
		mockRoutesFolderName := fmt.Sprintf("%v_mock_routes", configuration.EncodeName(mockService.Name))
		serviceRoutesFilePath := helper.JoinPath(servicesMockRouteFolderPath, fmt.Sprintf("%v.caddyfile", mockRoutesFolderName))
		notify.Wrench("Generating mock service %v routes caddy file", mockService.Name)
		mockRoutesCaddyFile := ""
		for _, mockRoute := range mockService.MockRoutes {
			notify.Debug(mockRoute.Regex)
			if mockRoute.Regex != "" {
				routeEndpoint := configuration.EncodeName(mockRoute.Name)
				mockRoutesCaddyFile += fmt.Sprintf("@%v_%v_endpoint {\n", mockRoutesFolderName, routeEndpoint)
				mockRoutesCaddyFile += fmt.Sprintf("  path_regexp %v\n", mockRoute.Regex)
				mockRoutesCaddyFile += "}\n"
				mockRoutesCaddyFile += "\n"
				mockRoutesCaddyFile += fmt.Sprintf("handle @%v_%v_endpoint {\n", mockRoutesFolderName, routeEndpoint)
				for _, headerMap := range mockRoute.Headers {
					for header, value := range headerMap {
						mockRoutesCaddyFile += fmt.Sprintf("  header %v \"%v\"\n", header, value)
					}
				}

				if mockRoute.Responds.ContentType != "" {
					mockRoutesCaddyFile += fmt.Sprintf("  header Content-Type \"%v\"\n", mockRoute.Responds.ContentType)
				}

				mockRoutesCaddyFile += "\n"
				if mockRoute.Responds.RawBody != "" {
					mockRoutesCaddyFile += fmt.Sprintf("  respond %q\n", mockRoute.Responds.RawBody)

				} else if mockRoute.Responds.Body != nil {
					body, err := json.Marshal(mockRoute.Responds.Body)

					if err != nil {
						return err
					}

					mockRoutesCaddyFile += fmt.Sprintf("  respond %q\n", body)

				} else {
					mockRoutesCaddyFile += "  respond \"OK\"\n"
				}
				mockRoutesCaddyFile += "}\n"
				mockRoutesCaddyFile += "\n"

				notify.Debug(mockRoutesCaddyFile)
				if err := helper.WriteToFile(mockRoutesCaddyFile, serviceRoutesFilePath); err != nil {
					return err
				}

				if config.Verbose() {
					notify.InfoWithIcon(icons.IconCheckMark, "Finished generating mock service %v routes caddy file", mockService.Name)
				}
			}
		}
	}

	return nil
}

func (svc *CaddyService) copyWebClient(client *configuration.SpaService) error {
	if err := svc.updateWebClientEnvironment(client); err != nil {
		return err
	}

	notify.Info("Starting to copy client SPA %v to the build folder", client.Name)
	basePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.SPA_PATH, configuration.EncodeName(client.Name))
	if err := helper.CopyDir(client.Path, basePath); err != nil {
		notify.FromError(err, "Could not copy %v Client SPA to build folder")
		return err
	}
	if config.Verbose() {
		notify.Info("Finished copying the client SPA %v to the build folder", client.Name)
	}
	return nil
}

func (svc *CaddyService) generateDockerfile() error {
	filePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, "dockerfile")
	notify.Wrench("Generating locally dockerfile")
	dockerFile := "FROM caddy:latest\n"
	dockerFile += "\n"
	dockerFile += "WORKDIR /etc/caddy\n"
	dockerFile += "\n"
	dockerFile += fmt.Sprintf("COPY ./%v /etc/caddy\n", common.CADDY_PATH)
	dockerFile += "RUN mkdir /etc/caddy/spa\n"
	dockerFile += "\n"
	if config.GlobalConfiguration.Network != nil && config.GlobalConfiguration.Network.CERTPath != "" && config.GlobalConfiguration.Network.PrivateKeyPath != "" {
		certPath := filepath.Base(config.GlobalConfiguration.Network.CERTPath)
		privKeyPath := filepath.Base(config.GlobalConfiguration.Network.PrivateKeyPath)
		baseTlsPath := configuration.EncodeName(helper.JoinPath(common.CADDY_PATH, common.TLS_PATH))

		dockerFile += fmt.Sprintf("COPY ./%v/%v /etc/caddy/%v\n", baseTlsPath, certPath, certPath)
		dockerFile += fmt.Sprintf("COPY ./%v/%v /etc/caddy/%v\n", baseTlsPath, privKeyPath, privKeyPath)

		dockerFile += "\n"
	}

	for _, client := range config.GetCurrentContext().SpaServices {
		if !client.UseReverseProxy {
			folderName := configuration.EncodeName(client.Name)
			basePath := configuration.EncodeName(helper.JoinPath(common.SPA_PATH, folderName))
			dockerFile += fmt.Sprintf("COPY ./%v /etc/caddy/spa/%v\n", basePath, configuration.EncodeName(client.Name))
			dockerFile += fmt.Sprintf("ENV %v_path=/etc/caddy/spa/%v\n", configuration.EncodeName(client.Name), configuration.EncodeName(client.Name))
		}

		dockerFile += "ENV root_path=/etc/caddy\n"

		dockerFile += "\n"
	}

	if err := helper.WriteToFile(dockerFile, filePath); err != nil {
		return err
	}

	return nil
}

func (svc *CaddyService) generateDockerComposeFile() error {
	filePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, "docker-compose.yml")
	notify.Wrench("Generating docker compose file")
	dockerComposeFile := "version: '3.7'\n"
	dockerComposeFile += "name: locally\n"
	dockerComposeFile += "services:\n"
	dockerComposeFile += "  locally:\n"
	dockerComposeFile += "    image: ${DOCKER_REGISTRY-}locally\n"
	dockerComposeFile += "    ports:\n"
	dockerComposeFile += "      - 80:80\n"
	dockerComposeFile += "      - 443:443\n"
	dockerComposeFile += "    build:\n"
	dockerComposeFile += "      context: '.'\n"
	dockerComposeFile += "      dockerfile: 'dockerfile'\n"

	if err := helper.WriteToFile(dockerComposeFile, filePath); err != nil {
		return err
	}

	return nil
}

func (svc *CaddyService) generateMainCaddyFile() error {
	filePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH, "Caddyfile")
	notify.Wrench("Generating main Caddyfile")
	caddyFile := "{\n"
	caddyFile += "  debug\n"
	caddyFile += "}\n"
	caddyFile += "\n"
	caddyFile += "(cors_headers) {\n"
	caddyFile += "  @origin{args.0} header Origin {args.0}\n"
	caddyFile += "  header @origin{args.0} ?Access-Control-Allow-Origin \"{args.0}\" defer\n"
	caddyFile += fmt.Sprintf("  header @origin{args.0} ?Access-Control-Allow-Methods \"%s\" defer\n", config.GlobalConfiguration.Cors.AllowedMethods)
	caddyFile += fmt.Sprintf("  header @origin{args.0} ?Access-Control-Allow-Headers \"%s\" defer\n", config.GlobalConfiguration.Cors.AllowedHeaders)
	caddyFile += "  header @origin{args.0} ?Vary Origin defer\n"
	caddyFile += "  header @origin{args.0} ?Access-Control-Allow-Credentials true defer\n"
	caddyFile += "}\n"
	caddyFile += "\n"
	caddyFile += fmt.Sprintf("import {$root_path}/%s/%s/*\n", common.CADDY_ROOT_SERVICES_PATH, common.CADDY_ROOT_SERVICES_HOSTS_PATH)
	caddyFile += "\n"
	caddyFile += fmt.Sprintf("%v.%v {\n", common.ExtractUri(config.GetCurrentContext().Configuration.RootURI), config.GlobalConfiguration.Network.DomainName)
	if config.GlobalConfiguration.Network != nil && config.GlobalConfiguration.Network.CERTPath != "" && config.GlobalConfiguration.Network.PrivateKeyPath != "" {
		certPath := filepath.Base(config.GlobalConfiguration.Network.CERTPath)
		privKeyPath := filepath.Base(config.GlobalConfiguration.Network.PrivateKeyPath)
		caddyFile += fmt.Sprintf("  tls {$root_path}/%v/%v {$root_path}/%v/%v\n", common.TLS_PATH, certPath, common.TLS_PATH, privKeyPath)
	}
	caddyFile += "\n"
	caddyFile += fmt.Sprintf("  import cors_headers https://%v.%v\n", config.GetCurrentContext().Configuration.RootURI, config.GlobalConfiguration.Network.DomainName)
	for _, tenant := range config.GetCurrentContext().Tenants {
		caddyFile += fmt.Sprintf("  import cors_headers https://%v.%v\n", tenant.URI, config.GlobalConfiguration.Network.DomainName)
	}

	for _, origin := range config.GlobalConfiguration.Cors.AllowedOrigins {
		if !strings.HasPrefix(origin, "http://") || !strings.HasPrefix(origin, "https://") {
			origin = fmt.Sprintf("https://%s", origin)
		}

		caddyFile += fmt.Sprintf("  import cors_headers %v\n", origin)
	}

	for _, webClient := range config.GetCurrentContext().SpaServices {
		if webClient.URI != "" && webClient.URI != "." {
			caddyFile += fmt.Sprintf("  import cors_headers https://%v.%v\n", webClient.URI, config.GlobalConfiguration.Network.DomainName)
		}
	}

	// Forcing locally to take ownership of the CORS response
	caddyFile += "\n"
	caddyFile += "  @options {\n"
	caddyFile += "    method OPTIONS\n"
	caddyFile += "  }\n"
	caddyFile += "\n"
	caddyFile += "  handle @options {\n"
	caddyFile += "    header ?Access-Control-Allow-Origin \"*\" defer\n"
	caddyFile += "    header ?Access-Control-Allow-Methods \"OPTIONS,HEAD,GET,POST,PUT,PATCH,DELETE\" defer\n"
	caddyFile += "    header ?Access-Control-Allow-Headers \"*\" defer\n"
	caddyFile += "    \n"
	caddyFile += "    respond 200\n"
	caddyFile += "  }\n"
	caddyFile += "\n"

	// Adding the mocker routes into the default caddy file
	caddyFile += fmt.Sprintf("  import {$root_path}/%s/*\n", common.CADDY_MOCK_ROUTES_PATH)

	caddyFile += fmt.Sprintf("  import {$root_path}/%s/%s/*\n", common.CADDY_ROOT_SERVICES_PATH, common.CADDY_ROOT_SERVICES_ROUTES_PATH)
	// caddyFile += fmt.Sprintf("  import {$root_path}/%s/*\n", common.CADDY_ROOT_SERVICES_PATH)
	caddyFile += "\n"
	defaultSpa := svc.getDefaultSpaService()
	if defaultSpa == nil {
		caddyFile += "  handle * {\n"
		caddyFile += "    respond 404\n"
		caddyFile += "  }\n"
	} else {
		if defaultSpa.UseReverseProxy && defaultSpa.ReverseProxyURI != "" {
			caddyFile += "  handle * {\n"
			caddyFile += fmt.Sprintf("    reverse_proxy %v\n", defaultSpa.ReverseProxyURI)
			caddyFile += "  }\n"
		} else {
			caddyFile += "  handle * {\n"
			caddyFile += "    respond 404\n"
			caddyFile += "  }\n"
		}
	}
	caddyFile += "}\n"
	caddyFile += "\n"
	caddyFile += fmt.Sprintf("import {$root_path}/%s/*\n", common.CADDY_TENANTS_PATH)
	caddyFile += fmt.Sprintf("import {$root_path}/%s/*\n", common.CADDY_UI_PATH)
	caddyFile += fmt.Sprintf("import {$root_path}/%s/*\n", common.CADDY_HOSTED_SERVICES_PATH)
	caddyFile += "\n"
	context := config.GetCurrentContext()
	caddyFile += fmt.Sprintf("%s {\n", context.Configuration.LocallyConfigService.Url)
	caddyFile += "  handle * {\n"
	caddyFile += fmt.Sprintf("    reverse_proxy %s\n", context.Configuration.LocallyConfigService.ReverseProxyUrl)
	caddyFile += "  }\n"
	caddyFile += "}\n"

	if err := helper.WriteToFile(caddyFile, filePath); err != nil {
		return err
	}

	return nil
}

func (svc *CaddyService) copyCertificates() error {
	if config.GlobalConfiguration.Network != nil && config.GlobalConfiguration.Network.CERTPath != "" && config.GlobalConfiguration.Network.PrivateKeyPath != "" {
		basePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.CADDY_PATH, common.TLS_PATH)
		certPath := filepath.Base(config.GlobalConfiguration.Network.CERTPath)
		privKeyPath := filepath.Base(config.GlobalConfiguration.Network.PrivateKeyPath)

		notify.InfoWithIcon(icons.IconClipboard, "Starting to copy SSL Certificate %v to the build folder", config.GlobalConfiguration.Network.CERTPath)
		if err := helper.CopyFile(config.GlobalConfiguration.Network.CERTPath, helper.JoinPath(basePath, certPath)); err != nil {
			notify.FromError(err, "Could not copy SSL Certificate %v to build folder", config.GlobalConfiguration.Network.CERTPath)
			return err
		}
		if config.Verbose() {
			notify.InfoWithIcon(icons.IconCheckMark, "Finished copying the SSL Certificate %v to the build folder", config.GlobalConfiguration.Network.CERTPath)
		}

		notify.InfoWithIcon(icons.IconClipboard, "Starting to copy SSL Certificate Private Key %v to the build folder", config.GlobalConfiguration.Network.PrivateKeyPath)
		if err := helper.CopyFile(config.GlobalConfiguration.Network.PrivateKeyPath, helper.JoinPath(basePath, privKeyPath)); err != nil {
			notify.FromError(err, "Could not copy SL Certificate Private Key %v to build folder", config.GlobalConfiguration.Network.PrivateKeyPath)
			return err
		}
		if config.Verbose() {
			notify.InfoWithIcon(icons.IconCheckMark, "Finished copying the SL Certificate Private Key %v to the build folder", config.GlobalConfiguration.Network.PrivateKeyPath)
		}
	}
	return nil
}

func (svc *CaddyService) generateServiceMockRouteFragment(name string, mockRoutes []*configuration.MockRoute) (string, error) {
	notify.Wrench("Generating service %s mock route fragment", name)
	caddyFragment := ""
	for _, mockRoute := range mockRoutes {
		notify.Debug(mockRoute.Regex)
		if mockRoute.Regex != "" {
			routeEndpoint := configuration.EncodeName(mockRoute.Name)
			caddyFragment += "  \n"
			caddyFragment += fmt.Sprintf("  @%v_%v_endpoint {\n", name, routeEndpoint)
			caddyFragment += fmt.Sprintf("    path_regexp %v\n", mockRoute.Regex)
			caddyFragment += "  }\n"
			caddyFragment += "  \n"
			caddyFragment += fmt.Sprintf("  handle @%v_%v_endpoint {\n", name, routeEndpoint)
			for _, headerMap := range mockRoute.Headers {
				for header, value := range headerMap {
					caddyFragment += fmt.Sprintf("    header %v \"%v\"\n", header, value)
				}
			}

			if mockRoute.Responds.ContentType != "" {
				caddyFragment += fmt.Sprintf("    header Content-Type \"%v\"\n", mockRoute.Responds.ContentType)
			}

			caddyFragment += "  \n"
			if mockRoute.Responds.RawBody != "" {
				caddyFragment += fmt.Sprintf("    respond %q\n", mockRoute.Responds.RawBody)

			} else if mockRoute.Responds.Body != nil {
				body, err := json.Marshal(mockRoute.Responds.Body)

				if err != nil {
					return "", err
				}

				caddyFragment += fmt.Sprintf("    respond %q\n", body)

			} else {
				caddyFragment += "    respond \"OK\"\n"
			}
			caddyFragment += "  }\n"
			caddyFragment += "  \n"
		}
	}

	if config.Verbose() {
		notify.InfoWithIcon(icons.IconCheckMark, "Finished generating service %s mock route fragment", name)
	}
	return caddyFragment, nil
}

func (svc *CaddyService) updateWebClientEnvironment(webClient *configuration.SpaService) error {
	// logger.Info("Updating WebClient Environment.json file")
	// if webClient.EnvironmentPath == "" {
	// 	return nil
	// }

	// baseUrl := fmt.Sprintf("https://%v.%v", config.GetCurrentLandscape().Configuration.RootURI, config.GetCurrentLandscape().Configuration.Domain)
	// basePath := helper.JoinPath(webClient.Path)
	// filePath := helper.JoinPath(basePath, webClient.EnvironmentPath)
	// if !helper.FileExists(filePath) {
	// 	return fmt.Errorf("file %v not found", filePath)
	// }

	// content, err := helper.ReadFromFile(filePath)
	// if err != nil {
	// 	return err
	// }

	// var environment map[string]interface{}
	// if err := json.Unmarshal(content, &environment); err != nil {
	// 	return err
	// }

	// // Replacing the keys in the current environment to replace localhost common entries
	// for key, val := range environment {
	// 	switch v := val.(type) {
	// 	case string:
	// 		v = strings.ReplaceAll(v, "http://localhost", baseUrl)
	// 		v = strings.ReplaceAll(v, "https://localhost", baseUrl)
	// 		environment[key] = v
	// 	}
	// }

	// for _, keyVal := range webClient.Environment {
	// 	for key, val := range keyVal {
	// 		environment[key] = val
	// 	}
	// }

	// updatedEnvironment, err := json.MarshalIndent(environment, "", "  ")
	// if err != nil {
	// 	return err
	// }

	// if err := helper.WriteToFile(string(updatedEnvironment), filePath); err != nil {
	// 	return err
	// }

	return nil
}
