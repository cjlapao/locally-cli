package config

import "time"

type Config struct {
	Debug bool         `json:"debug"`
	Items []ConfigItem `json:"items"`
}

func (c *Config) Get(key string) *ConfigItem {
	for _, item := range c.Items {
		if item.Key == key {
			return &item
		}
	}

	return nil
}

func (c *Config) IsDebug() bool {
	return c.GetBool(DebugKey, false)
}

func (c *Config) GetValue(key string, defaultValue interface{}) interface{} {
	item := c.Get(key)
	if item == nil || !item.IsSet() {
		return defaultValue
	}

	return item.Value
}

func (c *Config) GetString(key string, defaultValue string) string {
	item := c.Get(key)
	if item == nil || !item.IsSet() {
		return defaultValue
	}

	return item.GetString()
}

func (c *Config) GetBool(key string, defaultValue bool) bool {
	item := c.Get(key)
	if item == nil || !item.IsSet() {
		return defaultValue
	}

	return item.GetBool()
}

func (c *Config) GetInt(key string, defaultValue int) int {
	item := c.Get(key)
	if item == nil || !item.IsSet() {
		return defaultValue
	}

	return item.GetInt()
}

func (c *Config) GetDuration(key string, defaultValue time.Duration) time.Duration {
	item := c.Get(key)
	if item == nil || !item.IsSet() {
		return defaultValue
	}

	return item.GetDuration()
}

func (c *Config) Set(key string, value string) {
	for i, item := range c.Items {
		if item.Key == key {
			c.Items[i].Value = value
			return
		}
	}

	c.Items = append(c.Items, ConfigItem{Key: key, Value: value})
}
