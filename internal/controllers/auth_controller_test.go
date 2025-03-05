package controllers

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
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// MockEmailService is a mock implementation of the email service
type MockEmailService struct {
	mock.Mock
	IsConfiguredValue bool
}

// IsConfigured returns whether the email service is configured
func (m *MockEmailService) IsConfigured() bool {
	return m.IsConfiguredValue
}

// SendVerificationEmail sends a verification email
func (m *MockEmailService) SendVerificationEmail(email, token string) error {
	args := m.Called(email, token)
	return args.Error(0)
}

// setupTestDB sets up a test database
func setupTestDB(t *testing.T) *gorm.DB {
	// Use an in-memory SQLite database for testing
	db, err := database.InitTestDB()
	if err != nil {
		t.Fatalf("Failed to initialize test database: %v", err)
	}

	// Migrate the schema
	err = db.AutoMigrate(&models.User{})
	if err != nil {
		t.Fatalf("Failed to migrate schema: %v", err)
	}

	return db
}

// setupTestRouter sets up a test router
func setupTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *AuthController) {
	// Set up Gin
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Create a mock email service
	mockEmailService := &MockEmailService{
		IsConfiguredValue: true,
	}
	mockEmailService.On("SendVerificationEmail", mock.Anything, mock.Anything).Return(nil)

	// Create an auth instance
	authInstance, err := auth.New()
	if err != nil {
		t.Fatalf("Failed to create auth instance: %v", err)
	}

	// Create a test config
	testConfig := &config.Config{
		AppBaseURL: "http://localhost:8080",
	}

	// Create an auth controller
	authController := &AuthController{
		Auth:         authInstance,
		EmailService: mockEmailService,
		config:       testConfig,
	}

	return router, authController
}

// TestRegisterSuccess tests successful registration
func TestRegisterSuccess(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	// Set the global DB variable directly
	database.DB = db
	defer database.CloseDB()

	// Set up test router
	router, authController := setupTestRouter(t, db)

	// Set up the route
	router.POST("/register", authController.ProcessRegister)

	// Create a test request
	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("password", "password123")
	form.Add("confirm_password", "password123")
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/verification-pending", w.Header().Get("Location"))

	// Check that the user was created
	var user models.User
	result := db.Where("email = ?", "test@example.com").First(&user)
	assert.NoError(t, result.Error)
	assert.Equal(t, "test@example.com", user.Email)

	// Verify that the password is hashed (not stored in plain text)
	assert.NotEqual(t, "password123", user.Password)
	assert.True(t, strings.HasPrefix(user.Password, "$2a$"), "Password should be hashed with bcrypt")

	// Verify other user properties
	assert.False(t, user.Confirmed)
	assert.NotEmpty(t, user.ConfirmToken)
	assert.False(t, user.ConfirmTokenExpiry.IsZero())
}

// TestRegisterPasswordMismatch tests registration with mismatched passwords
func TestRegisterPasswordMismatch(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	// Set the global DB variable directly
	database.DB = db
	defer database.CloseDB()

	// Set up test router
	router, authController := setupTestRouter(t, db)

	// Set up the route
	router.POST("/register", authController.ProcessRegister)

	// Create a test request
	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("password", "password123")
	form.Add("confirm_password", "password456")
	req, _ := http.NewRequest("POST", "/register", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Passwords do not match")

	// Check that no user was created
	var count int64
	db.Model(&models.User{}).Count(&count)
	assert.Equal(t, int64(0), count)
}

// TestVerifyEmailSuccess tests successful email verification
func TestVerifyEmailSuccess(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	// Set the global DB variable directly
	database.DB = db
	defer database.CloseDB()

	// Set up test router
	router, authController := setupTestRouter(t, db)

	// Set up the route
	router.GET("/verify/:token", authController.VerifyEmail)

	// Create a test user with a confirmation token
	token := "test-token"
	user := models.User{
		Email:              "test@example.com",
		Password:           "password123",
		ConfirmToken:       token,
		ConfirmTokenExpiry: time.Now().Add(24 * time.Hour),
		Confirmed:          false,
	}
	db.Create(&user)

	// Create a test request
	req, _ := http.NewRequest("GET", "/verify/"+token, nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/login?verified=true", w.Header().Get("Location"))

	// Check that the user is now confirmed
	db.First(&user, user.ID)
	assert.True(t, user.Confirmed)
	assert.Empty(t, user.ConfirmToken)
	assert.True(t, user.ConfirmTokenExpiry.IsZero())
}

// TestVerifyEmailInvalidToken tests email verification with an invalid token
func TestVerifyEmailInvalidToken(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	// Set the global DB variable directly
	database.DB = db
	defer database.CloseDB()

	// Set up test router
	router, authController := setupTestRouter(t, db)

	// Set up the route
	router.GET("/verify/:token", authController.VerifyEmail)

	// Create a test request
	req, _ := http.NewRequest("GET", "/verify/invalid-token", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid or expired verification token")
}

// TestVerifyEmailExpiredToken tests email verification with an expired token
func TestVerifyEmailExpiredToken(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	// Set the global DB variable directly
	database.DB = db
	defer database.CloseDB()

	// Set up test router
	router, authController := setupTestRouter(t, db)

	// Set up the route
	router.GET("/verify/:token", authController.VerifyEmail)

	// Create a test user with an expired confirmation token
	token := "test-token"
	user := models.User{
		Email:              "test@example.com",
		Password:           "password123",
		ConfirmToken:       token,
		ConfirmTokenExpiry: time.Now().Add(-24 * time.Hour), // Expired
		Confirmed:          false,
	}
	db.Create(&user)

	// Create a test request
	req, _ := http.NewRequest("GET", "/verify/"+token, nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Your verification token has expired")

	// Check that the user was not confirmed
	var updatedUser models.User
	db.First(&updatedUser, user.ID)
	assert.False(t, updatedUser.Confirmed)
}

// TestLoginUnverifiedUser tests login with an unverified user
func TestLoginUnverifiedUser(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	// Set the global DB variable directly
	database.DB = db
	defer database.CloseDB()

	// Set up test router
	router, authController := setupTestRouter(t, db)

	// Set up the route
	router.POST("/login", authController.ProcessLogin)

	// Create a test user that is not confirmed
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	user := models.User{
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Confirmed: false,
	}
	db.Create(&user)

	// Create a test request
	form := url.Values{}
	form.Add("email", "test@example.com")
	form.Add("password", "password123")
	req, _ := http.NewRequest("POST", "/login", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Please verify your email before logging in")
}

// TestResendVerification tests resending the verification email
func TestResendVerification(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	// Set the global DB variable directly
	database.DB = db
	defer database.CloseDB()

	// Set up test router
	router, authController := setupTestRouter(t, db)

	// Set up the route
	router.POST("/resend-verification", authController.ResendVerification)

	// Create a test user that is not confirmed
	user := models.User{
		Email:              "test@example.com",
		Password:           "password123",
		ConfirmToken:       "old-token",
		ConfirmTokenExpiry: time.Now().Add(-24 * time.Hour), // Expired
		Confirmed:          false,
	}
	db.Create(&user)

	// Create a test request
	form := url.Values{}
	form.Add("email", "test@example.com")
	req, _ := http.NewRequest("POST", "/resend-verification", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/verification-pending", w.Header().Get("Location"))

	// Check that the user's token was updated
	var updatedUser models.User
	db.First(&updatedUser, user.ID)
	assert.NotEqual(t, "old-token", updatedUser.ConfirmToken)
	assert.True(t, updatedUser.ConfirmTokenExpiry.After(time.Now()))
}
