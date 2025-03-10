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

// TestEndToEndPaymentFlow tests the complete payment flow from pricing page to checkout to success
// This test ensures that the payment flow works correctly in test mode
func TestEndToEndPaymentFlow(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Record initial subscription state
	initialTier := user.SubscriptionTier
	initialExpiry := user.SubscriptionExpiresAt

	// Set up test router and controller
	router := gin.New()
	paymentController := controllers.NewPaymentController(db)

	// Set up the routes
	router.GET("/pricing", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		// Use the mock pricing page instead of the real one to avoid template issues
		payment_test_utils.MockPricingPage(c, user)
	})

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

	// STEP 1: Visit the pricing page
	t.Log("STEP 1: Visiting pricing page")
	req, _ := http.NewRequest("GET", "/pricing", nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the pricing page is displayed
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Choose Your Plan")

	// Verify that the pricing page contains checkout links, not direct Stripe links
	assert.Contains(t, w.Body.String(), "/checkout?tier=monthly")
	assert.Contains(t, w.Body.String(), "/checkout?tier=yearly")
	assert.Contains(t, w.Body.String(), "/checkout?tier=lifetime")
	assert.NotContains(t, w.Body.String(), "buy.stripe.com")

	// STEP 2: Submit checkout form for monthly subscription
	t.Log("STEP 2: Creating checkout session")
	form := url.Values{}
	form.Add("tier", "monthly")
	req, _ = http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the success page with a test session ID
	assert.Equal(t, http.StatusSeeOther, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "/payment/success?session_id=cs_test_")

	// Extract the session ID from the redirect URL
	sessionID := strings.Split(redirectURL, "session_id=")[1]
	t.Logf("Got session ID: %s", sessionID)

	// STEP 3: Follow the redirect to the success page
	t.Log("STEP 3: Handling payment success")
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the owner page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/owner", w.Header().Get("Location"))

	// STEP 4: Verify that the user's subscription was updated
	t.Log("STEP 4: Verifying subscription update")
	var updatedUser models.User
	err := db.First(&updatedUser, user.ID).Error
	assert.NoError(t, err)

	// Check that the subscription tier was updated to monthly
	assert.Equal(t, "monthly", updatedUser.SubscriptionTier)
	assert.NotEqual(t, initialTier, updatedUser.SubscriptionTier)

	// Check that the subscription expiry was updated
	assert.True(t, updatedUser.SubscriptionExpiresAt.After(initialExpiry))

	// Check that the subscription is not marked as canceled
	assert.False(t, updatedUser.SubscriptionCanceled)

	// STEP 5: Verify that a payment record was created
	t.Log("STEP 5: Verifying payment record")
	var payment models.Payment
	err = db.Where("user_id = ?", user.ID).First(&payment).Error
	assert.NoError(t, err)

	// Check payment details
	assert.Equal(t, user.ID, payment.UserID)
	assert.Equal(t, "subscription", payment.PaymentType)
	assert.Equal(t, "succeeded", payment.Status)
	assert.Contains(t, payment.Description, "Monthly")
	assert.Contains(t, payment.StripeID, "cs_test_")
}

// TestUpgradeSubscriptionFlow tests upgrading from one subscription to another
func TestUpgradeSubscriptionFlow(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test user with an existing subscription
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up an existing subscription
	expiryDate := time.Now().AddDate(0, 1, 0) // 1 month from now
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": expiryDate,
		"stripe_customer_id":      "cus_existing",
	})

	// Reload the user to get the updated values
	db.First(&user, user.ID)
	initialExpiry := user.SubscriptionExpiresAt

	// Set up test router and controller
	router := gin.New()
	paymentController := controllers.NewPaymentController(db)

	// Set up the routes
	router.POST("/checkout", func(c *gin.Context) {
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)
		paymentController.CreateCheckoutSession(c)
	})

	router.GET("/payment/success", func(c *gin.Context) {
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)
		paymentController.HandlePaymentSuccess(c)
	})

	// STEP 1: Test upgrading to yearly subscription
	t.Log("STEP 1: Upgrading to yearly subscription")
	form := url.Values{}
	form.Add("tier", "yearly")
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract the session ID from the redirect URL
	redirectURL := w.Header().Get("Location")
	sessionID := strings.Split(redirectURL, "session_id=")[1]

	// STEP 2: Test the payment success handler
	t.Log("STEP 2: Handling payment success for upgrade")

	// In test mode, we need to manually add the tier to the URL
	// This simulates the metadata that would be in the Stripe session
	successURL := "/payment/success?session_id=" + sessionID + "_yearly"
	req, _ = http.NewRequest("GET", successURL, nil)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// STEP 3: Verify that the user's subscription was updated
	t.Log("STEP 3: Verifying subscription upgrade")
	var updatedUser models.User
	err := db.First(&updatedUser, user.ID).Error
	assert.NoError(t, err)

	// Check that the subscription tier was updated to yearly
	assert.Equal(t, "yearly", updatedUser.SubscriptionTier)

	// Check that the subscription expiry was extended
	assert.True(t, updatedUser.SubscriptionExpiresAt.After(initialExpiry))

	// The new expiry should be at least 11 months after the initial expiry
	// (1 year from now minus the 1 month that was already there)
	minExpectedExpiry := initialExpiry.AddDate(0, 11, 0)
	assert.True(t, updatedUser.SubscriptionExpiresAt.After(minExpectedExpiry) ||
		updatedUser.SubscriptionExpiresAt.Equal(minExpectedExpiry))
}
