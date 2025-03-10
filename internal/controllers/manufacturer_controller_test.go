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

// setupManufacturerTest sets up the test environment for manufacturer tests
func setupManufacturerTest(t *testing.T) (*gin.Engine, *ManufacturerController) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutils.SetupTestDB()

	// Create a test admin user
	_, _ = testutils.CreateTestUser(database.TestDB, "admin@example.com", "password", true)

	// Set the mock user for authentication
	adminUser, _ := testutils.GetTestUser(database.TestDB, "admin@example.com")
	auth.MockUser = adminUser

	// Create a manufacturer controller
	manufacturerController := NewManufacturerController()

	// Create a test router
	router := gin.Default()

	return router, manufacturerController
}

// createTestManufacturer creates a test manufacturer in the database
func createTestManufacturer(t *testing.T, name, nickname, country string) *models.Manufacturer {
	manufacturer := models.Manufacturer{
		Name:     name,
		Nickname: nickname,
		Country:  country,
	}
	result := database.TestDB.Create(&manufacturer)
	assert.NoError(t, result.Error)
	assert.NotZero(t, manufacturer.ID)
	return &manufacturer
}

// TestManufacturerIndex tests the Index method
func TestManufacturerIndex(t *testing.T) {
	// Setup
	router, manufacturerController := setupManufacturerTest(t)

	// Create some test manufacturers
	createTestManufacturer(t, "Glock", "Austrian Perfection", "Austria")
	createTestManufacturer(t, "Smith & Wesson", "S&W", "USA")

	// Setup routes
	router.GET("/admin/manufacturers", manufacturerController.Index)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/manufacturers", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// The HTML response might have escaped characters, so check for parts of the names
	assert.Contains(t, w.Body.String(), "Glock")
	assert.Contains(t, w.Body.String(), "Austria")
	assert.Contains(t, w.Body.String(), "USA")
}

// TestManufacturerShow tests the Show method
func TestManufacturerShow(t *testing.T) {
	// Setup
	router, manufacturerController := setupManufacturerTest(t)

	// Create a test manufacturer
	manufacturer := createTestManufacturer(t, "Beretta", "Italian Classic", "Italy")

	// Setup routes
	router.GET("/admin/manufacturers/:id", manufacturerController.Show)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/manufacturers/"+strconv.Itoa(int(manufacturer.ID)), nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Beretta")
	assert.Contains(t, w.Body.String(), "Italian Classic")
	assert.Contains(t, w.Body.String(), "Italy")
}

// TestManufacturerNew tests the New method
func TestManufacturerNew(t *testing.T) {
	// Setup
	router, manufacturerController := setupManufacturerTest(t)

	// Setup routes
	router.GET("/admin/manufacturers/new", manufacturerController.New)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/manufacturers/new", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "New Manufacturer")
	assert.Contains(t, w.Body.String(), "form")
}

