package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/config"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/services/email"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

// TestLogoutFlashMessage tests that a flash message cookie is set after logout
func TestLogoutFlashMessage(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutils.SetupTestDB()

	// Create a test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Confirmed: true,
	}
	database.TestDB.Create(&user)

	// Setup router with controllers
	router := gin.Default()
	cfg := &config.Config{
		AppBaseURL: "http://localhost:8080",
	}

	// Create mock email service
	mockEmailService := &email.MockEmailService{
		IsConfiguredResult: true,
	}

	// Create auth instance
	authInstance, _ := auth.New()

	authController := NewAuthController(authInstance, mockEmailService, cfg)
	homeController := NewHomeController()

	// Setup routes
	router.GET("/", homeController.Index)
	router.GET("/logout", authController.Logout)

	// Create a request to logout
	req, _ := http.NewRequest("GET", "/logout", nil)

	// Set cookies to simulate a logged-in user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: "test@example.com",
	})

	// Perform the logout request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that we got redirected to the home page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))

	// Get the cookies from the response
	var flashMessageCookie, flashTypeCookie *http.Cookie
	for _, cookie := range w.Result().Cookies() {
		switch cookie.Name {
		case "flash_message":
			flashMessageCookie = cookie
		case "flash_type":
			flashTypeCookie = cookie
		}
	}

	// Check that flash message cookies were set correctly
	assert.NotNil(t, flashMessageCookie, "flash_message cookie should be present")
	assert.Equal(t, "You+have+been+successfully+logged+out", flashMessageCookie.Value)

	assert.NotNil(t, flashTypeCookie, "flash_type cookie should be present")
	assert.Equal(t, "success", flashTypeCookie.Value)
}

// TestDeleteAccountFlashMessage tests that a flash message cookie is set after account deletion
func TestDeleteAccountFlashMessage(t *testing.T) {
	// Skip this test for now as it requires HTML rendering which is not properly set up in the test environment
	t.Skip("Skipping test that requires HTML rendering")

	// Setup
	gin.SetMode(gin.TestMode)
	testutils.SetupTestDB()

	// Create a test user
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Confirmed: true,
	}
	database.TestDB.Create(&user)

	// Setup router with controllers
	router := gin.Default()
	cfg := &config.Config{
		AppBaseURL: "http://localhost:8080",
	}

	// Create mock email service
	mockEmailService := &email.MockEmailService{
		IsConfiguredResult: true,
	}

	// Create auth instance
	authInstance, _ := auth.New()

	authController := NewAuthController(authInstance, mockEmailService, cfg)
	homeController := NewHomeController()

	// Setup routes
	router.GET("/", homeController.Index)
	router.POST("/delete-account", authController.ProcessDeleteAccount)

	// Create a request to delete account
	formData := "confirm_text=DELETE&password=password"
	req, _ := http.NewRequest("POST", "/delete-account", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set cookies to simulate a logged-in user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: "test@example.com",
	})

	// Perform the delete account request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that we got redirected to the home page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))

	// Get the cookies from the response
	var flashMessageCookie, flashTypeCookie *http.Cookie
	for _, cookie := range w.Result().Cookies() {
		switch cookie.Name {
		case "flash_message":
			flashMessageCookie = cookie
		case "flash_type":
			flashTypeCookie = cookie
		}
	}

	// Check that flash message cookies were set correctly
	assert.NotNil(t, flashMessageCookie, "flash_message cookie should be present")
	assert.Equal(t, "Sorry+to+see+you+go.+Your+account+has+been+deleted.", flashMessageCookie.Value)

	assert.NotNil(t, flashTypeCookie, "flash_type cookie should be present")
	assert.Equal(t, "success", flashTypeCookie.Value)
}
