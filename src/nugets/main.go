package nugets

import (
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/notifications"
)

var config = configuration.Get()
var notify = notifications.Get()

const (
	ServiceName = "Nugets"
)
