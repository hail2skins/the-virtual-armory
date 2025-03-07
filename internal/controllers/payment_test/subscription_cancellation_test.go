package payment_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestCancelSubscriptionButton tests that the cancel subscription button is shown only for monthly/yearly subscriptions
func TestCancelSubscriptionButton(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Test cases for different subscription tiers
	testCases := []struct {
		name                   string
		subscriptionTier       string
		shouldShowCancelButton bool
	}{
		{"Free Tier", "free", false},
		{"Monthly Subscription", "monthly", true},
		{"Yearly Subscription", "yearly", true},
		{"Lifetime Subscription", "lifetime", false},
		{"Premium Lifetime Subscription", "premium_lifetime", false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test user with the specified subscription tier
			user := payment_test_utils.CreateTestUser(t, db)
			db.Model(&user).Updates(map[string]interface{}{
				"subscription_tier":       tc.subscriptionTier,
				"subscription_expires_at": time.Now().AddDate(1, 0, 0),
				"stripe_customer_id":      "cus_test_" + tc.subscriptionTier,
			})

			// Create some test payments
			payment := models.Payment{
				UserID:      user.ID,
				Amount:      1000,
				Currency:    "usd",
				PaymentType: "subscription",
				Status:      "succeeded",
				Description: tc.subscriptionTier + " Subscription",
				StripeID:    "pi_test_123",
			}
			db.Create(&payment)

			// Set up the router
			router := gin.Default()

			// Set up the controller
			paymentController := controllers.NewPaymentController(db)

			// Set up the route
			router.GET("/owner/payment-history", func(c *gin.Context) {
				// Set authentication cookies for the test
				c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
				c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

				paymentController.ShowPaymentHistory(c)
			})

			// Test accessing the payment history page
			req, _ := http.NewRequest("GET", "/owner/payment-history", nil)
			req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
			req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			// Check that the response is successful
			assert.Equal(t, http.StatusOK, w.Code)

			// Check for the presence of the cancel button based on subscription tier
			if tc.shouldShowCancelButton {
				assert.Contains(t, w.Body.String(), "Cancel Subscription")
			} else {
				assert.NotContains(t, w.Body.String(), "Cancel Subscription")
			}
		})
	}
}

// TestCancelSubscriptionEndpoint tests the subscription cancellation endpoint
func TestCancelSubscriptionEndpoint(t *testing.T) {
	// Set test environment
	t.Setenv("APP_ENV", "test")

	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user with an active subscription
	user := payment_test_utils.CreateTestUser(t, db)
	expirationDate := time.Now().AddDate(0, 1, 0) // 1 month from now
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": expirationDate,
		"stripe_customer_id":      "cus_test_monthly",
	})

	// Set up the router
	router := gin.Default()

	// Set up the controller
	paymentController := controllers.NewPaymentController(db)

	// Set up the route
	router.POST("/subscription/cancel", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.CancelSubscription(c)
	})

	// Test accessing the cancel subscription endpoint
	req, _ := http.NewRequest("POST", "/subscription/cancel", nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the payment history page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/owner/payment-history", w.Header().Get("Location"))

	// Check that the flash message cookie is set
	cookies := w.Result().Cookies()
	var flashMessageFound, flashTypeFound bool
	for _, cookie := range cookies {
		if cookie.Name == "flash_message" {
			assert.Contains(t, cookie.Value, "canceled")
			flashMessageFound = true
		}
		if cookie.Name == "flash_type" {
			assert.Equal(t, "success", cookie.Value)
			flashTypeFound = true
		}
	}
	assert.True(t, flashMessageFound, "Flash message cookie not found")
	assert.True(t, flashTypeFound, "Flash type cookie not found")

	// Reload the user from the database
	var updatedUser models.User
	db.First(&updatedUser, user.ID)

	// Check that the user's subscription is marked as will not renew but still active
	assert.Equal(t, "monthly", updatedUser.SubscriptionTier)
	assert.Equal(t, expirationDate.Format("2006-01-02"), updatedUser.SubscriptionExpiresAt.Format("2006-01-02"))
	assert.True(t, updatedUser.SubscriptionCanceled)
}

// TestCancelSubscriptionWithStripe tests the integration with Stripe for subscription cancellation
func TestCancelSubscriptionWithStripe(t *testing.T) {
	// Skip this test in test environment since we can't make real Stripe API calls
	t.Skip("Skipping test that requires Stripe API integration")

	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user with a Stripe subscription
	user := payment_test_utils.CreateTestUser(t, db)
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": time.Now().AddDate(0, 1, 0),
		"stripe_customer_id":      "cus_test_stripe",
		"stripe_subscription_id":  "sub_test_123",
	})

	// Set up the router
	router := gin.Default()

	// Set up the controller
	paymentController := controllers.NewPaymentController(db)

	// Set up the route
	router.POST("/subscription/cancel", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.CancelSubscription(c)
	})

	// Test accessing the cancel subscription endpoint
	req, _ := http.NewRequest("POST", "/subscription/cancel", nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the payment history page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/owner/payment-history", w.Header().Get("Location"))

	// Reload the user from the database
	var updatedUser models.User
	db.First(&updatedUser, user.ID)

	// Check that the user's subscription is marked as canceled
	assert.True(t, updatedUser.SubscriptionCanceled)
}

// TestCancelSubscriptionConfirmationPage tests the confirmation page for subscription cancellation
func TestCancelSubscriptionConfirmationPage(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user with an active subscription
	user := payment_test_utils.CreateTestUser(t, db)
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": time.Now().AddDate(0, 1, 0),
		"stripe_customer_id":      "cus_test_monthly",
	})

	// Set up the router
	router := gin.Default()

	// Set up the controller
	paymentController := controllers.NewPaymentController(db)

	// Set up the route
	router.GET("/subscription/cancel/confirm", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.ShowCancelConfirmation(c)
	})

	// Test accessing the cancel confirmation page
	req, _ := http.NewRequest("GET", "/subscription/cancel/confirm", nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the response is successful
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the page contains the expected content
	assert.Contains(t, w.Body.String(), "Cancel Subscription")
	assert.Contains(t, w.Body.String(), "Are you sure you want to cancel")
	assert.Contains(t, w.Body.String(), "You will continue to have access until")
	assert.Contains(t, w.Body.String(), user.SubscriptionExpiresAt.Format("January 2, 2006"))
}
