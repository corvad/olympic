package cascade

import (
	"fmt"
	"log"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type DBConnection struct {
	Address string
}

type DB struct {
	details DBConnection
	DB      *gorm.DB
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

func NewDatabase(dbConn DBConnection) (*DB, error) {
	if dbConn.Address == "" {
		return nil, fmt.Errorf("dbConn.Address cannot be empty")
	}

	db, err := gorm.Open(sqlite.Open(dbConn.Address), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// Migrate the schema
	err = db.AutoMigrate(&Account{}, &Login{}, &Link{}, &Query{})
	if err != nil {
		return nil, err
	}

	log.Println("Database connected and migrated.")

	return &DB{
		details: dbConn,
		DB:      db,
	}, nil
}

func (db *DB) Close() {
	sqlDB, err := db.DB.DB()
	if err != nil {
		log.Println("Error getting database object for closing: ", err)
	}
	err = sqlDB.Close()
	if err != nil {
		log.Println("Error closing database: ", err)
	}
	log.Println("Database connection closed.")
}
