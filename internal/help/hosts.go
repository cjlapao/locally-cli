package help

func ShowHelpForHostsCommand() {
	logger.Info("Usage: locally hosts [COMMAND]")
	logger.Info("")
	logger.Info("locally Hosts Service")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  update  \t\t updates the hosts file entries")
	logger.Info("  clean   \t\t removes the hosts file entries created by locally")
}

func ShowHelpForHostsUpdateCommand() {
	logger.Info("Usage: locally hosts update")
	logger.Info("")
	logger.Info("updates the hosts file entries with the required of of the config.yaml")
	logger.Info("")
}

func ShowHelpForHostsCleanCommand() {
	logger.Info("Usage: locally hosts clean")
	logger.Info("")
	logger.Info("cleans the hosts file entries created by the update command")
	logger.Info("")
}
