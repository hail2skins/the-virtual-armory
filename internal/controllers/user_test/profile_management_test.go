package user_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/services/email"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// setupTestDB sets up a test database with a test user
func setupTestDB(t *testing.T) (*gorm.DB, models.User) {
	// Set up a test database
	db, err := testutils.SetupTestDB()
	assert.NoError(t, err)

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	assert.NoError(t, err)

	// Create a test user
	user := models.User{
		Email:                 "test@example.com",
		Password:              string(hashedPassword),
		SubscriptionTier:      "free",
		SubscriptionExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days from now
		Confirmed:             true,
	}
	err = db.Create(&user).Error
	assert.NoError(t, err)

	return db, user
}

// setupRouter sets up a test router with the user controller
func setupRouter(db *gorm.DB) (*gin.Engine, *controllers.UserController) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Mock the HTML renderer for testing
	router.HTMLRender = &mockHTMLRender{}

	userController := controllers.NewUserController(db)

	// Add a middleware to set the user in the context for testing
	router.Use(func(c *gin.Context) {
		// Get the user email from the cookie
		email, err := c.Cookie("user_email")
		if err == nil && email != "" {
			// Get the user from the database
			var user models.User
			if err := db.Where("email = ?", email).First(&user).Error; err == nil {
				// Set the user in the context
				c.Set("user", user)
			}
		}
		c.Next()
	})

	// Set up routes
	router.GET("/profile", userController.ShowProfile)
	router.GET("/profile/edit", userController.EditProfile)
	router.POST("/profile/update", userController.UpdateProfile)
	router.GET("/profile/subscription", userController.ShowSubscription)
	router.GET("/profile/delete", userController.ShowDeleteAccount)
	router.POST("/profile/delete", userController.DeleteAccount)
	router.POST("/profile/reactivate", userController.ReactivateAccount)

	return router, userController
}

// mockHTMLRender is a mock implementation of the HTMLRender interface
type mockHTMLRender struct{}

// Instance returns a mock renderer instance
func (r *mockHTMLRender) Instance(name string, data interface{}) render.Render {
	// Convert the data to JSON for testing
	return &mockRender{data: data}
}

// mockRender is a mock implementation of the Render interface
type mockRender struct {
	data interface{}
}

// Render writes the mock data as JSON to the response
func (r *mockRender) Render(w http.ResponseWriter) error {
	// Set the content type to JSON
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	// Write the data as JSON
	return json.NewEncoder(w).Encode(r.data)
}

// WriteContentType writes the content type to the response
func (r *mockRender) WriteContentType(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
}

// TestShowProfile tests that the profile page is displayed correctly
func TestShowProfile(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up router
	router, _ := setupRouter(db)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile", nil)

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains the user's email and subscription tier
	body := w.Body.String()
	assert.Contains(t, body, user.Email)
	assert.Contains(t, body, user.SubscriptionTier)
}

// TestEditProfile tests that the edit profile page is displayed correctly
func TestEditProfile(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up router
	router, _ := setupRouter(db)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile/edit", nil)

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains the user's email
	body := w.Body.String()
	assert.Contains(t, body, user.Email)
}

// TestUpdateProfile tests that the profile update works correctly
func TestUpdateProfile(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up router
	router, _ := setupRouter(db)

	// Create form data
	form := url.Values{}
	form.Add("email", "updated@example.com")

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profile/update", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, "/profile", w.Header().Get("Location"))

	// Check that the user was updated in the database
	var updatedUser models.User
	err := db.Where("id = ?", user.ID).First(&updatedUser).Error
	assert.NoError(t, err)
	assert.Equal(t, "updated@example.com", updatedUser.Email)
}

// TestShowSubscription tests that the subscription page is displayed correctly
func TestShowSubscription(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up router
	router, _ := setupRouter(db)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile/subscription", nil)

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains the user's subscription tier
	body := w.Body.String()
	assert.Contains(t, body, "Free Plan")
}

// TestShowDeleteAccount tests that the delete account page is displayed correctly
func TestShowDeleteAccount(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up router
	router, _ := setupRouter(db)

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/profile/delete", nil)

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the response contains the delete account confirmation
	body := w.Body.String()
	assert.Contains(t, body, "Delete Your Account")
}

// TestDeleteAccount tests that a user can delete their account
func TestDeleteAccount(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up router
	router, _ := setupRouter(db)

	// Create form data
	form := url.Values{}
	form.Add("confirm_text", "DELETE")
	form.Add("password", "password123") // Use the plain text password that was hashed

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profile/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response (should redirect to home page)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))

	// Check that the user was soft-deleted in the database
	var deletedUser models.User
	err := db.Unscoped().Where("id = ?", user.ID).First(&deletedUser).Error
	assert.NoError(t, err)
	assert.False(t, deletedUser.DeletedAt.Time.IsZero(), "User should be soft-deleted")
}

// TestReactivateAccount tests that a soft-deleted account can be reactivated
func TestReactivateAccount(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Soft-delete the user
	err := db.Delete(&user).Error
	assert.NoError(t, err)

	// Set up router
	router, _ := setupRouter(db)

	// Create form data
	form := url.Values{}
	form.Add("email", user.Email)
	form.Add("password", user.Password) // Use the actual hashed password

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profile/reactivate", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response (should redirect to login page)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/login", w.Header().Get("Location"))

	// Verify the user was reactivated in the database
	var reactivatedUser models.User
	err = db.First(&reactivatedUser, user.ID).Error
	assert.NoError(t, err)
	assert.True(t, reactivatedUser.DeletedAt.Time.IsZero(), "DeletedAt should be a zero time value")
}

