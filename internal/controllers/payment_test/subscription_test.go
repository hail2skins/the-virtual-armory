package payment_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestPricingPageDisplay tests that the pricing page displays correctly
func TestPricingPageDisplay(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for the pricing page
	router.GET("/pricing", func(c *gin.Context) {
		paymentController.ShowPricingPage(c)
	})

	// Test accessing the pricing page
	req, _ := http.NewRequest("GET", "/pricing", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the pricing page is displayed
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Simple, transparent pricing")
}

// TestSubscriptionTiers tests that the pricing page displays different subscription tiers
func TestSubscriptionTiers(t *testing.T) {
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

	// Check that the pricing page is displayed with subscription tiers
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Free")
	assert.Contains(t, w.Body.String(), "Liking It")
	assert.Contains(t, w.Body.String(), "Loving It")
	assert.Contains(t, w.Body.String(), "Supporter")

	// Check that the pricing information is displayed
	assert.Contains(t, w.Body.String(), "$0")
	assert.Contains(t, w.Body.String(), "$5")
	assert.Contains(t, w.Body.String(), "$30")
	assert.Contains(t, w.Body.String(), "$100")

	// Check that the gun limits are displayed
	assert.Contains(t, w.Body.String(), "Store up to 2 guns")
	assert.Contains(t, w.Body.String(), "Unlimited guns/ammo")
}

// TestStripeWebhookHandling tests that Stripe webhooks are handled correctly
func TestStripeWebhookHandling(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

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
					"user_id": "` + fmt.Sprintf("%d", user.ID) + `",
					"subscription_tier": "monthly"
				}
			}
		}
	}`

	// Test sending a webhook
	req, _ := http.NewRequest("POST", "/webhook", strings.NewReader(webhookPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "test_signature")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the webhook was processed
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the user's subscription was updated
	var updatedUser models.User
	db.First(&updatedUser, user.ID)
	assert.Equal(t, "monthly", updatedUser.SubscriptionTier)
	assert.True(t, updatedUser.SubscriptionExpiresAt.After(time.Now()))
}

// TestPaymentSuccess tests that successful payments are handled correctly
func TestPaymentSuccess(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for payment success
	router.GET("/payment/success", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.HandlePaymentSuccess(c)
	})

	// Test accessing the payment success page
	req, _ := http.NewRequest("GET", "/payment/success?session_id=cs_test_123", nil)
	// Add authentication cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the owner page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/owner", w.Header().Get("Location"))
}

// TestPaymentCancellation tests that payment cancellations are handled correctly
func TestPaymentCancellation(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for payment cancellation
	router.GET("/payment/cancel", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.HandlePaymentCancellation(c)
	})

	// Test accessing the payment cancellation page
	req, _ := http.NewRequest("GET", "/payment/cancel", nil)
	// Add authentication cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the pricing page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/pricing", w.Header().Get("Location"))
}
