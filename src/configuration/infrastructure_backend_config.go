package configuration

import "time"

type InfrastructureBackendConfig struct {
	LastInitiated *time.Time                        `json:"lastInitiated,omitempty" yaml:"lastInitiated,omitempty"`
	Azure         *InfrastructureAzureBackendConfig `json:"azure,omitempty" yaml:"azure,omitempty"`
}
