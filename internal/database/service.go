// Package database provides a service for managing the database connection.
package database

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/cjlapao/locally-cli/internal/database/types"
	"github.com/cjlapao/locally-cli/internal/logging"
	_ "github.com/lib/pq"
	"github.com/sirupsen/logrus"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	instance *Service
	once     sync.Once
)

// Service represents the database service
type Service struct {
	db     *gorm.DB
	config *types.Config
}

// GetInstance returns the singleton instance of the database service
func GetInstance() *Service {
	return instance
}

// Initialize initializes the database service singleton with the given config
func Initialize(config *types.Config) error {
	var initErr error
	once.Do(func() {
		// Configure logging
		logLevel := logger.Silent
		if config.Debug {
			logLevel = logger.Info
		}

		gormConfig := &gorm.Config{
			Logger: logger.Default.LogMode(logLevel),
		}

		var db *gorm.DB
		var err error

		// Initialize database based on type
		switch config.Type {
		case types.SQLite:
			db, err = initializeSQLite(config, gormConfig)
		case types.PostgreSQL:
			db, err = initializePostgreSQL(config, gormConfig)
		default:
			initErr = fmt.Errorf("unsupported database type: %s", config.Type)
			return
		}

		if err != nil {
			initErr = err
			return
		}

		instance = &Service{
			db:     db,
			config: config,
		}

		logging.WithField("database_type", config.Type).Info("Database service initialized")
	})

	return initErr
}

// initializeSQLite initializes SQLite database connection
func initializeSQLite(config *types.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	// Convert to absolute path if relative
	absPath, err := filepath.Abs(config.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get absolute path: %w", err)
	}

	// Ensure the directory exists
	if err := ensureDir(absPath); err != nil {
		return nil, fmt.Errorf("failed to create database directory: %w", err)
	}

	logging.WithField("database_path", absPath).Info("Using SQLite database path")

	// Open database connection
	db, err := gorm.Open(sqlite.Open(absPath), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to SQLite database: %w", err)
	}

	return db, nil
}

// initializePostgreSQL initializes PostgreSQL database connection
func initializePostgreSQL(config *types.Config, gormConfig *gorm.Config) (*gorm.DB, error) {
	// test the connection to the postgres server
	if err := testConnection(config); err != nil {
		return nil, fmt.Errorf("failed to test PostgreSQL connection: %w", err)
	}

	// Check if database exists
	exists, err := checkDatabaseExists(config)
	if err != nil {
		return nil, fmt.Errorf("failed to check database existence: %w", err)
	}

	// If database doesn't exist, create it
	if !exists {
		logging.WithField("database_name", config.Database).Info("Database does not exist, creating...")
		if err := createDatabase(config); err != nil {
			return nil, fmt.Errorf("failed to create database: %w", err)
		}
		logging.WithField("database_name", config.Database).Info("Database created successfully")
	}

	// Build connection string for the connection with the database name
	dsn := buildPostgresConnectionString(config, config.Database)
	logging.WithFields(logrus.Fields{
		"host":     config.Host,
		"port":     config.Port,
		"database": config.Database,
	}).Info("Connecting to PostgreSQL database")

	// Open database connection
	db, err := gorm.Open(postgres.Open(dsn), gormConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL database: %w", err)
	}

	return db, nil
}

// GetDB returns the database connection
func (s *Service) GetDB() *gorm.DB {
	return s.db
}

// Close closes the database connection
func (s *Service) Close() error {
	if s.db != nil {
		sqlDB, err := s.db.DB()
		if err != nil {
			return fmt.Errorf("failed to get underlying sql.DB: %w", err)
		}
		return sqlDB.Close()
	}
	return nil
}

// Migrate runs database migrations
func (s *Service) Migrate() error {
	// Add all models here
	return nil
}

// ensureDir creates the directory for the database file if it doesn't exist
func ensureDir(dbPath string) error {
	dir := filepath.Dir(dbPath)
	return createDirIfNotExists(dir)
}

