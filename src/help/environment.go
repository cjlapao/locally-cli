package help

func ShowHelpForEnvironment() {
	logger.Info("Usage: locally env [variable_name] [OPTIONS]")
	logger.Info("")
	logger.Info("locally")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("\t --help \t shows command specific help")
	logger.Info("\t --list-all \t Shows all environment variables in all key vaults")
	logger.Info("")
}
