package environment

import (
	"fmt"
	"github.com/cjlapao/locally-cli/help"
	"github.com/cjlapao/locally-cli/icons"
	"os"

	"github.com/cjlapao/common-go/helper"
)

func Operations(variable string) {
	env := Get()

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
		for _, vault := range env.vaults {
			vaultName := vault.Name()
			keys, err := env.GetAll(vaultName)
			if err != nil {
				notify.Error(err.Error())
			}
			for _, key := range keys {
				notify.Info("%s.%s\n", vaultName, env.Replace(key))
			}
		}
		return
	}

	notify.InfoWithIcon(icons.IconMagnifyingGlass, "Trying to find the value for %s in the environment", variable)
	newvar := fmt.Sprintf("${{ %s }}", variable)

	result := env.Replace(newvar)

	if result == newvar {
		notify.Error("Environment variable with name %s was not found", variable)
		os.Exit(1)
	} else {
		notify.Success("Variable %s was found in the environment with value:\n%s", variable, result)
		os.Exit(0)
	}
}