// createDirIfNotExists creates a directory if it doesn't exist
func createDirIfNotExists(dir string) error {
	// Check if directory exists
	if _, err := os.Stat(dir); os.IsNotExist(err) {
		logging.WithField("directory", dir).Info("Creating directory")
		// Create directory with permissions rwxr-xr-x
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}
	return nil
}

// buildPostgresConnectionString creates a PostgreSQL connection string
func buildPostgresConnectionString(config *types.Config, dbName string) string {
	if config.Host == "" {
		config.Host = "localhost"
	}
	if config.Port == 0 {
		config.Port = 5432
	}

	sslMode := "disable"
	if config.SSLMode {
		switch config.Type {
		case types.SQLite:
			sslMode = "disable"
		case types.PostgreSQL:
			sslMode = "prefer"
		}
	}
	conn := fmt.Sprintf("host=%s user=%s password=%s port=%d",
		config.Host,
		config.Username,
		config.Password,
		config.Port,
	)
	if sslMode != "disable" {
		conn += fmt.Sprintf(" sslmode=%s", sslMode)
	}
	if dbName != "" {
		conn += fmt.Sprintf(" database=%s", dbName)
	}
	return conn
}

// checkDatabaseExists checks if the specified database exists
func checkDatabaseExists(config *types.Config) (bool, error) {
	// Connect to PostgreSQL server without specifying a database
	var dialector gorm.Dialector
	switch config.Type {
	case types.SQLite:
		return true, nil
	case types.PostgreSQL:
		dsn := buildPostgresConnectionString(config, "postgres")
		dialector = postgres.Open(dsn)
	default:
		return false, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	srv, err := gorm.Open(dialector, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		return false, fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}

	// Check if database exists
	var exists int
	switch config.Type {
	case types.PostgreSQL:
		srv.Raw("SELECT 1 FROM pg_database WHERE datname = ?", config.Database).Scan(&exists)
	case types.SQLite:
		exists = 1
	default:
		return false, fmt.Errorf("unsupported database type: %s", config.Type)
	}

	sqlDB, err := srv.DB()
	if err != nil {
		return false, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	return exists == 1, nil
}

// createDatabase creates the specified database
func createDatabase(config *types.Config) error {
	var dial gorm.Dialector
	switch config.Type {
	case types.SQLite:
		return nil
	case types.PostgreSQL:
		dsn := buildPostgresConnectionString(config, "postgres")
		dial = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	srv, err := gorm.Open(dial, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}

	// creating the database
	var stmt string
	switch config.Type {
	case types.PostgreSQL:
		stmt = fmt.Sprintf("CREATE DATABASE \"%s\"", config.Database)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err := srv.Exec(stmt).Error; err != nil {
		if !strings.Contains(err.Error(), "already exists") {
			return fmt.Errorf("failed to create database: %w", err)
		}
	}

	// Setting up initial privileges
	var privs string
	switch config.Type {
	case types.PostgreSQL:
		privs = "GRANT ALL PRIVILEGES ON DATABASE " + config.Database + " TO " + config.Username
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	if err := srv.Exec(privs).Error; err != nil {
		return fmt.Errorf("failed to grant privileges: %w", err)
	}

	return nil
}

// testConnection tests the connection to the postgres server
func testConnection(config *types.Config) error {
	var dial gorm.Dialector
	switch config.Type {
	case types.SQLite:
		return nil
	case types.PostgreSQL:
		dsn := buildPostgresConnectionString(config, "postgres")
		dial = postgres.Open(dsn)
	default:
		return fmt.Errorf("unsupported database type: %s", config.Type)
	}

	srv, err := gorm.Open(dial, &gorm.Config{SkipDefaultTransaction: true})
	if err != nil {
		return fmt.Errorf("failed to connect to PostgreSQL server: %w", err)
	}

	sqlDB, err := srv.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}
	defer sqlDB.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return fmt.Errorf("ping: %w", err)
	}

	return nil
}
