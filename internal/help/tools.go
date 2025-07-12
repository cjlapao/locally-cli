package help

func ShowHelpForToolsCommand() {
	logger.Info("Usage: locally tools [TOOL_NAME] [COMMAND]")
	logger.Info("")
	logger.Info("locally Tools")
	logger.Info("")
	logger.Info("Tools:")
	logger.Info("  ems     \t Useful tools to use in EMS")
	logger.Info("  base64  \t Useful tools to use in EMS")
}

func ShowHelpForEmsTool() {
	logger.Info("Usage: locally tools ems [COMMAND]")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  apikey  \t Generates a compatible authorization header for ems based on an api key")
}

func ShowHelpForEmsApiKeyTool() {
	logger.Info("Usage: locally tools ems apikey [FLAG]")
	logger.Info("")
	logger.Info("Flags:")
	logger.Info("  --key=<apikey>  \t Api Key to use when generating the header")
}

func ShowHelpForBase64Tool() {
	logger.Info("Usage: locally tools base64 [COMMAND] [VALUE]")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  encode  \t encodes string to base64")
	logger.Info("  decode  \t decodes string to base64")
}

func ShowHelpForAzureAcrGetToken() {
	logger.Info("Usage: locally tools azure_acr [COMMAND] [ACR_NAME]")
	logger.Info("")
	logger.Info("Flags:")
	logger.Info("  --scope=<scope>    \t The scope of the token, if empty [repository:*:metadata_read] will be used")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  get-token  \t gets a new token for the specified Azure Container Registry")
}
