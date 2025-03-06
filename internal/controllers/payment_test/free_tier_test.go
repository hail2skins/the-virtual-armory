package payment_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupTestDB sets up a test database and seeds it with necessary reference data
func setupTestDB(t *testing.T) *gorm.DB {
	// Set up a test database using SQLite in-memory
	db, err := testutils.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}

	// Set the global database variable to the test database
	// This ensures that any code using database.GetDB() will use our test database
	database.DB = db

	// Seed the database with necessary reference data if it doesn't exist
	seedTestDatabase(t, db)

	return db
}

// seedTestDatabase seeds the test database with necessary reference data
func seedTestDatabase(t *testing.T, db *gorm.DB) {
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

// setupTestRouter sets up a test router with the gun controller
func setupTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *controllers.GunController) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.Default()

	// For tests, we don't need to load actual templates
	// Just mock the HTML renderer to prevent panics
	router.HTMLRender = &testRenderer{}

	// Create the gun controller
	gunController := controllers.NewGunController(db)

	return router, gunController
}

// testRenderer is a simple mock HTML renderer for tests
type testRenderer struct{}

// Instance implements the HTMLRender interface
func (r *testRenderer) Instance(name string, data interface{}) render.Render {
	return &render.HTML{
		Template: template.Must(template.New("").Parse("<html>Mock template</html>")),
		Data:     data,
	}
}

// createTestUser creates a test user in the database
func createTestUser(t *testing.T, db *gorm.DB) *models.User {
	// Generate a unique email for each test
	uniqueEmail := fmt.Sprintf("test%d@example.com", time.Now().UnixNano())

	user := &models.User{
		Email:     uniqueEmail,
		Password:  "password",
		Confirmed: true,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("Failed to create test user: %v", err)
	}
	return user
}

// createTestWeaponType creates a test weapon type in the database
func createTestWeaponType(t *testing.T, db *gorm.DB) *models.WeaponType {
	// Check if a test weapon type already exists
	var weaponType models.WeaponType
	result := db.Where("type = ?", "Test Weapon Type").First(&weaponType)

	if result.Error == nil {
		// Weapon type already exists, return it
		return &weaponType
	}

	// Create a new weapon type
	weaponType = models.WeaponType{
		Type:     "Test Weapon Type",
		Nickname: "Test Nickname",
	}
	if err := db.Create(&weaponType).Error; err != nil {
		t.Fatalf("Failed to create test weapon type: %v", err)
	}
	return &weaponType
}

// createTestCaliber creates a test caliber in the database
func createTestCaliber(t *testing.T, db *gorm.DB) *models.Caliber {
	// Check if a test caliber already exists
	var caliber models.Caliber
	result := db.Where("caliber = ?", "Test Caliber").First(&caliber)

	if result.Error == nil {
		// Caliber already exists, return it
		return &caliber
	}

	// Create a new caliber
	caliber = models.Caliber{
		Caliber:  "Test Caliber",
		Nickname: "Test Nickname",
	}
	if err := db.Create(&caliber).Error; err != nil {
		t.Fatalf("Failed to create test caliber: %v", err)
	}
	return &caliber
}

// createTestManufacturer creates a test manufacturer in the database
func createTestManufacturer(t *testing.T, db *gorm.DB) *models.Manufacturer {
	// Check if a test manufacturer already exists
	var manufacturer models.Manufacturer
	result := db.Where("name = ?", "Test Manufacturer").First(&manufacturer)

	if result.Error == nil {
		// Manufacturer already exists, return it
		return &manufacturer
	}

	// Create a new manufacturer
	manufacturer = models.Manufacturer{
		Name:     "Test Manufacturer",
		Nickname: "Test Nickname",
		Country:  "Test Country",
	}
	if err := db.Create(&manufacturer).Error; err != nil {
		t.Fatalf("Failed to create test manufacturer: %v", err)
	}
	return &manufacturer
}

