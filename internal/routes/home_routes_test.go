package routes

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHomeRoutes(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	r := gin.New()

	// Register home routes
	RegisterHomeRoutes(r)

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