// TestManufacturerCreate tests the Create method
func TestManufacturerCreate(t *testing.T) {
	// Setup
	router, manufacturerController := setupManufacturerTest(t)

	// Setup routes
	router.POST("/admin/manufacturers", manufacturerController.Create)

	// Create form data
	form := url.Values{}
	form.Add("name", "Sig Sauer")
	form.Add("nickname", "Sig")
	form.Add("country", "Switzerland")

	// Create a request
	req, _ := http.NewRequest("POST", "/admin/manufacturers", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the manufacturers index
	assert.Equal(t, http.StatusFound, w.Code) // 302 Found is the actual status code
	assert.Equal(t, "/admin/manufacturers", w.Header().Get("Location"))

	// Check that the manufacturer was created in the database
	var manufacturer models.Manufacturer
	result := database.TestDB.Where("name = ?", "Sig Sauer").First(&manufacturer)
	assert.NoError(t, result.Error)
	assert.Equal(t, "Sig Sauer", manufacturer.Name)
	assert.Equal(t, "Sig", manufacturer.Nickname)
	assert.Equal(t, "Switzerland", manufacturer.Country)
}

// TestManufacturerEdit tests the Edit method
func TestManufacturerEdit(t *testing.T) {
	// Setup
	router, manufacturerController := setupManufacturerTest(t)

	// Create a test manufacturer
	manufacturer := createTestManufacturer(t, "Heckler & Koch", "H&K", "Germany")

	// Setup routes
	router.GET("/admin/manufacturers/:id/edit", manufacturerController.Edit)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/manufacturers/"+strconv.Itoa(int(manufacturer.ID))+"/edit", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// The response contains the form with the manufacturer data
	// HTML might have escaped characters, so check for parts of the content
	body := w.Body.String()
	assert.Contains(t, body, "Edit Manufacturer")
	assert.Contains(t, body, "Germany")
}

// TestManufacturerUpdate tests the Update method
func TestManufacturerUpdate(t *testing.T) {
	// Setup
	router, manufacturerController := setupManufacturerTest(t)

	// Create a test manufacturer
	manufacturer := createTestManufacturer(t, "FN Herstal", "FN", "Belgium")

	// Setup routes
	router.POST("/admin/manufacturers/:id", manufacturerController.Update)

	// Create form data
	form := url.Values{}
	form.Add("name", "FN Herstal Updated")
	form.Add("nickname", "FN Updated")
	form.Add("country", "Belgium Updated")

	// Create a request
	req, _ := http.NewRequest("POST", "/admin/manufacturers/"+strconv.Itoa(int(manufacturer.ID)), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the manufacturer details page
	assert.Equal(t, http.StatusFound, w.Code) // 302 Found is the actual status code
	assert.Equal(t, "/admin/manufacturers/"+strconv.Itoa(int(manufacturer.ID)), w.Header().Get("Location"))

	// Check that the manufacturer was updated in the database
	var updatedManufacturer models.Manufacturer
	result := database.TestDB.First(&updatedManufacturer, manufacturer.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, "FN Herstal Updated", updatedManufacturer.Name)
	assert.Equal(t, "FN Updated", updatedManufacturer.Nickname)
	assert.Equal(t, "Belgium Updated", updatedManufacturer.Country)
}

// TestManufacturerDelete tests the Delete method
func TestManufacturerDelete(t *testing.T) {
	// Setup
	router, manufacturerController := setupManufacturerTest(t)

	// Create a test manufacturer
	manufacturer := createTestManufacturer(t, "Colt", "American Classic", "USA")

	// Setup routes
	router.DELETE("/admin/manufacturers/:id", manufacturerController.Delete)

	// Create a request
	req, _ := http.NewRequest("DELETE", "/admin/manufacturers/"+strconv.Itoa(int(manufacturer.ID)), nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the manufacturers index
	assert.Equal(t, http.StatusFound, w.Code) // 302 Found is the actual status code
	assert.Equal(t, "/admin/manufacturers", w.Header().Get("Location"))

	// Check that the manufacturer was deleted from the database
	var deletedManufacturer models.Manufacturer
	result := database.TestDB.First(&deletedManufacturer, manufacturer.ID)
	assert.Error(t, result.Error) // Should return an error because the record is deleted
}

// TestManufacturerDeleteAlternative tests the alternative Delete method (POST)
func TestManufacturerDeleteAlternative(t *testing.T) {
	// Setup
	router, manufacturerController := setupManufacturerTest(t)

	// Create a test manufacturer
	manufacturer := createTestManufacturer(t, "Ruger", "American Value", "USA")

	// Setup routes
	router.POST("/admin/manufacturers/:id/delete", manufacturerController.Delete)

	// Create a request
	req, _ := http.NewRequest("POST", "/admin/manufacturers/"+strconv.Itoa(int(manufacturer.ID))+"/delete", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the manufacturers index
	assert.Equal(t, http.StatusFound, w.Code) // 302 Found is the actual status code
	assert.Equal(t, "/admin/manufacturers", w.Header().Get("Location"))

	// Check that the manufacturer was deleted from the database
	var deletedManufacturer models.Manufacturer
	result := database.TestDB.First(&deletedManufacturer, manufacturer.ID)
	assert.Error(t, result.Error) // Should return an error because the record is deleted
}
