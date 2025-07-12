package database

import (
	"time"

	"github.com/cjlapao/locally-cli/internal/database/types"
)

// Base model for all database models

// Device represents a managed device in Jamf
type Device struct {
	types.BaseModel
	Name         string    `gorm:"uniqueIndex" json:"name"`
	SerialNumber string    `gorm:"uniqueIndex" json:"serial_number"`
	UDID         string    `gorm:"uniqueIndex" json:"udid"`
	Model        string    `json:"model"`
	Status       string    `json:"status"`
	LastSeen     time.Time `json:"last_seen"`
}

type Authentication struct {
	types.BaseModel
	Username           string `json:"username"`
	Password           string `json:"password"`
	CurrentJwtToken    string `json:"current_jwt_token"`
	DeviceSerialNumber string `gorm:"uniqueIndex" json:"device_serial_number"`
}

// Configuration represents a configuration profile
type Configuration struct {
	types.BaseModel
	Name        string `gorm:"uniqueIndex" json:"name"`
	Identifier  string `json:"identifier"`
	Description string `json:"description"`
	Platform    string `json:"platform"`
	Scope       string `json:"scope"`
	Version     int    `json:"version"`
}

// DatabaseType represents the type of database
type DatabaseType string

const (
	SQLite     DatabaseType = "sqlite"
	PostgreSQL DatabaseType = "postgresql"
)

// Config represents database configuration
type Config struct {
	Type DatabaseType `json:"type"`

	// SQLite configuration
	StoragePath string `json:"storage_path"` // Path to the SQLite database file

	// PostgreSQL configuration
	Host     string `json:"host"`     // PostgreSQL host
	Port     int    `json:"port"`     // PostgreSQL port
	Database string `json:"database"` // PostgreSQL database name
	Username string `json:"username"` // PostgreSQL username
	Password string `json:"password"` // PostgreSQL password
	SSLMode  bool   `json:"ssl_mode"` // PostgreSQL SSL mode (disable, require, verify-ca, verify-full)

	// Common configuration
	Debug bool `json:"debug"` // Enable debug logging
}
