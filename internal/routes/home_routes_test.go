package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/mock"
)

// MockHomeRoutesEmailService is a mock implementation of the email service for home routes tests
type MockHomeRoutesEmailService struct {
	mock.Mock
}

// IsConfigured returns whether the email service is configured
func (m *MockHomeRoutesEmailService) IsConfigured() bool {
	args := m.Called()
	return args.Bool(0)
}

// SendVerificationEmail sends a verification email
func (m *MockHomeRoutesEmailService) SendVerificationEmail(email, token string) error {
	args := m.Called(email, token)
	return args.Error(0)
}

// SendPasswordResetEmail sends a password reset email
func (m *MockHomeRoutesEmailService) SendPasswordResetEmail(email, resetLink string) error {
	args := m.Called(email, resetLink)
	return args.Error(0)
}

// SendContactFormEmail sends a contact form email
func (m *MockHomeRoutesEmailService) SendContactFormEmail(name, email, subject, message string) error {
	args := m.Called(name, email, subject, message)
	return args.Error(0)
}

func TestHomeRoutes(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	r := gin.New()

	// Create a mock email service
	mockEmailService := new(MockHomeRoutesEmailService)
	mockEmailService.On("IsConfigured").Return(true)
	mockEmailService.On("SendContactFormEmail", mock.Anything, mock.Anything, mock.Anything, mock.Anything).Return(nil)

	// Register home routes
	RegisterHomeRoutes(r, mockEmailService)

	// Test cases
	tests := []struct {
		name       string
		method     string
		path       string
		wantStatus int
	}{
		{"Index", "GET", "/", http.StatusOK},
		{"About", "GET", "/about", http.StatusOK},
		{"Contact", "GET", "/contact", http.StatusOK},
	}

	// Run tests
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test HTTP request
			req, err := http.NewRequest(tt.method, tt.path, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Create a ResponseRecorder to record the response
			rr := httptest.NewRecorder()

			// Serve the HTTP request
			r.ServeHTTP(rr, req)

			// Check the status code
			if status := rr.Code; status != tt.wantStatus {
				t.Errorf("Handler returned wrong status code: got %v want %v", status, tt.wantStatus)
			}
		})
	}
}