func getExistingWeaponType(t *testing.T, db *gorm.DB) *models.WeaponType {
	var weaponType models.WeaponType
	result := db.First(&weaponType)
	if result.Error != nil {
		t.Fatalf("Failed to get existing weapon type: %v", result.Error)
	}
	return &weaponType
}

func getExistingCaliber(t *testing.T, db *gorm.DB) *models.Caliber {
	var caliber models.Caliber
	result := db.First(&caliber)
	if result.Error != nil {
		t.Fatalf("Failed to get existing caliber: %v", result.Error)
	}
	return &caliber
}

func getExistingManufacturer(t *testing.T, db *gorm.DB) *models.Manufacturer {
	var manufacturer models.Manufacturer
	result := db.First(&manufacturer)
	if result.Error != nil {
		t.Fatalf("Failed to get existing manufacturer: %v", result.Error)
	}
	return &manufacturer
}

// TestFreeTierGunLimit tests that users on the free tier can only create 2 guns
func TestFreeTierGunLimit(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Get existing reference data
	weaponType := payment_test_utils.GetExistingWeaponType(t, db)
	caliber := payment_test_utils.GetExistingCaliber(t, db)
	manufacturer := payment_test_utils.GetExistingManufacturer(t, db)

	// Set up test router and controller
	router, gunController := payment_test_utils.SetupGunTestRouter(t, db)

	// Set up the route for creating guns
	router.POST("/owner/guns", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		gunController.Create(c)
	})

	// Create the first gun (should succeed)
	formData1 := url.Values{
		"name":            {"Test Gun 1"},
		"serial_number":   {"SN12345"},
		"weapon_type_id":  {fmt.Sprintf("%d", weaponType.ID)},
		"caliber_id":      {fmt.Sprintf("%d", caliber.ID)},
		"manufacturer_id": {fmt.Sprintf("%d", manufacturer.ID)},
	}
	req1, _ := http.NewRequest("POST", "/owner/guns", strings.NewReader(formData1.Encode()))
	req1.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Add authentication cookies to the request
	req1.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req1.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Check that the gun was created successfully
	assert.Equal(t, http.StatusSeeOther, w1.Code) // Redirect after successful creation

	// Create the second gun (should succeed)
	formData2 := url.Values{
		"name":            {"Test Gun 2"},
		"serial_number":   {"SN67890"},
		"weapon_type_id":  {fmt.Sprintf("%d", weaponType.ID)},
		"caliber_id":      {fmt.Sprintf("%d", caliber.ID)},
		"manufacturer_id": {fmt.Sprintf("%d", manufacturer.ID)},
	}
	req2, _ := http.NewRequest("POST", "/owner/guns", strings.NewReader(formData2.Encode()))
	req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Add authentication cookies to the request
	req2.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req2.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Check that the second gun was created successfully
	assert.Equal(t, http.StatusSeeOther, w2.Code) // Redirect after successful creation

	// Try to create a third gun (should fail)
	formData3 := url.Values{
		"name":            {"Test Gun 3"},
		"serial_number":   {"SN13579"},
		"weapon_type_id":  {fmt.Sprintf("%d", weaponType.ID)},
		"caliber_id":      {fmt.Sprintf("%d", caliber.ID)},
		"manufacturer_id": {fmt.Sprintf("%d", manufacturer.ID)},
	}
	req3, _ := http.NewRequest("POST", "/owner/guns", strings.NewReader(formData3.Encode()))
	req3.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Add authentication cookies to the request
	req3.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req3.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w3 := httptest.NewRecorder()
	router.ServeHTTP(w3, req3)

	// Check that the user is redirected to the pricing page
	assert.Equal(t, http.StatusSeeOther, w3.Code)
	assert.Equal(t, "/pricing", w3.Header().Get("Location"))

	// Verify that only 2 guns were created
	var count int64
	db.Model(&models.Gun{}).Where("owner_id = ?", user.ID).Count(&count)
	assert.Equal(t, int64(2), count)
}

