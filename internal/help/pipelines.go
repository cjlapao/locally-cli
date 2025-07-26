package help

func ShowHelpForPipelinesCommand() {
	logger.Info("Usage: locally pipelines [COMMAND] [PIPELINE]")
	logger.Info("")
	logger.Info("Pipeline:")
	logger.Info("\t Name of the of the pipeline you will want to run")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("\t --help \t shows command specific help")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  run               \t\t runs the selected pipeline")
	logger.Info("  validate          \t\t validate the selected pipeline")
	logger.Info("  list              \t\t list all pipelines")
}

func ShowHelpForPipelineRunCommand() {
	logger.Info("Usage: locally pipeline run [PIPELINE]")
	logger.Info("")
	logger.Info("Runs the selected pipeline")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("")
}

func ShowHelpForPipelineValidateCommand() {
	logger.Info("Usage: locally pipeline validate [PIPELINE]")
	logger.Info("")
	logger.Info("Validates the selected pipeline")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("")
}

func ShowHelpForPipelineListCommand() {
	logger.Info("Usage: locally pipeline list")
	logger.Info("")
	logger.Info("Lists all pipelines in the config")
	logger.Info("")
}
