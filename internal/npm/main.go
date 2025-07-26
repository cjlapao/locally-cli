package npm

import (
	"github.com/cjlapao/locally-cli/internal/configuration"
	"github.com/cjlapao/locally-cli/internal/notifications"
)

var (
	notify = notifications.Get()
	config = configuration.Get()
)
