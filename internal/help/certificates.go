package help

func ShowHelpForCertificatesCommand() {
	logger.Info("Usage: locally certificates [command]")
	logger.Info("")
	logger.Info("locally Certificates")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("\t --help \t shows command specific help")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  generate   \t\t Generates the certificates based on the config file")
	logger.Info("  clean      \t\t Cleans generated certificates from the configuration file")
	logger.Info("")
}
