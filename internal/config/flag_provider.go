package config

import "flag"

type FlagProvider struct{}

func NewFlagProvider() *FlagProvider {
	return &FlagProvider{}
}

func (p *FlagProvider) Load(cfg *Config) error {
	for _, item := range cfg.Items {
		if item.FlagName != "" {
			// reading the flag value
			flagValue := flag.Lookup(item.FlagName)
			if flagValue != nil {
				item.Value = flagValue.Value.String()
			}
		}
	}

	return nil
}
