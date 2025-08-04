package docker

// import (
// 	"context"
// 	"encoding/base64"
// 	"encoding/json"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"io/fs"
// 	"net/http"
// 	"net/url"
// 	"os"
// 	"path/filepath"
// 	"regexp"
// 	"sort"
// 	"strconv"
// 	"strings"
// 	"time"

// 	"github.com/cjlapao/locally-cli/internal/azure_cli"
// 	"github.com/cjlapao/locally-cli/internal/common"
// 	"github.com/cjlapao/locally-cli/internal/configuration"
// 	"github.com/cjlapao/locally-cli/internal/context/docker_component"
// 	"github.com/cjlapao/locally-cli/internal/context/entities"
// 	"github.com/cjlapao/locally-cli/internal/context/git_component"
// 	"github.com/cjlapao/locally-cli/internal/environment"
// 	"github.com/cjlapao/locally-cli/internal/executer"
// 	"github.com/cjlapao/locally-cli/internal/git"
// 	"github.com/cjlapao/locally-cli/internal/helpers"
// 	"github.com/cjlapao/locally-cli/internal/icons"
// 	"github.com/cjlapao/locally-cli/internal/mappers"
// 	"github.com/cjlapao/locally-cli/internal/notifications"

// 	"github.com/cjlapao/common-go/helper"
// )

// var globalDockerService *DockerService

// const (
// 	DockerComposeVersion string = "3.7"
// )

// type DockerService struct {
// 	notify  *notifications.NotificationsService
// 	wrapper *DockerCommandWrapper
// }

// type DockerServiceOptions struct {
// 	Name              string
// 	ComponentName     string
// 	BuildDependencies bool
// 	RootFolder        string
// 	Path              string
// 	DockerRegistry    *docker_component.DockerRegistry
// 	DockerCompose     *docker_component.DockerCompose
// 	StdOutput         bool
// }

// func New() *DockerService {
// 	svc := DockerService{
// 		wrapper: GetWrapper(),
// 		notify:  notifications.New(ServiceName),
// 	}

// 	return &svc
// }

// func Get() *DockerService {
// 	if globalDockerService != nil {
// 		return globalDockerService
// 	}

// 	return New()
// }

// func (svc *DockerService) CheckForDocker(softFail bool) {
// 	config = configuration.Get()
// 	if !config.GlobalConfiguration.Tools.Checked.DockerChecked {
// 		notify.InfoWithIcon(icons.IconFlag, "Checking for docker tool in the system")
// 		if output, err := executer.ExecuteWithNoOutput(helpers.GetDockerPath(), "version", "-f", "json"); err != nil {
// 			if !softFail {
// 				notify.Error("Docker tool not found in system, this is required for the selected function")
// 				os.Exit(1)
// 			} else {
// 				notify.Warning("Docker tool not found in system, this might generate an error in the future")
// 			}
// 		} else {
// 			var jOutput DockerVersion
// 			if err := json.Unmarshal([]byte(output.StdOut), &jOutput); err != nil {
// 				if !softFail {
// 					notify.Error("Docker tool not found in system, this is required for the selected function")
// 					os.Exit(1)
// 				} else {
// 					notify.Warning("Docker tool not found in system, this might generate an error in the future")
// 				}
// 			}

// 			notify.Success("Docker tool found with client version %s and server version %s", jOutput.Client.Version, jOutput.Server.APIVersion)
// 		}
// 		config.GlobalConfiguration.Tools.Checked.DockerChecked = true
// 	}
// }

// func (svc *DockerService) CheckForDockerCompose(softFail bool) {
// 	config = configuration.Get()
// 	if !config.GlobalConfiguration.Tools.Checked.DockerComposeChecked {

// 		notify.InfoWithIcon(icons.IconFlag, "Checking for docker compose tool in the system")
// 		if output, err := executer.ExecuteWithNoOutput(helpers.GetDockerComposePath(), "version", "-f", "json"); err != nil {
// 			if !softFail {
// 				notify.Error("Docker compose tool not found in system, this is required for the selected function")
// 				os.Exit(1)
// 			} else {
// 				notify.Warning("Docker compose tool not found in system, this might generate an error in the future")
// 			}
// 		} else {
// 			var jOutput dockerComposeVersion
// 			if err := json.Unmarshal([]byte(output.StdOut), &jOutput); err != nil {
// 				if !softFail {
// 					notify.Error("Docker compose tool not found in system, this is required for the selected function")
// 					os.Exit(1)
// 				} else {
// 					notify.Warning("Docker compose tool not found in system, this might generate an error in the future")
// 				}
// 			}

// 			notify.Success("Docker compose tool found with version %s", jOutput.Version)
// 		}
// 		config.GlobalConfiguration.Tools.Checked.DockerComposeChecked = true
// 	}
// }

// func (svc *DockerService) BuildServiceContainer(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		if container.DockerRegistry != nil &&
// 			container.DockerRegistry.Enabled &&
// 			container.DockerCompose != nil {
// 			notify.Warning("Service %s has a docker compose configuration and will not be built", options.Name)
// 			continue
// 		}

// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		if err := svc.CheckForDockerComposeFile(container, containerPath); err != nil {
// 			return err
// 		}

// 		if container.DockerRegistry != nil && container.DockerRegistry.Enabled {
// 			if err := svc.wrapper.Login(container.DockerRegistry.Registry, container.DockerRegistry.Credentials.Username, container.DockerRegistry.Credentials.Password, container.DockerRegistry.Credentials.SubscriptionId, container.DockerRegistry.Credentials.TenantId); err != nil {
// 				return err
// 			}
// 		}

// 		serviceFound = true
// 		wrapper := GetWrapper()
// 		if err := wrapper.Build(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) RebuildServiceContainer(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container dependencies")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		if (container.Location != nil || container.Location.RootFolder != "") &&
// 			container.DockerCompose != nil && container.DockerCompose.Location != "" {
// 			notify.Warning("Service %s has a docker compose configuration and will not be built", options.Name)
// 			continue
// 		}

// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		getLatest := helper.GetFlagSwitch("get-latest", false)
// 		if getLatest && container.DockerRegistry != nil && container.DockerRegistry.Enabled {
// 			basePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.DOCKER_COMPOSE_PATH)
// 			serviceFolder := helper.JoinPath(basePath, common.EncodeName(container.Name))
// 			helper.DeleteAllFiles(serviceFolder)
// 		}

// 		if err := svc.CheckForDockerComposeFile(container, containerPath); err != nil {
// 			return err
// 		}

// 		if container.DockerRegistry != nil && container.DockerRegistry.Enabled {
// 			if err := svc.wrapper.Login(container.DockerRegistry.Registry, container.DockerRegistry.Credentials.Username, container.DockerRegistry.Credentials.Password, container.DockerRegistry.Credentials.SubscriptionId, container.DockerRegistry.Credentials.TenantId); err != nil {
// 				return err
// 			}
// 		}

// 		serviceFound = true
// 		wrapper := GetWrapper()
// 		if err := wrapper.Build(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}

