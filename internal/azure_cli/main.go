package azure_cli

import (
	"github.com/cjlapao/locally-cli/internal/notifications"
)

// TODO: Add the ability to create storage accounts and containers
var notify = notifications.Get()

const (
	ServiceName = "AzureCli"
)
