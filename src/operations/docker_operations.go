package operations

import (
	"fmt"
	"os"

	"github.com/cjlapao/locally-cli/common"
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/context/docker_component"
	"github.com/cjlapao/locally-cli/docker"
	"github.com/cjlapao/locally-cli/help"

	"github.com/cjlapao/common-go/helper"
)

func DockerOperations(subCommand string, options *docker.DockerServiceOptions) {
	dockerService := docker.Get()
	config := configuration.Get()
	context := config.GetCurrentContext()

	dockerService.CheckForDocker(false)
	dockerService.CheckForDockerCompose(false)

	if subCommand == "" && helper.GetFlagSwitch("help", false) {
		help.ShowHelpForDockerCommand()
		os.Exit(0)
	}

	switch subCommand {
	case "build":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerBuildCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.BuildServiceContainer(options); err != nil {
			notify.FromError(err, "There was an error building %v service", service)
		}
	case "rebuild":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerRebuildCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.RebuildServiceContainer(options); err != nil {
			notify.FromError(err, "There was an error rebuilding %v service", service)
		}
	case "up":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerUpCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.ServiceContainerUp(options); err != nil {
			notify.FromError(err, "There was an error bringing up %v service", service)
		}
	case "down":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerDownCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.ServiceContainerDown(options); err != nil {
			notify.FromError(err, "There was an error bringing down %v service", service)
		}
	case "start":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerStartCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.StartServiceContainer(options); err != nil {
			notify.FromError(err, "There was an error starting up %v service", service)
		}
	case "stop":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerStopCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.StopServiceContainer(options); err != nil {
			notify.FromError(err, "There was an error stopping %v service", service)
		}
	case "pause":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerPauseCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.PauseServiceContainer(options); err != nil {
			notify.FromError(err, "There was an error pausing %v service", service)
		}
	case "resume":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerResumeCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.ResumeServiceContainer(options); err != nil {
			notify.FromError(err, "There was an error resuming %v service", service)
		}
	case "status":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerStatusCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if service == "" {
			if err := dockerService.ListServiceContainer(options); err != nil {
				notify.FromError(err, "There was an error getting the list of services", service)
			}
		} else {
			if err := dockerService.ServiceContainerStatus(options); err != nil {
				notify.FromError(err, "There was an error getting the status of %v service", service)
			}
		}
	case "list":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerListCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name: service,
			}
		}

		if err := dockerService.ListServiceContainer(options); err != nil {
			notify.FromError(err, "There was an error getting the list services", service)
		}
	case "logs":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerLogCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.ServiceContainerLogs(options); err != nil {
			notify.FromError(err, "There was an error getting the logs for %v service", service)
		}
	case "generate":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerGenerateCommand()
			os.Exit(0)
		}

		service := common.VerifyCommand(helper.GetArgumentAt(2))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.GenerateServiceDockerComposeOverrideFile(options); err != nil {
			notify.FromError(err, "There was an error generating docker-compose override for %v service", service)
		}

	case "delete":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerDeleteCommand()
			os.Exit(0)
		}
		service := common.VerifyCommand(helper.GetArgumentAt(2))
		component := common.VerifyCommand(helper.GetArgumentAt(3))

		buildDependencies := helper.GetFlagSwitch("build-dependencies", false)
		// Forcing the build dependencies if the tag is on as we need to crawl through them anyway
		if context.HasTags() {
			buildDependencies = true
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				Name:              service,
				ComponentName:     component,
				BuildDependencies: buildDependencies,
			}
		}

		if err := dockerService.DeleteImage(options); err != nil {
			notify.FromError(err, "There was an error deleting images for %v service", service)
		}
	case "pull":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerListCommand()
			os.Exit(0)
		}

		if options == nil {
			options = &docker.DockerServiceOptions{
				DockerRegistry: &docker_component.DockerRegistry{
					Enabled:      true,
					Registry:     helper.GetFlagValue("registry", ""),
					ManifestPath: helper.GetFlagValue("manifest-path", ""),
					Tag:          helper.GetFlagValue("tag", ""),
				},
			}
		}

		notify.Debug("Options: %s", fmt.Sprintf("%v", options))
		notify.Debug("Docker Registry: %s", fmt.Sprintf("%v", options.DockerRegistry))
		notify.Debug("Docker Compose: %s", fmt.Sprintf("%v", options.DockerCompose))

		if err := dockerService.PullImage(options); err != nil {
			notify.FromError(err, "There was an error pulling the image %s from %s", options.DockerRegistry.ManifestPath, options.DockerRegistry.Registry)
		}
	case "generate-compose":
		if helper.GetFlagSwitch("help", false) {
			help.ShowHelpForDockerListCommand()
			os.Exit(0)
		}

		if options == nil {
			notify.Error("This command cannot be use by command line")
			return
		}

		if err := dockerService.GenerateServiceDockerComposeFile(options); err != nil {
			notify.FromError(err, "There was an error generating docker compose for the image %s from %s", options.DockerRegistry.ManifestPath, options.DockerRegistry.Registry)
		}
	default:
		help.ShowHelpForDockerCommand()
		os.Exit(0)
	}
}
