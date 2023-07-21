package infrastructure

import (
	"github.com/cjlapao/locally-cli/configuration"
	"github.com/cjlapao/locally-cli/notifications"

	"github.com/cjlapao/common-go/log"
)

const (
	MissingInitWhenApplying string = "err. 404"
)

const (
	TerraformServiceName string = "terraform"
)

var config *configuration.ConfigService
var logger = log.Get()
var notify = notifications.New(TerraformServiceName)
