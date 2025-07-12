package interfaces

type EnvironmentVault interface {
	Name() string
	Sync() (map[string]interface{}, error)
}
