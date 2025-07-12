package config

import (
	"os"
)

type EnvProvider struct{}

func NewEnvProvider() *EnvProvider {
	return &EnvProvider{}
}

// Load implements the Provider interface
func (p *EnvProvider) Load(cfg *Config) error {
	for _, item := range cfg.Items {
		if item.EnvName != "" {
			envValue := os.Getenv(item.EnvName)
			if envValue != "" {
				cfg.Set(item.Key, envValue)
			}
		}
	}

	return nil
}
