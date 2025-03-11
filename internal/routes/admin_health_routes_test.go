package routes

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupAdminHealthTestRouter creates a test router with the admin health routes registered
func setupAdminHealthTestRouter(t *testing.T) (*gin.Engine, *gorm.DB) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	r := gin.New()
	r.Use(gin.Recovery())

	// Setup test database
	db, err := testutils.SetupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	// Create real auth instance
	authInstance, err := auth.New()
	require.NoError(t, err, "Failed to create auth instance")

	// Register the admin detailed health route with real middleware
	adminGroup := r.Group("/admin")
	adminGroup.Use(authInstance.RequireAuth())
	adminGroup.Use(authInstance.RequireAdmin())

	// Register admin health routes
	adminGroup.GET("/detailed-health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":   "ok",
			"database": "connected",
			"system":   map[string]interface{}{"memory_usage_mb": 100},
		})
	})

	return r, db
}

// TestAdminDetailedHealthEndpointWithGuestUser tests that a guest user is redirected to login with a flash message
func TestAdminDetailedHealthEndpointWithGuestUser(t *testing.T) {
	r, db := setupAdminHealthTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create a test HTTP request without authentication
	req, err := http.NewRequest("GET", "/admin/detailed-health", nil)
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

// TestAdminDetailedHealthEndpointWithRegularUser tests that a regular user is redirected to owner page with an admin required message
func TestAdminDetailedHealthEndpointWithRegularUser(t *testing.T) {
	r, db := setupAdminHealthTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create a real regular user
	regularUser, err := testutils.CreateTestUser(db, "regular@example.com", "password123", false)
	require.NoError(t, err, "Failed to create regular user")

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/admin/detailed-health", nil)
	require.NoError(t, err, "Failed to create request")

	// Set up session for the regular user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: regularUser.Email,
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

// TestAdminDetailedHealthEndpointWithAdminUser tests that an admin user can access the endpoint
func TestAdminDetailedHealthEndpointWithAdminUser(t *testing.T) {
	r, db := setupAdminHealthTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create a real admin user
	adminUser, err := testutils.CreateTestUser(db, "admin@example.com", "password123", true)
	require.NoError(t, err, "Failed to create admin user")

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/admin/detailed-health", nil)
	require.NoError(t, err, "Failed to create request")

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: adminUser.Email,
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
	assert.Contains(t, rr.Body.String(), "status", "Response should contain status")
	assert.Contains(t, rr.Body.String(), "database", "Response should contain database status")
}
