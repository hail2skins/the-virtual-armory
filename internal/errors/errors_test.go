package errors

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestErrorTypes(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ValidationError",
			err:      NewValidationError("Invalid input"),
			expected: "Invalid input",
		},
		{
			name:     "AuthError",
			err:      NewAuthError("Unauthorized"),
			expected: "Unauthorized",
		},
		{
			name:     "NotFoundError",
			err:      NewNotFoundError("Resource not found"),
			expected: "Resource not found",
		},
		{
			name:     "PaymentError",
			err:      NewPaymentError("Payment failed", "CARD_DECLINED"),
			expected: "Payment failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

func TestHandleError(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		err            error
		expectedStatus int
		acceptHeader   string
	}{
		{
			name:           "ValidationError with JSON Accept",
			err:            NewValidationError("Invalid input"),
			expectedStatus: http.StatusBadRequest,
			acceptHeader:   "application/json",
		},
		{
			name:           "AuthError with JSON Accept",
			err:            NewAuthError("Unauthorized"),
			expectedStatus: http.StatusUnauthorized,
			acceptHeader:   "application/json",
		},
		{
			name:           "NotFoundError with JSON Accept",
			err:            NewNotFoundError("Resource not found"),
			expectedStatus: http.StatusNotFound,
			acceptHeader:   "application/json",
		},
		{
			name:           "PaymentError with JSON Accept",
			err:            NewPaymentError("Payment failed", "CARD_DECLINED"),
			expectedStatus: http.StatusBadRequest,
			acceptHeader:   "application/json",
		},
		{
			name:           "Standard error with JSON Accept",
			err:            errors.New("Standard error"),
			expectedStatus: http.StatusInternalServerError,
			acceptHeader:   "application/json",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)
			c.Request.Header.Set("Accept", tt.acceptHeader)

			// Call the error handler
			HandleError(c, tt.err)

			// Check the status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// For JSON responses, check the content type
			if tt.acceptHeader == "application/json" {
				assert.Contains(t, w.Header().Get("Content-Type"), "application/json")
			}
		})
	}
}
