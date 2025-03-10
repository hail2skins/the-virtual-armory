package gun_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// Counter for generating unique emails
var emailCounter = 0

// setupGunTest sets up the test environment for gun tests
func setupGunTest(t *testing.T) (*gin.Engine, *controllers.GunController, *models.User) {
	// Setup
	gin.SetMode(gin.TestMode)
	db, err := testutils.SetupTestDB()
	assert.NoError(t, err)

	// Set the global database variable to the test database
	database.DB = db

	// Create a test user with a unique email
	emailCounter++
	email := fmt.Sprintf("test%d@example.com", emailCounter)
	user, err := testutils.CreateTestUser(db, email, "password123", false)
	assert.NoError(t, err)

	// Set the mock user for authentication
	auth.MockUser = user

	// Create a gun controller
	gunController := controllers.NewGunController(db)

	// Create a test router
	router := gin.Default()

	return router, gunController, user
}

// Cleanup function to reset MockUser after each test
func cleanup() {
	auth.MockUser = nil
}

// createTestWeaponType creates a test weapon type in the database
func createTestWeaponType(t *testing.T) *models.WeaponType {
	weaponType := models.WeaponType{
		Type: "Rifle",
	}
	database.DB.Create(&weaponType)
	return &weaponType
}

// createTestCaliber creates a test caliber in the database
func createTestCaliber(t *testing.T) *models.Caliber {
	caliber := models.Caliber{
		Caliber:  "9mm",
		Nickname: "Nine",
	}
	database.DB.Create(&caliber)
	return &caliber
}

// createTestManufacturer creates a test manufacturer in the database
func createTestManufacturer(t *testing.T) *models.Manufacturer {
	manufacturer := models.Manufacturer{
		Name: "Test Manufacturer",
	}
	database.DB.Create(&manufacturer)
	return &manufacturer
}

func TestGunIndex(t *testing.T) {
	// Setup
	router, gunController, user := setupGunTest(t)
	defer cleanup()

	// Create a weapon type, caliber, and manufacturer for the test
	weaponType := createTestWeaponType(t)
	caliber := createTestCaliber(t)
	manufacturer := createTestManufacturer(t)

	// Create test guns for the user
	gun1 := models.Gun{
		Name:           "Test Gun 1",
		WeaponTypeID:   weaponType.ID,
		CaliberID:      caliber.ID,
		ManufacturerID: manufacturer.ID,
		OwnerID:        user.ID,
	}
	assert.NoError(t, models.CreateGun(database.DB, &gun1))

	gun2 := models.Gun{
		Name:           "Test Gun 2",
		WeaponTypeID:   weaponType.ID,
		CaliberID:      caliber.ID,
		ManufacturerID: manufacturer.ID,
		OwnerID:        user.ID,
	}
	assert.NoError(t, models.CreateGun(database.DB, &gun2))

	// Setup the route
	router.GET("/owner/guns", gunController.Index)

	// Create a request
	req, err := http.NewRequest("GET", "/owner/guns", nil)
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains the gun names
	assert.Contains(t, w.Body.String(), "Test Gun 1")
	assert.Contains(t, w.Body.String(), "Test Gun 2")
}

func TestGunShow(t *testing.T) {
	// Setup
	router, gunController, user := setupGunTest(t)
	defer cleanup()

	// Create a weapon type, caliber, and manufacturer for the test
	weaponType := createTestWeaponType(t)
	caliber := createTestCaliber(t)
	manufacturer := createTestManufacturer(t)

	// Create a test gun for the user
	gun := models.Gun{
		Name:           "Test Gun",
		WeaponTypeID:   weaponType.ID,
		CaliberID:      caliber.ID,
		ManufacturerID: manufacturer.ID,
		OwnerID:        user.ID,
	}
	assert.NoError(t, models.CreateGun(database.DB, &gun))

	// Setup the route
	router.GET("/owner/guns/:id", gunController.Show)

	// Create a request
	req, err := http.NewRequest("GET", "/owner/guns/"+strconv.FormatUint(uint64(gun.ID), 10), nil)
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains the gun name
	assert.Contains(t, w.Body.String(), "Test Gun")
}

