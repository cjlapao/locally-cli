package nugets

import (
	"fmt"
	"time"

	"github.com/cjlapao/locally-cli/executer"
	"github.com/cjlapao/locally-cli/helpers"
	"github.com/cjlapao/locally-cli/icons"

	"github.com/cjlapao/common-go/helper"
)

type NugetCommandWrapper struct {
	ToolPath string
	Output   string
}

func GetWrapper() *NugetCommandWrapper {
	return &NugetCommandWrapper{}
}

func (svc *NugetCommandWrapper) Pack(projectPath string, outputPath string, version string) error {

	if !helper.FileExists(projectPath) {
		return fmt.Errorf("project %s does not exists", projectPath)
	}

	if !helper.FileExists(outputPath) {
		return fmt.Errorf("output %s does not exists", outputPath)
	}

	version = svc.FormatVersion(version)

	args := make([]string, 0)
	args = append(args, "pack", projectPath, "--no-build", "--output", outputPath, fmt.Sprintf("/p:PackageVersion=%s", version))
	if helper.GetFlagSwitch("build", false) {
		notify.Rocket("Running Dotnet build for %s", icons.IconRocket, projectPath)
		_, err := executer.ExecuteAndWatch("dotnet", "build", projectPath)
		if err != nil {
			notify.FromError(err, "Something wrong running dotnet pack")
			return err
		}
	}

	notify.Rocket("Running Dotnet pack for %s", projectPath)
	output, err := executer.ExecuteAndWatch("dotnet", args...)

	if err != nil {
		notify.FromError(err, "Something wrong running dotnet pack")
		return err
	}

	svc.Output = output.StdOut

	return nil
}

func (svc *NugetCommandWrapper) Add(outputPath, packageName, version string) error {

	version = svc.FormatVersion(version)
	nugetPackage := fmt.Sprintf("%s.%s.nupkg", packageName, version)
	packagePath := helper.JoinPath(outputPath, nugetPackage)

	if !helper.FileExists(packagePath) {
		err := fmt.Errorf("package %s does not exists", packagePath)
		notify.FromError(err, "Error running nuget add")
		return err
	}

	if !helper.FileExists(outputPath) {
		err := fmt.Errorf("output %s does not exists", outputPath)
		notify.FromError(err, "Error running nuget add")
		return err
	}

	notify.Rocket("%s Running Nuget Add for package %s", packagePath)
	output, err := executer.ExecuteAndWatch(helpers.GetNugetPath(), "add", packagePath, "-Source", outputPath)

	if err != nil {
		notify.FromError(err, "Something wrong running nuget add")
		return err
	}

	svc.Output = output.GetAllOutput()

	return nil
}

func (svc *NugetCommandWrapper) GetPackageFilePath(outputPath, packageName, version string) string {
	version = svc.FormatVersion(version)
	nugetPackage := fmt.Sprintf("%s.%s.nupkg", packageName, version)
	packagePath := helper.JoinPath(outputPath, nugetPackage)
	return packagePath
}

func (svc *NugetCommandWrapper) Delete(outputPath, packageName, version string) error {

	version = svc.FormatVersion(version)
	nugetPackage := fmt.Sprintf("%s.%s.nupkg", packageName, version)
	packagePath := helper.JoinPath(outputPath, nugetPackage)

	if !helper.FileExists(packagePath) {
		err := fmt.Errorf("package %s does not exists", packagePath)
		notify.FromError(err, "Error running nuget add")
		return err
	}

	if !helper.FileExists(outputPath) {
		err := fmt.Errorf("output %s does not exists", outputPath)
		notify.FromError(err, "Error running nuget add")
		return err
	}

	notify.Rocket("%s Running Nuget Delete for package %s", packagePath)
	output, err := executer.ExecuteAndWatch(helpers.GetNugetPath(), "delete", packageName, version, "-Source", outputPath, "-NonInteractive")

	if err != nil {
		notify.FromError(err, "Something wrong running nuget delete")
		return err
	}

	svc.Output = output.GetAllOutput()

	return nil
}

func (svc *NugetCommandWrapper) FormatVersion(version string) string {
	if version == "" {
		version = "0.0"
	}

	version = fmt.Sprintf("%s.%s01-alpha", version, time.Now().Format("060102"))

	return version
}
