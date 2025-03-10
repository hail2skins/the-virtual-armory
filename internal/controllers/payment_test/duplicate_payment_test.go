package payment_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/middleware"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// init function to reset webhook stats before tests run
func init() {
	// Reset webhook stats to avoid affecting other tests
	middleware.ResetWebhookStats()
}

// TestRealWorldPaymentScenario simulates a real-world payment scenario where both
// checkout.session.completed and invoice.paid events are received for the same subscription.
// This test verifies that only one payment record is created in the database.
func TestRealWorldPaymentScenario(t *testing.T) {
	// Reset webhook stats to avoid affecting other tests
	middleware.ResetWebhookStats()

	// Setup test environment
	gin.SetMode(gin.TestMode)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Create a payment controller
	paymentController := controllers.NewPaymentController(db)

	// Create a router with the webhook handler
	router := gin.Default()
	router.POST("/webhook", middleware.WebhookMonitor(), paymentController.HandleStripeWebhook)

	// Update the user's stripe_customer_id to match the webhook
	db.Model(&user).Update("stripe_customer_id", "cus_test_customer")

	// Convert user ID to string for the client_reference_id
	userIDString := fmt.Sprintf("%d", user.ID)

	// Step 1: Send a checkout.session.completed webhook event
	checkoutEvent := map[string]interface{}{
		"id":     "evt_test_checkout",
		"object": "event",
		"type":   "checkout.session.completed",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":           "cs_test_checkout",
				"object":       "checkout.session",
				"amount_total": 500, // $5 monthly subscription
				"currency":     "usd",
				"customer": map[string]interface{}{
					"id": "cus_test_customer",
				},
				"client_reference_id": userIDString,
			},
		},
	}

	checkoutEventJSON, _ := json.Marshal(checkoutEvent)
	req1, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(checkoutEventJSON))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Stripe-Signature", "test_signature") // Add test signature
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// After the checkout.session.completed event, check how many payment records exist
	var paymentsAfterCheckout []models.Payment
	db.Where("user_id = ?", user.ID).Find(&paymentsAfterCheckout)

	// There should be exactly one payment record after checkout.session.completed
	assert.Equal(t, 1, len(paymentsAfterCheckout), "Expected exactly one payment record after checkout.session.completed, but found %d", len(paymentsAfterCheckout))

	// Step 2: Send an invoice.paid webhook event for the same subscription
	invoiceEvent := map[string]interface{}{
		"id":     "evt_test_invoice",
		"object": "event",
		"type":   "invoice.paid",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":          "in_test_invoice",
				"object":      "invoice",
				"amount_paid": 500, // $5 monthly subscription
				"currency":    "usd",
				"customer": map[string]interface{}{
					"id": "cus_test_customer",
				},
				"subscription": "sub_test_subscription",
			},
		},
	}

	invoiceEventJSON, _ := json.Marshal(invoiceEvent)
	req2, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(invoiceEventJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Stripe-Signature", "test_signature") // Add test signature
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Step 3: Verify that still only one payment record exists after both events
	var paymentsAfterInvoice []models.Payment
	db.Where("user_id = ?", user.ID).Find(&paymentsAfterInvoice)

	// Assert that there is still exactly one payment record after both events
	assert.Equal(t, 1, len(paymentsAfterInvoice), "Expected exactly one payment record after both events, but found %d", len(paymentsAfterInvoice))

	if len(paymentsAfterInvoice) > 0 {
		// Verify the payment details
		assert.Equal(t, int64(500), paymentsAfterInvoice[0].Amount)
		assert.Equal(t, "usd", paymentsAfterInvoice[0].Currency)
		assert.Equal(t, "subscription", paymentsAfterInvoice[0].PaymentType)
		assert.Equal(t, "succeeded", paymentsAfterInvoice[0].Status)
		assert.Equal(t, "Monthly Subscription", paymentsAfterInvoice[0].Description)
	}

	// Reset webhook stats again after the test
	middleware.ResetWebhookStats()
}

