package configuration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/cjlapao/locally-cli/interfaces"

	"github.com/cjlapao/common-go/guard"
	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

type Context struct {
	Source               string                       `json:"-" yaml:"-"`
	IsDefault            bool                         `json:"-" yaml:"-"`
	Name                 string                       `json:"name,omitempty" yaml:"name,omitempty"`
	RootConfigFilePath   string                       `json:"configPath,omitempty" yaml:"configPath,omitempty"`
	Configuration        *ContextConfiguration        `json:"configuration,omitempty" yaml:"configuration,omitempty"`
	EnvironmentVariables *EnvironmentVariables        `json:"environmentVariables,omitempty" yaml:"environmentVariables,omitempty"`
	Infrastructure       *Infrastructure              `json:"infrastructure,omitempty" yaml:"infrastructure,omitempty"`
	SpaServices          []*SpaService                `json:"spaServices,omitempty" yaml:"spaServices,omitempty"`
	BackendServices      []*BackendService            `json:"backendServices,omitempty" yaml:"backendServices,omitempty"`
	MockServices         []*MockService               `json:"mockServices,omitempty" yaml:"mockServices,omitempty"`
	NugetPackages        *NugetPackages               `json:"nugetPackages,omitempty" yaml:"nugetPackages,omitempty"`
	Tenants              []*Tenant                    `json:"tenants,omitempty" yaml:"tenants,omitempty"`
	Pipelines            []*Pipeline                  `json:"pipelines,omitempty" yaml:"pipelines,omitempty"`
	Credentials          *Credentials                 `json:"credentials,omitempty" yaml:"credentials,omitempty"`
	BackendConfig        *InfrastructureBackendConfig `json:"backendConfig,omitempty" yaml:"backendConfig,omitempty"`
	Fragments            []*Context                   `json:"-" yaml:"-"`
	registeredServices   []interfaces.LocallyService  `json:"-" yaml:"-"`
}

func (svc *Context) SaveEnvironmentVariables() error {
	var config Context
	format := "yaml"
	configContent, err := helper.ReadFromFile(svc.RootConfigFilePath)
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

	config.EnvironmentVariables = svc.EnvironmentVariables

	switch format {
	case "json":
		content, err := json.Marshal(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}
		helper.WriteToFile(string(content), svc.RootConfigFilePath)
	case "yaml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}

		helper.WriteToFile(b.String(), svc.RootConfigFilePath)
	}

	return nil
}

func (svc *Context) SaveCredentials() error {
	var config Context
	format := "yaml"
	configContent, err := helper.ReadFromFile(svc.RootConfigFilePath)
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

	config.Credentials = svc.Credentials

	switch format {
	case "json":
		content, err := json.Marshal(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}
		helper.WriteToFile(string(content), svc.RootConfigFilePath)
	case "yaml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}

		helper.WriteToFile(b.String(), svc.RootConfigFilePath)
	}

	return nil
}

func (svc *Context) SaveBackendConfig() error {
	var config Context
	format := "yaml"
	configContent, err := helper.ReadFromFile(svc.RootConfigFilePath)
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

	config.BackendConfig = svc.BackendConfig

	switch format {
	case "json":
		content, err := json.Marshal(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}
		helper.WriteToFile(string(content), svc.RootConfigFilePath)
	case "yaml":
		var b bytes.Buffer
		yamlEncoder := yaml.NewEncoder(&b)
		yamlEncoder.SetIndent(2)
		err := yamlEncoder.Encode(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
			return err
		}

		helper.WriteToFile(b.String(), svc.RootConfigFilePath)
	}

	return nil
}

func (svc *Context) Save() error {
	return svc.SaveFragment(svc)
}

func (svc *Context) GetOverrideFileName(filePath string) string {
	if strings.HasSuffix(filePath, OVERRIDE_CONFIG_FILE_MARKER+".yml") ||
		strings.HasSuffix(filePath, OVERRIDE_CONFIG_FILE_MARKER+".yaml") ||
		strings.HasSuffix(filePath, OVERRIDE_CONFIG_FILE_MARKER+".json") {
		// input is already a override file path
		return filePath
	}

	if strings.HasSuffix(filePath, ".yml") {
		return strings.TrimSuffix(filePath, ".yml") + OVERRIDE_CONFIG_FILE_MARKER + ".yml"
	}

	if strings.HasSuffix(filePath, ".yaml") {
		return strings.TrimSuffix(filePath, ".yaml") + OVERRIDE_CONFIG_FILE_MARKER + ".yaml"
	}

	if strings.HasSuffix(filePath, ".json") {
		return strings.TrimSuffix(filePath, ".json") + OVERRIDE_CONFIG_FILE_MARKER + ".json"
	}

	return filePath
}

func (svc *Context) SaveFragment(fragment *Context) error {

	filePathToSave := svc.GetOverrideFileName(fragment.Source)
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

func (svc *Context) AddRegisteredService(service interfaces.LocallyService) {
	found := false
	for _, existingService := range svc.registeredServices {
		if strings.EqualFold(existingService.GetName(), service.GetName()) {
			found = true
			break
		}
	}

	if !found {
		svc.registeredServices = append(svc.registeredServices, service)
	}
}

func (svc *Context) GetRegisteredService(serviceName string) interfaces.LocallyService {
	for _, existingService := range svc.registeredServices {
		if strings.EqualFold(existingService.GetName(), serviceName) {
			return existingService
		}
	}

	return nil
}

func (svc *Context) GetContainerFragmentByName(name string) *Context {
	config := Get()
	containers := config.GetDockerServices(name, true)

	if len(containers) == 0 {
		notify.Debug("Container %s was not found in the context", name)
		return nil
	}

	notify.Debug("Container %s was found in the context with source %s", name, containers[0].Source)

	return config.GetFragment(containers[0].Source)
}

func (svc *Context) AddDockerRegistryConfig(serviceName string, value *DockerRegistry) {
	for _, service := range svc.BackendServices {
		if strings.EqualFold(service.Name, serviceName) {
			service.DockerRegistry.Clone(value, false)
			return
		}
	}

	for _, service := range svc.SpaServices {
		if strings.EqualFold(service.Name, serviceName) {
			service.DockerRegistry = value
			return
		}
	}
	notify.Debug("Service %s was not found in this context", serviceName)
}

func (svc *Context) AddDockerComposeConfig(serviceName string, value *DockerCompose) {
	for _, service := range svc.BackendServices {
		if strings.EqualFold(service.Name, serviceName) {
			service.DockerCompose.Clone(value, false)
			return
		}
	}

	for _, service := range svc.SpaServices {
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

func SaveCurrentContext() error {
	config := Get()
	currentContext := config.GetCurrentContext()

	return currentContext.Save()
}
