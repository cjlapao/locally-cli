package help

func ShowHelpForProxyCommand() {
	logger.Info("Usage: locally proxy command [OPTIONS]")
	logger.Info("")
	logger.Info("locally Caddy Proxy")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("\t --help \t shows command specific help")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  generate            \t\t builds all the necessary files for the caddy proxy")
	logger.Info("  build-container     \t\t builds caddy container")
	logger.Info("  rebuild-container   \t\t rebuilds caddy container")
	logger.Info("  run                 \t\t runs caddy proxy from command line")
	logger.Info("  up                  \t\t creates the caddy proxy container service")
	logger.Info("  down                \t\t removes the caddy proxy container service")
	logger.Info("  start               \t\t starts the caddy proxy container service")
	logger.Info("  stop                \t\t stops the caddy proxy container service")
	logger.Info("  pause               \t\t pauses the caddy proxy container service")
	logger.Info("  resume              \t\t resumes the caddy proxy container service")
	logger.Info("  status              \t\t shows the status of the caddy proxy container service")
	logger.Info("  logs                \t\t gets the logs of the caddy proxy container service")
	logger.Info("")
}

func ShowHelpForProxyGenerateCommand() {
	logger.Info("Usage: locally proxy generate")
	logger.Info("")
	logger.Info("Generates all the files needed for caddy proxy ")
	logger.Info("")
}

func ShowHelpForProxyBuildCommand() {
	logger.Info("Usage: locally proxy build-container [OPTIONS]")
	logger.Info("")
	logger.Info("Builds Caddy Proxy Containers")
	logger.Info("")
	logger.Info("options:")
	logger.Info("  --no-cache          \t\t builds caddy container with no use of cache")
	logger.Info("  --force-recreate    \t\t force recreate the docker containers")
	logger.Info("  --force-clean       \t\t cleans the existing container before building")
	logger.Info("")
}

func ShowHelpForProxyRebuildCommand() {
	logger.Info("Usage: locally proxy rebuild-container [OPTIONS]")
	logger.Info("")
	logger.Info("Rebuilds Caddy Proxy Containers")
	logger.Info("")
	logger.Info("options:")
	logger.Info("  --no-cache          \t\t builds caddy container with no use of cache")
	logger.Info("  --force-recreate    \t\t force recreate the docker containers")
	logger.Info("  --force-clean       \t\t cleans the existing container before building")
	logger.Info("")
}

func ShowHelpForProxyRunCommand() {
	logger.Info("Usage: locally proxy run [OPTIONS]")
	logger.Info("")
	logger.Info("locally Proxy Run")
	logger.Info("")
	logger.Info("options:")
	logger.Info("  --generate          \t\t all the necessary files for the caddy proxy")
	logger.Info("")
}

func ShowHelpForProxyUpCommand() {
	logger.Info("Usage: locally proxy up [OPTIONS]")
	logger.Info("")
	logger.Info("Creates and Starts Caddy proxy")
	logger.Info("")
	logger.Info("options:")
	logger.Info("  --generate          \t\t all the necessary files for the caddy proxy")
	logger.Info("  --build             \t\t builds caddy container")
	logger.Info("  --no-cache          \t\t builds caddy container with no use of cache")
	logger.Info("  --force-recreate    \t\t force recreate the docker containers")
	logger.Info("  --force-clean       \t\t cleans the existing container before building")
	logger.Info("")
}

func ShowHelpForProxyDownCommand() {
	logger.Info("Usage: locally proxy down")
	logger.Info("")
	logger.Info("Stops and deletes Caddy Proxy container from docker")
	logger.Info("")
}

func ShowHelpForProxyStartCommand() {
	logger.Info("Usage: locally proxy start")
	logger.Info("")
	logger.Info("Starts Caddy Proxy container from docker")
	logger.Info("")
}

func ShowHelpForProxyStopCommand() {
	logger.Info("Usage: locally proxy stop")
	logger.Info("")
	logger.Info("Stops Caddy Proxy container from docker")
	logger.Info("")
}

func ShowHelpForProxyPauseCommand() {
	logger.Info("Usage: locally proxy pause")
	logger.Info("")
	logger.Info("Pauses Caddy Proxy container from docker")
	logger.Info("")
}

func ShowHelpForProxyResumeCommand() {
	logger.Info("Usage: locally proxy resume")
	logger.Info("")
	logger.Info("Resumes Caddy Proxy container from docker")
	logger.Info("")
}

func ShowHelpForProxyStatusCommand() {
	logger.Info("Usage: locally proxy status")
	logger.Info("")
	logger.Info("Shows the status Caddy Proxy container from docker")
	logger.Info("")
}

func ShowHelpForProxyLogsCommand() {
	logger.Info("Usage: locally proxy logs [OPTIONS]")
	logger.Info("")
	logger.Info("Shows the container logs for Caddy Proxy Container")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --follow          \t\t tails the logs for the container")
	logger.Info("")
}
