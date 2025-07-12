package database

import "gorm.io/gorm"

// DataStore defines the interface that all domain-specific stores must implement
type DataStore interface {
	// Migrate runs the store-specific migrations
	Migrate() error
	// GetDB returns the database connection
	GetDB() *gorm.DB
}

// BaseDataStore provides common functionality for all data stores
type BaseDataStore struct {
	db *gorm.DB
}

// NewBaseDataStore creates a new base data store
func NewBaseDataStore(db *gorm.DB) *BaseDataStore {
	return &BaseDataStore{
		db: db,
	}
}

// GetDB returns the database connection
func (s *BaseDataStore) GetDB() *gorm.DB {
	return s.db
}

// WithTransaction executes the given function within a transaction
func (s *BaseDataStore) WithTransaction(fn func(tx *gorm.DB) error) error {
	return s.db.Transaction(fn)
}
