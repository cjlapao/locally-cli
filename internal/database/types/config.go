package types

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
