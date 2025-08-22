package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/environment"
	"github.com/cjlapao/locally-cli/internal/help"
	"github.com/cjlapao/locally-cli/internal/lanes"
	"github.com/cjlapao/locally-cli/internal/notifications"
	"github.com/cjlapao/locally-cli/internal/operations"
	"github.com/cjlapao/locally-cli/internal/system"

	"github.com/cjlapao/common-go/helper"
	"github.com/cjlapao/common-go/log"
	"github.com/cjlapao/common-go/version"
)

var (
	releaseVersion = "0.0.1"
	versionSvc     = version.Get()
)

var (
	logger = log.Get()
	notify = notifications.Get()
)

func main() {
	SetVersion()
	ctx := appctx.NewContext(context.Background())
	getVersion := helper.GetFlagSwitch("version", false)
	if getVersion {
		PrintVersion()
		os.Exit(0)
	}
	versionSvc.PrintAnsiHeader()
	command := strings.ToLower(helper.GetCommandAt(0))
	subCommand := strings.ToLower(helper.GetCommandAt(1))

	if helper.GetFlagSwitch("help", false) && command == "" {
		help.ShowHelpForNoCommand()
		os.Exit(0)
	}

	if command == "" {
		help.ShowHelpForNoCommand()
		os.Exit(1)
	}

	if helper.GetFlagSwitch("debug", false) {
		logger.WithDebug()
	}

	config := configuration.Get()
	if err := config.Init(); err != nil {
		notify.Critical("There was a critical error loading the configuration file")
	}

	operationsService := operations.Get()
	if command == "api" {
		apiOperation := operationsService.GetOperation(operations.API_OPERATION_NAME)
		apiOperation.Run()
	}

	systemService := system.Get()
	systemService.Init()

	switch command {
	case "test":
		fmt.Println("Test operations are disabled")
		// tester.TestOperations(subCommand)
	case "keyvault":
		operations.AzureKeyvaultOperations(subCommand)
	case "nuget":
		operations.NugetOperations(subCommand)
	case "config":
		operations.ConfigOperations(subCommand)
	case "certificates":
		operations.CertificatesOperations(subCommand)
	case "docker":
		operations.DockerOperations(subCommand, nil)
	case "proxy":
		operations.ProxyOperations(subCommand)
	case "hosts":
		operations.HostsOperations(subCommand)
	case "tools":
		operations.ToolsOperations(subCommand)
	case "lanes":
		lanes.Operations(subCommand)
	case "env":
		environment.Operations(ctx, subCommand)
	case "infrastructure":
		stack := common.VerifyCommand(helper.GetArgumentAt(2))
		operations.InfrastructureOperations(subCommand, stack, nil)
	default:
		help.ShowHelpForNoCommand()
		os.Exit(0)
	}

	finalMessage := "Finished"
	if notify.HasErrors() {
		finalMessage += fmt.Sprintf(", found %d error(s)", notify.CountErrors())
	}
	if notify.HasWarning() {
		finalMessage += fmt.Sprintf(", found %d warning(s)", notify.CountWarnings())
	}

	if notify.HasErrors() {
		notify.Error(finalMessage)
	} else if notify.HasWarning() {
		notify.Warning(finalMessage)
	} else {
		notify.Success(finalMessage)
	}

	if !notify.HasErrors() || !notify.HasWarning() {
		os.Exit(0)
	} else {
		os.Exit(1)
	}
}

func SetVersion() {
	versionSvc.Name = "Locally"
	versionSvc.Author = "Carlos Lapao"
	versionSvc.License = "Fair Source License"
	strVer, err := version.FromString(releaseVersion)
	if err == nil {
		versionSvc.Major = strVer.Major
		versionSvc.Minor = strVer.Minor
		versionSvc.Build = strVer.Build
		versionSvc.Rev = strVer.Rev
	}
}

func PrintVersion() {
	format := helper.GetFlagValue("o", "json")
	switch strings.ToLower(format) {
	case "json":
		fmt.Println(versionSvc.PrintVersion(int(version.JSON)))
	case "yaml":
		fmt.Println(versionSvc.PrintVersion(int(version.Yaml)))
	default:
		fmt.Println("Please choose a valid format, this can be either json or yaml")
	}
}
