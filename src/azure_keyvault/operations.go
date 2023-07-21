package azure_keyvault

import (
	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/help"
	"os"

	"github.com/cjlapao/common-go/helper"
)

func Operations(subCommand string) {
	if subCommand == "" && helper.GetFlagSwitch("help", false) {
		help.ShowHelpForAzureKeyvaultCommand()
		os.Exit(0)
	}

	switch subCommand {
	case "sync":
		url := common.VerifyCommand(helper.GetArgumentAt(2))
		if url == "" && helper.GetFlagSwitch("help", false) {
			help.ShowHelpForAzureKeyvaultSyncKeyvaultCommand()
			os.Exit(0)
		}
		kv := New("", &AzureKeyVaultOptions{
			KeyVaultUri:  url,
			DecodeBase64: true,
		})

		if _, err := kv.Sync(); err != nil {
			notify.FromError(err, "Error failing to syn keyvault")
		}
	default:
		help.ShowHelpForAzureKeyvaultCommand()
		os.Exit(0)
	}
}
