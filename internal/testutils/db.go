package testutils

import (
	"log"

	"github.com/hail2skins/the-virtual-armory/internal/models"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var (
	TestDB *gorm.DB
)

// SetupTestDB initializes an in-memory SQLite database for testing
func SetupTestDB() (*gorm.DB, error) {
	// Use SQLite in-memory database for tests
	config := &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	}

	var err error
	TestDB, err = gorm.Open(sqlite.Open("file::memory:?cache=shared"), config)
	if err != nil {
		log.Printf("Failed to connect to test database: %v", err)
		return nil, err
	}

	// Auto migrate the schema
	err = TestDB.AutoMigrate(
		&models.User{},
		&models.Manufacturer{},
		&models.Caliber{},
		&models.WeaponType{},
		&models.Gun{},
		&models.Payment{},
	)
	if err != nil {
		log.Printf("Failed to migrate test database: %v", err)
		return nil, err
	}

	log.Printf("Connected to test database")
	return TestDB, nil
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(db *gorm.DB) {
	// Delete all records from all tables
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM manufacturers")
	db.Exec("DELETE FROM calibers")
	db.Exec("DELETE FROM weapon_types")
	db.Exec("DELETE FROM guns")
	db.Exec("DELETE FROM payments")
}

// CreateTestUser creates a test user in the database
func CreateTestUser(db *gorm.DB, email, password string, isAdmin bool) (*models.User, error) {
	user := &models.User{
		Email:    email,
		Password: password,
		IsAdmin:  isAdmin,
	}

	result := db.Create(user)
	if result.Error != nil {
		return nil, result.Error
	}

	return user, nil
}

// GetTestUser retrieves a user by email from the test database
func GetTestUser(db *gorm.DB, email string) (*models.User, error) {
	var user models.User
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		return nil, result.Error
	}
	return &user, nil
}
