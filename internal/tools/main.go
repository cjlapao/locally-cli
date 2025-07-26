package tools

import (
	"github.com/cjlapao/locally-cli/internal/notifications"

	"github.com/cjlapao/common-go/log"
)

var (
	logger = log.Get()
	notify = notifications.Get()
)