// 		if err := svc.ServiceContainerUp(options); err != nil {
// 			return err
// 		}
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) ServiceContainerUp(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		getLatest := helper.GetFlagSwitch("get-latest", false)
// 		if getLatest && container.DockerRegistry != nil && container.DockerRegistry.Enabled {
// 			notify.Info("Removing generated docker compose to get latest...")
// 			basePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.DOCKER_COMPOSE_PATH)
// 			serviceFolder := helper.JoinPath(basePath, common.EncodeName(container.Name))
// 			helper.DeleteAllFiles(serviceFolder)
// 		}

// 		if err := svc.CheckForDockerComposeFile(container, containerPath); err != nil {
// 			return err
// 		}

// 		if container.DockerRegistry != nil && container.DockerRegistry.Enabled {
// 			if err := svc.wrapper.Login(container.DockerRegistry.Registry, container.DockerRegistry.Credentials.Username, container.DockerRegistry.Credentials.Password, container.DockerRegistry.Credentials.SubscriptionId, container.DockerRegistry.Credentials.TenantId); err != nil {
// 				return err
// 			}
// 		}

// 		serviceFound = true
// 		wrapper := GetWrapper()
// 		if err := wrapper.Up(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) ServiceContainerDown(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		if err := svc.CheckForDockerComposeFile(container, containerPath); err != nil {
// 			return err
// 		}

// 		if container.DockerRegistry != nil && container.DockerRegistry.Enabled {
// 			if err := svc.wrapper.Login(container.DockerRegistry.Registry, container.DockerRegistry.Credentials.Username, container.DockerRegistry.Credentials.Password, container.DockerRegistry.Credentials.SubscriptionId, container.DockerRegistry.Credentials.TenantId); err != nil {
// 				return err
// 			}
// 		}

// 		serviceFound = true
// 		wrapper := GetWrapper()

// 		images, err := wrapper.GetServiceImages(containerPath, container.Name, options.ComponentName)
// 		if err != nil {
// 			return err
// 		}

// 		if err := wrapper.Down(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}

// 		if len(images) == 0 {
// 			notify.Warning("No image was found to delete, there should be at least one")
// 		}

// 		for _, image := range images {
// 			if err := wrapper.RemoveImage(image.Repository, image.Tag); err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) StartServiceContainer(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		if err := svc.CheckForDockerComposeFile(container, containerPath); err != nil {
// 			return err
// 		}

// 		if container.DockerRegistry != nil && container.DockerRegistry.Enabled {
// 			if err := svc.wrapper.Login(container.DockerRegistry.Registry, container.DockerRegistry.Credentials.Username, container.DockerRegistry.Credentials.Password, container.DockerRegistry.Credentials.SubscriptionId, container.DockerRegistry.Credentials.TenantId); err != nil {
// 				return err
// 			}
// 		}

// 		serviceFound = true
// 		wrapper := GetWrapper()
// 		if err := wrapper.Start(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) StopServiceContainer(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		serviceFound = true

// 		wrapper := GetWrapper()
// 		if err := wrapper.Stop(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) PauseServiceContainer(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		serviceFound = true
// 		wrapper := GetWrapper()
// 		if err := wrapper.Pause(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) ResumeServiceContainer(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		serviceFound = true
// 		wrapper := GetWrapper()
// 		if err := wrapper.Resume(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) ServiceContainerStatus(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		serviceFound = true
// 		wrapper := GetWrapper()
// 		if err := wrapper.Status(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) ListServiceContainer(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	wrapper := GetWrapper()

// 	if options.Name == "" {
// 		if err := wrapper.List(""); err != nil {
// 			return err
// 		}
// 		return nil
// 	}

// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, service := range containers {
// 		if err := wrapper.List(service.Name); err != nil {
// 			return err
// 		}
// 		serviceFound = true
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) ServiceContainerLogs(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, true)

