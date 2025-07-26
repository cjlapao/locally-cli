package environment

import (
	"fmt"
	"os"

	"github.com/cjlapao/locally-cli/internal/appctx"
	"github.com/cjlapao/locally-cli/internal/help"
	"github.com/cjlapao/locally-cli/internal/icons"

	"github.com/cjlapao/common-go/helper"
)

func Operations(ctx *appctx.AppContext, variable string) {
	env := GetInstance()

	listAll := helper.GetFlagSwitch("list-all", false)
	if variable == "" && helper.GetFlagSwitch("help", false) {
		help.ShowHelpForEnvironment()
		os.Exit(0)
	}

	if variable == "" && !listAll {
		notify.Error("Variable cannot be empty")
		return
	}

	if listAll {
		for _, vaultName := range env.ListVaults(ctx) {
			variables, exists := env.GetAllVariables(ctx, vaultName)
			if !exists {
				notify.Error("Vault %s not found", vaultName)
				continue
			}
			for key, value := range variables {
				notify.Info("%s.%s: %v", vaultName, key, value)
			}
		}
		return
	}

	notify.InfoWithIcon(icons.IconMagnifyingGlass, "Trying to find the value for %s in the environment", variable)
	newvar := fmt.Sprintf("${{ %s }}", variable)

	result := env.Replace(ctx, newvar)

	if result == newvar {
		notify.Error("Environment variable with name %s was not found", variable)
		os.Exit(1)
	} else {
		notify.Success("Variable %s was found in the environment with value:\n%s", variable, result)
		os.Exit(0)
	}
}
