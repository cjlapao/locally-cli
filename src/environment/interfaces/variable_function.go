package interfaces

type VariableFunction interface {
	Name() string
	New() VariableFunction
	Exec(value string, args ...string) string
}
