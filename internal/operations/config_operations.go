package operations

import (
	"os"

	"github.com/cjlapao/locally-cli/internal/caddy"
	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/docker"
	"github.com/cjlapao/locally-cli/internal/help"
	"github.com/cjlapao/locally-cli/internal/system"

	"github.com/cjlapao/common-go/helper"
)

func ConfigOperations(subCommand string) {
	config := configuration.Get()
	systemService := system.Get()
	dockerService := docker.Get()

	if subCommand == "" && helper.GetFlagSwitch("help", false) {
		help.ShowHelpForConfigCommand()
		os.Exit(0)
	}

	switch subCommand {
	case "set-context":
		component := common.VerifyCommand(helper.GetArgumentAt(2))
		if component == "" || helper.GetFlagSwitch("help", false) {
			help.ShowHelpForConfigSetContextCommand()
			os.Exit(0)
		}

		if err := config.SetCurrentContext(component); err != nil {
			notify.Error(err.Error())
			os.Exit(1)
		}

		for _, service := range config.GetCurrentContext().BackendServices {
			notify.Reset()
			options := &docker.DockerServiceOptions{
				Name:              service.Name,
				BuildDependencies: false,
			}

			dockerService.GenerateServiceDockerComposeOverrideFile(options)
		}

		if notify.HasErrors() {
			notify.Error("There was an error generating docker-compose override")
		}
		systemService.CheckFolders(true)
		caddy := caddy.Get()
		caddy.GenerateDockerFiles()
		if err := caddy.GenerateCaddyFiles(); err != nil {
			notify.FromError(err, "There was an error generating caddy files")
		}
	case "list":
		component := common.VerifyCommand(helper.GetArgumentAt(2))
		if component == "" && helper.GetFlagSwitch("help", false) {
			help.ShowHelpForConfigListCommand()
			os.Exit(0)
		}
		switch component {
		case "spa-services":
			config.ListAllSPAServices()
		case "backend-services":
			config.ListAllBackendServices()
		case "tenants":
			config.ListAllTenants()
		case "mock-services":
			config.ListAllMockServices()
		case "nuget-packages":
			config.ListAllNugetPackages()
		case "infrastructure":
			config.ListAllInfrastructureStacks()
		case "pipelines":
			config.ListAllPipelines()
		default:
			config.ListAllServices()
		}
	case "current-context":
		config.PrintCurrentContext()
	case "list-fragments":
		config.PrintContextFragments()
	case "clean":
		HandleCleanCommand(config)

	default:
		help.ShowHelpForConfigCommand()
		os.Exit(0)
	}
}

func HandleCleanCommand(config *configuration.ConfigService) {
	if helper.GetFlagSwitch("help", false) {
		help.ShowHelpForConfigCleanCommand()
		os.Exit(0)
	}
	component := common.VerifyCommand(helper.GetArgumentAt(2))
	if component == "" {
		config.CleanContextConfigurationCurrent()
	} else if component == "--all" {
		config.CleanContextConfigurationAll()
	} else {
		help.ShowHelpForConfigCleanCommand()
		os.Exit(0)
	}
}
