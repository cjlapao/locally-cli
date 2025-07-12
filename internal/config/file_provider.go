package config

import (
	"encoding/json"
	"fmt"
	"os"
)

// FileProvider provides configuration from a JSON file
type FileProvider struct {
	filePath string
}

// NewFileProvider creates a new file provider
func NewFileProvider(filePath string) *FileProvider {
	return &FileProvider{
		filePath: filePath,
	}
}

// Load implements the Provider interface
func (p *FileProvider) Load(cfg *Config) error {
	if p.filePath == "" {
		return nil
	}

	// Check if file exists
	if _, err := os.Stat(p.filePath); os.IsNotExist(err) {
		return fmt.Errorf("config file not found: %s", p.filePath)
	}

	// Read file
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	// Parse JSON
	var fileConfig []ConfigItem
	if err := json.Unmarshal(data, &fileConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	cfg.Items = append(cfg.Items, fileConfig...)

	return nil
}

func (p *FileProvider) IsAvailable() bool {
	if p.filePath == "" {
		return false
	}
	_, err := os.Stat(p.filePath)
	return err == nil
}
