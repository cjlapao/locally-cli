package hosts

import "github.com/cjlapao/locally-cli/internal/notifications"

var notify = notifications.Get()

const (
	ServiceName = "Hosts"
)

const (
	START_SECTION   string = "# locally Section"
	END_SECTION     string = "# End locally Section"
	locally_COMMENT string = "added by locally"
)
