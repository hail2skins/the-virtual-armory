package payment_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/stretchr/testify/assert"
)

// TestPaymentSuccessRedirectToOwner tests that the payment success handler redirects to /owner
// NEVER CHANGE THIS REDIRECT - IT MUST ALWAYS GO TO /owner
// This test is specifically designed to fail if someone changes the redirect destination
// to ensure that users are always redirected to the owner page after a successful payment
func TestPaymentSuccessRedirectToOwner(t *testing.T) {
	// Set test environment
	t.Setenv("APP_ENV", "test")

	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer payment_test_utils.CleanupTestDB(t, db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up the router
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(payment_test_utils.AuthMiddlewareMock(user))

	// Set up the payment controller
	paymentController := controllers.NewPaymentController(db)
	router.GET("/payment/success", paymentController.HandlePaymentSuccess)

	// Create a test request with a test session ID
	req, _ := http.NewRequest("GET", "/payment/success?session_id=cs_test_123", nil)
	resp := httptest.NewRecorder()

	// Serve the request
	router.ServeHTTP(resp, req)

	// Assert that the response is a redirect to /owner
	// NEVER CHANGE THIS ASSERTION - THE REDIRECT MUST ALWAYS GO TO /owner
	assert.Equal(t, http.StatusSeeOther, resp.Code)
	assert.Equal(t, "/owner", resp.Header().Get("Location"),
		"CRITICAL TEST FAILURE: Payment success must redirect to /owner - NEVER change this redirect!")
}
