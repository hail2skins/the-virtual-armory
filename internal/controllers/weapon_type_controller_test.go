package controllers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// setupWeaponTypeTest sets up the test environment for weapon type tests
func setupWeaponTypeTest(t *testing.T) (*gin.Engine, *WeaponTypeController) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutils.SetupTestDB()

	// Create a test admin user
	_, _ = testutils.CreateTestUser(database.TestDB, "admin@example.com", "password", true)

	// Set the mock user for authentication
	adminUser, _ := testutils.GetTestUser(database.TestDB, "admin@example.com")
	auth.MockUser = adminUser

	// Create a weapon type controller
	weaponTypeController := NewWeaponTypeController(database.TestDB)

	// Create a test router
	router := gin.Default()

	return router, weaponTypeController
}

// createTestWeaponType creates a test weapon type in the database
func createTestWeaponType(t *testing.T, typeName, nickname string) *models.WeaponType {
	weaponType := models.WeaponType{
		Type:     typeName,
		Nickname: nickname,
	}
	result := database.TestDB.Create(&weaponType)
	assert.NoError(t, result.Error)
	assert.NotZero(t, weaponType.ID)
	return &weaponType
}

// TestWeaponTypeIndex tests the Index method
func TestWeaponTypeIndex(t *testing.T) {
	// Setup
	router, weaponTypeController := setupWeaponTypeTest(t)

	// Create some test weapon types
	createTestWeaponType(t, "Pistol", "Handgun")
	createTestWeaponType(t, "Rifle", "Long gun")

	// Setup routes
	router.GET("/admin/weapon-types", weaponTypeController.Index)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/weapon-types", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// The HTML response might have escaped characters, so check for parts of the names
	body := w.Body.String()
	assert.Contains(t, body, "Pistol")
	assert.Contains(t, body, "Rifle")
}

// TestWeaponTypeShow tests the Show method
func TestWeaponTypeShow(t *testing.T) {
	// Setup
	router, weaponTypeController := setupWeaponTypeTest(t)

	// Create a test weapon type
	weaponType := createTestWeaponType(t, "Shotgun", "Scattergun")

	// Setup routes
	router.GET("/admin/weapon-types/:id", weaponTypeController.Show)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/weapon-types/"+strconv.Itoa(int(weaponType.ID)), nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// The response contains the weapon type details
	body := w.Body.String()
	assert.Contains(t, body, "Shotgun")
	assert.Contains(t, body, "Scattergun")
}

// TestWeaponTypeNew tests the New method
func TestWeaponTypeNew(t *testing.T) {
	// Setup
	router, weaponTypeController := setupWeaponTypeTest(t)

	// Setup routes
	router.GET("/admin/weapon-types/new", weaponTypeController.New)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/weapon-types/new", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// The response contains the form for creating a new weapon type
	body := w.Body.String()
	assert.Contains(t, body, "New Weapon Type")
	assert.Contains(t, body, "form")
}

