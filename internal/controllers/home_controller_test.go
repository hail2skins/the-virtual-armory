package controllers

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// MockHomeEmailService is a mock implementation of the email service for home controller tests
type MockHomeEmailService struct {
	mock.Mock
}

// IsConfigured returns whether the email service is configured
func (m *MockHomeEmailService) IsConfigured() bool {
	args := m.Called()
	return args.Bool(0)
}

// SendVerificationEmail sends a verification email
func (m *MockHomeEmailService) SendVerificationEmail(email, token string) error {
	args := m.Called(email, token)
	return args.Error(0)
}

// SendPasswordResetEmail sends a password reset email
func (m *MockHomeEmailService) SendPasswordResetEmail(email, resetLink string) error {
	args := m.Called(email, resetLink)
	return args.Error(0)
}

// SendContactFormEmail sends a contact form email
func (m *MockHomeEmailService) SendContactFormEmail(name, email, subject, message string) error {
	args := m.Called(name, email, subject, message)
	return args.Error(0)
}

func TestHomeController_Index(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	r := gin.New()

	// Create a mock email service
	mockEmailService := new(MockHomeEmailService)
	mockEmailService.On("IsConfigured").Return(true)
	mockEmailService.On("SendContactFormEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Create a new home controller
	controller := NewHomeController(mockEmailService)

	// Register the index route
	r.GET("/", controller.Index)

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestHomeController_About(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	r := gin.New()

	// Create a mock email service
	mockEmailService := new(MockHomeEmailService)
	mockEmailService.On("IsConfigured").Return(true)

	// Create a new home controller
	controller := NewHomeController(mockEmailService)

	// Register the about route
	r.GET("/about", controller.About)

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/about", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}

func TestHomeController_Contact(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	r := gin.New()

	// Create a mock email service
	mockEmailService := new(MockHomeEmailService)
	mockEmailService.On("IsConfigured").Return(true)

	// Create a new home controller
	controller := NewHomeController(mockEmailService)

	// Register the contact route
	r.GET("/contact", controller.Contact)

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/contact", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Check the status code
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}
}
