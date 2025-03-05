package database

import (
	"log"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testDB *gorm.DB

// InitTestDB initializes a test database using SQLite in-memory
func InitTestDB() (*gorm.DB, error) {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	testDB = db
	return db, nil
}

// SetDB sets the test database
func SetDB(db *gorm.DB) {
	testDB = db
}

// CloseDB closes the test database connection
func CloseDB() {
	if testDB != nil {
		sqlDB, err := testDB.DB()
		if err != nil {
			log.Printf("Error getting SQL DB: %v", err)
			return
		}
		sqlDB.Close()
	}
}
