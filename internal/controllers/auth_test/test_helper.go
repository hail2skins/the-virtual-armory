package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// MockAuthService is a mock implementation of the Auth service
type MockAuthService struct {
	mock.Mock
}

// Middleware mocks the Auth.Middleware method
func (m *MockAuthService) Middleware() gin.HandlerFunc {
	args := m.Called()
	return args.Get(0).(gin.HandlerFunc)
}

// RequireAuth mocks the Auth.RequireAuth method
func (m *MockAuthService) RequireAuth() gin.HandlerFunc {
	args := m.Called()
	return args.Get(0).(gin.HandlerFunc)
}

// RequireAdmin mocks the Auth.RequireAdmin method
func (m *MockAuthService) RequireAdmin() gin.HandlerFunc {
	args := m.Called()
	return args.Get(0).(gin.HandlerFunc)
}

// SetupTestRouter sets up a test router with the auth controller
func SetupTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, *controllers.AuthController, *MockAuthService) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a test router
	router := gin.New()
	router.Use(gin.Recovery())

	// Set up the test database
	db, err := testutils.SetupTestDB()
	require.NoError(t, err)

	// Create a mock auth service
	mockAuth := new(MockAuthService)

	// Create the auth controller with the mock auth service
	authController := &controllers.AuthController{
		Auth: &auth.Auth{}, // We'll use a minimal Auth struct
	}

	return router, db, authController, mockAuth
}

// CreateTestRequest creates a test HTTP request
func CreateTestRequest(method, path string, body interface{}) (*http.Request, *httptest.ResponseRecorder) {
	var reqBody *bytes.Buffer

	if body != nil {
		switch v := body.(type) {
		case url.Values:
			reqBody = bytes.NewBufferString(v.Encode())
		case string:
			reqBody = bytes.NewBufferString(v)
		default:
			jsonBytes, _ := json.Marshal(body)
			reqBody = bytes.NewBuffer(jsonBytes)
		}
	} else {
		reqBody = bytes.NewBufferString("")
	}

	req, _ := http.NewRequest(method, path, reqBody)
	if body != nil {
		if _, ok := body.(url.Values); ok {
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		} else if _, ok := body.(string); !ok {
			req.Header.Set("Content-Type", "application/json")
		}
	}

	w := httptest.NewRecorder()
	return req, w
}

// CreateFormRequest creates a test HTTP request with form data
func CreateFormRequest(method, path string, formData url.Values) (*http.Request, *httptest.ResponseRecorder) {
	req, _ := http.NewRequest(method, path, strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	return req, w
}

// CleanupTest cleans up after a test
func CleanupTest(db *gorm.DB) {
	testutils.CleanupTestDB(db)
}

// CreateTestUser creates a test user
func CreateTestUser(t *testing.T, db *gorm.DB, email, password string, isAdmin bool) *models.User {
	user, err := testutils.CreateTestUser(db, email, password, isAdmin)
	require.NoError(t, err)
	return user
}
