package npm

import (
	"fmt"
	"os"
	"strings"

	"github.com/cjlapao/locally-cli/internal/common"
	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/executer"
	"github.com/cjlapao/locally-cli/internal/helpers"
	"github.com/cjlapao/locally-cli/internal/icons"
)

type NpmCommandWrapper struct {
	Output  string
	Version string
}

func GetWrapper() *NpmCommandWrapper {
	config = configuration.Get()
	return &NpmCommandWrapper{}
}

func (svc *NpmCommandWrapper) CheckForNpm(softFail bool) {
	config = configuration.Get()
	if !config.GlobalConfiguration.Tools.Checked.NpmChecked {
		notify.InfoWithIcon(icons.IconFlag, "Checking for npm in the system")
		if output, err := executer.ExecuteWithNoOutput(helpers.GetNpmPath(), "-v"); err != nil {
			if !softFail {
				notify.Error("Npm tool not found in system, this is required for the selected function")
				os.Exit(1)
			} else {
				notify.Warning("Npm tool not found in system, this might generate an error in the future")
			}
		} else {
			svc.Output = output.GetAllOutput()
			svc.Version = strings.Split(output.StdOut, "\n")[0]
			notify.Success("Npm tool found with version %s", output.StdOut)
		}
		config.GlobalConfiguration.Tools.Checked.NpmChecked = true
	}
}

func (svc *NpmCommandWrapper) CI(workingDir string, minVersion string) error {
	svc.versionCheck(minVersion)
	svc.init()
	if output, err := executer.ExecuteWithNoOutput(helpers.GetNpmPath(), "ci", "--prefix", workingDir); err != nil {
		return svc.error(output, err, "There was an error running npm install")
	} else {
		svc.Output = output.GetAllOutput()
		notify.Success("executed CI with: %s", output.StdOut)
	}

	return nil
}

func (svc *NpmCommandWrapper) Install(workingDir string, minVersion string) error {
	svc.versionCheck(minVersion)
	svc.init()
	if output, err := executer.ExecuteWithNoOutput(helpers.GetNpmPath(), "install", "--force", "--prefix", workingDir); err != nil {
		return svc.error(output, err, "There was an error running npm install")
	} else {
		svc.success("Install", output)
	}

	return nil
}

func (svc *NpmCommandWrapper) Publish(workingDir string, minVersion string) error {
	svc.versionCheck(minVersion)
	svc.init()
	if output, err := executer.ExecuteWithNoOutput(helpers.GetNpmPath(), "publish", "--prefix", workingDir); err != nil {
		return svc.error(output, err, "There was an error running npm publish")
	} else {
		svc.success("Publish", output)
	}

	return nil
}

func (svc *NpmCommandWrapper) Custom(customCommand string, workingDir string, minVersion string) error {
	svc.versionCheck(minVersion)
	svc.init()
	commands := strings.Split(customCommand, " ")
	commands = append(commands, "--prefix")
	commands = append(commands, workingDir)
	if output, err := executer.ExecuteWithNoOutput(helpers.GetNpmPath(), commands...); err != nil {
		errorString := fmt.Sprintf("There was an error running npm custom [%s]", customCommand)
		return svc.error(output, err, errorString)
	} else {
		command := fmt.Sprintf("Custom [%s]", customCommand)
		svc.success(command, output)
	}

	return nil
}

func (svc *NpmCommandWrapper) init() error {
	if output, err := executer.ExecuteWithNoOutput(helpers.GetNpmPath(), "init", "-y"); err != nil {
		return svc.error(output, err, "There was an error initialising npm for locally")
	} else {
		svc.success("Init", output)
	}

	return nil
}

func (svc *NpmCommandWrapper) error(output executer.ExecuteOutput, err error, errorString string) error {
	notify.FromError(err, errorString)

	if common.IsDebug() {
		notify.Debug(output.StdErr)
	}

	return err
}

func (svc *NpmCommandWrapper) success(command string, output executer.ExecuteOutput) {
	svc.Output = output.GetAllOutput()

	if common.IsDebug() {
		notify.Success("executed %s with: %s", command, output.StdOut)
	} else {
		notify.Success("executed %s", command)
	}
}

func (svc *NpmCommandWrapper) versionCheck(minVersion string) {
	if minVersion != "" && minVersion > svc.Version {
		notify.Error("current npm version %s, please upgrade to %s", svc.Version, minVersion)
		os.Exit(0)
	}
}
