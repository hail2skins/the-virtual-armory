package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupAdminErrorMetricsTestRouter creates a test router with the admin error metrics routes registered
func setupAdminErrorMetricsTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, *testutils.TestUsers) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	r := gin.New()
	r.Use(gin.Recovery())

	// Setup test database
	db, err := testutils.SetupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	// Create test users
	testUsers := testutils.CreateTestUsers()

	// Create real auth instance
	authInstance, err := auth.New()
	require.NoError(t, err, "Failed to create auth instance")

	// Create admin controller
	adminController := controllers.NewAdminController()

	// Register the admin error metrics route with real middleware
	adminGroup := r.Group("/admin")
	adminGroup.Use(authInstance.RequireAuth())
	adminGroup.Use(authInstance.RequireAdmin())

	// Register admin error metrics route
	adminGroup.GET("/error-metrics", adminController.ErrorMetrics)

	// Save users to database
	err = db.Create(&testUsers.Admin).Error
	require.NoError(t, err, "Failed to create admin user")
	err = db.Create(&testUsers.Unsubscribed).Error
	require.NoError(t, err, "Failed to create regular user")

	return r, db, testUsers
}

// TestAdminErrorMetricsEndpointWithGuestUser tests that a guest user is redirected to login with a flash message
func TestAdminErrorMetricsEndpointWithGuestUser(t *testing.T) {
	r, db, _ := setupAdminErrorMetricsTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create a test HTTP request without authentication
	req, err := http.NewRequest("GET", "/admin/error-metrics", nil)
	require.NoError(t, err, "Failed to create request")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Check that the guest user is redirected to login
	assert.Equal(t, http.StatusFound, rr.Code, "Guest user should be redirected")
	assert.Equal(t, "/login", rr.Header().Get("Location"), "Guest user should be redirected to login")

	// Check for flash message cookies
	cookies := rr.Result().Cookies()
	var flashMessage, flashType string
	for _, cookie := range cookies {
		if cookie.Name == "flash_message" {
			flashMessage = cookie.Value
			flashMessage = strings.ReplaceAll(flashMessage, "+", " ")
			t.Logf("Flash message: %q", flashMessage) // Debug output
		} else if cookie.Name == "flash_type" {
			flashType = cookie.Value
			t.Logf("Flash type: %q", flashType) // Debug output
		}
	}

	// Assert the exact flash message and type
	assert.NotEmpty(t, flashMessage, "Flash message should be set")
	assert.Equal(t, "You do not have permission to access that page", flashMessage, "Flash message should match exactly")
	assert.Equal(t, "error", flashType, "Flash type should be error")
}

// TestAdminErrorMetricsEndpointWithRegularUser tests that a regular user is redirected to owner page with an admin required message
func TestAdminErrorMetricsEndpointWithRegularUser(t *testing.T) {
	r, db, testUsers := setupAdminErrorMetricsTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/admin/error-metrics", nil)
	require.NoError(t, err, "Failed to create request")

	// Set up session for the regular user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Unsubscribed.Email,
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Check that the regular user is redirected to owner page
	assert.Equal(t, http.StatusFound, rr.Code, "Regular user should be redirected")
	assert.Equal(t, "/owner", rr.Header().Get("Location"), "Regular user should be redirected to owner page")

	// Check for flash message cookies
	cookies := rr.Result().Cookies()
	var flashMessage, flashType string
	for _, cookie := range cookies {
		if cookie.Name == "flash_message" {
			flashMessage = cookie.Value
			flashMessage = strings.ReplaceAll(flashMessage, "+", " ")
			t.Logf("Flash message: %q", flashMessage) // Debug output
		} else if cookie.Name == "flash_type" {
			flashType = cookie.Value
			t.Logf("Flash type: %q", flashType) // Debug output
		}
	}

	// Assert the exact flash message and type
	assert.NotEmpty(t, flashMessage, "Flash message should be set")
	assert.Equal(t, "You must be an administrator to access this page", flashMessage, "Flash message should match exactly")
	assert.Equal(t, "error", flashType, "Flash type should be error")
}

// TestAdminErrorMetricsEndpointWithAdminUser tests that an admin user can access the endpoint
func TestAdminErrorMetricsEndpointWithAdminUser(t *testing.T) {
	r, db, testUsers := setupAdminErrorMetricsTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/admin/error-metrics", nil)
	require.NoError(t, err, "Failed to create request")

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Check that the admin user can access the endpoint
	assert.Equal(t, http.StatusOK, rr.Code, "Admin user should be able to access the endpoint")
	assert.Contains(t, rr.Body.String(), "Error Metrics Dashboard", "Response should contain dashboard title")
	assert.Contains(t, rr.Body.String(), "Error Rates", "Response should contain error rates section")
	assert.Contains(t, rr.Body.String(), "Recent Errors", "Response should contain recent errors section")
}
