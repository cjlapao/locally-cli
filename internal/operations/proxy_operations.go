package operations

// import (
// 	"os"

// 	"github.com/cjlapao/locally-cli/internal/caddy"
// 	"github.com/cjlapao/locally-cli/internal/help"
// 	"github.com/cjlapao/locally-cli/internal/system"

// 	"github.com/cjlapao/common-go/helper"
// )

// func ProxyOperations(subCommand string) {
// 	caddySvc := caddy.Get()
// 	systemSvc := system.Get()
// 	caddySvc.CheckForCaddy(false)
// 	if subCommand == "" && helper.GetFlagSwitch("help", false) {
// 		help.ShowHelpForProxyCommand()
// 		os.Exit(0)
// 	}
// 	switch subCommand {
// 	case "generate":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyGenerateCommand()
// 			os.Exit(0)
// 		}
// 		systemSvc.CheckFolders(true)
// 		caddySvc.GenerateDockerFiles()
// 		if err := caddySvc.GenerateCaddyFiles(); err != nil {
// 			notify.FromError(err, "There was an error generating caddy files")
// 		}
// 	case "build-container":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyBuildCommand()
// 			os.Exit(0)
// 		}
// 		systemSvc.CheckFolders(false)
// 		caddySvc.GenerateDockerFiles()
// 		if err := caddySvc.BuildContainer(); err != nil {
// 			notify.FromError(err, "There was an error building locally container")
// 		}
// 	case "rebuild-container":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyRebuildCommand()
// 			os.Exit(0)
// 		}
// 		systemSvc.CheckFolders(true)
// 		caddySvc.GenerateDockerFiles()
// 		if err := caddySvc.RebuildContainer(); err != nil {
// 			notify.FromError(err, "There was an error rebuilding locally container")
// 		}
// 	case "run":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyRunCommand()
// 			os.Exit(0)
// 		}
// 		if helper.GetFlagSwitch("generate", false) {
// 			systemSvc.CheckFolders(true)
// 			caddySvc.GenerateDockerFiles()
// 			if err := caddySvc.GenerateCaddyFiles(); err != nil {
// 				notify.FromError(err, "There was an error generating caddy files")
// 			}
// 		}
// 		caddySvc := caddy.GetWrapper()
// 		if err := caddySvc.Run(); err != nil {
// 			notify.FromError(err, "There was an error running proxy")
// 		}
// 	case "up":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyUpCommand()
// 			os.Exit(0)
// 		}
// 		if helper.GetFlagSwitch("generate", false) {
// 			systemSvc.CheckFolders(true)
// 			caddySvc.GenerateDockerFiles()
// 			if err := caddySvc.GenerateCaddyFiles(); err != nil {
// 				notify.FromError(err, "There was an error generating caddy files")
// 			}
// 		}
// 		if helper.GetFlagSwitch("build", false) {
// 			if err := caddySvc.BuildContainer(); err != nil {
// 				notify.FromError(err, "There was an error building proxy container")
// 			}
// 		}
// 		if err := caddySvc.ContainerUp(); err != nil {
// 			notify.FromError(err, "There was an error bringing up proxy container")
// 		}
// 	case "down":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyDownCommand()
// 			os.Exit(0)
// 		}
// 		if err := caddySvc.ContainerDown(); err != nil {
// 			notify.FromError(err, "There was an error bringing down proxy container")
// 		}
// 	case "start":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyStartCommand()
// 			os.Exit(0)
// 		}
// 		if err := caddySvc.StartContainer(); err != nil {
// 			notify.FromError(err, "There was an error starting up proxy container")
// 		}
// 	case "stop":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyStopCommand()
// 			os.Exit(0)
// 		}
// 		if err := caddySvc.StoplocallyContainer(); err != nil {
// 			notify.FromError(err, "There was an error stopping proxy container")
// 		}
// 	case "pause":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyPauseCommand()
// 			os.Exit(0)
// 		}
// 		if err := caddySvc.PauseContainer(); err != nil {
// 			notify.FromError(err, "There was an error pausing proxy container")
// 		}
// 	case "resume":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyResumeCommand()
// 			os.Exit(0)
// 		}
// 		if err := caddySvc.ResumeContainer(); err != nil {
// 			notify.FromError(err, "There was an error resuming proxy container")
// 		}
// 	case "status":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyStatusCommand()
// 			os.Exit(0)
// 		}
// 		if err := caddySvc.ContainerStatus(); err != nil {
// 			notify.FromError(err, "There was an error getting the status for Caddy Proxy container")
// 		}
// 	case "logs":
// 		if helper.GetFlagSwitch("help", false) {
// 			help.ShowHelpForProxyBuildCommand()
// 			os.Exit(0)
// 		}
// 		if err := caddySvc.ContainerLogs(); err != nil {
// 			notify.FromError(err, "There was an error getting caddy container logs")
// 		}
// 	default:
// 		help.ShowHelpForProxyCommand()
// 		os.Exit(0)
// 	}
// }
