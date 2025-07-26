package interfaces

import "github.com/cjlapao/locally-cli/pkg/diagnostics"

type EnvironmentVariableFunction interface {
	Name() string
	New() EnvironmentVariableFunction
	Exec(value string, args ...string) (string, *diagnostics.Diagnostics)
}
