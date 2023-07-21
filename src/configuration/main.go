package configuration

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/icons"
	"github.com/cjlapao/locally-cli/notifications"

	"github.com/cjlapao/common-go/helper"
	"gopkg.in/yaml.v3"
)

// This is used to track changes to the context configuration schema. Users are warned if a mismatch is found.
// Note that this only applies to changes in context configuration schema and not the global configuration
var schemaVersion = "0.0.3"

const (
	DEFAULT_CONTEXT_INFRASTRUCTURE_FOLDER     string = "infrastructure"
	DEFAULT_CONTEXT_SERVICE_FOLDER            string = "services"
	DEFAULT_CONTEXT_SERVICE_BACKEND_FOLDER    string = "backends"
	DEFAULT_CONTEXT_SERVICE_MOCKS_FOLDER      string = "mocks"
	DEFAULT_CONTEXT_SERVICE_WEBCLIENTS_FOLDER string = "webclients"
	locally_CONFIG_FOLDER                     string = "configuration"
)

const (
	OVERRIDE_CONFIG_FILE_MARKER string = ".override"
)

type ConfigService struct {
	configFilename      string
	GlobalConfiguration *GlobalConfiguration
}

var globalConfigurationService *ConfigService
var notify = notifications.Get()

func Get() *ConfigService {
	if globalConfigurationService != nil {
		return globalConfigurationService
	}

	globalConfigurationService = New()
	return globalConfigurationService
}

func New() *ConfigService {
	globalConfigurationService = &ConfigService{
		GlobalConfiguration: &GlobalConfiguration{
			Tools: &Tools{
				Checked: &CheckedTools{},
			},
		},
	}
	globalConfigurationService.GlobalConfiguration.verbose = helper.GetFlagSwitch("v", false)

	return globalConfigurationService
}

func (svc *ConfigService) Init() error {
	err := svc.loadConfiguration()
	svc.GetCurrentContext()

	return err
}

func (svc *ConfigService) GetCurrentContext() *Context {
	if svc.GlobalConfiguration.CurrentContext != "" {
		return svc.GetContext(svc.GlobalConfiguration.CurrentContext)
	} else {
		context := svc.GetDefaultContext()
		if context != nil {
			return context
		} else {
			notify.Error("No context found, exiting")
			os.Exit(1)
		}
	}

	notify.Error("No context found, exiting")
	os.Exit(1)
	return nil
}

func (svc *ConfigService) GetContext(name string) *Context {
	for _, context := range svc.GlobalConfiguration.Contexts {
		if strings.EqualFold(context.Name, name) {
			svc.GlobalConfiguration.CurrentContext = context.Name
			return context
		}
	}

	return nil
}

func (svc *ConfigService) GetDefaultContext() *Context {
	for _, context := range svc.GlobalConfiguration.Contexts {
		if context.IsDefault {
			svc.GlobalConfiguration.CurrentContext = context.Name
			svc.SaveConfigFile()
			return context
		}
	}

	if len(svc.GlobalConfiguration.Contexts) > 0 {
		svc.GlobalConfiguration.Contexts[0].IsDefault = true
		svc.GlobalConfiguration.CurrentContext = svc.GlobalConfiguration.Contexts[0].Name
		svc.SaveConfigFile()
		return svc.GlobalConfiguration.Contexts[0]
	}

	return nil
}

func (svc *ConfigService) loadConfiguration() error {

	// Read the global config file first. locally has one global config file which specifies:
	//   - List of contexts
	//   - Global config which applies to all contexts (e.g. path to required tools, certificate files etc)
	if err := svc.loadGlobalConfiguration(); err != nil {
		return err
	}

	svc.initializeToolsDefaults()
	svc.initializeCorsDefaults()

	// Go through the contexts specified in the global config and load them all.
	// For each context, the global config specifies:
	//  - Name of the context
	//  - Path to the context config file
	if err := svc.loadContextConfigurationAllContexts(); err != nil {
		return err
	}

	return nil
}

func (svc *ConfigService) loadGlobalConfiguration() error {
	file := svc.getGlobalConfigFile()
	if file != "" {
		notify.InfoWithIcon(icons.IconBook, "Loading global configuration from %v", file)
		content, err := helper.ReadFromFile(file)
		if err != nil {
			notify.FromError(err, "There was an error reading the global configuration file")
			return err
		}
		if err := yaml.Unmarshal(content, &svc.GlobalConfiguration); err != nil {
			if err := json.Unmarshal(content, &svc.GlobalConfiguration); err != nil {
				notify.FromError(err, "There was an error parsing the global configuration file")
				return err
			} else {
				svc.GlobalConfiguration.format = "json"
			}
		} else {
			svc.GlobalConfiguration.format = "yaml"
		}
	} else {
		notify.Error("No global configuration file was found")
		return errors.New("no global configuration file was found")
	}

	return nil
}

func (svc *ConfigService) getDefaultOutputFolder(context *Context) (string, error) {
	basePath := context.RootConfigFilePath
	fileInfo, err := os.Stat(basePath)
	if err != nil {
		return "", err
	}

	if !fileInfo.IsDir() {
		basePath = filepath.Dir(basePath)
	}

	return helper.JoinPath(basePath, common.DEFAULT_locally_OUTPUT_PATH), nil
}

