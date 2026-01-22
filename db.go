package cascade

import (
	"fmt"
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
	LinkID    uint // Link ID
	QueriedBy uint // Account ID
}

type Link struct {
	gorm.Model
	Url      string
	ShortUrl string
	Count    int
	Owner    uint
}

type Login struct {
	gorm.Model
	AccountID    uint
	IpAddress    string
	UserAgent    string
	RefreshToken string
	ExpiresAt    time.Time
}

type Account struct {
	gorm.Model
	Email    string
	Verified bool
	Password string
}

func OpenDB(sqliteDBPath string) (*DB, error) {
	if sqliteDBPath == "" {
		return nil, fmt.Errorf("sqliteDBPath cannot be empty")
	}

	db, err := gorm.Open(sqlite.Open(sqliteDBPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	err = db.AutoMigrate(&Account{}, &Login{}, &Link{}, &Query{})
	if err != nil {
		return nil, err
	}

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
