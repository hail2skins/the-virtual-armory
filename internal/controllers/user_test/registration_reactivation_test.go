package user_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/config"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/services/email"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupTestDBWithDeletedUser sets up a test database with a soft-deleted user
func setupTestDBWithDeletedUser(t *testing.T) (*gorm.DB, models.User) {
	// Set up a test database
	db, err := testutils.SetupTestDB()
	assert.NoError(t, err)

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	// Create a test user
	user := models.User{
		Email:                 "deleted@example.com",
		Password:              string(hashedPassword),
		SubscriptionTier:      "free",
		SubscriptionExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days from now
		Confirmed:             true,
	}
	err = db.Create(&user).Error
	assert.NoError(t, err)

	// Soft-delete the user
	err = db.Delete(&user).Error
	assert.NoError(t, err)

	return db, user
}

// TestRegisterWithDeletedEmail tests the registration flow when a user tries to register with an email that belongs to a soft-deleted account
func TestRegisterWithDeletedEmail(t *testing.T) {
	// Set up test database with a soft-deleted user
	db, deletedUser := setupTestDBWithDeletedUser(t)
	defer testutils.CleanupTestDB(db)

	// Set the test database for this test
	database.TestDB = db

	// Set up router with auth controller
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up the HTML renderer for testing
	router.HTMLRender = &mockHTMLRender{}

	// Create a mock email service
	mockEmailService := &email.MockEmailService{}

	// Create the auth controller with a minimal Auth struct
	authController := controllers.NewAuthController(&auth.Auth{}, mockEmailService, &config.Config{})

	// Register routes
	router.POST("/register", authController.ProcessRegister)
	router.GET("/reactivate", authController.ReactivateAccount)
	router.POST("/reactivate", authController.ProcessReactivation)

	// Test case 1: User tries to register with an email that belongs to a soft-deleted account
	t.Run("DetectsDeletedAccount", func(t *testing.T) {
		// Create form data for registration
		form := url.Values{}
		form.Add("email", deletedUser.Email)
		form.Add("password", "new_password")
		form.Add("confirm_password", "new_password")

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Serve the request
		router.ServeHTTP(w, req)

		// Assert response (should redirect to reactivation page)
		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/reactivate?email="+deletedUser.Email, w.Header().Get("Location"))
	})

	// Test case 2: User chooses to reactivate their account
	t.Run("ReactivatesAccount", func(t *testing.T) {
		// Create form data for reactivation
		form := url.Values{}
		form.Add("email", deletedUser.Email)
		form.Add("password", "password123") // Use the plain text password that was hashed

		// Create a test request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/reactivate", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Serve the request
		router.ServeHTTP(w, req)

		// Assert response (should redirect to owner page)
		assert.Equal(t, http.StatusSeeOther, w.Code)
		assert.Equal(t, "/owner", w.Header().Get("Location"))

		// Verify the user was reactivated in the database
		var reactivatedUser models.User
		err := db.Unscoped().First(&reactivatedUser, deletedUser.ID).Error
		assert.NoError(t, err)
		assert.True(t, reactivatedUser.DeletedAt.Time.IsZero())
	})
}