// TestWeaponTypeCreate tests the Create method
func TestWeaponTypeCreate(t *testing.T) {
	// Setup
	router, weaponTypeController := setupWeaponTypeTest(t)

	// Setup routes
	router.POST("/admin/weapon-types", weaponTypeController.Create)

	// Create form data
	form := url.Values{}
	form.Add("type", "Submachine Gun")
	form.Add("nickname", "SMG")

	// Create a request
	req, _ := http.NewRequest("POST", "/admin/weapon-types", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the weapon types index
	assert.Equal(t, http.StatusSeeOther, w.Code) // 303 See Other is the actual status code
	assert.Equal(t, "/admin/weapon-types", w.Header().Get("Location"))

	// Check that the weapon type was created in the database
	var weaponType models.WeaponType
	result := database.TestDB.Where("type = ?", "Submachine Gun").First(&weaponType)
	assert.NoError(t, result.Error)
	assert.Equal(t, "Submachine Gun", weaponType.Type)
	assert.Equal(t, "SMG", weaponType.Nickname)
}

// TestWeaponTypeEdit tests the Edit method
func TestWeaponTypeEdit(t *testing.T) {
	// Setup
	router, weaponTypeController := setupWeaponTypeTest(t)

	// Create a test weapon type
	weaponType := createTestWeaponType(t, "Carbine", "Short rifle")

	// Setup routes
	router.GET("/admin/weapon-types/:id/edit", weaponTypeController.Edit)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/weapon-types/"+strconv.Itoa(int(weaponType.ID))+"/edit", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// The response contains the form with the weapon type data
	body := w.Body.String()
	assert.Contains(t, body, "Edit Weapon Type")
	assert.Contains(t, body, "Carbine")
}

// TestWeaponTypeUpdate tests the Update method
func TestWeaponTypeUpdate(t *testing.T) {
	// Setup
	router, weaponTypeController := setupWeaponTypeTest(t)

	// Create a test weapon type
	weaponType := createTestWeaponType(t, "Revolver", "Wheel gun")

	// Setup routes
	router.POST("/admin/weapon-types/:id", weaponTypeController.Update)

	// Create form data
	form := url.Values{}
	form.Add("type", "Revolver Updated")
	form.Add("nickname", "Wheel gun Updated")

	// Create a request
	req, _ := http.NewRequest("POST", "/admin/weapon-types/"+strconv.Itoa(int(weaponType.ID)), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the weapon type details page
	assert.Equal(t, http.StatusSeeOther, w.Code) // 303 See Other is the actual status code
	assert.Equal(t, "/admin/weapon-types/"+strconv.Itoa(int(weaponType.ID)), w.Header().Get("Location"))

	// Check that the weapon type was updated in the database
	var updatedWeaponType models.WeaponType
	result := database.TestDB.First(&updatedWeaponType, weaponType.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, "Revolver Updated", updatedWeaponType.Type)
	assert.Equal(t, "Wheel gun Updated", updatedWeaponType.Nickname)
}

// TestWeaponTypeDelete tests the Delete method
func TestWeaponTypeDelete(t *testing.T) {
	// Setup
	router, weaponTypeController := setupWeaponTypeTest(t)

	// Create a test weapon type
	weaponType := createTestWeaponType(t, "Machine Gun", "MG")

	// Setup routes
	router.DELETE("/admin/weapon-types/:id", weaponTypeController.Delete)

	// Create a request
	req, _ := http.NewRequest("DELETE", "/admin/weapon-types/"+strconv.Itoa(int(weaponType.ID)), nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the weapon types index
	assert.Equal(t, http.StatusSeeOther, w.Code) // 303 See Other is the actual status code
	assert.Equal(t, "/admin/weapon-types", w.Header().Get("Location"))

	// Check that the weapon type was deleted from the database
	var deletedWeaponType models.WeaponType
	result := database.TestDB.First(&deletedWeaponType, weaponType.ID)
	assert.Error(t, result.Error) // Should return an error because the record is deleted
}

// TestWeaponTypeDeleteAlternative tests the alternative Delete method (POST)
func TestWeaponTypeDeleteAlternative(t *testing.T) {
	// Setup
	router, weaponTypeController := setupWeaponTypeTest(t)

	// Create a test weapon type
	weaponType := createTestWeaponType(t, "Assault Rifle", "AR")

	// Setup routes
	router.POST("/admin/weapon-types/:id/delete", weaponTypeController.Delete)

	// Create a request
	req, _ := http.NewRequest("POST", "/admin/weapon-types/"+strconv.Itoa(int(weaponType.ID))+"/delete", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the weapon types index
	assert.Equal(t, http.StatusSeeOther, w.Code) // 303 See Other is the actual status code
	assert.Equal(t, "/admin/weapon-types", w.Header().Get("Location"))

	// Check that the weapon type was deleted from the database
	var deletedWeaponType models.WeaponType
	result := database.TestDB.First(&deletedWeaponType, weaponType.ID)
	assert.Error(t, result.Error) // Should return an error because the record is deleted
}