func TestGunNew(t *testing.T) {
	// Setup
	router, gunController, _ := setupGunTest(t)
	defer cleanup()

	// Create test weapon types, calibers, and manufacturers
	createTestWeaponType(t)
	createTestCaliber(t)
	createTestManufacturer(t)

	// Setup the route
	router.GET("/owner/guns/new", gunController.New)

	// Create a request
	req, err := http.NewRequest("GET", "/owner/guns/new", nil)
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Since we're using HTML templates, we can't easily check the content
	// Just verify that the response is not empty
	assert.NotEmpty(t, w.Body.String())
}

func TestGunCreate(t *testing.T) {
	// Setup
	router, gunController, user := setupGunTest(t)
	defer cleanup()

	// Create a weapon type, caliber, and manufacturer for the test
	weaponType := createTestWeaponType(t)
	caliber := createTestCaliber(t)
	manufacturer := createTestManufacturer(t)

	// Setup the route
	router.POST("/owner/guns", gunController.Create)

	// Create form data
	form := url.Values{}
	form.Add("name", "New Test Gun")
	form.Add("weapon_type_id", strconv.FormatUint(uint64(weaponType.ID), 10))
	form.Add("caliber_id", strconv.FormatUint(uint64(caliber.ID), 10))
	form.Add("manufacturer_id", strconv.FormatUint(uint64(manufacturer.ID), 10))
	form.Add("acquired", time.Now().Format("2006-01-02"))

	// Create a request
	req, err := http.NewRequest("POST", "/owner/guns", strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code (should be a redirect)
	assert.Equal(t, http.StatusSeeOther, w.Code)

	// Check the redirect location
	assert.Equal(t, "/owner/guns", w.Header().Get("Location"))

	// Verify the gun was created in the database
	var guns []models.Gun
	assert.NoError(t, database.DB.Where("owner_id = ?", user.ID).Find(&guns).Error)
	assert.Equal(t, 1, len(guns))
	assert.Equal(t, "New Test Gun", guns[0].Name)
}

func TestGunEdit(t *testing.T) {
	// Setup
	router, gunController, user := setupGunTest(t)
	defer cleanup()

	// Create a weapon type, caliber, and manufacturer for the test
	weaponType := createTestWeaponType(t)
	caliber := createTestCaliber(t)
	manufacturer := createTestManufacturer(t)

	// Create a test gun for the user
	gun := models.Gun{
		Name:           "Test Gun",
		WeaponTypeID:   weaponType.ID,
		CaliberID:      caliber.ID,
		ManufacturerID: manufacturer.ID,
		OwnerID:        user.ID,
	}
	assert.NoError(t, models.CreateGun(database.DB, &gun))

	// Setup the route
	router.GET("/owner/guns/:id/edit", gunController.Edit)

	// Create a request
	req, err := http.NewRequest("GET", "/owner/guns/"+strconv.FormatUint(uint64(gun.ID), 10)+"/edit", nil)
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains the gun name and form elements
	assert.Contains(t, w.Body.String(), "Test Gun")
	assert.Contains(t, w.Body.String(), "Edit Gun")
	assert.Contains(t, w.Body.String(), "form")
}

func TestGunUpdate(t *testing.T) {
	// Setup
	router, gunController, user := setupGunTest(t)
	defer cleanup()

	// Create a weapon type, caliber, and manufacturer for the test
	weaponType := createTestWeaponType(t)
	caliber := createTestCaliber(t)
	manufacturer := createTestManufacturer(t)

	// Create a test gun for the user
	gun := models.Gun{
		Name:           "Test Gun",
		WeaponTypeID:   weaponType.ID,
		CaliberID:      caliber.ID,
		ManufacturerID: manufacturer.ID,
		OwnerID:        user.ID,
	}
	assert.NoError(t, models.CreateGun(database.DB, &gun))

	// Setup the route
	router.POST("/owner/guns/:id", gunController.Update)

	// Create form data
	form := url.Values{}
	form.Add("name", "Updated Test Gun")
	form.Add("weapon_type_id", strconv.FormatUint(uint64(weaponType.ID), 10))
	form.Add("caliber_id", strconv.FormatUint(uint64(caliber.ID), 10))
	form.Add("manufacturer_id", strconv.FormatUint(uint64(manufacturer.ID), 10))
	form.Add("acquired", time.Now().Format("2006-01-02"))

	// Create a request
	req, err := http.NewRequest("POST", "/owner/guns/"+strconv.FormatUint(uint64(gun.ID), 10), strings.NewReader(form.Encode()))
	assert.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code (should be a redirect)
	assert.Equal(t, http.StatusSeeOther, w.Code)

	// Check the redirect location
	assert.Equal(t, "/owner/guns/"+strconv.FormatUint(uint64(gun.ID), 10), w.Header().Get("Location"))

	// Verify the gun was updated in the database
	var updatedGun models.Gun
	assert.NoError(t, database.DB.First(&updatedGun, gun.ID).Error)
	assert.Equal(t, "Updated Test Gun", updatedGun.Name)
}

func TestGunDelete(t *testing.T) {
	// Setup
	router, gunController, user := setupGunTest(t)
	defer cleanup()

	// Create a weapon type, caliber, and manufacturer for the test
	weaponType := createTestWeaponType(t)
	caliber := createTestCaliber(t)
	manufacturer := createTestManufacturer(t)

	// Create a test gun for the user
	gun := models.Gun{
		Name:           "Test Gun",
		WeaponTypeID:   weaponType.ID,
		CaliberID:      caliber.ID,
		ManufacturerID: manufacturer.ID,
		OwnerID:        user.ID,
	}
	assert.NoError(t, models.CreateGun(database.DB, &gun))

	// Setup the route
	router.POST("/owner/guns/:id/delete", gunController.Delete)

	// Create a request
	req, err := http.NewRequest("POST", "/owner/guns/"+strconv.FormatUint(uint64(gun.ID), 10)+"/delete", nil)
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code (should be a redirect)
	assert.Equal(t, http.StatusSeeOther, w.Code)

	// Check the redirect location
	assert.Equal(t, "/owner/guns", w.Header().Get("Location"))

	// Verify the gun was deleted from the database
	var count int64
	database.DB.Model(&models.Gun{}).Where("id = ?", gun.ID).Count(&count)
	assert.Equal(t, int64(0), count)
}

func TestSearchCalibers(t *testing.T) {
	// Setup
	router, gunController, _ := setupGunTest(t)
	defer cleanup()

	// Create test calibers
	caliber1 := models.Caliber{
		Caliber: "9mm Parabellum",
	}
	assert.NoError(t, database.DB.Create(&caliber1).Error)

	caliber2 := models.Caliber{
		Caliber: "45 ACP",
	}
	assert.NoError(t, database.DB.Create(&caliber2).Error)

	// Setup the route
	router.GET("/api/calibers/search", gunController.SearchCalibers)

	// Test case 1: Search with query
	req1, err := http.NewRequest("GET", "/api/calibers/search?q=9mm", nil)
	assert.NoError(t, err)

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	assert.Equal(t, http.StatusOK, w1.Code)

	// Check that the response is not empty
	assert.NotEmpty(t, w1.Body.String())

	// Test case 2: Search with empty query (should return popular calibers)
	req2, err := http.NewRequest("GET", "/api/calibers/search?q=", nil)
	assert.NoError(t, err)

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)

	// Check that the response is not empty
	assert.NotEmpty(t, w2.Body.String())
}
