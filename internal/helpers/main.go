package helpers

import (
	"strconv"

	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/executer"
	"github.com/cjlapao/locally-cli/internal/icons"

	"github.com/cjlapao/common-go/log"
)

func GetCaddyPath() string {
	config := configuration.Get()
	if config == nil || config.GlobalConfiguration == nil || config.GlobalConfiguration.Tools == nil || config.GlobalConfiguration.Tools.Caddy == nil {
		return "caddy"
	}

	if config.GlobalConfiguration.Tools.Caddy.Path != "" {
		return config.GlobalConfiguration.Tools.Caddy.Path
	} else {
		return "caddy"
	}
}

func GetNugetPath() string {
	config := configuration.Get()
	if config == nil || config.GlobalConfiguration == nil || config.GlobalConfiguration.Tools == nil || config.GlobalConfiguration.Tools.Nuget == nil {
		return "nuget"
	}

	if config.GlobalConfiguration.Tools.Nuget.Path != "" {
		return config.GlobalConfiguration.Tools.Nuget.Path
	} else {
		return "nuget"
	}
}

func GetDockerPath() string {
	config := configuration.Get()
	if config == nil || config.GlobalConfiguration == nil || config.GlobalConfiguration.Tools == nil || config.GlobalConfiguration.Tools.Docker == nil {
		return "docker"
	}

	if config.GlobalConfiguration.Tools.Docker.DockerPath != "" {
		return config.GlobalConfiguration.Tools.Docker.DockerPath
	} else {
		return "docker"
	}
}

func GetDockerComposePath() string {
	config := configuration.Get()
	if config == nil || config.GlobalConfiguration == nil || config.GlobalConfiguration.Tools == nil || config.GlobalConfiguration.Tools.Docker == nil {
		return "docker-compose"
	}

	if config.GlobalConfiguration.Tools.Docker.ComposerPath != "" {
		return config.GlobalConfiguration.Tools.Docker.ComposerPath
	} else {
		return "docker-compose"
	}
}

func GetTerraformPath() string {
	config := configuration.Get()
	if config == nil || config.GlobalConfiguration == nil || config.GlobalConfiguration.Tools == nil || config.GlobalConfiguration.Tools.Terraform == nil {
		return "terraform"
	}

	if config.GlobalConfiguration.Tools.Terraform.Path != "" {
		return config.GlobalConfiguration.Tools.Terraform.Path
	} else {
		return "terraform"
	}
}

func GetDotnetPath() string {
	config := configuration.Get()
	if config == nil || config.GlobalConfiguration == nil || config.GlobalConfiguration.Tools == nil || config.GlobalConfiguration.Tools.Dotnet == nil {
		return "dotnet"
	}
	if config.GlobalConfiguration.Tools.Dotnet.Path != "" {
		return config.GlobalConfiguration.Tools.Dotnet.Path
	} else {
		return "dotnet"
	}
}

func GetAzureCliPath() string {
	config := configuration.Get()
	if config == nil || config.GlobalConfiguration == nil || config.GlobalConfiguration.Tools == nil || config.GlobalConfiguration.Tools.AzureCli == nil {
		return "az"
	}
	if config.GlobalConfiguration.Tools.AzureCli.Path != "" {
		return config.GlobalConfiguration.Tools.AzureCli.Path
	} else {
		return "az"
	}
}

func GetNpmPath() string {
	config := configuration.Get()
	if config == nil || config.GlobalConfiguration == nil || config.GlobalConfiguration.Tools == nil || config.GlobalConfiguration.Tools.Npm == nil {
		return "npm"
	}

	if config.GlobalConfiguration.Tools.Npm.Path != "" {
		return config.GlobalConfiguration.Tools.Npm.Path
	} else {
		return "npm"
	}
}

func Retry(name string, command string, arguments []string, verbose bool) (string, error) {
	config := configuration.Get()
	logger := log.Get()

	retryCount := common.DEFAULT_RETRY_COUNT
	var err error
	var output executer.ExecuteOutput

	if config.GlobalConfiguration != nil && config.GlobalConfiguration.Tools != nil && config.GlobalConfiguration.Tools.Docker != nil && config.GlobalConfiguration.Tools.Docker.BuildRetries > 0 {
		retryCount = config.GlobalConfiguration.Tools.Docker.BuildRetries
	}

	for retryCount >= 0 {
		if verbose {
			output, err = executer.Execute(command, arguments...)
		} else {
			output, err = executer.ExecuteWithNoOutput(command, arguments...)
		}

		if err != nil {
			if retryCount == 0 {
				logger.Exception(err, "%s Something wrong running %v", icons.IconRevolvingLight, command)
				return output.GetAllOutput(), err
			} else {
				logger.Warn("%s There was an error running %v, retrying %v more time(s)", icons.IconWarning, name, strconv.Itoa(retryCount))
				retryCount -= 1
			}
		} else {
			break
		}
	}

	return output.GetAllOutput(), nil
}
