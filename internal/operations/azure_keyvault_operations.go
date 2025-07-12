package operations

import (
	"os"

	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/help"
	"github.com/cjlapao/locally-cli/internal/vaults/azure_keyvault"

	"github.com/cjlapao/common-go/helper"
)

func AzureKeyvaultOperations(subCommand string) {
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
		kv := azure_keyvault.New("", &azure_keyvault.AzureKeyVaultOptions{
			KeyVaultUri:  url,
			DecodeBase64: true,
		})

		if _, err := kv.Sync(); err != nil {
			notify.FromError(err, "Error failed to sync keyvault")
		}
	default:
		help.ShowHelpForAzureKeyvaultCommand()
		os.Exit(0)
	}
}
