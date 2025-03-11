package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/auth"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/partials"
	"github.com/stretchr/testify/assert"
)

func TestErrorTemplates(t *testing.T) {
	// Set up Gin in test mode
	gin.SetMode(gin.TestMode)

	// Create a test recorder and context
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	// Test cases
	testCases := []struct {
		name     string
		template func(string) error
		message  string
	}{
		{
			name: "Auth Error Template",
			template: func(msg string) error {
				component := auth.Error(msg)
				return component.Render(c.Request.Context(), w)
			},
			message: "Test auth error",
		},
		{
			name: "Partials Error Template",
			template: func(msg string) error {
				component := partials.Error(msg)
				return component.Render(c.Request.Context(), w)
			},
			message: "Test partials error",
		},
	}

	// Run tests
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Reset the recorder
			w = httptest.NewRecorder()
			c, _ = gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)

			// Render the template
			err := tc.template(tc.message)

			// Check that there was no error
			assert.NoError(t, err)

			// Check that the response contains the error message
			assert.Contains(t, w.Body.String(), tc.message)

			// Check that the status code is 200 (OK)
			assert.Equal(t, http.StatusOK, w.Code)
		})
	}
}
