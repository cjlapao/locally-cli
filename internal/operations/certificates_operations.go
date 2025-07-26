package operations

import (
	"os"

	"github.com/cjlapao/locally-cli/internal/certificates"
	"github.com/cjlapao/locally-cli/internal/help"

	"github.com/cjlapao/common-go/helper"
)

func CertificatesOperations(subCommand string) {
	if subCommand == "" && helper.GetFlagSwitch("help", false) {
		help.ShowHelpForCertificatesCommand()
		os.Exit(0)
	}

	switch subCommand {
	case "generate":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForCertificatesCommand()
			os.Exit(0)
		}

		certificates.GenerateCertificates()
	case "clean":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForCertificatesCommand()
			os.Exit(0)
		}

		certificates.CleanConfig()
	default:
		help.ShowHelpForCertificatesCommand()
		os.Exit(0)
	}
}
