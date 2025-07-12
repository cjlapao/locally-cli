package caddy

import (
	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/notifications"
)

var (
	config *configuration.ConfigService
	notify = notifications.Get()
)

const (
	ServiceName = "Caddy"
)
