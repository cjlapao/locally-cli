package context

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/context/docker_component"
	context_entities "github.com/cjlapao/locally-cli/internal/context/entities"
	"github.com/cjlapao/locally-cli/internal/context/infrastructure_component"
	"github.com/cjlapao/locally-cli/internal/context/mock_component"
	"github.com/cjlapao/locally-cli/internal/context/nuget_package_component"
	"github.com/cjlapao/locally-cli/internal/context/pipeline_component"
	"github.com/cjlapao/locally-cli/internal/context/service_component"
	"github.com/cjlapao/locally-cli/internal/entities"
	"github.com/cjlapao/locally-cli/internal/interfaces"
	"github.com/google/uuid"

	"github.com/cjlapao/common-go/guard"
	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

type Context struct {
	ID                   string                                                `json:"id,omitempty" yaml:"id,omitempty"`
	IsValid              bool                                                  `json:"-" yaml:"-"`
	IsEnabled            bool                                                  `json:"isEnabled" yaml:"isEnabled"`
	Source               string                                                `json:"-" yaml:"-"`
	IsDefault            bool                                                  `json:"-" yaml:"-"`
	Name                 string                                                `json:"name,omitempty" yaml:"name,omitempty"`
	RootConfigFilePath   string                                                `json:"configPath,omitempty" yaml:"configPath,omitempty"`
	Configuration        *context_entities.ContextConfiguration                `json:"configuration,omitempty" yaml:"configuration,omitempty"`
	EnvironmentVariables *entities.EnvironmentVariables                        `json:"environmentVariables,omitempty" yaml:"environmentVariables,omitempty"`
	Infrastructure       *infrastructure_component.Infrastructure              `json:"infrastructure,omitempty" yaml:"infrastructure,omitempty"`
	SpaServices          []*service_component.SpaService                       `json:"spaServices,omitempty" yaml:"spaServices,omitempty"`
	BackendServices      []*service_component.BackendService                   `json:"backendServices,omitempty" yaml:"backendServices,omitempty"`
	MockServices         []*mock_component.MockService                         `json:"mockServices,omitempty" yaml:"mockServices,omitempty"`
	NugetPackages        *nuget_package_component.NugetPackages                `json:"nugetPackages,omitempty" yaml:"nugetPackages,omitempty"`
	Tenants              []*context_entities.Tenant                            `json:"tenants,omitempty" yaml:"tenants,omitempty"`
	Pipelines            []*pipeline_component.Pipeline                        `json:"pipelines,omitempty" yaml:"pipelines,omitempty"`
	Credentials          *context_entities.Credentials                         `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	BackendConfig        *infrastructure_component.InfrastructureBackendConfig `json:"backendConfig,omitempty" yaml:"backendConfig,omitempty"`
	Fragments            []*Context                                            `json:"-" yaml:"-"`
	registeredServices   []interfaces.LocallyService                           `json:"-" yaml:"-"`
}

func (ctx *Context) TestConnection() error {
	return nil
}

// TODO: check what we still need from the old code and move it to the new context

func (ctx *Context) SaveEnvironmentVariables() error {
	var config Context
	format := "yaml"
	configContent, err := helper.ReadFromFile(ctx.RootConfigFilePath)
	if err != nil {
		notify.FromError(err, "There was an error reading the configuration file")
		return err
	}

	if err := yaml.Unmarshal(configContent, &config); err != nil {
		if err := json.Unmarshal(configContent, &config); err != nil {
			notify.FromError(err, "There was an error reading the configuration file")
			return err
		} else {
			format = "json"
		}
	} else {
		format = "yaml"
	}

	config.EnvironmentVariables = ctx.EnvironmentVariables

	switch format {
	case "json":
		content, err := json.Marshal(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}
		helper.WriteToFile(string(content), ctx.RootConfigFilePath)
	case "yaml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}

		helper.WriteToFile(b.String(), ctx.RootConfigFilePath)
	}

	return nil
}

func (ctx *Context) SaveCredentials() error {
	var config Context
	format := "yaml"
	configContent, err := helper.ReadFromFile(ctx.RootConfigFilePath)
	if err != nil {
		notify.FromError(err, "There was an error reading the configuration file")
		return err
	}

	if err := yaml.Unmarshal(configContent, &config); err != nil {
		if err := json.Unmarshal(configContent, &config); err != nil {
			notify.FromError(err, "There was an error reading the configuration file")
			return err
		} else {
			format = "json"
		}
	} else {
		format = "yaml"
	}

	config.Credentials = ctx.Credentials

	switch format {
	case "json":
		content, err := json.Marshal(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}
		helper.WriteToFile(string(content), ctx.RootConfigFilePath)
	case "yaml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}

		helper.WriteToFile(b.String(), ctx.RootConfigFilePath)
	}

	return nil
}

func (ctx *Context) SaveBackendConfig() error {
	var config Context
	format := "yaml"
	configContent, err := helper.ReadFromFile(ctx.RootConfigFilePath)
	if err != nil {
		notify.FromError(err, "There was an error reading the configuration file")
		return err
	}

	if err := yaml.Unmarshal(configContent, &config); err != nil {
		if err := json.Unmarshal(configContent, &config); err != nil {
			notify.FromError(err, "There was an error reading the configuration file")
			return err
		} else {
			format = "json"
		}
	} else {
		format = "yaml"
	}

	config.BackendConfig = ctx.BackendConfig

	switch format {
	case "json":
		content, err := json.Marshal(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}
		helper.WriteToFile(string(content), ctx.RootConfigFilePath)
	case "yaml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}

		helper.WriteToFile(b.String(), ctx.RootConfigFilePath)
	}

	return nil
}

func (ctx *Context) Save() error {
	return ctx.SaveFragment(ctx)
}

func (ctx *Context) GetOverrideFileName(filePath string) string {
	if strings.HasSuffix(filePath, common.OVERRIDE_CONFIG_FILE_MARKER+".yml") ||
		strings.HasSuffix(filePath, common.OVERRIDE_CONFIG_FILE_MARKER+".yaml") ||
		strings.HasSuffix(filePath, common.OVERRIDE_CONFIG_FILE_MARKER+".json") {
		// input is already a override file path
		return filePath
	}

	if strings.HasSuffix(filePath, ".yml") {
		return strings.TrimSuffix(filePath, ".yml") + common.OVERRIDE_CONFIG_FILE_MARKER + ".yml"
	}

	if strings.HasSuffix(filePath, ".yaml") {
		return strings.TrimSuffix(filePath, ".yaml") + common.OVERRIDE_CONFIG_FILE_MARKER + ".yaml"
	}

	if strings.HasSuffix(filePath, ".json") {
		return strings.TrimSuffix(filePath, ".json") + common.OVERRIDE_CONFIG_FILE_MARKER + ".json"
	}

	return filePath
}

func (ctx *Context) SaveFragment(fragment *Context) error {
	if ctx.ID == "" {
		ctx.ID = uuid.New().String()
	}

	filePathToSave := ctx.GetOverrideFileName(fragment.Source)
	notify.Debug("Saving fragment on path %s", filePathToSave)

	format := "json"

	if strings.HasSuffix(filePathToSave, ".yml") ||
		strings.HasSuffix(filePathToSave, ".yaml") {
		format = "yaml"
	}

	switch format {
	case "json":
		content, err := json.Marshal(fragment)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}
		helper.WriteToFile(string(content), filePathToSave)
	case "yaml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode(fragment)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}

		helper.WriteToFile(b.String(), filePathToSave)
	}

	return nil
}

func (ctx *Context) AddRegisteredService(service interfaces.LocallyService) {
	found := false
	for _, existingService := range ctx.registeredServices {
		if strings.EqualFold(existingService.GetName(), service.GetName()) {
			found = true
			break
		}
	}

	if !found {
		ctx.registeredServices = append(ctx.registeredServices, service)
	}
}

func (ctx *Context) GetRegisteredService(serviceName string) interfaces.LocallyService {
	for _, existingService := range ctx.registeredServices {
		if strings.EqualFold(existingService.GetName(), serviceName) {
			return existingService
		}
	}

	return nil
}

func (ctx *Context) GetContainerFragmentByName(name string) *Context {
	containers := ctx.GetDockerServices(name, true)

	if len(containers) == 0 {
		notify.Debug("Container %s was not found in the context", name)
		return nil
	}

	notify.Debug("Container %s was found in the context with source %s", name, containers[0].Source)

	return ctx.GetFragment(containers[0].Source)
}

func (ctx *Context) AddDockerRegistryConfig(serviceName string, value *docker_component.DockerRegistry) {
	for _, service := range ctx.BackendServices {
		if strings.EqualFold(service.Name, serviceName) {
			service.DockerRegistry.Clone(value, false)
			return
		}
	}

	for _, service := range ctx.SpaServices {
		if strings.EqualFold(service.Name, serviceName) {
			service.DockerRegistry = value
			return
		}
	}
	notify.Debug("Service %s was not found in this context", serviceName)
}

func (ctx *Context) AddDockerComposeConfig(serviceName string, value *docker_component.DockerCompose) {
	for _, service := range ctx.BackendServices {
		if strings.EqualFold(service.Name, serviceName) {
			service.DockerCompose.Clone(value, false)
			return
		}
	}

	for _, service := range ctx.SpaServices {
		if strings.EqualFold(service.Name, serviceName) {
			service.DockerCompose = value
			return
		}
	}

	notify.Debug("Service %s was not found in this context", serviceName)
}

func UpdateValueIfEmpty[T interface{}](source T, dest T) T {
	if err := guard.EmptyOrNil(source); err == nil {
		if err := guard.EmptyOrNil(dest); err == nil {
			notify.Debug("updating %s to %s", fmt.Sprintf("%v", source), fmt.Sprintf("%v", dest))
			source = dest
			return dest
		}
	}

	return source
}

func (ctx *Context) SaveCurrentContext() error {
	return ctx.Save()
}

func (ctx *Context) GetDockerServices(name string, ignoreTags bool) []*docker_component.DockerContainer {
	result := make([]*docker_component.DockerContainer, 0)

	var backendServices []*service_component.BackendService
	var frontendServices []*service_component.SpaService

	if ignoreTags {
		notify.Debug("Ignoring flags, getting one by one")
		backendServices = ctx.BackendServices
		frontendServices = ctx.SpaServices
	} else {
		backendServices = ctx.GetBackendServicesByTags()
		frontendServices = ctx.GetSpaServicesByTags()
	}

	for _, service := range backendServices {
		if !service.HasPath() {
			continue
		}

		container := docker_component.DockerContainer{
			Name:           service.Name,
			Location:       service.Location,
			Repository:     service.Repository,
			DependsOn:      service.DependsOn,
			RequiredBy:     service.RequiredBy,
			Source:         service.Source,
			DockerRegistry: service.DockerRegistry,
			DockerCompose:  service.DockerCompose,
			Tags:           service.Tags,
			Components:     make([]*docker_component.DockerContainer, 0),
		}

		for _, component := range service.Components {
			componentContainer := docker_component.DockerContainer{
				Name:                 component.Name,
				Source:               component.Source,
				DependsOn:            component.DependsOn,
				RequiredBy:           component.RequiredBy,
				EnvironmentVariables: component.EnvironmentVariables,
				BuildArguments:       component.BuildArguments,
				ManifestPath:         component.ManifestPath,
				ManifestTag:          component.ManifestTag,
				Tags:                 service.Tags,
			}

			container.Components = append(container.Components, &componentContainer)
		}

		result = append(result, &container)
	}

	for _, service := range frontendServices {
		if !service.HasPath() {
			continue
		}

		container := docker_component.DockerContainer{
			Name:           service.Name,
			Location:       service.Location,
			Repository:     service.Repository,
			DependsOn:      service.DependsOn,
			RequiredBy:     service.RequiredBy,
			Source:         service.Source,
			DockerCompose:  service.DockerCompose,
			DockerRegistry: service.DockerRegistry,
			Components:     make([]*docker_component.DockerContainer, 0),
			Tags:           service.Tags,
		}

		componentContainer := docker_component.DockerContainer{
			Name:                 service.Name,
			DependsOn:            service.DependsOn,
			RequiredBy:           service.RequiredBy,
			EnvironmentVariables: service.EnvironmentVariables,
			BuildArguments:       service.BuildArguments,
			Tags:                 service.Tags,
		}

		if service.DockerRegistry != nil && service.DockerRegistry.ManifestPath != "" {
			componentContainer.ManifestPath = service.DockerRegistry.ManifestPath
		}

		container.Components = append(container.Components, &componentContainer)

		result = append(result, &container)
	}

	if helper.GetFlagSwitch("all", false) || ctx.HasTags() && !ignoreTags {
		return result
	} else {
		if name == "" {
			return make([]*docker_component.DockerContainer, 0)
		}

		filteredContainers := make([]*docker_component.DockerContainer, 0)
		for _, container := range result {
			if strings.EqualFold(common.EncodeName(container.Name), name) {
				filteredContainers = append(filteredContainers, container)
			}
		}

		return filteredContainers
	}
}

func (ctx *Context) GetFragment(source string) *Context {
	for _, s := range ctx.Fragments {
		if strings.EqualFold(source, s.Source) {
			return s
		}
	}

	return nil
}

func (ctx *Context) AddAzureCredential(cred *entities.AzureCredentials) {
	if ctx.Credentials == nil {
		ctx.Credentials = &context_entities.Credentials{}
	}

	ctx.Credentials.Azure = cred
}
