package flash

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSetMessage(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// Call the function
	SetMessage(ctx, "Test message", "success")

	// Check the cookies
	cookies := w.Result().Cookies()

	var flashMessageCookie, flashTypeCookie *http.Cookie
	for _, cookie := range cookies {
		switch cookie.Name {
		case "flash_message":
			flashMessageCookie = cookie
		case "flash_type":
			flashTypeCookie = cookie
		}
	}

	// Assert that the cookies were set correctly
	assert.NotNil(t, flashMessageCookie, "flash_message cookie should be present")
	assert.Equal(t, "Test+message", flashMessageCookie.Value)
	assert.Equal(t, 5, flashMessageCookie.MaxAge)

	assert.NotNil(t, flashTypeCookie, "flash_type cookie should be present")
	assert.Equal(t, "success", flashTypeCookie.Value)
	assert.Equal(t, 5, flashTypeCookie.MaxAge)
}

func TestSetMessageWithMaxAge(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(w)

	// Call the function with a custom MaxAge
	SetMessageWithMaxAge(ctx, "Test message", "error", 60)

	// Check the cookies
	cookies := w.Result().Cookies()

	var flashMessageCookie, flashTypeCookie *http.Cookie
	for _, cookie := range cookies {
		switch cookie.Name {
		case "flash_message":
			flashMessageCookie = cookie
		case "flash_type":
			flashTypeCookie = cookie
		}
	}

	// Assert that the cookies were set correctly
	assert.NotNil(t, flashMessageCookie, "flash_message cookie should be present")
	assert.Equal(t, "Test+message", flashMessageCookie.Value)
	assert.Equal(t, 60, flashMessageCookie.MaxAge)

	assert.NotNil(t, flashTypeCookie, "flash_type cookie should be present")
	assert.Equal(t, "error", flashTypeCookie.Value)
	assert.Equal(t, 60, flashTypeCookie.MaxAge)
}

func TestSetMessageWithDifferentTypes(t *testing.T) {
	// Test different message types
	messageTypes := []string{"success", "error", "warning", "info"}

	for _, msgType := range messageTypes {
		// Setup
		gin.SetMode(gin.TestMode)
		w := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(w)

		// Call the function
		SetMessage(ctx, "Test "+msgType+" message", msgType)

		// Check the cookies
		cookies := w.Result().Cookies()

		var flashTypeCookie *http.Cookie
		for _, cookie := range cookies {
			if cookie.Name == "flash_type" {
				flashTypeCookie = cookie
				break
			}
		}

		// Assert that the cookie was set with the correct type
		assert.NotNil(t, flashTypeCookie, "flash_type cookie should be present")
		assert.Equal(t, msgType, flashTypeCookie.Value)
	}
}
