package caddy

import (
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/notifications"
)

var config *configuration.ConfigService
var notify = notifications.Get()

const (
	ServiceName = "Caddy"
)
