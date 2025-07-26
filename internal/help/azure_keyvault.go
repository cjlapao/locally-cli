package help

func ShowHelpForAzureKeyvaultCommand() {
	logger.Info("Usage: locally keyvault command")
	logger.Info("")
	logger.Info("locally keyvault")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("\t --help \t shows command specific help")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  sync        \t\t synchronizes a keyvault secrets into the configuration file")
	logger.Info("")
}

func ShowHelpForAzureKeyvaultSyncKeyvaultCommand() {
	logger.Info("Usage: locally keyvault sync [KEYVAULT_URL]")
	logger.Info("")
	logger.Info("Syncs the selected azure keyvault into the configuration file")
	logger.Info("")
}
