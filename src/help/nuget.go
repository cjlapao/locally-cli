package help

func ShowHelpForNugetCommand() {
	logger.Info("Usage: locally docker [COMMAND] [PACKAGE] [OPTIONS]")
	logger.Info("")
	logger.Info("locally Nuget Service")
	logger.Info("")
	logger.Info("Package:")
	logger.Info("\t Name of the package as in the config.yaml, if not present then the command \n\t will be applied to all packages in the service")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("\t --help \t shows command specific help")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  generate     \t\t generates the packages")
}

func ShowHelpForNugetGenerateCommand() {
	logger.Info("Usage: locally nuget generate [package] [OPTIONS]")
	logger.Info("")
	logger.Info("generates nuget package and adds it to the local feed")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --tag=<tag_name>     \t\t select similar services or service components based on their tags")
	logger.Info("  --build              \t\t builds the package before packing it")
	logger.Info("")
}