// 	for _, container := range containers {
// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		wrapper := GetWrapper()
// 		if err := wrapper.Logs(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}
// 		serviceFound = true
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) GenerateServiceDockerComposeOverrideFile(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		filePath := helper.JoinPath(containerPath, "docker-compose.override.yaml")
// 		notify.Wrench("Generating docker compose overwrite file for %v service", container.Name)
// 		dockerComposeFile := fmt.Sprintf("version: '%s'\n", DockerComposeVersion)
// 		dockerComposeFile += fmt.Sprintf("name: %s\n", container.Name)
// 		dockerComposeFile += "services:\n"
// 		if len(container.Components) > 0 {
// 			for _, component := range container.Components {
// 				dockerComposeFile += svc.generateDockerComposeServiceOverride(component)
// 			}
// 		} else {
// 			dockerComposeFile += svc.generateDockerComposeServiceOverride(container)
// 		}

// 		if err := helper.WriteToFile(dockerComposeFile, filePath); err != nil {
// 			return err
// 		}
// 		serviceFound = true
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) GenerateServiceDockerComposeFile(options *DockerServiceOptions) error {
// 	var dependencyError error

// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, true)

// 	containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 	if dependencyError != nil {
// 		notify.FromError(dependencyError, "Building docker container")
// 		return dependencyError
// 	}

// 	basePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.DOCKER_COMPOSE_PATH)
// 	if !helper.FileExists(basePath) {
// 		notify.Hammer("Creating %s folder", basePath)
// 		if !helper.CreateDirectory(basePath, fs.ModePerm) {
// 			err := fmt.Errorf("error creating the %v folder", basePath)
// 			return err
// 		}
// 	}

// 	for _, container := range containers {
// 		if options.DockerCompose == nil && container.DockerCompose == nil {
// 			err := errors.New("docker compose is nil")
// 			notify.Error(err.Error())
// 			return err
// 		}

// 		if options.DockerCompose != nil {
// 			container.DockerCompose = options.DockerCompose
// 		}
// 		if options.DockerRegistry != nil {
// 			container.DockerRegistry.Clone(options.DockerRegistry, false)
// 		}

// 		notify.Debug("registry: %s,  Manifest Path: %s", container.DockerRegistry.Registry, container.DockerRegistry.ManifestPath)

// 		serviceFolder := helper.JoinPath(basePath, common.EncodeName(container.Name))
// 		if !helper.FileExists(serviceFolder) {
// 			notify.Hammer("Creating %s folder", serviceFolder)
// 			if !helper.CreateDirectory(serviceFolder, fs.ModePerm) {
// 				err := fmt.Errorf("error creating the %v folder", serviceFolder)
// 				return err
// 			}
// 		}

// 		dockerComposePath := helper.JoinPath(serviceFolder, "docker-compose")
// 		if helper.FileExists(fmt.Sprintf("%s.yml", dockerComposePath)) {
// 			if err := helper.DeleteFile(fmt.Sprintf("%s.yml", dockerComposePath)); err != nil {
// 				return err
// 			}
// 			notify.InfoWithIcon(icons.IconBomb, "Docker compose found in the repo folder, removing")
// 		}
// 		if helper.FileExists(fmt.Sprintf("%s.yaml", dockerComposePath)) {
// 			if err := helper.DeleteFile(fmt.Sprintf("%s.yaml", dockerComposePath)); err != nil {
// 				return err
// 			}
// 			notify.InfoWithIcon(icons.IconBomb, "Docker compose found in the repo folder, removing")
// 		}

// 		compose := svc.generateContainerDockerCompose(container)
// 		svc.CheckComposeFolder(container)
// 		if err := helper.WriteToFile(compose, fmt.Sprintf("%s.yaml", dockerComposePath)); err != nil {
// 			return err
// 		}

// 		container.DockerCompose.Location = fmt.Sprintf("%s.yaml", dockerComposePath)

// 		serviceFound = true
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) DeleteImage(options *DockerServiceOptions) error {
// 	serviceFound := false
// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, !options.BuildDependencies)

// 	// Building dependencies
// 	if options.BuildDependencies {
// 		var dependencyError error
// 		containers, dependencyError = config.GetDockerContainerDependencies(containers)
// 		if dependencyError != nil {
// 			notify.FromError(dependencyError, "Building docker container")
// 			return dependencyError
// 		} else {
// 			if options.Name == "" {
// 				notify.Hammer("Container dependencies were built successfully")
// 			} else {
// 				notify.Hammer("Container %s dependencies were built successfully", options.Name)
// 			}
// 		}
// 	}

// 	for _, container := range containers {
// 		containerPath, pathError := svc.getPath(container, options)
// 		if pathError != nil {
// 			return pathError
// 		}
// 		notify.Debug("Using Path: %s", containerPath)

// 		serviceFound = true
// 		wrapper := GetWrapper()
// 		images, err := wrapper.GetServiceImages(containerPath, container.Name, options.ComponentName)
// 		if err != nil {
// 			return err
// 		}

// 		if err := wrapper.Stop(containerPath, container.Name, options.ComponentName); err != nil {
// 			return err
// 		}

// 		if len(images) == 0 {
// 			notify.Warning("No image was found to delete, there should be at least one")
// 		}

// 		for _, image := range images {
// 			if err := wrapper.RemoveImage(image.Repository, image.Tag); err != nil {
// 				return err
// 			}
// 		}
// 	}

// 	if !serviceFound {
// 		return fmt.Errorf("service %v was not found in the configuration file", options.Name)
// 	}

// 	return nil
// }

// func (svc *DockerService) RunEfMigrations(context, baseImage, imageName, repoUrl, projectPath,
// 	startupProjectPath string, dbConnectionString string, arguments, environmentVars map[string]string,
// ) error {
// 	wrapper := GetWrapper()

// 	if imageName == "" {
// 		imageName = "test.image"
// 	}

// 	if projectPath == "" {
// 		return errors.New("projectPath cannot be empty")
// 	}

// 	args := make([]string, 0)
// 	for key, value := range arguments {
// 		notify.Debug("%s: %s", key, value)
// 		args = append(args, key)
// 	}

// 	dockerFile := svc.generateEfMigrationDockerFile(baseImage, imageName, repoUrl, projectPath, startupProjectPath, dbConnectionString, args, environmentVars)
// 	dockerComposeFile := svc.generateEfMigrationDockerComposeFile(imageName, imageName, environmentVars)

// 	pipelinesFolder := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.PIPELINES_PATH)
// 	if !helper.FileExists(pipelinesFolder) {
// 		notify.Hammer("Creating %s folder", pipelinesFolder)
// 		if !helper.CreateDirectory(pipelinesFolder, fs.ModePerm) {
// 			err := fmt.Errorf("error creating the %v folder", pipelinesFolder)
// 			return err
// 		}
// 	}

// 	migrationFolder := helper.JoinPath(pipelinesFolder, imageName)
// 	if !helper.FileExists(migrationFolder) {
// 		notify.Hammer("Creating %s folder", migrationFolder)
// 		if !helper.CreateDirectory(migrationFolder, fs.ModePerm) {
// 			err := fmt.Errorf("error creating the %v folder", migrationFolder)
// 			return err
// 		}
// 	}

// 	dockerFilePath := helper.JoinPath(migrationFolder, fmt.Sprintf("%s.dockerfile", common.EncodeName(imageName)))
// 	helper.WriteToFile(dockerFile, dockerFilePath)
// 	notify.Debug("docker file %s\ncontent:\n %s", dockerFilePath, dockerFile)

// 	dockerComposeFilePath := helper.JoinPath(migrationFolder, "docker-compose.yaml")
// 	helper.WriteToFile(dockerComposeFile, dockerComposeFilePath)
// 	notify.Debug("docker compose file %s\ncontent:\n %s", dockerComposeFilePath, dockerComposeFile)

// 	options := BuildImageOptions{
// 		Name:       imageName,
// 		Tag:        "latest",
// 		Context:    context,
// 		Parameters: arguments,
// 		FilePath:   dockerFilePath,
// 		UseCache:   false,
// 	}

// 	notify.Debug("Using options: %v", options)
// 	err := wrapper.BuildImage(options)
// 	if err != nil {
// 		return err
// 	}

// 	if err := wrapper.Up(migrationFolder, imageName, ""); err != nil {
// 		return err
// 	}

// 	notify.Rocket("Checking if image %v is running", imageName)

// 	count := 0
// 	for {
// 		count += 1

// 		r, err := wrapper.IsRunning(imageName)
// 		if err != nil {
// 			return err
// 		}
// 		if !r {
// 			break
// 		}

// 		duration := (time.Second * 10) * time.Duration(count)
// 		notify.Info("Migrations %s are still running, waiting.. [%s]", imageName, duration)
// 		time.Sleep(time.Second * 10)

// 		if count > 30 {
// 			notify.Error("Migrations are taking longer than expected")
// 			break
// 		}
// 	}

// 	if !common.IsDebug() {
// 		if err := wrapper.Down(migrationFolder, imageName, ""); err != nil {
// 			return err
// 		}

// 		if err := wrapper.RemoveImage(options.Name, options.Tag); err != nil {
// 			return err
// 		}

// 		if err := helper.DeleteFile(dockerFilePath); err != nil {
// 			return err
// 		}
// 		if err := helper.DeleteFile(dockerComposeFilePath); err != nil {
// 			return err
// 		}
// 		if err := helper.DeleteFile(migrationFolder); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (svc *DockerService) RunDotnetContainer(command, context, baseImage, imageName, repoUrl, projectPath string, arguments, environmentVars map[string]string, buildArguments []string) error {
// 	wrapper := GetWrapper()

// 	if command == "" {
// 		return errors.New("command cannot be empty")
// 	}
// 	if imageName == "" {
// 		imageName = "dotnet.image"
// 	}

// 	imageName = fmt.Sprintf("dotnet_%s", imageName)
// 	if projectPath == "" {
// 		return errors.New("projectPath cannot be empty")
// 	}

// 	args := make([]string, 0)
// 	for key, value := range arguments {
// 		notify.Debug("%s: %s", key, value)
// 		args = append(args, key)
// 	}

// 	dockerFile := svc.generateDotnetContainerDockerFile(command, baseImage, imageName, repoUrl, projectPath, args, buildArguments, environmentVars)
// 	dockerComposeFile := svc.generateDotnetContainerComposeFile(imageName, imageName, environmentVars)

// 	pipelinesFolder := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.PIPELINES_PATH)
// 	if !helper.FileExists(pipelinesFolder) {
// 		notify.Hammer("Creating %s folder", pipelinesFolder)
// 		if !helper.CreateDirectory(pipelinesFolder, fs.ModePerm) {
// 			err := fmt.Errorf("error creating the %v folder", pipelinesFolder)
// 			return err
// 		}
// 	}

// 	runFolder := helper.JoinPath(pipelinesFolder, imageName)
// 	if !helper.FileExists(runFolder) {
// 		notify.Hammer("Creating %s folder", runFolder)
// 		if !helper.CreateDirectory(runFolder, fs.ModePerm) {
// 			err := fmt.Errorf("error creating the %v folder", runFolder)
// 			return err
// 		}
// 	}

// 	dockerFilePath := helper.JoinPath(runFolder, fmt.Sprintf("%s.dockerfile", common.EncodeName(imageName)))
// 	helper.WriteToFile(dockerFile, dockerFilePath)
// 	notify.Debug("docker file %s\ncontent:\n %s", dockerFilePath, dockerFile)

// 	dockerComposeFilePath := helper.JoinPath(runFolder, "docker-compose.yaml")
// 	helper.WriteToFile(dockerComposeFile, dockerComposeFilePath)
// 	notify.Debug("docker compose file %s\ncontent:\n %s", dockerComposeFilePath, dockerComposeFile)

// 	options := BuildImageOptions{
// 		Name:       imageName,
// 		Tag:        "latest",
// 		Context:    context,
// 		Parameters: arguments,
// 		FilePath:   dockerFilePath,
// 		UseCache:   false,
// 	}

// 	notify.Debug("Using options: %v", options)
// 	err := wrapper.BuildImage(options)
// 	if err != nil {
// 		return err
// 	}

// 	if err := wrapper.Up(runFolder, imageName, ""); err != nil {
// 		return err
// 	}

// 	notify.Rocket("Checking if image %v is running", imageName)

// 	count := 0
// 	for {
// 		count += 1

// 		r, err := wrapper.IsRunning(imageName)
// 		if err != nil {
// 			return err
// 		}
// 		if !r {
// 			break
// 		}

// 		duration := (time.Second * 10) * time.Duration(count)
// 		notify.Info("Migrations %s are still running, waiting.. [%s]", imageName, duration)
// 		time.Sleep(time.Second * 10)

// 		if count > 30 {
// 			notify.Error("Migrations are taking longer than expected")
// 			break
// 		}
// 	}

// 	if !common.IsDebug() {
// 		if err := wrapper.Down(runFolder, imageName, ""); err != nil {
// 			return err
// 		}

// 		if err := wrapper.RemoveImage(options.Name, options.Tag); err != nil {
// 			return err
// 		}

// 		if err := helper.DeleteFile(dockerFilePath); err != nil {
// 			return err
// 		}

// 		if err := helper.DeleteFile(dockerComposeFilePath); err != nil {
// 			return err
// 		}

// 		if err := helper.DeleteFile(runFolder); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (svc *DockerService) PullImage(options *DockerServiceOptions) error {
// 	dockerWrapper := GetWrapper()
// 	env := environment.GetInstance()
// 	if options == nil {
// 		return errors.New("options cannot be nil")
// 	}

// 	if options.DockerRegistry == nil {
// 		return errors.New("docker registry cannot be nil")
// 	}

// 	registry := options.DockerRegistry

// 	notify.Debug(registry.Registry)
// 	notify.Debug(env.Replace(registry.Registry))
// 	registryName := common.GetHostFromUrl(env.Replace(registry.Registry))
// 	basePath := env.Replace(registry.BasePath)
// 	manifestPath := env.Replace(registry.ManifestPath)
// 	notify.Debug(basePath)
// 	notify.Debug(manifestPath)
// 	if basePath != "" {
// 		manifestPath = fmt.Sprintf("%s/%s", strings.Trim(basePath, "/"), strings.Trim(manifestPath, "/"))
// 	}
// 	notify.Debug(manifestPath)
// 	notify.Debug(registryName)
// 	username := ""
// 	password := ""
// 	subscriptionId := ""
// 	tenantId := ""
// 	if registry.Credentials != nil {
// 		username = env.Replace(registry.Credentials.Username)
// 		password = env.Replace(registry.Credentials.Password)
// 		subscriptionId = env.Replace(registry.Credentials.SubscriptionId)
// 		tenantId = env.Replace(registry.Credentials.TenantId)

// 		if err := dockerWrapper.Login(registryName, username, password, subscriptionId, tenantId); err != nil {
// 			return err
// 		}
// 	}

// 	if registry.Tag == "" {
// 		var imageErr error
// 		registry.Tag, imageErr = svc.GetLatestImageTag(registryName, manifestPath, username, password, subscriptionId, tenantId)
// 		if imageErr != nil {
// 			return imageErr
// 		}
// 	}

// 	ctx := config.GetCurrentContext()
// 	containers := ctx.GetDockerServices(options.Name, true)
// 	if len(containers) > 0 {
// 		containers[0].DockerRegistry.Clone(registry, false)
// 		// currentContext := config.GetCurrentContext()
// 		// fragment := currentContext.GetContainerFragmentByName(options.Name)
// 		// if fragment != nil {
// 		// currentContext.SaveFragment(fragment)
// 		// }
// 	}

// 	if err := dockerWrapper.Pull(registryName, manifestPath, registry.Tag); err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (worker DockerService) GetLatestImageTag(registry, imagePath, username, password, subscriptionId, tenantId string) (string, error) {
// 	env := environment.GetInstance()
// 	azureCli := azure_cli.Get()
// 	registry = env.Replace(registry)
// 	imagePath = env.Replace(imagePath)
// 	username = env.Replace(username)
// 	password = env.Replace(password)
// 	subscriptionId = env.Replace(subscriptionId)
// 	tenantId = env.Replace(tenantId)
// 	host := ""
// 	if !strings.HasPrefix(registry, "http://") && !strings.HasPrefix(registry, "https://") {
// 		host += "https://"
// 	}

// 	host += strings.Trim(registry, "/")
// 	host += fmt.Sprintf("/v2/%s/tags/list", strings.Trim(imagePath, "/"))

// 	notify.Debug("Registry: %s -> ImagePath %s", registry, imagePath)

// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)

// 	defer cancel()

// 	req, err := http.NewRequestWithContext(ctx, "GET", host, nil)
// 	if err != nil {
// 		return "", err
// 	}

// 	if username != "" && password != "" {
// 		if username == "00000000-0000-0000-0000-000000000000" {
// 			notify.Debug("Seems %s is an Azure ACR Oauth user, authenticating using oauth2 token", registry)
// 			token, err := azureCli.ExchangeRefreshTokenForAccessToken(registry, "", subscriptionId, tenantId)
// 			if err != nil {
// 				return "", err
// 			}
// 			req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))
// 			notify.Debug("Authentication Header: %s", fmt.Sprintf("Bearer %s", token))
// 		} else {
// 			auth := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", username, password)))
// 			req.Header.Add("Authorization", fmt.Sprintf("Basic %s", auth))
// 			notify.Debug("Added header Authorization: Basic %s ", auth)
// 		}
// 	}

// 	client := &http.Client{}

// 	notify.Debug("Trying to get the latest tags from %s", host)
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

// 	body, err := io.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", err
// 	}

// 	notify.Debug("Parsing the response body %s", fmt.Sprintf("%v", string(body)))
// 	var manifestTags docker_component.DockerRegistryTagList
// 	if err := json.Unmarshal(body, &manifestTags); err != nil {
// 		return "", err
// 	}

// 	for _, tag := range manifestTags.Tags {
// 		if strings.EqualFold(tag, "latest") {
// 			return "latest", nil
// 		}
// 	}

// 	sort.Slice(manifestTags.Tags, func(i, j int) bool {
// 		return manifestTags.Tags[i] > manifestTags.Tags[j]
// 	})

// 	notify.Debug("Ordered Tags: %s", fmt.Sprintf("%v", manifestTags.Tags))

// 	return manifestTags.Tags[0], nil
// }

// func (svc *DockerService) generateDockerComposeServiceOverride(component *docker_component.DockerContainer) string {
// 	env := environment.GetInstance()
// 	context := config.GetCurrentContext()
// 	notify.Hammer("Building %s container fragment", component.Name)
// 	caddyFragment := ""
// 	caddyFragment += fmt.Sprintf("  %v:\n", component.Name)
// 	caddyFragment += "    extra_hosts:\n"
// 	caddyFragment += fmt.Sprintf("      - %v.%v:host-gateway\n", config.GetCurrentContext().Configuration.RootURI, config.GlobalConfiguration.Network.DomainName)
// 	caddyFragment += "      - host.docker.internal:host-gateway\n"
// 	if context.Configuration.LocallyConfigService != nil {
// 		parsedUrl, err := url.Parse(context.Configuration.LocallyConfigService.Url)
// 		if err == nil {
// 			caddyFragment += fmt.Sprintf("      - %s:host-gateway\n", parsedUrl.Host)
// 		}
// 	}
// 	if len(component.BuildArguments) > 0 {
// 		caddyFragment += "    build:\n"
// 		caddyFragment += "      args:\n"
// 		for _, argument := range component.BuildArguments {
// 			val := argument
// 			val = env.Replace(val)
// 			val = strings.ReplaceAll(val, "\r\n", "")
// 			val = strings.ReplaceAll(val, "\n", "")
// 			val = strings.ReplaceAll(val, "$", "$$")

// 			if argument != "" {
// 				re := regexp.MustCompile("^[a-zA-Z0-9]{1}.*$")
// 				if !re.MatchString(val) {
// 					singleQuote := regexp.MustCompile("^.*['].*$")
// 					doubleQuote := regexp.MustCompile("^.*[\"].*$")
// 					if singleQuote.MatchString(val) && doubleQuote.MatchString(val) {
// 						val = strings.ReplaceAll(val, "'", "''")
// 						val = fmt.Sprintf("'%v'", val)
// 					} else if singleQuote.MatchString(val) {
// 						val = fmt.Sprintf("\"%v\"", val)
// 					} else if doubleQuote.MatchString(val) {
// 						val = fmt.Sprintf("'%v'", val)
// 					} else {
// 						val = fmt.Sprintf("'%v'", val)
// 					}
// 				}
// 				caddyFragment += fmt.Sprintf("        - %v\n", val)
// 			}
// 		}
// 	}
// 	if len(component.EnvironmentVariables) > 0 {
// 		caddyFragment += "    environment:\n"
// 		for key, value := range component.EnvironmentVariables {
// 			val := value
// 			val = env.Replace(val)
// 			val = strings.ReplaceAll(val, "\r\n", "")
// 			val = strings.ReplaceAll(val, "\n", "")
// 			val = strings.ReplaceAll(val, "$", "$$")

// 			if val != "" {
// 				re := regexp.MustCompile("^[a-zA-Z0-9]{1}.*$")
// 				if !re.MatchString(val) {
// 					singleQuote := regexp.MustCompile("^.*['].*$")
// 					doubleQuote := regexp.MustCompile("^.*[\"].*$")
// 					if singleQuote.MatchString(val) && doubleQuote.MatchString(val) {
// 						val = strings.ReplaceAll(val, "'", "''")
// 						val = fmt.Sprintf("'%v'", val)
// 					} else if singleQuote.MatchString(val) {
// 						val = strings.ReplaceAll(val, "\"", "\"")
// 						val = fmt.Sprintf("\"%v\"", val)
// 					} else if doubleQuote.MatchString(val) {
// 						val = strings.ReplaceAll(val, "'", "''")
// 						val = fmt.Sprintf("'%v'", val)
// 					} else {
// 						val = fmt.Sprintf("'%v'", val)
// 					}
// 				}
// 				caddyFragment += fmt.Sprintf("      %v: %v\n", key, val)
// 			}
// 		}
// 	}

// 	return caddyFragment
// }

// func (svc *DockerService) generateContainerDockerCompose(container *docker_component.DockerContainer) string {
// 	env := environment.GetInstance()

// 	dockerCompose := container.DockerCompose
// 	dockerRegistry := container.DockerRegistry
// 	version := container.DockerCompose.Version
// 	if version == "" {
// 		version = DockerComposeVersion
// 	}
// 	registry := dockerRegistry.Registry
// 	basePath := dockerRegistry.BasePath
// 	manifestPath := dockerRegistry.ManifestPath

// 	notify.Debug("T1: %s", basePath)
// 	notify.Debug("T2: %s", manifestPath)
// 	if basePath != "" {
// 		manifestPath = fmt.Sprintf("%s/%s", strings.Trim(basePath, "/"), strings.Trim(manifestPath, "/"))
// 	}

// 	notify.Hammer("Building %s container fragment", container.Name)
// 	dockerComposeFileContent := ""
// 	dockerComposeFileContent += fmt.Sprintf("version: '%s'\n", version)
// 	dockerComposeFileContent += fmt.Sprintf("name: %v\n", container.Name)
// 	dockerComposeFileContent += "services:\n"
// 	for _, component := range container.Components {
// 		dockerComposeService := dockerCompose.Services[component.Name]
// 		image := ""
// 		if dockerComposeService != nil && dockerComposeService.Image != "" {
// 			image = dockerComposeService.Image
// 		} else if image == "" && dockerRegistry != nil &&
// 			dockerRegistry.Enabled &&
// 			registry != "" &&
// 			manifestPath != "" {
// 			if component.ManifestPath != "" {
// 				image = fmt.Sprintf("%s/%s/%s", strings.Trim(registry, "/"), strings.Trim(basePath, "/"), strings.Trim(component.ManifestPath, "/"))
// 			} else {
// 				image = fmt.Sprintf("%s/%s", strings.Trim(registry, "/"), strings.Trim(manifestPath, "/"))
// 			}
// 			if container.ManifestTag != "" {
// 				image += fmt.Sprintf(":%s", container.ManifestTag)
// 			} else {
// 				imageManifestPath := strings.Trim(manifestPath, "/")
// 				if component.ManifestPath != "" {
// 					imageManifestPath = fmt.Sprintf("%s/%s", strings.Trim(basePath, "/"), strings.Trim(component.ManifestPath, "/"))
// 				}

// 				tag, err := svc.GetLatestImageTag(dockerRegistry.Registry, imageManifestPath, dockerRegistry.Credentials.Username, dockerRegistry.Credentials.Password, dockerRegistry.Credentials.SubscriptionId, dockerRegistry.Credentials.TenantId)
// 				if err == nil {
// 					image += fmt.Sprintf(":%s", tag)
// 					notify.Debug("Image tag set to %s", tag)
// 				} else {
// 					notify.Warning("Could not get the latest tag for %s on registry %s", container.Name, image)
// 				}
// 			}

// 			image = strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(env.Replace(image), "https://", ""), "http://", ""), "//", "/")
// 		} else {
// 			image = fmt.Sprintf("${DOCKER_REGISTRY-}%s", common.EncodeName(component.Name))
// 		}
// 		args := component.BuildArguments
// 		if dockerComposeService != nil && dockerComposeService.Build != nil && len(dockerComposeService.Build.Args) > 0 {
// 			args = dockerComposeService.Build.Args
// 		}

// 		envVars := component.EnvironmentVariables
// 		if dockerComposeService != nil && dockerComposeService.Build != nil && len(dockerComposeService.Environment) > 0 {
// 			envVars = map[string]string{}
// 			for key, val := range dockerComposeService.Environment {
// 				switch v := val.(type) {
// 				case string:
// 					envVars[key] = env.Replace(v)
// 				case int:
// 					envVars[key] = strconv.Itoa(v)
// 				case bool:
// 					envVars[key] = strconv.FormatBool(v)
// 				}
// 			}
// 		}

// 		dockerComposeFileContent += fmt.Sprintf("  %s:\n", env.Replace(component.Name))
// 		dockerComposeFileContent += fmt.Sprintf("    image: %s\n", strings.ToLower(env.Replace(image))) // convert image to lowercase as otherwise it might cause issues
// 		dockerComposeFileContent += "    extra_hosts:\n"
// 		dockerComposeFileContent += fmt.Sprintf("      - %v.%v:host-gateway\n", config.GetCurrentContext().Configuration.RootURI, config.GlobalConfiguration.Network.DomainName)
// 		dockerComposeFileContent += "      - host.docker.internal:host-gateway\n"
// 		context := config.GetCurrentContext()
// 		if context.Configuration.LocallyConfigService != nil {
// 			parsedUrl, err := url.Parse(context.Configuration.LocallyConfigService.Url)
// 			if err == nil {
// 				dockerComposeFileContent += fmt.Sprintf("      - %s:host-gateway\n", parsedUrl.Host)
// 			}
// 		}
// 		if len(args) > 0 {
// 			dockerComposeFileContent += "    build:\n"
// 			dockerComposeFileContent += "      args:\n"
// 			for _, arguments := range component.BuildArguments {
// 				val := arguments
// 				val = env.Replace(val)
// 				val = strings.ReplaceAll(val, "\r\n", "")
// 				val = strings.ReplaceAll(val, "\n", "")
// 				val = strings.ReplaceAll(val, "$", "$$")

// 				if val != "" {
// 					re := regexp.MustCompile("^[a-zA-Z0-9]{1}.*$")
// 					if !re.MatchString(val) {
// 						singleQuote := regexp.MustCompile("^.*['].*$")
// 						doubleQuote := regexp.MustCompile("^.*[\"].*$")
// 						if singleQuote.MatchString(val) && doubleQuote.MatchString(val) {
// 							val = strings.ReplaceAll(val, "'", "''")
// 							val = fmt.Sprintf("'%v'", val)
// 						} else if singleQuote.MatchString(val) {
// 							val = fmt.Sprintf("\"%v\"", val)
// 						} else if doubleQuote.MatchString(val) {
// 							val = fmt.Sprintf("'%v'", val)
// 						} else {
// 							val = fmt.Sprintf("'%v'", val)
// 						}
// 					}
// 					dockerComposeFileContent += fmt.Sprintf("        - %v\n", val)

// 				}
// 			}
// 		}

// 		if dockerComposeService != nil && len(dockerComposeService.Volumes) > 0 {
// 			dockerComposeFileContent += "    volumes:\n"
// 			for _, volume := range dockerComposeService.Volumes {
// 				dockerComposeFileContent += fmt.Sprintf("      - %s\n", env.Replace(volume))
// 			}
// 		}
// 		dockerComposeFileContent += "    networks:\n"
// 		dockerComposeFileContent += "      default:\n"
// 		dockerComposeFileContent += "        aliases:\n"
// 		dockerComposeFileContent += fmt.Sprintf("          - %s.dev\n", component.Name)

// 		if dockerComposeService != nil && len(dockerComposeService.Ports) > 0 {
// 			dockerComposeFileContent += "    ports:\n"
// 			for _, port := range dockerComposeService.Ports {
// 				dockerComposeFileContent += fmt.Sprintf("      - %s\n", env.Replace(port))
// 			}
// 		}

// 		if len(envVars) > 0 {
// 			dockerComposeFileContent += "    environment:\n"
// 			for key, value := range component.EnvironmentVariables {
// 				val := value
// 				val = env.Replace(val)
// 				val = strings.ReplaceAll(val, "\r\n", "")
// 				val = strings.ReplaceAll(val, "\n", "")
// 				val = strings.ReplaceAll(val, "$", "$$")

// 				if value != "" {
// 					re := regexp.MustCompile("^[a-zA-Z0-9]{1}.*$")
// 					if !re.MatchString(val) {
// 						singleQuote := regexp.MustCompile("^.*['].*$")
// 						doubleQuote := regexp.MustCompile("^.*[\"].*$")
// 						if singleQuote.MatchString(val) && doubleQuote.MatchString(val) {
// 							val = strings.ReplaceAll(val, "'", "''")
// 							val = fmt.Sprintf("'%v'", val)
// 						} else if singleQuote.MatchString(val) {
// 							val = strings.ReplaceAll(val, "\"", "\"")
// 							val = fmt.Sprintf("\"%v\"", val)
// 						} else if doubleQuote.MatchString(val) {
// 							val = strings.ReplaceAll(val, "'", "''")
// 							val = fmt.Sprintf("'%v'", val)
// 						} else {
// 							val = fmt.Sprintf("'%v'", val)
// 						}
// 					}
// 					dockerComposeFileContent += fmt.Sprintf("      %v: %v\n", key, val)
// 				}
// 			}
// 		}
// 	}

// 	return dockerComposeFileContent
// }

// func (svc *DockerService) generateEfMigrationDockerFile(baseImage, imageName, repoUrl, projectPath,
// 	startupProjectPath string, dbConnectionString string, arguments []string, environmentVars map[string]string,
// ) string {
// 	notify.Hammer("Building %s ef migration docker image", imageName)
// 	if baseImage == "" {
// 		baseImage = "mcr.microsoft.com/dotnet/sdk:6.0-focal"
// 	}

// 	dockerFile := ""
// 	dockerFile += fmt.Sprintf("FROM %v\n", baseImage)
// 	dockerFile += "\n"
// 	dockerFile += "# Setting up the argument parameters\n"
// 	for _, argument := range arguments {
// 		dockerFile += fmt.Sprintf("ARG %s\n", argument)
// 	}
// 	// Setting the authentication for the private feeds
// 	dockerFile += "\n"
// 	dockerFile += fmt.Sprintf("%s\n", `ENV VSS_NUGET_EXTERNAL_FEED_ENDPOINTS="{\"endpointCredentials\": [{\"endpoint\":\"https://example.pkgs.visualstudio.com/_packaging/Uno/nuget/v3/index.json\",  \"password\":\"${FEED_ACCESSTOKEN}\"}]}"`)
// 	dockerFile += fmt.Sprintf("%s\n", `ENV PATH="${PATH}:/root/.dotnet/tools"`)
// 	dockerFile += fmt.Sprintf("%s\n", `ENV NUGET_CREDENTIALPROVIDER_SESSIONTOKENCACHE_ENABLED=true`)
// 	dockerFile += "\n"

// 	dockerFile += "\n"
// 	dockerFile += "WORKDIR /repo\n"
// 	// Forcing a new DNS resolution to avoid issues
// 	// dockerFile += "RUN echo \"nameserver 8.8.8.8\" | tee /etc/resolv.conf > /dev/null\n"

// 	dockerFile += "\n"
// 	dockerFile += "RUN apt update && apt -y install git\n"
// 	dockerFile += fmt.Sprintf("RUN git clone %s /repo\n", repoUrl)

// 	dockerFile += "\n"

// 	dockerFile += "RUN curl -L https://raw.githubusercontent.com/Microsoft/artifacts-credprovider/master/helpers/installcredprovider.sh  | sh\n"

// 	dockerFile += "\n"
// 	dockerFile += "RUN dotnet --version\n"
// 	dockerFile += "RUN dotnet tool install --global dotnet-ef --version 6.0.8\n"
// 	dockerFile += "RUN dotnet ef --version\n"

// 	dockerFile += "\n"
// 	for key, value := range environmentVars {
// 		encodedValue, err := json.Marshal(value)
// 		if err == nil {
// 			dockerFile += fmt.Sprintf("ENV %s=%s\n", key, string(encodedValue))
// 		}
// 	}

// 	dockerFile += "\n"
// 	projectPath = strings.TrimPrefix(projectPath, "/")
// 	projectPath = strings.TrimPrefix(projectPath, "\\")
// 	dockerFile += fmt.Sprintf("RUN dotnet build /repo/%s\n", projectPath)

// 	dockerFile += fmt.Sprintf("ENTRYPOINT [\"dotnet\",\"ef\", \"database\", \"update\", \"--project\", \"/repo/%s\"", strings.TrimPrefix(projectPath, "/"))

// 	if startupProjectPath != "" {
// 		dockerFile += fmt.Sprintf(", \"--startup-project\", \"%s\"", strings.TrimPrefix(startupProjectPath, "/"))
// 	}

// 	if dbConnectionString != "" {
// 		dockerFile += fmt.Sprintf(", \"--connection\", \"%s\"", dbConnectionString)
// 	}

// 	dockerFile += "]\n"

// 	return dockerFile
// }

// func (svc *DockerService) generateEfMigrationDockerComposeFile(name, imageName string, environmentVars map[string]string) string {
// 	env := environment.GetInstance()
// 	notify.Hammer("Building %s ef migration container fragment", name)
// 	dockerComposeFile := ""
// 	dockerComposeFile += "version: '3.7'\n"
// 	dockerComposeFile += fmt.Sprintf("name: %v\n", name)
// 	dockerComposeFile += "services:\n"
// 	dockerComposeFile += "  container:\n"
// 	dockerComposeFile += fmt.Sprintf("    image: %s\n", imageName)
// 	if len(environmentVars) > 0 {
// 		dockerComposeFile += "    environment:\n"
// 	}
// 	for key, value := range environmentVars {
// 		val := value
// 		val = env.Replace(val)
// 		val = strings.ReplaceAll(val, "\r\n", "")
// 		val = strings.ReplaceAll(val, "\n", "")
// 		val = strings.ReplaceAll(val, "$", "$$")

// 		if value != "" {
// 			re := regexp.MustCompile("^[a-zA-Z0-9]{1}.*$")
// 			if !re.MatchString(val) {
// 				singleQuote := regexp.MustCompile("^.*['].*$")
// 				doubleQuote := regexp.MustCompile("^.*[\"].*$")
// 				if singleQuote.MatchString(val) && doubleQuote.MatchString(val) {
// 					val = strings.ReplaceAll(val, "'", "''")
// 					val = fmt.Sprintf("'%v'", val)
// 				} else if singleQuote.MatchString(val) {
// 					val = strings.ReplaceAll(val, "\"", "\"")
// 					val = fmt.Sprintf("\"%v\"", val)
// 				} else if doubleQuote.MatchString(val) {
// 					val = strings.ReplaceAll(val, "'", "''")
// 					val = fmt.Sprintf("'%v'", val)
// 				} else {
// 					val = fmt.Sprintf("'%v'", val)
// 				}
// 			}
// 			dockerComposeFile += fmt.Sprintf("      %v: %v\n", key, val)
// 		}
// 	}

// 	return dockerComposeFile
// }

// func (svc *DockerService) generateDotnetContainerDockerFile(command, baseImage, imageName, repoUrl, projectPath string, arguments, runArguments []string, environmentVars map[string]string) string {
// 	notify.Hammer("Building %s %s docker image", imageName, command)
// 	if baseImage == "" {
// 		baseImage = "mcr.microsoft.com/dotnet/sdk:6.0-focal"
// 	}

// 	dockerFile := ""
// 	dockerFile += fmt.Sprintf("FROM %v\n", baseImage)
// 	dockerFile += "\n"
// 	dockerFile += "# Setting up the argument parameters\n"
// 	for _, argument := range arguments {
// 		dockerFile += fmt.Sprintf("ARG %s\n", argument)
// 	}
// 	// Setting the authentication for the private feeds
// 	dockerFile += "\n"
// 	dockerFile += fmt.Sprintf("%s\n", `ENV VSS_NUGET_EXTERNAL_FEED_ENDPOINTS="{\"endpointCredentials\": [{\"endpoint\":\"https://example.pkgs.visualstudio.com/_packaging/Uno/nuget/v3/index.json\",  \"password\":\"${FEED_ACCESSTOKEN}\"}]}"`)
// 	dockerFile += fmt.Sprintf("%s\n", `ENV NUGET_CREDENTIALPROVIDER_SESSIONTOKENCACHE_ENABLED=true`)
// 	dockerFile += "\n"

// 	dockerFile += "\n"
// 	dockerFile += "WORKDIR /repo\n"
// 	// Forcing a new DNS resolution to avoid issues
// 	// dockerFile += "RUN echo \"nameserver 8.8.8.8\" | tee /etc/resolv.conf > /dev/null\n"

// 	dockerFile += "\n"
// 	dockerFile += "RUN apt update && apt -y install git\n"
// 	dockerFile += fmt.Sprintf("RUN git clone %s /repo\n", repoUrl)

// 	dockerFile += "\n"

// 	dockerFile += "RUN curl -L https://raw.githubusercontent.com/Microsoft/artifacts-credprovider/master/helpers/installcredprovider.sh  | sh\n"

// 	dockerFile += "\n"
// 	dockerFile += "RUN dotnet --version\n"

// 	dockerFile += "\n"
// 	for key, value := range environmentVars {
// 		encodedValue, err := json.Marshal(value)
// 		if err == nil {
// 			dockerFile += fmt.Sprintf("ENV %s=%s\n", key, string(encodedValue))
// 		}
// 	}

// 	dockerFile += "\n"
// 	projectPath = strings.TrimPrefix(projectPath, "/")
// 	projectPath = strings.TrimPrefix(projectPath, "\\")
// 	if !strings.EqualFold(command, "build") {
// 		dockerFile += fmt.Sprintf("RUN dotnet build /repo/%s\n", projectPath)
// 	}
// 	args := make([]string, 0)
// 	args = append(args, "\"dotnet\"")
// 	args = append(args, fmt.Sprintf("\"%s\"", command))
// 	if projectPath != "" && !strings.EqualFold(command, "build") {
// 		args = append(args, "\"--project\"")
// 		args = append(args, fmt.Sprintf("\"/repo/%s\"", projectPath))
// 	}

// 	for _, val := range runArguments {
// 		runArgParts := strings.Split(val, " ")
// 		if len(runArgParts) == 1 {
// 			args = append(args, fmt.Sprintf("\"%s\"", strings.TrimSpace(runArgParts[0])))
// 		} else {
// 			args = append(args, fmt.Sprintf("\"%s\"", strings.TrimSpace(runArgParts[0])))
// 			args = append(args, fmt.Sprintf("\"%s\"", strings.TrimSpace(strings.Join(runArgParts[1:], " "))))
// 		}
// 	}
// 	dockerFile += fmt.Sprintf("ENTRYPOINT [%s]\n", strings.Join(args, ","))

// 	return dockerFile
// }

// func (svc *DockerService) generateDotnetContainerComposeFile(name, imageName string, environmentVars map[string]string) string {
// 	env := environment.GetInstance()
// 	notify.Hammer("Building %s build and run container fragment", name)
// 	dockerComposeFile := ""
// 	dockerComposeFile += "version: '3.7'\n"
// 	dockerComposeFile += fmt.Sprintf("name: %v\n", name)
// 	dockerComposeFile += "services:\n"
// 	dockerComposeFile += "  container:\n"
// 	dockerComposeFile += fmt.Sprintf("    image: %s\n", imageName)
// 	if len(environmentVars) > 0 {
// 		dockerComposeFile += "    environment:\n"
// 	}
// 	for key, value := range environmentVars {
// 		val := value
// 		val = env.Replace(val)
// 		val = strings.ReplaceAll(val, "\r\n", "")
// 		val = strings.ReplaceAll(val, "\n", "")
// 		val = strings.ReplaceAll(val, "$", "$$")

// 		if value != "" {
// 			re := regexp.MustCompile("^[a-zA-Z0-9]{1}.*$")
// 			if !re.MatchString(val) {
// 				singleQuote := regexp.MustCompile("^.*['].*$")
// 				doubleQuote := regexp.MustCompile("^.*[\"].*$")
// 				if singleQuote.MatchString(val) && doubleQuote.MatchString(val) {
// 					val = strings.ReplaceAll(val, "'", "''")
// 					val = fmt.Sprintf("'%v'", val)
// 				} else if singleQuote.MatchString(val) {
// 					val = strings.ReplaceAll(val, "\"", "\"")
// 					val = fmt.Sprintf("\"%v\"", val)
// 				} else if doubleQuote.MatchString(val) {
// 					val = strings.ReplaceAll(val, "'", "''")
// 					val = fmt.Sprintf("'%v'", val)
// 				} else {
// 					val = fmt.Sprintf("'%v'", val)
// 				}
// 			}
// 			dockerComposeFile += fmt.Sprintf("      %v: %v\n", key, val)
// 		}
// 	}

// 	return dockerComposeFile
// }

// func (svc *DockerService) getPath(container *docker_component.DockerContainer, options *DockerServiceOptions) (string, error) {
// 	notify.Debug(container.Name)
// 	returnPath := ""

// 	env := environment.GetInstance()
// 	if options == nil {
// 		err := fmt.Errorf("options cannot be nil")
// 		return "", err
// 	}

// 	if container.Location == nil {
// 		container.Location = &entities.Location{}
// 	}

// 	if container.Location.Path == "" {
// 		container.Location.Path = "/"
// 	}

// 	if options.DockerCompose == nil {
// 		if container.DockerCompose != nil {
// 			options.DockerCompose = container.DockerCompose
// 		} else {
// 			options.DockerCompose = &docker_component.DockerCompose{}
// 		}
// 	}

// 	if options.DockerRegistry == nil {
// 		if container.DockerRegistry != nil {
// 			options.DockerRegistry = container.DockerRegistry
// 		} else {
// 			options.DockerRegistry = &docker_component.DockerRegistry{}
// 		}
// 	}

// 	if container.Repository == nil {
// 		container.Repository = &git_component.GitCloneRepository{}
// 	}
// 	if container.DockerCompose == nil {
// 		container.DockerCompose = &docker_component.DockerCompose{}
// 	}
// 	if container.DockerRegistry == nil {
// 		container.DockerRegistry = &docker_component.DockerRegistry{}
// 	}

// 	if container.Repository.Enabled {
// 		returnPath = ""
// 	}

// 	if options.DockerCompose.Location != "" {
// 		returnPath = filepath.Dir(options.DockerCompose.Location)
// 	} else if container.DockerCompose.Location != "" {
// 		returnPath = filepath.Dir(container.DockerCompose.Location)
// 	}

// 	if returnPath == "" && (options.DockerRegistry.Enabled || container.DockerRegistry.Enabled) {
// 		basePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.DOCKER_COMPOSE_PATH)
// 		serviceFolder := helper.JoinPath(basePath, common.EncodeName(container.Name))
// 		returnPath = serviceFolder
// 	}

// 	if returnPath == "" {
// 		clonedPath, err := svc.clone(container)
// 		if err != nil {
// 			return "", err
// 		}
// 		if clonedPath != "" {
// 			returnPath = helper.JoinPath(clonedPath, env.Replace(container.Location.Path))
// 		} else {
// 			returnPath = helper.JoinPath(env.Replace(container.Location.RootFolder), env.Replace(container.Location.Path))
// 		}
// 	}

// 	returnPath = strings.Trim(returnPath, "\\")
// 	notify.Debug("Got the return path of %s for docker service", returnPath)

// 	return returnPath, nil
// }

// func (svc *DockerService) CanClone(container *docker_component.DockerContainer) bool {
// 	if container == nil {
// 		notify.Debug("Container is nil, ignoring")
// 		return false
// 	}

// 	if container.Location != nil && container.Location.Path != "" && container.Location.RootFolder != "" {
// 		if container.Repository != nil && !container.Repository.Enabled {
// 			currentPath := helper.JoinPath(container.Location.RootFolder, container.Location.Path)
// 			notify.Debug("Container has a path defined: %s, ignoring", currentPath)
// 			return false
// 		} else {
// 			notify.Debug("Container has a path defined but repo is enabled")
// 		}
// 	}

// 	if container.Repository == nil {
// 		notify.Debug("Container has no repository defined, ignoring")
// 		return false
// 	}

// 	if !container.Repository.Enabled {
// 		notify.Debug("Container repository was not enabled, ignoring")
// 		return false
// 	}

// 	if container.Repository.Url == "" {
// 		notify.Debug("Container has no repository url defined, ignoring")
// 		return false
// 	}

// 	return true
// }

// func (svc *DockerService) clone(container *docker_component.DockerContainer) (string, error) {
// 	env := environment.GetInstance()

// 	if container.Repository == nil {
// 		return "", nil
// 	}

// 	destination := env.Replace(container.Repository.Destination)
// 	if helper.DirectoryExists(destination) {
// 		notify.Debug("Destination folder %s already exists, ignoring", destination)
// 		return destination, nil
// 	}

// 	cleanRepo := helper.GetFlagSwitch("clean-repo", false)

// 	if svc.CanClone(container) {
// 		git := git.Get()

// 		if container.Repository.Credentials != nil {
// 			mappers.DecodeGitCredentials(container.Repository.Credentials)
// 			if err := git.CloneWithCredentials(container.Repository.Url, destination, container.Repository.Credentials, cleanRepo); err != nil {
// 				return "", err
// 			}
// 		} else {
// 			if err := git.Clone(container.Repository.Url, destination, false); err != nil {
// 				return "", err
// 			}
// 		}

// 		return destination, nil
// 	} else {
// 		return "", nil
// 	}
// }

// func (svc *DockerService) CheckComposeFolder(container *docker_component.DockerContainer) error {
// 	basePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.DOCKER_COMPOSE_PATH)
// 	if !helper.FileExists(basePath) {
// 		notify.Hammer("Creating %s folder", basePath)
// 		if !helper.CreateDirectory(basePath, fs.ModePerm) {
// 			err := fmt.Errorf("error creating the %v folder", basePath)
// 			return err
// 		}
// 	}
// 	serviceFolder := helper.JoinPath(basePath, common.EncodeName(container.Name))
// 	if !helper.FileExists(serviceFolder) {
// 		notify.Hammer("Creating %s folder", serviceFolder)
// 		if !helper.CreateDirectory(serviceFolder, fs.ModePerm) {
// 			err := fmt.Errorf("error creating the %v folder", serviceFolder)
// 			return err
// 		}
// 	}

// 	return nil
// }

// func (svc *DockerService) CheckForDockerComposeFile(container *docker_component.DockerContainer, containerPath string) error {
// 	fileList := []string{
// 		"docker-compose.yaml",
// 		"docker-compose.yml",
// 	}
// 	if helper.GetFlagSwitch("clean", false) {
// 		notify.Debug("Starting to clean the current docker-compose")
// 		basePath := helper.JoinPath(config.GetCurrentContext().Configuration.OutputPath, common.DOCKER_COMPOSE_PATH, common.EncodeName(container.Name))
// 		for _, file := range fileList {
// 			filePath := helper.JoinPath(basePath, file)
// 			notify.Debug("Testing for file %s", filePath)
// 			if helper.FileExists(filePath) {
// 				notify.Debug("Deleting file %s", filePath)
// 				helper.DeleteFile(filePath)
// 			}
// 		}
// 	}

// 	dockerComposePath := helper.JoinPath(containerPath, "docker-compose")
// 	if !helper.FileExists(fmt.Sprintf("%s.yml", dockerComposePath)) && !helper.FileExists(fmt.Sprintf("%s.yaml", dockerComposePath)) {
// 		notify.Hammer("Docker compose not found in the repo folder, trying to generate a default one")
// 		compose := svc.generateContainerDockerCompose(container)
// 		svc.CheckComposeFolder(container)
// 		if err := helper.WriteToFile(compose, fmt.Sprintf("%s.yml", dockerComposePath)); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
