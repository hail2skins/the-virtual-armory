package payment_test_utils

import (
	"fmt"
	"html/template"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// TestRenderer is a simple mock HTML renderer for tests
type TestRenderer struct{}

// Instance implements the HTMLRender interface
func (r *TestRenderer) Instance(name string, data interface{}) render.Render {
	// Try to extract content from the data map
	var content string
	if dataMap, ok := data.(gin.H); ok {
		if contentVal, exists := dataMap["content"]; exists {
			content = fmt.Sprintf("%v", contentVal)
		}
	}

	// If no content was found, use a default template
	if content == "" {
		content = "Mock template"
	}

	return &render.HTML{
		Template: template.Must(template.New("").Parse("<html>" + content + "</html>")),
		Data:     data,
	}
}

// SetupTestDB sets up a test database
func SetupTestDB(t *testing.T) *gorm.DB {
	// Set up a test database using SQLite in-memory
	db, err := testutils.SetupTestDB()
	assert.NoError(t, err)

	// Set the global database variable to the test database
	// This ensures that any code using database.GetDB() will use our test database
	database.DB = db

	// Seed the database with necessary reference data if it doesn't exist
	SeedTestDatabase(t, db)

	return db
}

// CleanupTestDB cleans up the test database
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	testutils.CleanupTestDB(db)
}

// SeedTestDatabase seeds the test database with necessary reference data
func SeedTestDatabase(t *testing.T, db *gorm.DB) {
	// Check if we already have weapon types
	var count int64
	db.Model(&models.WeaponType{}).Count(&count)
	if count == 0 {
		// Seed weapon types
		weaponTypes := []models.WeaponType{
			{Type: "Handgun", Nickname: "Pistol"},
			{Type: "Rifle", Nickname: "Long gun"},
			{Type: "Shotgun", Nickname: "Scatter gun"},
		}
		for _, wt := range weaponTypes {
			if err := db.Create(&wt).Error; err != nil {
				t.Fatalf("Failed to seed weapon type: %v", err)
			}
		}
	}

	// Check if we already have calibers
	db.Model(&models.Caliber{}).Count(&count)
	if count == 0 {
		// Seed calibers
		calibers := []models.Caliber{
			{Caliber: "9mm", Nickname: "Nine"},
			{Caliber: ".45 ACP", Nickname: "Forty-five"},
			{Caliber: "5.56x45mm", Nickname: "NATO"},
		}
		for _, c := range calibers {
			if err := db.Create(&c).Error; err != nil {
				t.Fatalf("Failed to seed caliber: %v", err)
			}
		}
	}

	// Check if we already have manufacturers
	db.Model(&models.Manufacturer{}).Count(&count)
	if count == 0 {
		// Seed manufacturers
		manufacturers := []models.Manufacturer{
			{Name: "Glock", Country: "Austria"},
			{Name: "Smith & Wesson", Country: "USA"},
			{Name: "Colt", Country: "USA"},
		}
		for _, m := range manufacturers {
			if err := db.Create(&m).Error; err != nil {
				t.Fatalf("Failed to seed manufacturer: %v", err)
			}
		}
	}
}

// CreateTestUser creates a test user
func CreateTestUser(t *testing.T, db *gorm.DB) *models.User {
	// Generate a unique email for each test
	uniqueEmail := fmt.Sprintf("test%d@example.com", time.Now().UnixNano())

	user := &models.User{
		Email:                 uniqueEmail,
		Password:              "password",
		SubscriptionTier:      "free",
		SubscriptionExpiresAt: time.Now(),
		Confirmed:             true,
	}
	err := db.Create(user).Error
	assert.NoError(t, err)
	return user
}

// AuthMiddlewareMock creates a middleware that mocks authentication
func AuthMiddlewareMock(user *models.User) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Set the user in the context
		c.Set("user", user)
		c.Next()
	}
}

// CreateTestPayment creates a test payment
func CreateTestPayment(t *testing.T, db *gorm.DB, user *models.User) *models.Payment {
	payment := &models.Payment{
		UserID:      user.ID,
		Amount:      500, // $5.00
		Currency:    "usd",
		PaymentType: "subscription",
		Status:      "succeeded",
		Description: "Monthly subscription",
		StripeID:    "pi_test123",
	}
	err := db.Create(payment).Error
	assert.NoError(t, err)
	return payment
}

// GetExistingWeaponType gets an existing weapon type from the database
func GetExistingWeaponType(t *testing.T, db *gorm.DB) *models.WeaponType {
	var weaponType models.WeaponType
	result := db.First(&weaponType)
	if result.Error != nil {
		t.Fatalf("Failed to get existing weapon type: %v", result.Error)
	}
	return &weaponType
}

// GetExistingCaliber gets an existing caliber from the database
func GetExistingCaliber(t *testing.T, db *gorm.DB) *models.Caliber {
	var caliber models.Caliber
	result := db.First(&caliber)
	if result.Error != nil {
		t.Fatalf("Failed to get existing caliber: %v", result.Error)
	}
	return &caliber
}

// GetExistingManufacturer gets an existing manufacturer from the database
func GetExistingManufacturer(t *testing.T, db *gorm.DB) *models.Manufacturer {
	var manufacturer models.Manufacturer
	result := db.First(&manufacturer)
	if result.Error != nil {
		t.Fatalf("Failed to get existing manufacturer: %v", result.Error)
	}
	return &manufacturer
}

// SetupPricingTestRouter sets up a test router with the payment controller for pricing tests
func SetupPricingTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *controllers.PaymentController) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.Default()

	// For tests, we don't need to load actual templates
	// Just mock the HTML renderer to prevent panics
	router.HTMLRender = &TestRenderer{}

	// Create the payment controller
	paymentController := controllers.NewPaymentController(db)

	return router, paymentController
}

// SetupGunTestRouter sets up a test router with the gun controller
func SetupGunTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *controllers.GunController) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.Default()

	// For tests, we don't need to load actual templates
	// Just mock the HTML renderer to prevent panics
	router.HTMLRender = &TestRenderer{}

	// Create the gun controller
	gunController := controllers.NewGunController(db)

	return router, gunController
}
