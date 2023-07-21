package operations

import (
	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/help"
	"github.com/cjlapao/locally-cli/icons"
	"github.com/cjlapao/locally-cli/tools"
	"os"

	"github.com/cjlapao/common-go/helper"
)

func ToolsOperations(subCommand string) {
	if subCommand == "" && helper.GetFlagSwitch("help", false) {
		help.ShowHelpForToolsCommand()
		os.Exit(0)
	}

	switch subCommand {
	case "ems":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForEmsTool()
			os.Exit(0)
		}
		toolArgument := common.VerifyCommand(helper.GetArgumentAt(2))
		switch toolArgument {
		case "apikey":
			apiKey := helper.GetFlagValue("key", "")
			if helper.GetFlagSwitch("help", false) {
				help.ShowHelpForEmsApiKeyTool()
				os.Exit(0)
			}

			notify.Debug("ApiKey Flag: %s", apiKey)
			ems := tools.GetEmsServiceTool()
			if result, err := ems.GenerateEmsApiKeyHeader(apiKey); err != nil {
				notify.FromError(err, "Error generating key")
			} else {
				notify.InfoWithIcon(icons.IconKey, "API %s", result)
			}
		}
	case "base64":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForBase64Tool()
			os.Exit(0)
		}

		toolArgument := common.VerifyCommand(helper.GetArgumentAt(2))
		value := common.VerifyCommand(helper.GetArgumentAt(3))
		if value == "" || toolArgument == "" {
			help.ShowHelpForBase64Tool()
			os.Exit(0)
		}

		switch toolArgument {
		case "decode":
			decoder := tools.Base64Tool{}
			result, err := decoder.Decode(value)
			if err != nil {
				notify.Error("Could not decode string")
			}

			notify.Success("%s", result)

		case "encode":
			decoder := tools.Base64Tool{}
			result := decoder.Encode(value)
			if result == "" {
				notify.Error("Could not encode string")
			}

			notify.Success("%s", result)
		default:
			help.ShowHelpForBase64Tool()
			os.Exit(0)
		}
	default:
		help.ShowHelpForToolsCommand()
		os.Exit(0)
	}
}