// TestSinglePaymentForSubscription verifies that only one payment record is created
// when a subscription is purchased, even if both checkout.session.completed and
// invoice.paid events are received
func TestSinglePaymentForSubscription(t *testing.T) {
	// Reset webhook stats to avoid affecting other tests
	middleware.ResetWebhookStats()

	// Setup test environment
	gin.SetMode(gin.TestMode)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Create a payment controller
	paymentController := controllers.NewPaymentController(db)

	// Create a router with the webhook handler
	router := gin.Default()
	router.POST("/webhook", middleware.WebhookMonitor(), paymentController.HandleStripeWebhook)

	// Update the user's stripe_customer_id to match the webhook
	db.Model(&user).Update("stripe_customer_id", "cus_test_customer")

	// Convert user ID to string for the client_reference_id
	userIDString := fmt.Sprintf("%d", user.ID)

	// Step 1: Send a checkout.session.completed webhook event
	checkoutEvent := map[string]interface{}{
		"id":     "evt_test_checkout",
		"object": "event",
		"type":   "checkout.session.completed",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":           "cs_test_checkout",
				"object":       "checkout.session",
				"amount_total": 500, // $5 monthly subscription
				"currency":     "usd",
				"customer": map[string]interface{}{
					"id": "cus_test_customer",
				},
				"client_reference_id": userIDString,
			},
		},
	}

	checkoutEventJSON, _ := json.Marshal(checkoutEvent)
	req1, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(checkoutEventJSON))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Stripe-Signature", "test_signature") // Add test signature
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Step 2: Send an invoice.paid webhook event for the same subscription
	invoiceEvent := map[string]interface{}{
		"id":     "evt_test_invoice",
		"object": "event",
		"type":   "invoice.paid",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":          "in_test_invoice",
				"object":      "invoice",
				"amount_paid": 500, // $5 monthly subscription
				"currency":    "usd",
				"customer": map[string]interface{}{
					"id": "cus_test_customer",
				},
				"subscription": "sub_test_subscription",
			},
		},
	}

	invoiceEventJSON, _ := json.Marshal(invoiceEvent)
	req2, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(invoiceEventJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Stripe-Signature", "test_signature") // Add test signature
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Step 3: Verify that only one payment record was created
	var payments []models.Payment
	db.Where("user_id = ?", user.ID).Find(&payments)

	// Assert that there is exactly one payment record
	assert.Equal(t, 1, len(payments), "Expected exactly one payment record, but found %d", len(payments))

	if len(payments) > 0 {
		// Verify the payment details
		assert.Equal(t, int64(500), payments[0].Amount)
		assert.Equal(t, "usd", payments[0].Currency)
		assert.Equal(t, "subscription", payments[0].PaymentType)
		assert.Equal(t, "succeeded", payments[0].Status)
		assert.Equal(t, "Monthly Subscription", payments[0].Description)
	}

	// Reset webhook stats again after the test
	middleware.ResetWebhookStats()
}

// TestSinglePaymentForUpgrade verifies that only one payment record is created
// when a subscription is upgraded, even if both checkout.session.completed and
// invoice.paid events are received
func TestSinglePaymentForUpgrade(t *testing.T) {
	// Reset webhook stats to avoid affecting other tests
	middleware.ResetWebhookStats()

	// Setup test environment
	gin.SetMode(gin.TestMode)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user with an existing monthly subscription
	user := payment_test_utils.CreateTestUser(t, db)

	// Set the user as having a monthly subscription
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": time.Now().AddDate(0, 1, 0), // 1 month from now
		"stripe_customer_id":      "cus_test_customer",
	})

	// Create a payment controller
	paymentController := controllers.NewPaymentController(db)

	// Create a router with the webhook handler
	router := gin.Default()
	router.POST("/webhook", middleware.WebhookMonitor(), paymentController.HandleStripeWebhook)

	// Convert user ID to string for the client_reference_id
	userIDString := fmt.Sprintf("%d", user.ID)

	// Step 1: Send a checkout.session.completed webhook event for a yearly subscription upgrade
	checkoutEvent := map[string]interface{}{
		"id":     "evt_test_checkout_yearly",
		"object": "event",
		"type":   "checkout.session.completed",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":           "cs_test_checkout_yearly",
				"object":       "checkout.session",
				"amount_total": 3000, // $30 yearly subscription
				"currency":     "usd",
				"customer": map[string]interface{}{
					"id": "cus_test_customer",
				},
				"client_reference_id": userIDString,
			},
		},
	}

	checkoutEventJSON, _ := json.Marshal(checkoutEvent)
	req1, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(checkoutEventJSON))
	req1.Header.Set("Content-Type", "application/json")
	req1.Header.Set("Stripe-Signature", "test_signature") // Add test signature
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)
	assert.Equal(t, http.StatusOK, w1.Code)

	// Step 2: Send an invoice.paid webhook event for the same subscription upgrade
	invoiceEvent := map[string]interface{}{
		"id":     "evt_test_invoice_yearly",
		"object": "event",
		"type":   "invoice.paid",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":          "in_test_invoice_yearly",
				"object":      "invoice",
				"amount_paid": 3000, // $30 yearly subscription
				"currency":    "usd",
				"customer": map[string]interface{}{
					"id": "cus_test_customer",
				},
				"subscription": "sub_test_subscription_yearly",
			},
		},
	}

	invoiceEventJSON, _ := json.Marshal(invoiceEvent)
	req2, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(invoiceEventJSON))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set("Stripe-Signature", "test_signature") // Add test signature
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)
	assert.Equal(t, http.StatusOK, w2.Code)

	// Step 3: Verify that only one payment record was created for the upgrade
	var payments []models.Payment
	db.Where("user_id = ? AND amount = ?", user.ID, 3000).Find(&payments)

	// Assert that there is exactly one payment record for the upgrade
	assert.Equal(t, 1, len(payments), "Expected exactly one payment record for the upgrade, but found %d", len(payments))

	if len(payments) > 0 {
		// Verify the payment details
		assert.Equal(t, int64(3000), payments[0].Amount)
		assert.Equal(t, "usd", payments[0].Currency)
		assert.Equal(t, "subscription", payments[0].PaymentType)
		assert.Equal(t, "succeeded", payments[0].Status)
		assert.Equal(t, "Yearly Subscription", payments[0].Description)
	}

	// Step 4: Verify that the user's subscription was updated correctly
	var updatedUser models.User
	db.First(&updatedUser, user.ID)
	assert.Equal(t, "yearly", updatedUser.SubscriptionTier)

	// Reset webhook stats again after the test
	middleware.ResetWebhookStats()
}