func (svc *ConfigService) doOutputFolderChecks(context *Context) error {
	if context.Configuration.OutputPath == "" {
		defaultOutputFolder, err := svc.getDefaultOutputFolder(context)
		if err != nil {
			return err
		}

		if !helper.DirectoryExists(defaultOutputFolder) {
			notify.Hammer("Creating the default output folder %s", defaultOutputFolder)
			if helper.CreateDirectory(defaultOutputFolder, fs.ModePerm) {
				notify.InfoWithIcon(icons.IconCheckMark, "Default output folder created")
			} else {
				notify.Critical("There was an error creating the default output folder")
			}
		}
		context.Configuration.OutputPath = defaultOutputFolder
	}

	return nil
}

func (svc *ConfigService) doConfigurationSchemaChecks(context *Context) {
	if context.Configuration.SchemaVersion == "" {
		notify.InfoWithIcon(icons.IconWarning, "%s", "##########################################################################################################################################")
		notify.InfoWithIcon(icons.IconWarning, "Context configuration schema version not specified for context %s in file %s", context.Name, context.RootConfigFilePath)
		notify.InfoWithIcon(icons.IconWarning, "Context configuration schema version check could not be performed. You may encounter issues. Update your configuration to stop seeing this warning")
		notify.InfoWithIcon(icons.IconWarning, "%s", "##########################################################################################################################################")
	} else if context.Configuration.SchemaVersion != schemaVersion {
		notify.InfoWithIcon(icons.IconWarning, "%s", "##########################################################################################################################################")
		notify.InfoWithIcon(icons.IconWarning, "Context configuration schema version mismatch for context %s in file %s", context.Name, context.RootConfigFilePath)
		notify.InfoWithIcon(icons.IconWarning, "Expected schema version=[%s], found=[%s]. You may encounter issues. Update your configuration files to stop seeing this warning", schemaVersion, context.Configuration.SchemaVersion)
		notify.InfoWithIcon(icons.IconWarning, "%s", "##########################################################################################################################################")
	}
}

func (svc *ConfigService) doConfigFolderChecks(context *Context) error {
	if context.Configuration.ConfigFolder == "" {
		basePath := context.RootConfigFilePath
		fileInfo, err := os.Stat(basePath)
		if err != nil {
			return err
		}

		if !fileInfo.IsDir() {
			basePath = filepath.Dir(basePath)
		}

		defaultPath := helper.JoinPath(basePath, DEFAULT_CONTEXT_SERVICE_FOLDER)
		if !helper.DirectoryExists(defaultPath) {
			notify.Hammer("No service folder found, creating default %s folder in %s", DEFAULT_CONTEXT_SERVICE_FOLDER, context.RootConfigFilePath)
			helper.CreateDirectory(defaultPath, fs.ModePerm)
		}

		backendPath := helper.JoinPath(defaultPath, DEFAULT_CONTEXT_SERVICE_BACKEND_FOLDER)
		if !helper.DirectoryExists(backendPath) {
			notify.Hammer("No service backend folder found, creating default %s folder in %s", DEFAULT_CONTEXT_SERVICE_BACKEND_FOLDER, defaultPath)
			helper.CreateDirectory(backendPath, fs.ModePerm)
		}

		mocksPath := helper.JoinPath(defaultPath, DEFAULT_CONTEXT_SERVICE_MOCKS_FOLDER)
		if !helper.DirectoryExists(mocksPath) {
			notify.Hammer("No service backend folder found, creating default %s folder in %s", DEFAULT_CONTEXT_SERVICE_MOCKS_FOLDER, defaultPath)
			helper.CreateDirectory(mocksPath, fs.ModePerm)
		}

		webclientsPath := helper.JoinPath(defaultPath, DEFAULT_CONTEXT_SERVICE_WEBCLIENTS_FOLDER)
		if !helper.DirectoryExists(webclientsPath) {
			notify.Hammer("No service backend folder found, creating default %s folder in %s", DEFAULT_CONTEXT_SERVICE_WEBCLIENTS_FOLDER, defaultPath)
			helper.CreateDirectory(webclientsPath, fs.ModePerm)
		}

		context.Configuration.ConfigFolder = defaultPath
	}

	return nil
}

func (svc *ConfigService) CleanContextConfigurationAll() error {
	for _, context := range svc.GlobalConfiguration.Contexts {
		if err := svc.cleanContextConfiguration(context); err != nil {
			return err
		}
	}

	return nil
}

func (svc *ConfigService) CleanContextConfigurationCurrent() error {
	context := svc.GetCurrentContext()
	return svc.cleanContextConfiguration(context)
}

