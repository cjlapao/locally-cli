package nugets

// import (
// 	"os"
// 	"strconv"
// 	"strings"

// 	"github.com/cjlapao/locally-cli/internal/configuration"
// 	"github.com/cjlapao/locally-cli/internal/executer"
// 	"github.com/cjlapao/locally-cli/internal/helpers"
// 	"github.com/cjlapao/locally-cli/internal/icons"
// 	"github.com/cjlapao/locally-cli/internal/notifications"

// 	"github.com/cjlapao/common-go/helper"
// )

// var globalNugetService *NugetService

// type NugetService struct {
// 	notify  *notifications.NotificationsService
// 	wrapper *NugetCommandWrapper
// }

// func New() *NugetService {
// 	svc := NugetService{
// 		wrapper: GetWrapper(),
// 		notify:  notifications.New(ServiceName),
// 	}

// 	return &svc
// }

// func Get() *NugetService {
// 	if globalNugetService != nil {
// 		return globalNugetService
// 	}

// 	return New()
// }

// func (svc *NugetService) CheckForDotnet(softFail bool) {
// 	config = configuration.Get()
// 	if !config.GlobalConfiguration.Tools.Checked.DotnetChecked {
// 		notify.InfoWithIcon(icons.IconFlag, "Checking for dotnet tool in the system")
// 		if output, err := executer.ExecuteWithNoOutput(helpers.GetDotnetPath(), "--version"); err != nil {
// 			if !softFail {
// 				notify.Error("Dotnet tool not found in system, this is required for the selected function")
// 				os.Exit(1)
// 			} else {
// 				notify.Warning("Dotnet tool not found in system, this might generate an error in the future")
// 			}
// 		} else {
// 			notify.Success("Dotnet tool found with version %s", output)
// 		}
// 		config.GlobalConfiguration.Tools.Checked.DotnetChecked = true
// 	}
// }

// func (svc *NugetService) CheckForNuget(softFail bool) {
// 	config = configuration.Get()
// 	if !config.GlobalConfiguration.Tools.Checked.NugetChecked {
// 		notify.InfoWithIcon(icons.IconFlag, "Checking for nuget tool in the system")
// 		if output, err := executer.ExecuteWithNoOutput(helpers.GetNugetPath(), "help"); err != nil {
// 			if !softFail {
// 				notify.Error("Nuget tool not found in system, this is required for the selected function")
// 				os.Exit(1)
// 			} else {
// 				notify.Warning("Nuget tool not found in system, this might generate an error in the future")
// 			}
// 		} else {
// 			version := strings.TrimSpace(strings.Trim(strings.ReplaceAll(strings.Split(output.StdOut, "\n")[0], "NuGet Version: ", ""), "\r"))
// 			notify.Success("Nuget tool found with version %s", version)
// 		}
// 		config.GlobalConfiguration.Tools.Checked.NugetChecked = true
// 	}
// }

// func (svc *NugetService) GenerateNugetPackages(name string, tags ...string) {
// 	notify.Wrench("Generating Nuget Packages")
// 	config := configuration.Get()
// 	context := config.GetCurrentContext()
// 	if context.NugetPackages == nil {
// 		notify.Warning("No nuget packages found to generate")
// 		return
// 	}
// 	nugetPackages := context.NugetPackages.Packages
// 	if len(tags) > 0 || name == "" {
// 		notify.Wrench("Generating %s packages", strconv.Itoa(len(nugetPackages)))
// 		if len(tags) > 0 {
// 			nugetPackages = context.GetNugetPackagesByTags()
// 		}
// 		for _, pkg := range nugetPackages {
// 			notify.Wrench("Generating nuget package for %s", pkg.Name)
// 			svc := GetWrapper()
// 			packagePath := svc.GetPackageFilePath(context.NugetPackages.OutputSource, pkg.Name, pkg.MajorVersion)
// 			pkgFolder := strings.ToLower(helper.JoinPath(context.NugetPackages.OutputSource, pkg.Name))
// 			if helper.DirectoryExists(helper.JoinPath(context.NugetPackages.OutputSource, pkg.Name)) {
// 				versionFolder := helper.JoinPath(pkgFolder, svc.FormatVersion(pkg.MajorVersion))
// 				if helper.DirectoryExists(versionFolder) {
// 					notify.InfoWithIcon(icons.IconToilet, "Deleting local version %s", versionFolder)
// 					helper.DeleteAllFiles(versionFolder)
// 				}
// 			}

// 			if helper.FileExists(packagePath) {
// 				svc.Delete(context.NugetPackages.OutputSource, pkg.Name, pkg.MajorVersion)
// 			}

// 			if err := svc.Pack(pkg.ProjectFile, context.NugetPackages.OutputSource, pkg.MajorVersion); err == nil {
// 				svc.Add(context.NugetPackages.OutputSource, pkg.Name, pkg.MajorVersion)
// 			}
// 		}
// 	} else {
// 		found := false
// 		for _, pkg := range nugetPackages {
// 			if strings.EqualFold(pkg.Name, name) {
// 				notify.Wrench("Generating nuget package for %s", pkg.Name)
// 				svc := GetWrapper()
// 				packagePath := svc.GetPackageFilePath(context.NugetPackages.OutputSource, pkg.Name, pkg.MajorVersion)

// 				pkgFolder := strings.ToLower(helper.JoinPath(context.NugetPackages.OutputSource, pkg.Name))
// 				if helper.DirectoryExists(helper.JoinPath(context.NugetPackages.OutputSource, pkg.Name)) {
// 					versionFolder := helper.JoinPath(pkgFolder, svc.FormatVersion(pkg.MajorVersion))
// 					if helper.DirectoryExists(versionFolder) {
// 						notify.InfoWithIcon(icons.IconToilet, "Deleting local version %s", versionFolder)
// 						helper.DeleteAllFiles(versionFolder)
// 					}
// 				}

// 				if helper.FileExists(packagePath) {
// 					svc.Delete(context.NugetPackages.OutputSource, pkg.Name, pkg.MajorVersion)
// 				}

// 				if err := svc.Pack(pkg.ProjectFile, context.NugetPackages.OutputSource, pkg.MajorVersion); err == nil {
// 					svc.Add(context.NugetPackages.OutputSource, pkg.Name, pkg.MajorVersion)
// 				}
// 				found = true
// 				break
// 			}
// 		}
// 		if !found {
// 			notify.Error("Nuget Package with name %s was not found", name)
// 		}
// 	}
// }
