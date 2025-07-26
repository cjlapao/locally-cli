package help

//TODO: IMplement correct help
func ShowHelpForInfrastructureCommand() {
	logger.Info("Usage: locally infrastructure [COMMAND] [STACK] [OPTIONS]")
	logger.Info("")
	logger.Info("locally Infrastructure Service")
	logger.Info("")
	logger.Info("Stack:")
	logger.Info("\t Name of the stack as in the config.yaml, if not present then the command \n\t will be applied to all stacks in the infrastructure")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("\t --help \t shows command specific help")
	logger.Info("\t --tag=<tag_name> \t select similar stacks based on their tags")
	logger.Info("\t --reconfigure \t reconfigures the backend")
	logger.Info("\t --clean \t Cleans all of the terraform init configuration")
	logger.Info("\t --clean-repo \t Deletes the current repository and clones it again")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  init-backend \t\t Initializes the terraform folder for the stack")
	logger.Info("  init         \t\t Initializes the terraform folder for the stack")
	logger.Info("  validate     \t\t Validates the terraform folder for the stack")
	logger.Info("  plan         \t\t Plans the terraform folder for the stack")
	logger.Info("  apply        \t\t Applies the changes in the plan for the stack")
	logger.Info("  output       \t\t Outputs the values of the stack to the global variables")
	logger.Info("  down         \t\t Destroys a stack/stacks")
	logger.Info("  up           \t\t Chain command to bring a stack infrastructure up")
	logger.Info("  refresh      \t\t Refresh the configuration files with the changes from a stack")
	logger.Info("  graph        \t\t Builds a dependency graph for the stack")
}

func ShowHelpForInfrastructureUpCommand() {
	logger.Info("Usage: locally locally infrastructure up [STACK] [OPTIONS]")
	logger.Info("")
	logger.Info("Brings a infrastructure stack up with all its dependencies")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --help \t shows command specific help")
	logger.Info("  --tag=<tag_name>      \t\t select similar stacks based on their tags")
	logger.Info("  --build-dependencies  \t\t builds the selected stack and all of its dependency chart")
	logger.Info("")
}

func ShowHelpForInfrastructureInitCommand() {
	logger.Info("Usage: locally locally infrastructure up [STACK] [OPTIONS]")
	logger.Info("")
	logger.Info("Brings a infrastructure stack up with all its dependencies")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("  --help \t shows command specific help")
	logger.Info("  --tag=<tag_name>      \t\t select similar stacks based on their tags")
	logger.Info("  --build-dependencies  \t\t builds the selected stack and all of its dependency chart")
	logger.Info("  --reconfigure         \t\t restarts the configuration of the stack from fresh")
	logger.Info("  --clean               \t\t cleans the terraform folders before running the init")
	logger.Info("")
}