func (svc *ConfigService) cleanContextConfiguration(context *Context) error {

	// locally also saves any dynamically determined configuration related to a default fragment config file in a corresponding .override.<ext> file.
	// Once an override file is created that file is used going forward and the corresponding default config file is ignored.
	// Sometimes we need to "reset" the configuration back to default. This method can be used in this case to deletes all
	// override files found in the current context. On the next run, locally will load the default config files
	if err := svc.deleteOverrideFiles(context.Configuration.ConfigFolder); err != nil {
		return err
	}

	// During operation, locally also persists some configuration in the root context config file (e.g. the lastInitiated and accessKey under backendConfig)
	// We decided not to create an override file for this root config file as this is a user visible file and users set configuration in this file and
	// creating an override file for this file can cause confusion. So during clean up we have delete specific configuration items from the file
	return svc.cleanContextConfigurationRootFile(context)
}

func (svc *ConfigService) cleanContextConfigurationRootFile(context *Context) error {

	if context.BackendConfig != nil {
		context.BackendConfig.LastInitiated = nil
		context.BackendConfig.Azure.AccessKey = ""
		context.SaveBackendConfig()
	}

	return nil
}

func (svc *ConfigService) isOverrideConfigFile(fileNameOrPath string) bool {
	return strings.HasSuffix(fileNameOrPath, OVERRIDE_CONFIG_FILE_MARKER+".yml") ||
		strings.HasSuffix(fileNameOrPath, OVERRIDE_CONFIG_FILE_MARKER+".yaml") ||
		strings.HasSuffix(fileNameOrPath, OVERRIDE_CONFIG_FILE_MARKER+".json")
}

func (svc *ConfigService) deleteOverrideFiles(folderPath string) error {

	// Recurse through the sub-folders and delete all the override files found

	if helper.FileExists(folderPath) {
		if svc.GlobalConfiguration.verbose {
			notify.InfoWithIcon(icons.IconInfo, "Checking for any override config files in the folder %s", folderPath)
		}
		if files, err := os.ReadDir(folderPath); err == nil {
			for _, file := range files {
				if file.IsDir() {
					if err := svc.deleteOverrideFiles(helper.JoinPath(folderPath, file.Name())); err != nil {
						return err
					}
				}

				if !file.IsDir() && svc.isOverrideConfigFile(file.Name()) {
					filePathToDelete := helper.JoinPath(folderPath, file.Name())
					notify.InfoWithIcon(icons.IconBomb, "Deleting override config file %s", filePathToDelete)
					if err := helper.DeleteFile(filePathToDelete); err != nil {
						return err
					}
				}
			}
		} else {
			notify.Error("Failed to delete override files from folder %s", folderPath)
			return errors.New(fmt.Sprintf("Failed to delete override files from folder %s", folderPath))
		}
	}

	return nil
}

func (svc *ConfigService) loadContextConfigurationAllContexts() error {

	// Configuration for a context is split into:
	//   - A root config file
	//   - Various config files under "services" folder (e.g. config files for infrastructure, backends, webclients, pipelines etc)
	//     Think of these config files as "fragments" that make up the whole configuration for a context

	for _, context := range svc.GlobalConfiguration.Contexts {
		if context.RootConfigFilePath == "" {
			notify.InfoWithIcon(icons.IconWarning, "Skipping context %s as config file path was not specified", context.Name)
			continue
		}

		if err := svc.loadContextConfigurationRootFile(context); err != nil {
			return err
		}

		if context.Configuration == nil {
			notify.InfoWithIcon(icons.IconWarning, "Skipping context %s as configuration could not be loaded", context.Name)
			continue
		}

		svc.doConfigurationSchemaChecks(context)

		if err := svc.doOutputFolderChecks(context); err != nil {
			notify.InfoWithIcon(icons.IconWarning, "Skipping context %s due to error related to output folder", context.Name)
			continue
		}

		if err := svc.doConfigFolderChecks(context); err != nil {
			return err
		}

		if err := svc.loadContextConfigurationFragments(context, context.Configuration.ConfigFolder); err != nil {
			return err
		}

		// Setting the default config service internal url to be used later
		if context.Configuration.LocallyConfigService == nil {
			context.Configuration.LocallyConfigService = &LocallyConfigService{
				Url:             "http://config-service.dev-ops.svc.cluster.local",
				ReverseProxyUrl: "host.docker.internal:5510",
			}
		}

		if context.BackendServices == nil {
			context.BackendServices = make([]*BackendService, 0)
		}
		if context.SpaServices == nil {
			context.SpaServices = make([]*SpaService, 0)
		}
	}

	return nil
}

func (svc *ConfigService) initializeToolsDefaults() {
	if svc.GlobalConfiguration.Tools == nil {
		svc.GlobalConfiguration.Tools = &Tools{
			Checked: &CheckedTools{
				DockerChecked:        false,
				DockerComposeChecked: false,
				CaddyChecked:         false,
				NugetChecked:         false,
				DotnetChecked:        false,
				GitChecked:           false,
				TerraformChecked:     false,
				AzureCliChecked:      false,
				NpmChecked:           false,
			},
		}
	}
}

