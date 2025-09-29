package random

import (
	"fmt"
	"strconv"

	"github.com/cjlapao/locally-cli/internal/environment/interfaces"
	"github.com/cjlapao/locally-cli/internal/notifications"
	"github.com/cjlapao/locally-cli/pkg/diagnostics"

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

func (worker RandomValueFunction) New() interfaces.EnvironmentVariableFunction {
	return RandomValueFunction{
		name: "random.func",
	}
}

func (worker RandomValueFunction) Name() string {
	return worker.name
}

func (worker RandomValueFunction) Exec(value string, args ...string) (string, *diagnostics.Diagnostics) {
	diag := diagnostics.New("random.func")
	defer diag.Complete()

	if value == "" {
		diag.AddError("INVALID_ARGUMENT", "Invalid argument", "random.func", map[string]interface{}{
			"argument": value,
		})
		return value, diag
	}

	if len(args) <= 1 {
		return value, diag
	}

	if args[0] == "random" {
		notify.Debug("Executing Random Function")
		length, err := strconv.Atoi(args[1])
		if err != nil {
			diag.AddError("INVALID_ARGUMENT", "Invalid argument", "random.func", map[string]interface{}{
				"argument": args[1],
			})
			return value, diag
		}
		r := cryptorand.GetRandomString(length)
		return fmt.Sprintf("%s%s", value, r), diag
	}

	return value, diag
}
