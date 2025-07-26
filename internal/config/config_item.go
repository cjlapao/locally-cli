// Package config provides a configuration system for the application.
// It allows for setting and getting configuration values from various sources.
// The configuration is stored in a map and can be accessed using the Get() method.
// The configuration can be set using the Set() method.
// The configuration can be checked if it is set using the IsSet() method.
// The configuration can be checked if it is a flag using the IsFlagSet() method.
// The configuration can be checked if it is an environment variable using the IsEnvSet() method.
// The configuration can be checked if it is a boolean using the GetBool() method.
package config

import (
	"strconv"
	"strings"
	"time"
)

type ConfigItem struct {
	Key      string
	Value    string
	FlagName string
	EnvName  string
}

func (c *ConfigItem) Get() string {
	return c.Value
}

func (c *ConfigItem) Set(value string) {
	c.Value = value
}

func (c *ConfigItem) IsSet() bool {
	return c.Value != ""
}

func (c *ConfigItem) IsFlagSet() bool {
	return c.FlagName != ""
}

func (c *ConfigItem) IsEnvSet() bool {
	return c.EnvName != ""
}

func (c *ConfigItem) GetBool() bool {
	if c == nil {
		return false
	}
	return c.Value == "true" || c.Value == "1" || c.Value == "yes" || c.Value == "on"
}

func (c *ConfigItem) GetInt() int {
	if c == nil {
		return 0
	}
	v, err := strconv.Atoi(c.Value)
	if err != nil {
		return 0
	}
	return v
}

func (c *ConfigItem) GetDuration() time.Duration {
	if c == nil {
		return 0
	}
	dur, err := time.ParseDuration(c.Value)
	if err != nil {
		return 0
	}
	return dur
}

func (c *ConfigItem) GetString() string {
	if c == nil {
		return ""
	}
	return c.Value
}

func (c *ConfigItem) GetStringSlice() []string {
	if c == nil {
		return []string{}
	}
	return strings.Split(c.Value, ",")
}
