package interfaces

import "github.com/cjlapao/locally-cli/pkg/diagnostics"

type EnvironmentVault interface {
	Name() string
	Sync() (map[string]interface{}, *diagnostics.Diagnostics)
	Get(key string) (interface{}, bool)
	Set(key string, value interface{}) *diagnostics.Diagnostics
	Remove(key string) *diagnostics.Diagnostics
}
