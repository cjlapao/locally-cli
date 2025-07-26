package help

import (
	"github.com/cjlapao/common-go/log"
)

var logger = log.Get()

func ShowHelpForNoCommand() {
	logger.Info("Usage: locally [COMMAND]")
	logger.Info("")
	logger.Info("locally")
	logger.Info("")
	logger.Info("Options:")
	logger.Info("\t --help \t shows command specific help")
	logger.Info("")
	logger.Info("Commands:")
	logger.Info("  config          \t\t Sets the configuration context and switch between them")
	logger.Info("  certificates    \t\t Generates self signed valid chain certificates for local development")
	logger.Info("  docker          \t\t Controls services docker operations from generation to life cycle")
	logger.Info("  env             \t\t allows to query configuration variables")
	logger.Info("  keyvault        \t\t Allows manual synchronization of an azure keyvault into configuration")
	logger.Info("  hosts           \t\t Controls system host file changes to help generate custom entries")
	logger.Info("  infrastructure  \t\t Builds the required infrastructure for the services based on the stacks")
	logger.Info("  pipelines       \t\t Running integrated pipelines for easy manage of services")
	logger.Info("  proxy           \t\t Controls caddy proxy service allowing to generate/update configuration")
	logger.Info("  nuget           \t\t builds nuget packages and adds them to a local feed")
	logger.Info("  tools           \t\t Some useful developers tools")
	logger.Info("")
}