func (svc *ConfigService) initializeCorsDefaults() {
	if svc.GlobalConfiguration.Cors == nil {
		svc.GlobalConfiguration.Cors = &Cors{
			AllowedMethods: "OPTIONS,HEAD,GET,POST,PUT,PATCH,DELETE",
			AllowedHeaders: "*",
			AllowedOrigins: make([]string, 0),
		}
	} else {
		if svc.GlobalConfiguration.Cors.AllowedHeaders == "" {
			svc.GlobalConfiguration.Cors.AllowedHeaders = "*"
		}

		if svc.GlobalConfiguration.Cors.AllowedMethods == "" {
			svc.GlobalConfiguration.Cors.AllowedMethods = "OPTIONS,HEAD,GET,POST,PUT,PATCH,DELETE"
		}
	}
}

func (svc *ConfigService) loadContextConfigurationRootFile(context *Context) error {
	if svc.GlobalConfiguration.verbose {
		notify.Info("Loading configuration file %s for context %s", context.RootConfigFilePath, context.Name)
	}

	path, err := os.Executable()
	if err != nil {
		return err
	}

	if strings.HasPrefix(context.RootConfigFilePath, ".\\") {
		context.RootConfigFilePath = strings.ReplaceAll(context.RootConfigFilePath, ".\\", "")
		context.RootConfigFilePath = helper.JoinPath(filepath.Dir(path), context.RootConfigFilePath)
	}
	if strings.HasPrefix(context.RootConfigFilePath, "./") {
		context.RootConfigFilePath = strings.ReplaceAll(context.RootConfigFilePath, "./", "")
		context.RootConfigFilePath = helper.JoinPath(filepath.Dir(path), context.RootConfigFilePath)
	}

	if !helper.FileExists(context.RootConfigFilePath) {
		err := fmt.Errorf("the context configuration file %s for context %s does not exist,", context.RootConfigFilePath, context.Name)
		notify.Error(err.Error())
		return err
	}

	context.Source = context.RootConfigFilePath

	content, err := helper.ReadFromFile(context.RootConfigFilePath)

	if err != nil {
		notify.FromError(err, "There was an error reading the context configuration file")
		return err
	}
	if err := yaml.Unmarshal(content, context); err != nil {
		if err := json.Unmarshal(content, context); err != nil {
			notify.FromError(err, "There was an error reading the context configuration file")
			return err
		}
	}

	return nil
}

func (svc *ConfigService) sanitizeConfigFolderPath(context *Context, folderPath string) (string, error) {
	if context.Configuration.ConfigFolder == "" || !helper.FileExists(context.Configuration.ConfigFolder) {
		return "", fmt.Errorf("config folder %s does not exists, skipping loading", context.Configuration.ConfigFolder)
	}

	// If the provided folder path is not a folder, then get the parent folder
	fileInfo, err := os.Stat(folderPath)
	if err != nil {
		return "", err
	}

	if !fileInfo.IsDir() {
		notify.Debug(folderPath)
		folderPath = filepath.Base(folderPath)
	}

	return folderPath, nil
}

func (svc *ConfigService) loadContextConfigurationFragment(context *Context, folderPath string, fileName string) error {
	var configFile Context
	content, err := helper.ReadFromFile(helper.JoinPath(folderPath, fileName))
	if err != nil {
		notify.FromError(err, "There was an error reading the configuration file %s for %s context", fileName, context.Name)
		return err
	}

	if err := yaml.Unmarshal(content, &configFile); err != nil {
		if err := json.Unmarshal(content, &configFile); err != nil {
			notify.FromError(err, "There was an error reading the configuration file %s for %s context", fileName, context.Name)
			return err
		} else {
			configFile.Source = helper.JoinPath(folderPath, fileName)
		}
	} else {
		configFile.Source = helper.JoinPath(folderPath, fileName)
	}

	if configFile.Source != "" {
		if svc.GlobalConfiguration.verbose {
			notify.Info("Loading content of config file %s for %s context", fileName, context.Name)
		}
		if configFile.Configuration != nil {
			context.Configuration = configFile.Configuration
		}

		if configFile.EnvironmentVariables != nil {
			// Sync the several variables
			for key, value := range configFile.EnvironmentVariables.Global {
				context.EnvironmentVariables.Global[key] = value
			}
			for key, value := range configFile.EnvironmentVariables.KeyVault {
				context.EnvironmentVariables.KeyVault[key] = value
			}
			for key, value := range configFile.EnvironmentVariables.Terraform {
				context.EnvironmentVariables.Terraform[key] = value
			}
		}

		for _, pipeline := range configFile.Pipelines {
			pipeline.source = configFile.Source
		}
		context.Pipelines = append(context.Pipelines, configFile.Pipelines...)

		for _, m := range configFile.SpaServices {
			m.source = configFile.Source
		}
		context.SpaServices = append(context.SpaServices, configFile.SpaServices...)

		for _, m := range configFile.Tenants {
			m.source = configFile.Source
		}
		context.Tenants = append(context.Tenants, configFile.Tenants...)

		for _, m := range configFile.BackendServices {
			m.Source = configFile.Source
		}
		context.BackendServices = append(context.BackendServices, configFile.BackendServices...)

		for _, m := range configFile.MockServices {
			m.source = configFile.Source
		}
		context.MockServices = append(context.MockServices, configFile.MockServices...)

		if configFile.NugetPackages != nil {
			if context.NugetPackages == nil {
				context.NugetPackages = &NugetPackages{
					Packages: make([]*NugetPackage, 0),
				}
			}

			configFile.NugetPackages.source = configFile.Source
			if configFile.NugetPackages.OutputSource != "" {
				context.NugetPackages.OutputSource = configFile.NugetPackages.OutputSource
			}

			for _, m := range configFile.NugetPackages.Packages {
				m.source = configFile.Source
			}
			context.NugetPackages.Packages = append(context.NugetPackages.Packages, configFile.NugetPackages.Packages...)
		}

		if configFile.Infrastructure != nil {
			if context.Infrastructure == nil {
				context.Infrastructure = &Infrastructure{
					Stacks: make([]*InfrastructureStack, 0),
				}
			}
			configFile.Infrastructure.Source = configFile.Source
			if configFile.Infrastructure.ConfigFile != "" {
				context.Infrastructure.ConfigFile = configFile.Infrastructure.ConfigFile
				if err := svc.loadContextConfigurationFragments(context, context.Infrastructure.ConfigFile); err != nil {
					return err
				}
			}

			for _, m := range configFile.Infrastructure.Stacks {
				m.Source = configFile.Source
			}
			context.Infrastructure.Stacks = append(context.Infrastructure.Stacks, configFile.Infrastructure.Stacks...)
		}

		if context.Fragments == nil {
			context.Fragments = make([]*Context, 0)
		}

		context.Fragments = append(context.Fragments, &configFile)
	}

	return nil
}

