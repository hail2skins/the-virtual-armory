package database

import (
	"fmt"
	"log"
	"os"

	"github.com/hail2skins/the-virtual-armory/internal/database/seed"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	DB *gorm.DB
)

// InitGORM initializes the GORM database connection
func InitGORM() (*gorm.DB, error) {
	// If DB is already initialized, return it
	if DB != nil {
		return DB, nil
	}

	// Get database connection details from environment variables
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USERNAME")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_DATABASE")
	schema := os.Getenv("DB_SCHEMA")

	// Create DSN string
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s search_path=%s sslmode=disable",
		host, user, password, dbname, port, schema)

	// Set up GORM configuration
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	}

	// Connect to the database
	var err error
	DB, err = gorm.Open(postgres.Open(dsn), config)
	if err != nil {
		log.Printf("Failed to connect to database: %v", err)
		return nil, err
	}

	// Auto migrate the schema
	err = DB.AutoMigrate(
		&models.User{},
		&models.Manufacturer{},
		&models.Caliber{},
		&models.WeaponType{},
		&models.Gun{},
	)
	if err != nil {
		log.Printf("Failed to migrate database: %v", err)
		return nil, err
	}

	log.Printf("Connected to database: %s", dbname)

	// Run database seeds
	seed.RunSeeds(DB)

	return DB, nil
}

// GetDB returns the GORM database connection
func GetDB() *gorm.DB {
	if DB == nil {
		var err error
		DB, err = InitGORM()
		if err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}
	}
	return DB
}
