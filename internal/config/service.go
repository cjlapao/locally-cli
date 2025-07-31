package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

var (
	instance *ConfigService
	once     sync.Once
)

// Provider interface for configuration providers
type Provider interface {
	Load(cfg *Config) error
}

// ConfigService represents the configuration service
type ConfigService struct {
	config      *Config
	storagePath string
	providers   []Provider
	isLoaded    bool
}

// GetInstance returns the singleton instance of the config service
func GetInstance() *ConfigService {
	return instance
}

// Initialize initializes the config service singleton with the given providers
func Initialize(providers ...Provider) (*ConfigService, error) {
	var initErr error
	once.Do(func() {
		service := &ConfigService{
			config:      DefaultConfig(),
			providers:   providers,
			storagePath: "",
		}

		// Add providers in order of priority (lowest to highest)
		service.AddProvider(NewEnvProvider())
		service.AddProvider(NewFlagProvider())

		configFlag := flag.String("config", "", "Path to configuration file (JSON or YAML)")
		flag.Parse()
		service.AddProvider(NewFileProvider(*configFlag))

		// Load configuration from all providers
		if err := service.Load(); err != nil {
			initErr = fmt.Errorf("failed to load configuration: %w", err)
			return
		}

		// Get user folder
		userFolder, err := os.UserHomeDir()
		if err != nil {
			initErr = fmt.Errorf("failed to get user home directory: %w", err)
			return
		}
		storagePath := filepath.Join(userFolder, SystemStoragePath)
		// creating the folder if it doesn't exist
		if err := os.MkdirAll(storagePath, 0o755); err != nil {
			initErr = fmt.Errorf("failed to create locally storage directory: %w", err)
			return
		}
		service.storagePath = storagePath

		instance = service
	})

	return instance, initErr
}

// Load loads configuration from all providers
func (s *ConfigService) Load() error {
	if s.isLoaded {
		return nil
	}

	for _, provider := range s.providers {
		if err := provider.Load(s.config); err != nil {
			return fmt.Errorf("failed to load config from provider: %w", err)
		}
	}

	s.isLoaded = true
	return nil
}

// Get returns the current configuration
func (s *ConfigService) Get() *Config {
	return s.config
}

// Reload reloads configuration from all providers
func (s *ConfigService) Reload() error {
	s.isLoaded = false
	return s.Load()
}

// AddProvider adds a new configuration provider
func (s *ConfigService) AddProvider(provider Provider) {
	s.providers = append(s.providers, provider)
	s.isLoaded = false
}

func (s *ConfigService) SetConfigFile(path string) {
	s.providers = append(s.providers, NewFileProvider(path))
}

func (s *ConfigService) SetStoragePath(path string) {
	s.config.Items = append(s.config.Items, ConfigItem{
		Key:   "database.storage_path",
		Value: path,
	})
}

func (s *ConfigService) GetStoragePath() (string, error) {
	storagePath := s.config.GetString(DatabaseStoragePathKey, "")
	if storagePath == "" {
		s.config.Set(DatabaseStoragePathKey, s.storagePath)
		return s.storagePath, nil
	}

	return storagePath, nil
}
