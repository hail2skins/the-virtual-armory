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

// setupCaliberTest sets up the test environment for caliber tests
func setupCaliberTest(t *testing.T) (*gin.Engine, *CaliberController) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutils.SetupTestDB()

	// Create a test admin user
	_, _ = testutils.CreateTestUser(database.TestDB, "admin@example.com", "password", true)

	// Set the mock user for authentication
	adminUser, _ := testutils.GetTestUser(database.TestDB, "admin@example.com")
	auth.MockUser = adminUser

	// Create a caliber controller
	caliberController := NewCaliberController()

	// Create a test router
	router := gin.Default()

	return router, caliberController
}

// createTestCaliber creates a test caliber in the database
func createTestCaliber(t *testing.T) *models.Caliber {
	caliber := models.Caliber{
		Caliber:  "9mm",
		Nickname: "Nine",
	}
	database.TestDB.Create(&caliber)
	return &caliber
}

// TestCaliberIndex tests the Index method
func TestCaliberIndex(t *testing.T) {
	// Setup
	router, caliberController := setupCaliberTest(t)

	// Create some test calibers
	createTestCaliber(t)
	createTestCaliber(t)

	// Setup routes
	router.GET("/admin/calibers", caliberController.Index)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/calibers", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "9mm")
	assert.Contains(t, w.Body.String(), "Nine")
}

// TestCaliberShow tests the Show method
func TestCaliberShow(t *testing.T) {
	// Setup
	router, caliberController := setupCaliberTest(t)

	// Create a test caliber with a unique name
	caliber := models.Caliber{
		Caliber:  "9mm Test Show",
		Nickname: "Nine Show",
	}
	result := database.TestDB.Create(&caliber)
	assert.NoError(t, result.Error)
	assert.NotZero(t, caliber.ID)

	// Setup routes
	router.GET("/admin/calibers/:id", caliberController.Show)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/calibers/"+strconv.Itoa(int(caliber.ID)), nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "9mm Test Show")
	assert.Contains(t, w.Body.String(), "Nine Show")
}

// TestCaliberNew tests the New method
func TestCaliberNew(t *testing.T) {
	// Setup
	router, caliberController := setupCaliberTest(t)

	// Setup routes
	router.GET("/admin/calibers/new", caliberController.New)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/calibers/new", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "New Caliber")
	assert.Contains(t, w.Body.String(), "form")
}

// TestCaliberCreate tests the Create method
func TestCaliberCreate(t *testing.T) {
	// Setup
	router, caliberController := setupCaliberTest(t)

	// Setup routes
	router.POST("/admin/calibers", caliberController.Create)

	// Create form data
	form := url.Values{}
	form.Add("caliber", "45 ACP")
	form.Add("nickname", "Forty-Five")

	// Create a request
	req, _ := http.NewRequest("POST", "/admin/calibers", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the calibers index
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/admin/calibers", w.Header().Get("Location"))

	// Check that the caliber was created in the database
	var caliber models.Caliber
	result := database.TestDB.Where("caliber = ?", "45 ACP").First(&caliber)
	assert.NoError(t, result.Error)
	assert.Equal(t, "45 ACP", caliber.Caliber)
	assert.Equal(t, "Forty-Five", caliber.Nickname)
}

// TestCaliberEdit tests the Edit method
func TestCaliberEdit(t *testing.T) {
	// Setup
	router, caliberController := setupCaliberTest(t)

	// Create a test caliber with a unique name
	caliber := models.Caliber{
		Caliber:  "9mm Test Edit",
		Nickname: "Nine Edit",
	}
	result := database.TestDB.Create(&caliber)
	assert.NoError(t, result.Error)
	assert.NotZero(t, caliber.ID)

	// Setup routes
	router.GET("/admin/calibers/:id/edit", caliberController.Edit)

	// Create a request
	req, _ := http.NewRequest("GET", "/admin/calibers/"+strconv.Itoa(int(caliber.ID))+"/edit", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Edit Caliber")
	assert.Contains(t, w.Body.String(), "9mm Test Edit")
	assert.Contains(t, w.Body.String(), "Nine Edit")
}

// TestCaliberUpdate tests the Update method
func TestCaliberUpdate(t *testing.T) {
	// Setup
	router, caliberController := setupCaliberTest(t)

	// Create a test caliber with a unique name
	caliber := models.Caliber{
		Caliber:  "9mm Test Update",
		Nickname: "Nine Update",
	}
	result := database.TestDB.Create(&caliber)
	assert.NoError(t, result.Error)
	assert.NotZero(t, caliber.ID)

	// Setup routes
	router.POST("/admin/calibers/:id", caliberController.Update)

	// Create form data
	form := url.Values{}
	form.Add("caliber", "9mm Luger Update")
	form.Add("nickname", "Nine Luger Update")

	// Create a request
	req, _ := http.NewRequest("POST", "/admin/calibers/"+strconv.Itoa(int(caliber.ID)), strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the calibers index
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/admin/calibers", w.Header().Get("Location"))

	// Check that the caliber was updated in the database
	var updatedCaliber models.Caliber
	result = database.TestDB.First(&updatedCaliber, caliber.ID)
	assert.NoError(t, result.Error)
	assert.Equal(t, "9mm Luger Update", updatedCaliber.Caliber)
	assert.Equal(t, "Nine Luger Update", updatedCaliber.Nickname)
}

// TestCaliberDelete tests the Delete method
func TestCaliberDelete(t *testing.T) {
	// Setup
	router, caliberController := setupCaliberTest(t)

	// Create a test caliber with a unique name
	caliber := models.Caliber{
		Caliber:  "9mm Test Delete",
		Nickname: "Nine Delete",
	}
	result := database.TestDB.Create(&caliber)
	assert.NoError(t, result.Error)
	assert.NotZero(t, caliber.ID)

	// Setup routes
	router.DELETE("/admin/calibers/:id", caliberController.Delete)

	// Create a request
	req, _ := http.NewRequest("DELETE", "/admin/calibers/"+strconv.Itoa(int(caliber.ID)), nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the calibers index
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/admin/calibers", w.Header().Get("Location"))

	// Check that the caliber was deleted from the database
	var deletedCaliber models.Caliber
	result = database.TestDB.First(&deletedCaliber, caliber.ID)
	assert.Error(t, result.Error) // Should return an error because the record is deleted
}

// TestCaliberDeleteAlternative tests the alternative Delete method (POST)
func TestCaliberDeleteAlternative(t *testing.T) {
	// Setup
	router, caliberController := setupCaliberTest(t)

	// Create a test caliber with a unique name
	caliber := models.Caliber{
		Caliber:  "9mm Test Delete Alt",
		Nickname: "Nine Delete Alt",
	}
	result := database.TestDB.Create(&caliber)
	assert.NoError(t, result.Error)
	assert.NotZero(t, caliber.ID)

	// Setup routes
	router.POST("/admin/calibers/:id/delete", caliberController.Delete)

	// Create a request
	req, _ := http.NewRequest("POST", "/admin/calibers/"+strconv.Itoa(int(caliber.ID))+"/delete", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the calibers index
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/admin/calibers", w.Header().Get("Location"))

	// Check that the caliber was deleted from the database
	var deletedCaliber models.Caliber
	result = database.TestDB.First(&deletedCaliber, caliber.ID)
	assert.Error(t, result.Error) // Should return an error because the record is deleted
}