func (svc *ConfigService) isConfigFile(fileName string) bool {
	// note that this returns true for both default and override config file
	return strings.HasSuffix(fileName, ".yml") ||
		strings.HasSuffix(fileName, ".yaml") ||
		strings.HasSuffix(fileName, ".json")
}

func (svc *ConfigService) isDefaultConfigFile(fileName string) bool {
	return (strings.HasSuffix(fileName, ".yml") && !strings.HasSuffix(fileName, OVERRIDE_CONFIG_FILE_MARKER+".yml")) ||
		(strings.HasSuffix(fileName, ".yaml") && !strings.HasSuffix(fileName, OVERRIDE_CONFIG_FILE_MARKER+".yaml")) ||
		(strings.HasSuffix(fileName, ".json") && !strings.HasSuffix(fileName, OVERRIDE_CONFIG_FILE_MARKER+".json"))
}

func (svc *ConfigService) overrideConfigFileExists(folderPath string, fileName string) bool {

	// fileName is expected to be a default config file i.e. does not have extension .override.<ext>

	if strings.HasSuffix(fileName, ".yml") {
		overrideFilePath := strings.TrimSuffix(helper.JoinPath(folderPath, fileName), ".yml") + OVERRIDE_CONFIG_FILE_MARKER + ".yml"
		return helper.FileExists(overrideFilePath)
	}

	if strings.HasSuffix(fileName, ".yaml") {
		overrideFilePath := strings.TrimSuffix(helper.JoinPath(folderPath, fileName), ".yaml") + OVERRIDE_CONFIG_FILE_MARKER + ".yaml"
		return helper.FileExists(overrideFilePath)
	}

	if strings.HasSuffix(fileName, ".json") {
		overrideFilePath := strings.TrimSuffix(helper.JoinPath(folderPath, fileName), ".json") + OVERRIDE_CONFIG_FILE_MARKER + ".json"
		return helper.FileExists(overrideFilePath)
	}

	return false
}

func (svc *ConfigService) loadContextConfigurationFragments(context *Context, folderPath string) error {

	// Recurse through the sub-folders and load all the fragment config files found

	folderPath, err := svc.sanitizeConfigFolderPath(context, folderPath)
	if err != nil {
		return err
	}

	if helper.FileExists(folderPath) {
		if svc.GlobalConfiguration.verbose {
			notify.InfoWithIcon(icons.IconBook, "Loading content from the configuration folder %s", folderPath)
		}
		if files, err := os.ReadDir(folderPath); err == nil {
			for _, file := range files {
				if file.IsDir() {
					if err := svc.loadContextConfigurationFragments(context, helper.JoinPath(folderPath, file.Name())); err != nil {
						return err
					}
				}

				// Load the fragment config file. If an override file exists for it, then load that instead of the default file.
				// Note that this is NOT a "merge override" operation where individual config items in the default file
				// are overridden. If an override file exists then the default file will be completely ignored.

				if !file.IsDir() && svc.isConfigFile(file.Name()) {
					if svc.isDefaultConfigFile(file.Name()) && svc.overrideConfigFileExists(folderPath, file.Name()) {
						// Ignore this default config file as its override file exists which will be loaded instead
						continue
					}

					if err := svc.loadContextConfigurationFragment(context, folderPath, file.Name()); err != nil {
						return err
					}
				}
			}
		} else {
			notify.Error("No configuration file was found")
			return errors.New("no configuration file was found")
		}
	}

	return nil
}

