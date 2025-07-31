package database

import (
	"database/sql"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type sqliteDialector struct {
	Conn *sql.DB
}

func (s sqliteDialector) Name() string {
	return "sqlite"
}

func (s sqliteDialector) Initialize(db *gorm.DB) error {
	return sqlite.Dialector{
		Conn: s.Conn,
	}.Initialize(db)
}
