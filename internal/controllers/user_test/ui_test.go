package user_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/config"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// testRenderer is a mock HTML renderer for testing template names
type testRenderer struct {
	captureTemplateName func(name string, data interface{})
}

// Instance implements the HTMLRender interface
func (r *testRenderer) Instance(name string, data interface{}) render.Render {
	if r.captureTemplateName != nil {
		r.captureTemplateName(name, data)
	}
	return &mockRender{data: data}
}

// TestProfilePageUI tests that the profile page has the necessary UI elements
func TestProfilePageUI(t *testing.T) {
	// Set up test database with a test user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set the test database for this test
	database.TestDB = db

	// Set up router with user controller
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create the user controller
	userController := controllers.NewUserController(db)

	// Set up auth middleware and cookies
	router.Use(func(c *gin.Context) {
		// Set cookies to simulate authentication
		c.Request.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
		c.Request.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
		c.Next()
	})

	// Register routes
	router.GET("/profile", userController.ShowProfile)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile", nil)

	// Add cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains HTML content
	body := w.Body.String()
	assert.Contains(t, body, "<!doctype html>")
	assert.Contains(t, body, user.Email)
}

// TestDeleteAccountPageUI tests that the delete account page has the necessary UI elements
func TestDeleteAccountPageUI(t *testing.T) {
	// Set up test database with a test user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set the test database for this test
	database.TestDB = db

	// Set up router with user controller
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create the user controller
	userController := controllers.NewUserController(db)

	// Set up auth middleware and cookies
	router.Use(func(c *gin.Context) {
		// Set cookies to simulate authentication
		c.Request.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
		c.Request.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
		c.Next()
	})

	// Register routes
	router.GET("/profile/delete", userController.ShowDeleteAccount)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile/delete", nil)

	// Add cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains HTML content
	body := w.Body.String()
	assert.Contains(t, body, "<!doctype html>")
	assert.Contains(t, body, "Delete Your Account")
}

// TestEditProfilePageUI tests that the edit profile page has the necessary UI elements
func TestEditProfilePageUI(t *testing.T) {
	// Set up test database with a test user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set the test database for this test
	database.TestDB = db

	// Set up router with user controller
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create the user controller
	userController := controllers.NewUserController(db)

	// Set up auth middleware and cookies
	router.Use(func(c *gin.Context) {
		// Set cookies to simulate authentication
		c.Request.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
		c.Request.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
		c.Next()
	})

	// Register routes
	router.GET("/profile/edit", userController.EditProfile)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile/edit", nil)

	// Add cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains the expected elements
	body := w.Body.String()
	assert.Contains(t, body, "Edit Profile")
	assert.Contains(t, body, "Email Address")
	assert.Contains(t, body, "Password")
	assert.Contains(t, body, "Reset Password")
	assert.Contains(t, body, "Save Changes")
}

// TestReactivateAccountPageUI tests that the reactivate account page has the necessary UI elements
func TestReactivateAccountPageUI(t *testing.T) {
	// Skip this test for now until we can properly mock the templ rendering
	t.Skip("Skipping reactivation test until we can properly mock the templ rendering")

	// Set up test database with a soft-deleted user
	db, err := testutils.SetupTestDB()
	assert.NoError(t, err)
	defer testutils.CleanupTestDB(db)

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	// Create a test user
	user := models.User{
		Email:            "deleted@example.com",
		Password:         string(hashedPassword),
		SubscriptionTier: "free",
		Confirmed:        true,
	}
	err = db.Create(&user).Error
	assert.NoError(t, err)

	// Soft-delete the user
	err = db.Delete(&user).Error
	assert.NoError(t, err)

	// Set the test database for this test
	database.TestDB = db

	// Set up router with auth controller
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create the auth controller
	authController := controllers.NewAuthController(&auth.Auth{}, nil, &config.Config{})

	// Register routes
	router.GET("/reactivate", authController.ReactivateAccount)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/reactivate?email=deleted@example.com", nil)

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// For this test, we're just checking that the handler doesn't crash
	// The actual HTML content is handled by the template engine in the real app
}

// TestAuthProfilePageUI tests that the auth profile page has a link to the user profile management
func TestAuthProfilePageUI(t *testing.T) {
	// Set up test database with a test user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set the test database for this test
	database.TestDB = db

	// Set up router with auth controller
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create the auth controller
	authController := controllers.NewAuthController(&auth.Auth{}, nil, &config.Config{})

	// Set up auth middleware and cookies
	router.Use(func(c *gin.Context) {
		// Set cookies to simulate authentication
		c.Request.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
		c.Request.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
		c.Next()
	})

	// Register routes
	router.GET("/owner", authController.Profile)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/owner", nil)

	// Add cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Get the response body
	body := w.Body.String()

	// Check for a single "Manage Profile" link
	assert.Contains(t, body, "Manage Profile")
	assert.Contains(t, body, "href=\"/profile\"")

	// Should NOT contain direct links to these pages from the owner page
	assert.NotContains(t, body, "href=\"/profile/edit\"")
	assert.NotContains(t, body, "href=\"/profile/subscription\"")
	assert.NotContains(t, body, "href=\"/profile/delete\"")
}

// TestEditProfilePageEmailWarning tests that the edit profile page contains a warning about email verification
func TestEditProfilePageEmailWarning(t *testing.T) {
	// Set up test database with a user
	db, _ := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile/edit", nil)

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: "test@example.com"})

	// Set up router
	router := gin.Default()
	router.HTMLRender = &mockHTMLRender{}
	userController := controllers.NewUserController(db)
	router.GET("/profile/edit", userController.EditProfile)

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains a warning about email verification
	body := w.Body.String()
	assert.Contains(t, body, "If you change your email, you will need to verify it again")
}