func (svc *ConfigService) getGlobalConfigFile() string {
	if file := helper.GetFlagValue("file", ""); file != "" {
		return file
	}
	if file := helper.GetFlagValue("f", ""); file != "" {
		return file
	}

	return svc.getDefaultGlobalConfigFile()
}

func (svc *ConfigService) getDefaultGlobalConfigFile() string {

	exPath := common.GetExeDirectoryPath()

	// Try these file names in order
	fileNames := []string{
		"locally-config.personal.yml",
		"locally-config.personal.yaml",
		"locally-config.yml",
		"locally-config.yaml",
		"config.personal.yml",
		"config.personal.yaml",
		"config.yml",
		"config.yaml",
		"config.personal.json",
		"config.json",
	}

	for _, fileName := range fileNames {
		if helper.FileExists(helper.JoinPath(exPath, fileName)) {
			svc.configFilename = helper.JoinPath(exPath, fileName)
			return svc.configFilename
		}
	}

	return ""
}

func (svc *ConfigService) SetCurrentContext(context string) error {
	if svc.GetContext(context) != nil {
		svc.GlobalConfiguration.CurrentContext = context
		svc.SaveConfigFile()
		notify.Success("Current context changed to %s", context)
		return nil
	}

	return fmt.Errorf("context %s was not found in the current configuration file", context)
}

func (svc *ConfigService) SaveConfigFile() {
	var config GlobalConfiguration
	configContent, err := helper.ReadFromFile(svc.configFilename)
	if err != nil {
		notify.FromError(err, "There was an error reading the configuration file")
	}

	switch svc.GlobalConfiguration.format {
	case "json":
		if err := json.Unmarshal(configContent, &config); err != nil {
			notify.FromError(err, "There was an error reading the configuration file")
		}
		config.CurrentContext = svc.GlobalConfiguration.CurrentContext
		if svc.GlobalConfiguration.Cors != nil {
			config.Cors = svc.GlobalConfiguration.Cors
		}

		// if svc.Configuration.CertificateGenerator != nil {
		// 	config.CertificateGenerator = svc.Configuration.CertificateGenerator
		// }
		content, err := json.MarshalIndent(config, "", "  ")
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
		}

		helper.WriteToFile(string(content), svc.configFilename)
	default:
		if err := yaml.Unmarshal(configContent, &config); err != nil {
			notify.FromError(err, "There was an error reading the configuration file")
		}
		config.CurrentContext = svc.GlobalConfiguration.CurrentContext
		if svc.GlobalConfiguration.Cors != nil {
			config.Cors = svc.GlobalConfiguration.Cors
		}

		// if svc.Configuration.CertificateGenerator != nil {
		// 	config.CertificateGenerator = svc.Configuration.CertificateGenerator
		// }

		content, err := yaml.Marshal(config)
		if err != nil {
			notify.FromError(err, "Unable to set the current context in configuration file")
		}

		helper.WriteToFile(string(content), svc.configFilename)
	}
}

