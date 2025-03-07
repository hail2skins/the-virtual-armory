package payment_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestPaymentFeedback tests that the user receives appropriate feedback after payment
func TestPaymentFeedback(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user with a subscription
	user := payment_test_utils.CreateTestUser(t, db)
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": time.Now().AddDate(0, 1, 0),
	})

	// Set up the router
	router := gin.Default()

	// Set up the controller
	authController := controllers.NewAuthController(nil, nil, nil)

	// Set up the route
	router.GET("/owner", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		// Call the Profile function
		authController.Profile(c)
	})

	// Test accessing the owner page with flash message
	req, _ := http.NewRequest("GET", "/owner", nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
	req.AddCookie(&http.Cookie{Name: "flash_message", Value: "Your payment was successful"})
	req.AddCookie(&http.Cookie{Name: "flash_type", Value: "success"})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the response contains the subscription info
	assert.Equal(t, http.StatusOK, w.Code)

	// Since the flash message is handled by JavaScript, we can't check for it directly
	// Instead, we'll check for the subscription information
	assert.Contains(t, w.Body.String(), "Current Plan: Liking It")
}

// TestRedirectToOwnerAfterPayment tests that the user is redirected to the owner page after payment
func TestRedirectToOwnerAfterPayment(t *testing.T) {
	// Set test environment
	t.Setenv("APP_ENV", "test")

	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up the router
	router := gin.Default()

	// Set up the controller
	paymentController := controllers.NewPaymentController(db)

	// Set up the route
	router.GET("/payment/success", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.HandlePaymentSuccess(c)
	})

	// Test accessing the payment success page
	req, _ := http.NewRequest("GET", "/payment/success?session_id=cs_test_123", nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the owner page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/owner", w.Header().Get("Location"))

	// Check that the flash message cookie is set
	cookies := w.Result().Cookies()
	var flashMessageFound, flashTypeFound bool
	for _, cookie := range cookies {
		if cookie.Name == "flash_message" {
			assert.Contains(t, cookie.Value, "successful")
			assert.Contains(t, cookie.Value, "subscription")
			flashMessageFound = true
		}
		if cookie.Name == "flash_type" {
			assert.Equal(t, "success", cookie.Value)
			flashTypeFound = true
		}
	}
	assert.True(t, flashMessageFound, "Flash message cookie not found")
	assert.True(t, flashTypeFound, "Flash type cookie not found")
}
