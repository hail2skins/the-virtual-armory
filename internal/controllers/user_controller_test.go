package controllers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// Mock the email service for testing
type UserControllerMockEmailService struct {
	mock.Mock
}

// IsConfigured mocks the IsConfigured method
func (m *UserControllerMockEmailService) IsConfigured() bool {
	args := m.Called()
	return args.Bool(0)
}

// SendVerificationEmail mocks the SendVerificationEmail method
func (m *UserControllerMockEmailService) SendVerificationEmail(email, token string) error {
	args := m.Called(email, token)
	return args.Error(0)
}

// SendPasswordResetEmail mocks the SendPasswordResetEmail method
func (m *UserControllerMockEmailService) SendPasswordResetEmail(email, resetLink string) error {
	args := m.Called(email, resetLink)
	return args.Error(0)
}

// SendContactFormEmail mocks the SendContactFormEmail method
func (m *UserControllerMockEmailService) SendContactFormEmail(name, email, subject, message string) error {
	args := m.Called(name, email, subject, message)
	return args.Error(0)
}

// MockUserController extends UserController with a mock getCurrentUser method
type MockUserController struct {
	*UserController
	mockUser *models.User
}

// getCurrentUser returns the mock user instead of looking it up
func (c *MockUserController) getCurrentUser(ctx *gin.Context) (*models.User, error) {
	return c.mockUser, nil
}

// setupUserControllerTest sets up a test environment for the user controller
func setupUserControllerTest(t *testing.T) (*gin.Engine, *UserController, *UserControllerMockEmailService) {
	// Setup
	gin.SetMode(gin.TestMode)
	testutils.SetupTestDB()

	// Create a mock email service
	mockEmailService := new(UserControllerMockEmailService)
	mockEmailService.On("IsConfigured").Return(true)
	mockEmailService.On("SendVerificationEmail", mock.Anything, mock.Anything).Return(nil)
	mockEmailService.On("SendPasswordResetEmail", mock.Anything, mock.Anything).Return(nil)
	mockEmailService.On("SendContactFormEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create a user controller with the mock email service
	userController := NewUserControllerWithEmailService(database.TestDB, mockEmailService)

	// Create a test user
	user := createTestUser(t)

	// Set the mock user for authentication
	auth.MockUser = user

	// Create a test router
	router := gin.Default()

	return router, userController, mockEmailService
}

// createTestUser creates a test user in the database
func createTestUser(t *testing.T) *models.User {
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Email:     "test@example.com",
		Password:  string(hashedPassword),
		Confirmed: true,
	}
	database.TestDB.Create(&user)
	return &user
}

// TestShowProfile tests the ShowProfile method
func TestShowProfile(t *testing.T) {
	// Setup
	router, userController, _ := setupUserControllerTest(t)

	// Setup routes
	router.GET("/profile", func(ctx *gin.Context) {
		userController.ShowProfile(ctx)
	})

	// Create a request
	req, _ := http.NewRequest("GET", "/profile", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), auth.MockUser.Email)
}

// TestEditProfile tests the EditProfile method
func TestEditProfile(t *testing.T) {
	// Setup
	router, userController, _ := setupUserControllerTest(t)

	// Setup routes
	router.GET("/profile/edit", func(ctx *gin.Context) {
		userController.EditProfile(ctx)
	})

	// Create a request
	req, _ := http.NewRequest("GET", "/profile/edit", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), auth.MockUser.Email)
	assert.Contains(t, w.Body.String(), "Edit Profile")
}

// TestUpdateProfile tests the UpdateProfile method
func TestUpdateProfile(t *testing.T) {
	// Setup
	router, userController, mockEmailService := setupUserControllerTest(t)

	// Setup routes
	router.POST("/profile", func(ctx *gin.Context) {
		userController.UpdateProfile(ctx)
	})

	// Create a request with form data
	formData := "email=newemail@example.com"
	req, _ := http.NewRequest("POST", "/profile", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Setup mock expectations
	mockEmailService.On("SendVerificationEmail", "newemail@example.com", mock.Anything).Return(nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/verification-pending", w.Header().Get("Location"))

	// Verify that the email service was called
	mockEmailService.AssertCalled(t, "SendVerificationEmail", "newemail@example.com", mock.Anything)
}

// TestShowSubscription tests the ShowSubscription method
func TestShowSubscription(t *testing.T) {
	// Setup
	router, userController, _ := setupUserControllerTest(t)

	// Setup routes
	router.GET("/profile/subscription", func(ctx *gin.Context) {
		userController.ShowSubscription(ctx)
	})

	// Create a request
	req, _ := http.NewRequest("GET", "/profile/subscription", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Subscription Management")
	assert.Contains(t, w.Body.String(), "Current Plan")
}

// TestShowDeleteAccount tests the ShowDeleteAccount method
func TestShowDeleteAccount(t *testing.T) {
	// Setup
	router, userController, _ := setupUserControllerTest(t)

	// Setup routes
	router.GET("/profile/delete", func(ctx *gin.Context) {
		userController.ShowDeleteAccount(ctx)
	})

	// Create a request
	req, _ := http.NewRequest("GET", "/profile/delete", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Delete Your Account")
}

// TestDeleteAccount tests the DeleteAccount method
func TestDeleteAccount(t *testing.T) {
	// Setup
	router, userController, _ := setupUserControllerTest(t)

	// Create a test user with a bcrypt-hashed password
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	user := models.User{
		Email:    "delete-test@example.com",
		Password: string(hashedPassword),
	}
	database.TestDB.Create(&user)

	// Update the mock user to use the delete-test user
	auth.MockUser = &user

	// Setup routes
	router.POST("/profile/delete", func(ctx *gin.Context) {
		userController.DeleteAccount(ctx)
	})

	// Create a request with form data
	formData := "confirm_text=DELETE&password=password"
	req, _ := http.NewRequest("POST", "/profile/delete", strings.NewReader(formData))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - it should redirect to the home page
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))

	// Check that a flash message was set
	var flashMessageCookie, flashTypeCookie *http.Cookie
	for _, cookie := range w.Result().Cookies() {
		switch cookie.Name {
		case "flash_message":
			flashMessageCookie = cookie
		case "flash_type":
			flashTypeCookie = cookie
		}
	}

	assert.NotNil(t, flashMessageCookie, "flash_message cookie should be present")
	assert.Contains(t, flashMessageCookie.Value, "deleted")

	assert.NotNil(t, flashTypeCookie, "flash_type cookie should be present")
	assert.Equal(t, "success", flashTypeCookie.Value)

	// Check that the user was soft-deleted
	var deletedUser models.User
	result := database.TestDB.Unscoped().Where("email = ?", user.Email).First(&deletedUser)
	assert.NoError(t, result.Error)
	assert.False(t, deletedUser.DeletedAt.Time.IsZero(), "User should be soft-deleted")
}
