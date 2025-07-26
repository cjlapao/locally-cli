package operations

import (
	"os"

	"github.com/cjlapao/locally-cli/internal/help"
	"github.com/cjlapao/locally-cli/internal/hosts"

	"github.com/cjlapao/common-go/helper"
)

func HostsOperations(subCommand string) {
	hostsSvc := hosts.Get()

	if subCommand == "" && helper.GetFlagSwitch("help", false) {
		help.ShowHelpForHostsCommand()
		os.Exit(0)
	}

	switch subCommand {
	case "update":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForHostsUpdateCommand()
			os.Exit(0)
		}
		if err := hostsSvc.GenerateHostsEntries(); err != nil {
			notify.FromError(err, "There was an error building Services Containers")
		}
	case "clean":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForHostsCleanCommand()
			os.Exit(0)
		}
		if err := hostsSvc.Clean(); err != nil {
			notify.FromError(err, "There was an error building Services Containers")
		}
	default:
		help.ShowHelpForHostsCommand()
		os.Exit(0)
	}
}
