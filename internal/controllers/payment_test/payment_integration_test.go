package payment_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestPaymentIntegration tests the complete payment flow
func TestPaymentIntegration(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Verify user starts with free tier
	assert.Equal(t, "free", user.SubscriptionTier)

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the routes
	router.POST("/checkout", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.CreateCheckoutSession(c)
	})

	router.GET("/payment/success", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.HandlePaymentSuccess(c)
	})

	// Step 1: Submit checkout form for monthly subscription
	formData := url.Values{
		"tier": {"monthly"},
	}
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that we're redirected to the success page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "/payment/success?session_id=cs_test_")

	// Step 2: Follow the redirect to the success page
	req, _ = http.NewRequest("GET", redirectURL, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the owner page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/owner", w.Header().Get("Location"))

	// Step 3: Verify that the user's subscription has been updated
	var updatedUser models.User
	err := db.First(&updatedUser, user.ID).Error
	assert.NoError(t, err)

	// Check that the subscription tier has been updated to monthly
	assert.Equal(t, "monthly", updatedUser.SubscriptionTier)

	// Check that the subscription expiration date is in the future
	assert.True(t, updatedUser.SubscriptionExpiresAt.After(time.Now()))

	// Step 4: Verify that a payment record was created
	var payments []models.Payment
	err = db.Where("user_id = ?", user.ID).Find(&payments).Error
	assert.NoError(t, err)
	assert.Equal(t, 1, len(payments))

	// Check payment details
	payment := payments[0]
	assert.Equal(t, user.ID, payment.UserID)
	assert.Equal(t, int64(500), payment.Amount) // $5.00
	assert.Equal(t, "usd", payment.Currency)
	assert.Equal(t, "subscription", payment.PaymentType)
	assert.Equal(t, "succeeded", payment.Status)
	assert.Contains(t, payment.Description, "monthly")
}

// TestPaymentHistoryPage tests the payment history page
func TestPaymentHistoryPage(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Create a test payment
	payment := models.Payment{
		UserID:      user.ID,
		Amount:      500, // $5.00
		Currency:    "usd",
		PaymentType: "subscription",
		Status:      "succeeded",
		Description: "Monthly Subscription",
		StripeID:    "pi_test_123",
	}
	err := db.Create(&payment).Error
	assert.NoError(t, err)

	// Set up test router and controller
	router := gin.Default()
	paymentController := controllers.NewPaymentController(db)

	// Set up the route for payment history
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

	// Check that the payment history page is displayed
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Payment History")
	assert.Contains(t, w.Body.String(), "Monthly Subscription")
	assert.Contains(t, w.Body.String(), "$5.00")
}
