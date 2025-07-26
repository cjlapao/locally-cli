package types

import "gorm.io/gorm"

type DataStore interface {
	GetDB() *gorm.DB
	Migrate() error
	WithTransaction(fn func(tx *gorm.DB) error) error
}
