package olympic

import (
	"fmt"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DB struct {
	sqliteDBPath string
	DB           *gorm.DB
}

type Query struct {
	gorm.Model
	queriedBy uint // Account ID
}

type Link struct {
	gorm.Model
	url     string
	count   int
	owner   Account
	queries []uint // Click IDs of clicks on this link
}

type Login struct {
	gorm.Model
	accountID uint
	timestamp time.Time
	ipAddress string
	userAgent string
}

type Account struct {
	gorm.Model
	email    string
	verified bool
	password string
	logins   []uint // login IDs
}

func OpenDB(sqliteDBPath string) (*DB, error) {
	if sqliteDBPath == "" {
		return nil, fmt.Errorf("sqliteDBPath cannot be empty")
	}

	//check if db already exists
	file, err := os.Stat(sqliteDBPath)
	if err != nil {
		return nil, fmt.Errorf("failed to stat database file: %w", err)
	}

	if file.IsDir() {
		return nil, fmt.Errorf("the provided sqliteDBPath is a directory, not a file")
	}

	db, err := gorm.Open(sqlite.Open(sqliteDBPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	db.AutoMigrate(&Account{}, &Login{})

	
	return &DB{
		sqliteDBPath: sqliteDBPath,
		DB:           db,
	}, nil
}

func (db *DB) Close() error {
	sqlDB, err := db.DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get db: %w", err)
	}
	err = sqlDB.Close()
	if err != nil {
		return fmt.Errorf("failed to close database: %w", err)
	}
	return nil
}