// TestUpdateProfileWithEmailChange tests that updating a profile with a new email triggers email verification
func TestUpdateProfileWithEmailChange(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a mock email service
	mockEmailService := &email.MockEmailService{
		IsConfiguredResult: true,
	}

	// Set up router with the mock email service
	router, _ := setupRouterWithEmailService(db, mockEmailService)

	// Create a test request with a new email
	w := httptest.NewRecorder()
	form := url.Values{}
	form.Add("email", "newemail@example.com")
	req, _ := http.NewRequest("POST", "/profile/update", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response - should redirect to verification pending page
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/verification-pending", w.Header().Get("Location"))

	// Assert that the email service was called to send a verification email
	assert.True(t, mockEmailService.SendVerificationEmailCalled)
	assert.Equal(t, "newemail@example.com", mockEmailService.SendVerificationEmailEmail)
	assert.NotEmpty(t, mockEmailService.SendVerificationEmailToken)

	// Check that the user's email was updated in the database
	var updatedUser models.User
	db.First(&updatedUser, user.ID)
	assert.Equal(t, "newemail@example.com", updatedUser.Email)
	assert.False(t, updatedUser.Confirmed)
	assert.NotEmpty(t, updatedUser.ConfirmToken)
	assert.False(t, updatedUser.ConfirmTokenExpiry.IsZero())
}

// setupRouterWithEmailService sets up a router with a mock email service
func setupRouterWithEmailService(db *gorm.DB, emailService email.EmailService) (*gin.Engine, *controllers.UserController) {
	// Create a new router
	router := gin.Default()

	// Set up HTML renderer
	router.HTMLRender = &mockHTMLRender{}

	// Create user controller with email service
	userController := controllers.NewUserControllerWithEmailService(db, emailService)

	// Set up routes
	router.GET("/profile", userController.ShowProfile)
	router.GET("/profile/edit", userController.EditProfile)
	router.POST("/profile/update", userController.UpdateProfile)
	router.GET("/profile/subscription", userController.ShowSubscription)
	router.GET("/profile/delete", userController.ShowDeleteAccount)
	router.POST("/profile/delete", userController.DeleteAccount)
	router.POST("/profile/reactivate", userController.ReactivateAccount)

	return router, userController
}

// TestDeleteAccountSuccess tests that a user can successfully delete their account
func TestDeleteAccountSuccess(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up router
	router, _ := setupRouter(db)

	// Create form data
	form := url.Values{}
	form.Add("confirm_text", "DELETE")
	form.Add("password", "password123") // This matches the password set in setupTestDB

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profile/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response (should redirect to home page)
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))

	// Verify the user was soft-deleted in the database
	var deletedUser models.User
	err := db.Unscoped().First(&deletedUser, user.ID).Error
	assert.NoError(t, err)
	assert.False(t, deletedUser.DeletedAt.Time.IsZero(), "DeletedAt should not be a zero time value")
}

// TestDeleteAccountInvalidConfirmation tests that account deletion fails with invalid confirmation text
func TestDeleteAccountInvalidConfirmation(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up router
	router, _ := setupRouter(db)

	// Create form data with wrong confirmation text
	form := url.Values{}
	form.Add("confirm_text", "delete") // lowercase, should be "DELETE"
	form.Add("password", "password123")

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profile/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response (should render the delete account page with an error)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Delete Your Account") // Page title
	// The error message is set as a flash cookie, not directly in the HTML
	// So we don't check for it in the body

	// Verify the user was NOT deleted in the database
	var existingUser models.User
	err := db.First(&existingUser, user.ID).Error
	assert.NoError(t, err)
	assert.True(t, existingUser.DeletedAt.Time.IsZero(), "User should not be deleted")
}

// TestDeleteAccountInvalidPassword tests that account deletion fails with invalid password
func TestDeleteAccountInvalidPassword(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up router
	router, _ := setupRouter(db)

	// Create form data with wrong password
	form := url.Values{}
	form.Add("confirm_text", "DELETE")
	form.Add("password", "wrongpassword")

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profile/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response (should render the delete account page with an error)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Delete Your Account") // Page title
	// The error message is set as a flash cookie, not directly in the HTML
	// So we don't check for it in the body

	// Verify the user was NOT deleted in the database
	var existingUser models.User
	err := db.First(&existingUser, user.ID).Error
	assert.NoError(t, err)
	assert.True(t, existingUser.DeletedAt.Time.IsZero(), "User should not be deleted")
}

// TestDeleteAccountEmptyFields tests that account deletion fails when fields are empty
func TestDeleteAccountEmptyFields(t *testing.T) {
	// Set up test database with a user
	db, user := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up router
	router, _ := setupRouter(db)

	// Create form data with empty fields
	form := url.Values{}
	form.Add("confirm_text", "")
	form.Add("password", "")

	// Create a test request
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/profile/delete", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Set cookies to simulate authentication
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	// Serve the request
	router.ServeHTTP(w, req)

	// Assert response (should render the delete account page with errors)
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Delete Your Account") // Page title
	// The error messages are set as flash cookies, not directly in the HTML
	// So we don't check for them in the body

	// Verify the user was NOT deleted in the database
	var existingUser models.User
	err := db.First(&existingUser, user.ID).Error
	assert.NoError(t, err)
	assert.True(t, existingUser.DeletedAt.Time.IsZero(), "User should not be deleted")
}