func (svc *ConfigService) ListAllServices() {
	context := svc.GetCurrentContext()
	mockServices := context.MockServices

	spaServices := svc.GetSpaServicesByTags()
	backendServices := svc.GetBackendServicesByTags()
	nugetPackages := svc.GetNugetPackagesByTags()
	stacks := svc.GetInfrastructureStacksByTags()

	notify.InfoWithIcon(icons.IconClipboard, "Listing all  %s context services:", svc.GlobalConfiguration.CurrentContext)

	notify.InfoWithIcon(icons.IconFolder, "Pipelines:")
	BuildDependencyGraph(context.Pipelines, false)

	if len(context.Pipelines) > 0 {
		for _, pipeline := range context.Pipelines {
			notify.InfoIndentIcon(icons.IconFolder, "%s", "  ", pipeline.Name)
			for _, job := range pipeline.Jobs {
				notify.InfoIndentIcon(icons.IconFolder, "%s", "    ", job.Name)
				for _, step := range job.Steps {
					notify.InfoIndentIcon(icons.IconPage, "%s", "      ", step.Name)
				}
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No Pipelines found", "  ")
	}

	notify.InfoWithIcon(icons.IconFolder, "Infrastructure Stacks:")
	if len(stacks) > 0 {
		for _, stack := range stacks {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", stack.Name, stack.Source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", stack.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No Infrastructure Stacks found", "  ")
	}

	notify.InfoWithIcon(icons.IconFolder, "Backend Services:")
	if len(backendServices) > 0 {
		for _, backendService := range backendServices {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", backendService.Name, backendService.Source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", backendService.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No Backend Services found", "  ")
	}

	notify.InfoWithIcon(icons.IconFolder, "SPA Services:")
	if len(spaServices) > 0 {
		for _, spaService := range spaServices {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, " %s in %s", "  ", spaService.Name, spaService.source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", spaService.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No SPA Services found", "  ")
	}

	notify.InfoWithIcon(icons.IconFolder, "Tenants:")
	if len(context.Tenants) > 0 {
		for _, tenant := range context.Tenants {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", tenant.Name, tenant.source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", tenant.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No Tenants found", "  ")
	}

	notify.InfoWithIcon(icons.IconFolder, "Mock Services:")
	if len(mockServices) > 0 {
		for _, mockService := range mockServices {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", mockService.Name, mockService.source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", mockService.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No Mock Services found", "  ")
	}

	notify.InfoWithIcon(icons.IconFolder, "Nuget Packages:")
	if len(nugetPackages) > 0 {
		for _, nugetPackage := range nugetPackages {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", nugetPackage.Name, nugetPackage.source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", nugetPackage.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No Nuget Packages found", "  ")
	}
}

func (svc *ConfigService) ListAllBackendServices() {
	backendServices := svc.GetBackendServicesByTags()

	notify.InfoWithIcon(icons.IconClipboard, "Listing all  %s context backend services:", svc.GlobalConfiguration.CurrentContext)
	if len(backendServices) > 0 {
		for _, service := range backendServices {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", service.Name, service.Source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", service.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No backend services found", "  ")
	}
}

func (svc *ConfigService) ListAllSPAServices() {
	spaServices := svc.GetSpaServicesByTags()

	notify.InfoWithIcon(icons.IconClipboard, "Listing all %s context SPA services", svc.GlobalConfiguration.CurrentContext)
	if len(spaServices) > 0 {
		for _, service := range spaServices {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", service.Name, service.source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", service.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No spa services found", "  ")
	}
}

func (svc *ConfigService) ListAllTenants() {
	context := svc.GetCurrentContext()
	notify.InfoWithIcon(icons.IconClipboard, "Listing all %s context tenants:", svc.GlobalConfiguration.CurrentContext)
	if len(context.Tenants) > 0 {
		for _, tenant := range context.Tenants {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", tenant.Name, tenant.source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", tenant.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No tenants found", "  ")
	}
}

func (svc *ConfigService) ListAllMockServices() {
	context := svc.GetCurrentContext()
	notify.InfoWithIcon(icons.IconClipboard, "Listing all %s context mock services", svc.GlobalConfiguration.CurrentContext)
	if len(context.Tenants) > 0 {
		for _, service := range context.MockServices {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", service.Name, service.source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", service.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No tenants found", "  ")
	}
}

func (svc *ConfigService) ListAllPipelines() {
	context := svc.GetCurrentContext()
	notify.InfoWithIcon(icons.IconClipboard, "Listing all %s Pipelines", svc.GlobalConfiguration.CurrentContext)
	notify.InfoWithIcon(icons.IconFolder, "Pipelines:")
	BuildDependencyGraph(context.Pipelines, false)
	if len(context.Pipelines) > 0 {
		for _, pipeline := range context.Pipelines {
			notify.InfoIndentIcon(icons.IconFolder, "%s", "  ", pipeline.Name)
			for _, job := range pipeline.Jobs {
				notify.InfoIndentIcon(icons.IconFolder, "%s", "    ", job.Name)
				for _, step := range job.Steps {
					notify.InfoIndentIcon(icons.IconPage, "%s", "      ", step.Name)
				}
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No pipelines found", "  ")
	}
}

func (svc *ConfigService) ListAllNugetPackages() {
	context := svc.GetCurrentContext()

	if context.NugetPackages == nil {
		notify.Warning("No nuget packages configuration found")
		return
	}

	nugetPackages := svc.GetNugetPackagesByTags()

	notify.InfoWithIcon(icons.IconClipboard, "Listing all %s context nuget packages:", svc.GlobalConfiguration.CurrentContext)

	if len(context.Pipelines) > 0 {
		for _, nugetPackage := range nugetPackages {
			if svc.Verbose() {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", nugetPackage.Name, nugetPackage.source)
			} else {
				notify.InfoIndentIcon(icons.IconBlackSquare, "%s", "  ", nugetPackage.Name)
			}
		}
	} else {
		notify.InfoIndentIcon(icons.IconBlackSquare, "No nuget packages found", "  ")
	}
}

func (svc *ConfigService) ListAllInfrastructureStacks() {
	context := svc.GetCurrentContext()

	if svc.Debug() {
		result, _ := json.Marshal(context.Infrastructure)
		notify.Debug(string(result))
	}
	if context.Infrastructure == nil {
		notify.Info("No infrastructure stacks configuration found")
		return
	}

	stacks := svc.GetInfrastructureStacksByTags()

	notify.InfoWithIcon(icons.IconClipboard, "Listing all %s context infrastructure stacks:", svc.GlobalConfiguration.CurrentContext)

	for _, stack := range stacks {
		if svc.Verbose() {
			notify.InfoIndentIcon(icons.IconBlackSquare, "%s in %s", "  ", stack.Name, stack.Source)
		} else {
			notify.InfoIndentIcon(icons.IconBlackSquare, "%s", stack.Name)
		}
	}
}

func (svc *ConfigService) PrintCurrentContext() {
	currentContext := svc.GlobalConfiguration.CurrentContext

	notify.InfoWithIcon(icons.IconBell, "Current context is set to %s", currentContext)
}

func (svc *ConfigService) GetDockerServices(name string, ignoreTags bool) []*DockerContainer {
	result := make([]*DockerContainer, 0)

	var backendServices []*BackendService
	var frontendServices []*SpaService

	if ignoreTags {
		notify.Debug("Ignoring flags, getting one by one")
		backendServices = svc.GetCurrentContext().BackendServices
		frontendServices = svc.GetCurrentContext().SpaServices
	} else {
		backendServices = svc.GetBackendServicesByTags()
		frontendServices = svc.GetSpaServicesByTags()
	}

	for _, service := range backendServices {
		if !service.HasPath() {
			continue
		}

		container := DockerContainer{
			Name:           service.Name,
			Location:       service.Location,
			Repository:     service.Repository,
			DependsOn:      service.DependsOn,
			RequiredBy:     service.RequiredBy,
			Source:         service.Source,
			DockerRegistry: service.DockerRegistry,
			DockerCompose:  service.DockerCompose,
			Tags:           service.Tags,
			Components:     make([]*DockerContainer, 0),
		}

		for _, component := range service.Components {
			componentContainer := DockerContainer{
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

		container := DockerContainer{
			Name:           service.Name,
			Location:       service.Location,
			Repository:     service.Repository,
			DependsOn:      service.DependsOn,
			RequiredBy:     service.RequiredBy,
			Source:         service.source,
			DockerCompose:  service.DockerCompose,
			DockerRegistry: service.DockerRegistry,
			Components:     make([]*DockerContainer, 0),
			Tags:           service.Tags,
		}

		componentContainer := DockerContainer{
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

	if helper.GetFlagSwitch("all", false) || svc.HasTags() && !ignoreTags {
		return result
	} else {
		if name == "" {
			return make([]*DockerContainer, 0)
		}

		filteredContainers := make([]*DockerContainer, 0)
		for _, container := range result {
			if strings.EqualFold(EncodeName(container.Name), name) {
				filteredContainers = append(filteredContainers, container)
			}
		}

		return filteredContainers
	}
}

func (svc *ConfigService) Verbose() bool {
	return helper.GetFlagSwitch("verbose", false)
}

func (svc *ConfigService) Debug() bool {
	return helper.GetFlagSwitch("debug", false)
}

func (svc *ConfigService) PrintContextFragments() {
	context := svc.GetCurrentContext()
	notify.InfoWithIcon(icons.IconMagnifyingGlass, "Listing all %s context fragments", context.Name)
	for _, c := range context.Fragments {
		notify.InfoWithIcon(icons.IconBook, "%s", c.Source)
	}
}

func (svc *ConfigService) GetFragment(source string) *Context {
	context := svc.GetCurrentContext()
	for _, s := range context.Fragments {
		if strings.EqualFold(source, s.Source) {
			return s
		}
	}

	return nil
}

func (svc *ConfigService) GetFragmentInfrastructureStack(fragment *Context, name string) *InfrastructureStack {
	if fragment == nil || fragment.Infrastructure == nil || len(fragment.Infrastructure.Stacks) == 0 {
		return nil
	}

	for _, s := range fragment.Infrastructure.Stacks {
		if strings.EqualFold(s.Name, name) {
			return s
		}
	}

	return nil
}

func (svc *ConfigService) GetInfrastructureDependencies(stacks []*InfrastructureStack) ([]*InfrastructureStack, error) {
	context := svc.GetCurrentContext()
	for {
		added := false
		for _, stack := range stacks {
			for _, dependencyName := range stack.DependsOn {
				needStack := context.Infrastructure.GetStackByName(dependencyName)
				if needStack == nil {
					//lint:ignore ST1005 #
					err := fmt.Errorf("Cannot find the required dependency %s for stack %s in the configuration file", dependencyName, stack.Name)
					return nil, err
				}

				found := false
				for _, s := range stacks {
					if strings.EqualFold(s.Name, needStack.Name) {
						found = true
						break
					}
				}

				if !found {
					stacks = append(stacks, needStack)
					added = true
					break
				}
			}
		}

		if !added {
			break
		}
	}

	BuildDependencyTree(stacks)

	return stacks, nil
}

func (svc *ConfigService) GetDockerContainerDependencies(containers []*DockerContainer) ([]*DockerContainer, error) {
	context := svc.GetCurrentContext()
	for {
		added := false
		for _, container := range containers {
			for _, dependencyName := range container.DependsOn {
				fmt.Print(dependencyName)
				needContainer := context.GetContainerFragmentByName(dependencyName)
				if needContainer == nil {
					//lint:ignore ST1005 #
					err := fmt.Errorf("Cannot find the required dependency %s for container %s in the configuration file", dependencyName, container.Name)
					return nil, err
				}
			}
			for _, component := range container.Components {
				for _, dependencyName := range component.DependsOn {
					fmt.Print(dependencyName)
					needContainer := context.GetContainerFragmentByName(dependencyName)
					if needContainer == nil {
						//lint:ignore ST1005 #
						err := fmt.Errorf("Cannot find the required dependency %s for container %s in the configuration file", dependencyName, container.Name)
						return nil, err
					}
				}
			}
			BuildDependencyTree(container.Components)
		}

		if !added {
			break
		}
	}

	BuildDependencyTree(containers)

	return containers, nil
}
