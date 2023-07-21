package operations

import (
	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/help"
	"github.com/cjlapao/locally-cli/nugets"
	"os"

	"github.com/cjlapao/common-go/helper"
)

func NugetOperations(subCommand string) {
	nugetSvc := nugets.Get()

	nugetSvc.CheckForDotnet(false)
	nugetSvc.CheckForNuget(false)

	if subCommand == "" && helper.GetFlagSwitch("help", false) {
		help.ShowHelpForNugetCommand()
		os.Exit(0)
	}

	switch subCommand {
	case "generate":
		pkgName := common.VerifyCommand(helper.GetArgumentAt(2))
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForNugetGenerateCommand()
			os.Exit(0)
		}
		tags := helper.GetFlagArrayValue("tag")
		nugetSvc.GenerateNugetPackages(pkgName, tags...)
	default:
		help.ShowHelpForNugetCommand()
		os.Exit(0)
	}
}
