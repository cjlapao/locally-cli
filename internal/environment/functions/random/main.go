package random

import (
	"fmt"
	"strconv"

	"github.com/cjlapao/locally-cli/internal/environment/interfaces"
	"github.com/cjlapao/locally-cli/internal/notifications"

	cryptorand "github.com/cjlapao/common-go-cryptorand"
)

var notify = notifications.Get()

const (
	ErrorInvalidParameters = "500"
	ErrorInvalidConnection = "501"
)

type RandomValueFunction struct {
	name string
}

func (worker RandomValueFunction) New() interfaces.VariableFunction {
	return RandomValueFunction{
		name: "random.func",
	}
}

func (worker RandomValueFunction) Name() string {
	return worker.name
}

func (worker RandomValueFunction) Exec(value string, args ...string) string {
	if value == "" {
		return value
	}

	if len(args) <= 1 {
		return value
	}

	if args[0] == "random" {
		notify.Debug("Executing Random Function")
		length, err := strconv.Atoi(args[1])
		if err != nil {
			return value
		}
		r := cryptorand.GetRandomString(length)
		return fmt.Sprintf("%s%s", value, r)
	}

	return value
}
