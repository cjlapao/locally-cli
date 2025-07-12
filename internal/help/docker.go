package help

func ShowHelpForDockerCommand() {
	logger.Info("Usage: locally docker [COMMAND] [SERVICE] [COMPONENT] [OPTIONS]")
	logger.Info("")
	logger.Info("locally Docker Service")
	logger.Info("")
	logger.Info("Service:")
	logger.Info("\t Name of the service as in the config.yaml, this is mandatory for almost all commands")
	logger.Info("")
	logger.Info("Component:")
	logger.Info("\t Name of the service component as in the config.yaml, if not present then the command \n\t will be applied to all components in the service")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("\t --help \t Shows command specific help")
	logger.Info("\t --tag=<tag_name> \t Select similar stacks based on their tags")
	logger.Info("\t --get-latest \t Updates the docker compose with the latest tag")
	logger.Info("\t --clean-repo \t Deletes the current repository and clones it again")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  build               \t\t builds service or service component container")
	logger.Info("  rebuild             \t\t rebuilds service or service component container")
	logger.Info("  delete              \t\t deletes images from a service or service component")
	logger.Info("  up                  \t\t creates and starts service container")
	logger.Info("  down                \t\t stops and removes service container")
	logger.Info("  start               \t\t starts service or service component container")
	logger.Info("  stop                \t\t stops service or service component container")
	logger.Info("  pause               \t\t pauses service or service component container")
	logger.Info("  resume              \t\t resumes service or service component container")
	logger.Info("  status              \t\t shows the status of the service container")
	logger.Info("  list                \t\t gets all the components running in the service")
	logger.Info("  logs                \t\t gets the logs of the service or service component container")
	logger.Info("  generate            \t\t generates the service or service component container")
}

func ShowHelpForDockerGenerateCommand() {
	logger.Info("Usage: locally docker generate [SERVICE] [OPTIONS]")
	logger.Info("")
	logger.Info("generates the service or service component container")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --all          \t\t generates the docker compose override for all services")
	logger.Info("")
}

func ShowHelpForDockerBuildCommand() {
	logger.Info("Usage: locally docker build [SERVICE] [COMPONENT] [OPTIONS]")
	logger.Info("")
	logger.Info("Builds Service or Service Component Containers")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
	logger.Info("  --no-cache          \t\t builds service container or service component docker container with no use of cache")
	logger.Info("  --force-recreate    \t\t force recreate the service docker containers")
	logger.Info("  --force-clean       \t\t cleans the existing service docker container before building")
	logger.Info("")
}

func ShowHelpForDockerRebuildCommand() {
	logger.Info("Usage: locally docker rebuild [SERVICE] [COMPONENT] [OPTIONS]")
	logger.Info("")
	logger.Info("Rebuilds Service or Service Component Containers")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
	logger.Info("  --no-cache          \t\t builds service container or service component docker container with no use of cache")
	logger.Info("  --force-recreate    \t\t force recreate the service container or service component docker containers")
	logger.Info("  --force-clean       \t\t cleans the existing service container or service component docker container")
	logger.Info("")
}

func ShowHelpForDockerUpCommand() {
	logger.Info("Usage: locally docker up [SERVICE] [COMPONENT] [OPTIONS]")
	logger.Info("")
	logger.Info("Creates and starts service docker container")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --all          \t\t bring all services up")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
}

func ShowHelpForDockerDownCommand() {
	logger.Info("Usage: locally docker down [SERVICE] [COMPONENT]")
	logger.Info("")
	logger.Info("Stops and deletes Caddy Proxy container from docker")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
}

func ShowHelpForDockerStartCommand() {
	logger.Info("Usage: locally docker start [SERVICE] [COMPONENT]")
	logger.Info("")
	logger.Info("Starts service or service component docker container")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
}

func ShowHelpForDockerStopCommand() {
	logger.Info("Usage: locally docker stop [SERVICE] [COMPONENT]")
	logger.Info("")
	logger.Info("Stops service or service component docker container")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
}

func ShowHelpForDockerPauseCommand() {
	logger.Info("Usage: locally docker pause [SERVICE] [COMPONENT]")
	logger.Info("")
	logger.Info("Pauses service or service component docker container")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
}

func ShowHelpForDockerResumeCommand() {
	logger.Info("Usage: locally docker resume [SERVICE] [COMPONENT]")
	logger.Info("")
	logger.Info("Resumes service or service component docker container")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
}

func ShowHelpForDockerStatusCommand() {
	logger.Info("Usage: locally docker status [SERVICE] [COMPONENT]")
	logger.Info("")
	logger.Info("Shows the status of service or service component docker container")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
}

func ShowHelpForDockerLogCommand() {
	logger.Info("Usage: locally docker logs [SERVICE] [COMPONENT] [OPTIONS]")
	logger.Info("")
	logger.Info("Shows the container(s) logs for service or service component docker container")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --follow          \t\t tails the logs for the container")
	logger.Info("")
}

func ShowHelpForDockerListCommand() {
	logger.Info("Usage: locally docker list [SERVICE]")
	logger.Info("")
	logger.Info("Gets all the components running in the service")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
}

func ShowHelpForDockerDeleteCommand() {
	logger.Info("Usage: locally docker delete [SERVICE] [COMPONENT]")
	logger.Info("")
	logger.Info("Deletes a service images or a container image")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
}
