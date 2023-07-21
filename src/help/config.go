package help

func ShowHelpForConfigCommand() {
	logger.Info("Usage: locally config command [OPTIONS]")
	logger.Info("")
	logger.Info("locally Config")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("\t --help \t shows command specific help")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  set-context         \t\t sets the current context to work with")
	logger.Info("  list                \t\t lists all the services in the configuration, you can further filter them, use list help to know more")
	logger.Info("  current-context     \t\t shows the current context")
	logger.Info("  clean [--all]       \t\t cleans the current context or all contexts")
	logger.Info("")
}

func ShowHelpForConfigSetContextCommand() {
	logger.Info("Usage: locally config set-context [context name]")
	logger.Info("")
	logger.Info("Sets the current context to work with, this needs to exist in your configuration file")
	logger.Info("or you will get an error message back. context can be use to hold different configurations")
	logger.Info("either to build them or to run docker based on it")
	logger.Info("")
}

func ShowHelpForConfigCleanCommand() {
	logger.Info("Usage: locally config clean [--all]")
	logger.Info("")
	logger.Info("Cleans the current context if --all option is not specified")
	logger.Info("If --all option is specified then all contexts are cleaned up")
	logger.Info("")
}

func ShowHelpForConfigListCommand() {
	logger.Info("Usage: locally config list [COMMAND]")
	logger.Info("")
	logger.Info("Lists all services in the current context or a specific subset if added the current command")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("                    \t\t list all of the services")
	logger.Info("  spa-services      \t\t list all SPA services")
	logger.Info("  backend-services  \t\t list all backend services")
	logger.Info("  tenants           \t\t list all tenants")
	logger.Info("  mock-services     \t\t list all mock services")
	logger.Info("  infrastructure    \t\t list all infrastructure")
	logger.Info("  pipelines         \t\t list all pipelines")
	logger.Info("")
}

func ShowHelpForConfigSyncKeyvaultCommand() {
	logger.Info("Usage: locally config sync-keyvault [KEYVAULT_URL]")
	logger.Info("")
	logger.Info("Syncs the selected azure keyvault into the configuration file")
	logger.Info("")
}
