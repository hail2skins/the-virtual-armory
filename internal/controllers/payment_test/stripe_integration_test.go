package payment_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestStripeCheckoutCreation tests that a Stripe checkout session is created correctly
func TestStripeCheckoutCreation(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for checkout
	router.POST("/checkout", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.CreateCheckoutSession(c)
	})

	// Test creating a checkout session for monthly subscription
	form := url.Values{}
	form.Add("tier", "monthly")
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Add authentication cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the success page with a test session ID
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/payment/success?session_id=cs_test_")
}

// TestStripeWebhookSignatureVerification tests that Stripe webhook signatures are verified
func TestStripeWebhookSignatureVerification(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for the Stripe webhook
	router.POST("/webhook", func(c *gin.Context) {
		paymentController.HandleStripeWebhook(c)
	})

	// Create a test webhook payload
	webhookPayload := `{
		"id": "evt_test123",
		"type": "checkout.session.completed",
		"data": {
			"object": {
				"id": "cs_test_123",
				"customer": "cus_test_123",
				"metadata": {
					"user_id": "999",
					"subscription_tier": "monthly"
				}
			}
		}
	}`

	// Test sending a webhook with test signature
	req, _ := http.NewRequest("POST", "/webhook", strings.NewReader(webhookPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "test_signature")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// In test mode with test_signature, this should succeed with a 200 status code
	assert.Equal(t, http.StatusOK, w.Code)
}

// TestEnhancedPricingPageFeatures tests that the enhanced pricing page displays all features correctly
func TestEnhancedPricingPageFeatures(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for the pricing page
	router.GET("/pricing", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.ShowPricingPage(c)
	})

	// Test accessing the pricing page as a logged-in user
	req, _ := http.NewRequest("GET", "/pricing", nil)
	// Add authentication cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the pricing page is displayed with enhanced features
	assert.Equal(t, http.StatusOK, w.Code)

	// Check for recommended plan highlight
	assert.Contains(t, w.Body.String(), "Best Value")

	// Check for current plan indicator
	assert.Contains(t, w.Body.String(), "Current Plan")
}

// TestPaymentHistoryDisplay tests that payment history is displayed correctly
func TestPaymentHistoryDisplay(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Create some payment history records
	payment1 := models.Payment{
		UserID:      user.ID,
		Amount:      999, // $9.99
		Currency:    "usd",
		PaymentType: "subscription",
		Status:      "succeeded",
		Description: "Monthly Subscription",
	}
	db.Create(&payment1)

	payment2 := models.Payment{
		UserID:      user.ID,
		Amount:      999, // $9.99
		Currency:    "usd",
		PaymentType: "subscription",
		Status:      "succeeded",
		Description: "Monthly Subscription",
	}
	db.Create(&payment2)

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for the payment history page
	router.GET("/payment/history", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.ShowPaymentHistory(c)
	})

	// Test accessing the payment history page
	req, _ := http.NewRequest("GET", "/payment/history", nil)
	// Add authentication cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the payment history is displayed
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Payment History")
	assert.Contains(t, w.Body.String(), "Monthly Subscription")
	assert.Contains(t, w.Body.String(), "$9.99")
}
