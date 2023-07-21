package pipelines

import (
	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/help"
	"os"

	"github.com/cjlapao/common-go/helper"
)

func Operations(subCommand string) {
	pipelinesService := Get()

	if subCommand == "" && helper.GetFlagSwitch("help", false) {
		help.ShowHelpForPipelinesCommand()
		help.ShowHelpForPipelineRunCommand()
		os.Exit(0)
	}

	pipeline := common.VerifyCommand(helper.GetArgumentAt(2))

	switch subCommand {
	case "run":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForPipelineRunCommand()
			os.Exit(0)
		}
		if err := pipelinesService.Validate(pipeline); err == nil {
			if err := pipelinesService.Run(pipeline); err != nil {
				notify.Error("There was an error executing the requested pipeline %s", pipeline)
			}
		} else {
			notify.Error("There was an error validating the requested pipeline %s", pipeline)
		}
	case "validate":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForPipelineValidateCommand()
			os.Exit(0)
		}
		if err := pipelinesService.Validate(pipeline); err != nil {
			notify.Error("There was an error validating the requested pipeline %s", pipeline)
		}
	case "list":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForPipelineListCommand()
			os.Exit(0)
		}

		config := configuration.Get()
		config.ListAllPipelines()
	default:
		help.ShowHelpForPipelinesCommand()
		os.Exit(0)
	}

}