// TestExpiredSubscriptionRevertToFreeTier tests that users with expired subscriptions revert to the free tier
func TestExpiredSubscriptionRevertToFreeTier(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user with a subscription
	user := createTestUser(t, db)

	// Set up subscription data
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": time.Now().Add(-24 * time.Hour), // Expired 1 day ago
	})

	// Get existing reference data
	weaponType := getExistingWeaponType(t, db)
	caliber := getExistingCaliber(t, db)
	manufacturer := getExistingManufacturer(t, db)

	// Set up test router and controller
	router, gunController := setupTestRouter(t, db)

	// Create 3 guns for the user directly in the database
	// (bypassing the controller to simulate a user who had a subscription)
	for i := 1; i <= 3; i++ {
		gun := models.Gun{
			Name:           "Test Gun " + fmt.Sprintf("%d", i),
			WeaponTypeID:   weaponType.ID,
			CaliberID:      caliber.ID,
			ManufacturerID: manufacturer.ID,
			OwnerID:        user.ID,
		}
		db.Create(&gun)
	}

	// Set up the route for listing guns
	router.GET("/owner/guns", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		gunController.Index(c)
	})

	// Test accessing the guns list
	req, _ := http.NewRequest("GET", "/owner/guns", nil)
	// Add authentication cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// The response should only include the first 2 guns
	assert.Equal(t, http.StatusOK, w.Code)

	// The actual assertion for the number of guns will depend on how the view is implemented
	// For now, we'll check the database to ensure the user has 3 guns total
	var totalGuns int64
	db.Model(&models.Gun{}).Where("owner_id = ?", user.ID).Count(&totalGuns)
	assert.Equal(t, int64(3), totalGuns)

	// But the controller should only return 2 guns for display
	// This will be implemented in the controller later
}

// TestExpiredSubscriptionShowsLimitedGuns tests that a user with an expired subscription
// only sees their first 2 guns with an indication that there are more
func TestExpiredSubscriptionShowsLimitedGuns(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user with an expired subscription
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up subscription data - expired 1 day ago
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": time.Now().Add(-24 * time.Hour),
	})

	// Get existing reference data
	weaponType := payment_test_utils.GetExistingWeaponType(t, db)
	caliber := payment_test_utils.GetExistingCaliber(t, db)
	manufacturer := payment_test_utils.GetExistingManufacturer(t, db)

	// Create 5 guns for the user directly in the database
	// (bypassing the controller to simulate a user who had a subscription)
	for i := 1; i <= 5; i++ {
		gun := models.Gun{
			Name:           fmt.Sprintf("Test Gun %d", i),
			WeaponTypeID:   weaponType.ID,
			CaliberID:      caliber.ID,
			ManufacturerID: manufacturer.ID,
			OwnerID:        user.ID,
		}
		db.Create(&gun)
	}

	// Set up test router and controller
	router, gunController := payment_test_utils.SetupGunTestRouter(t, db)

	// Set up the route for listing guns
	router.GET("/owner/guns", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		gunController.Index(c)
	})

	// Test accessing the guns list
	req, _ := http.NewRequest("GET", "/owner/guns", nil)
	// Add authentication cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the response is successful
	assert.Equal(t, http.StatusOK, w.Code)

	// Check the response body
	body := w.Body.String()

	// Should contain the first two guns
	assert.Contains(t, body, "Test Gun 1")
	assert.Contains(t, body, "Test Gun 2")

	// Should NOT contain the other guns
	assert.NotContains(t, body, "Test Gun 3")
	assert.NotContains(t, body, "Test Gun 4")
	assert.NotContains(t, body, "Test Gun 5")

	// Should contain a message about having more guns
	assert.Contains(t, body, "You have 3 more guns")
	assert.Contains(t, body, "Please re-subscribe to see all your guns")
}
