package payment_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/stretchr/testify/assert"
)

// setupPaymentTest sets up the test environment for payment tests
func setupPaymentTest(t *testing.T) (*gin.Engine, *controllers.PaymentController, *models.User) {
	// Setup
	gin.SetMode(gin.TestMode)
	db := payment_test_utils.SetupTestDB(t)

	// Create a test user with a unique email
	user := payment_test_utils.CreateTestUser(t, db)

	// Set the mock user for authentication
	auth.MockUser = user

	// Create a payment controller
	paymentController := controllers.NewPaymentController(db)

	// Create a test router
	router := gin.Default()

	// Set HTML renderer
	router.HTMLRender = &payment_test_utils.TestRenderer{}

	return router, paymentController, user
}

// Cleanup function to reset MockUser after each test
func cleanup() {
	auth.MockUser = nil
}

func TestShowPricingPage(t *testing.T) {
	// Setup
	router, paymentController, _ := setupPaymentTest(t)
	defer cleanup()

	// Setup the route
	router.GET("/pricing", paymentController.ShowPricingPage)

	// Test case 1: Not logged in
	auth.MockUser = nil // Explicitly set to nil for this test
	req1, err := http.NewRequest("GET", "/pricing", nil)
	assert.NoError(t, err)

	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	// Check the status code
	assert.Equal(t, http.StatusOK, w1.Code)
	// Just verify that the response is not empty
	assert.NotEmpty(t, w1.Body.String())

	// Test case 2: Logged in
	// Restore the mock user by setting up a new test
	_, _, user := setupPaymentTest(t)
	auth.MockUser = user
	req2, err := http.NewRequest("GET", "/pricing", nil)
	assert.NoError(t, err)

	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	// Check the status code
	assert.Equal(t, http.StatusOK, w2.Code)
	// Just verify that the response is not empty
	assert.NotEmpty(t, w2.Body.String())
}

func TestCreateCheckoutSession(t *testing.T) {
	// Setup
	router, paymentController, user := setupPaymentTest(t)
	defer cleanup()

	// Save original environment variables
	originalAppEnv := os.Getenv("APP_ENV")
	originalBaseURL := os.Getenv("APP_BASE_URL")

	// Set environment variables for testing
	os.Setenv("APP_ENV", "test")
	os.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Restore environment variables after the test
	defer func() {
		os.Setenv("APP_ENV", originalAppEnv)
		os.Setenv("APP_BASE_URL", originalBaseURL)
	}()

	// Setup the route
	router.POST("/checkout", paymentController.CreateCheckoutSession)

	// Create form data for each subscription tier
	tiers := []string{"monthly", "yearly", "lifetime", "premium_lifetime"}

	for _, tier := range tiers {
		form := url.Values{}
		form.Add("tier", tier)

		// Create a request
		req, err := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
		assert.NoError(t, err)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

		// Create a response recorder
		w := httptest.NewRecorder()

		// Perform the request
		router.ServeHTTP(w, req)

		// Check the status code (should be a redirect)
		assert.Equal(t, http.StatusSeeOther, w.Code)

		// In test mode, it should redirect to a test URL with the user ID
		expectedURL := "http://localhost:3000/payment/success?session_id=cs_test_" + strconv.FormatUint(uint64(user.ID), 10)
		assert.Equal(t, expectedURL, w.Header().Get("Location"))
	}
}

func TestHandlePaymentSuccess(t *testing.T) {
	// Setup
	router, paymentController, user := setupPaymentTest(t)
	defer cleanup()

	// Setup the route
	router.GET("/payment/success", paymentController.HandlePaymentSuccess)

	// Create a request with a test session ID
	sessionID := "cs_test_" + strconv.FormatUint(uint64(user.ID), 10)
	req, err := http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code (should be a redirect or OK)
	assert.True(t, w.Code == http.StatusSeeOther || w.Code == http.StatusOK)
}

func TestHandlePaymentCancellation(t *testing.T) {
	// Setup
	router, paymentController, _ := setupPaymentTest(t)
	defer cleanup()

	// Setup the route
	router.GET("/payment/cancel", paymentController.HandlePaymentCancellation)

	// Create a request
	req, err := http.NewRequest("GET", "/payment/cancel", nil)
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code (should be a redirect or OK)
	assert.True(t, w.Code == http.StatusSeeOther || w.Code == http.StatusOK)
}

func TestShowPaymentHistory(t *testing.T) {
	// Setup
	router, paymentController, user := setupPaymentTest(t)
	defer cleanup()

	// Create some test payments for the user
	payment1 := models.Payment{
		UserID:      user.ID,
		Amount:      500, // $5.00
		Currency:    "usd",
		PaymentType: "subscription",
		Status:      "succeeded",
		Description: "Monthly subscription",
		StripeID:    "pi_test_1",
	}
	assert.NoError(t, database.DB.Create(&payment1).Error)

	payment2 := models.Payment{
		UserID:      user.ID,
		Amount:      3000, // $30.00
		Currency:    "usd",
		PaymentType: "subscription",
		Status:      "succeeded",
		Description: "Yearly subscription",
		StripeID:    "pi_test_2",
	}
	assert.NoError(t, database.DB.Create(&payment2).Error)

	// Setup the route
	router.GET("/owner/payment-history", paymentController.ShowPaymentHistory)

	// Create a request
	req, err := http.NewRequest("GET", "/owner/payment-history", nil)
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Just verify that the response is not empty
	assert.NotEmpty(t, w.Body.String())
}

func TestShowCancelConfirmation(t *testing.T) {
	// Setup
	router, paymentController, user := setupPaymentTest(t)
	defer cleanup()

	// Update the user to have a subscription
	user.SubscriptionTier = "monthly"
	user.StripeSubscriptionID = "sub_test_123"
	assert.NoError(t, database.DB.Save(&user).Error)

	// Setup the route
	router.GET("/subscription/cancel/confirm", paymentController.ShowCancelConfirmation)

	// Create a request
	req, err := http.NewRequest("GET", "/subscription/cancel/confirm", nil)
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Just verify that the response is not empty
	assert.NotEmpty(t, w.Body.String())
}

func TestCancelSubscription(t *testing.T) {
	// Setup
	router, paymentController, user := setupPaymentTest(t)
	defer cleanup()

	// Update the user to have a subscription
	user.SubscriptionTier = "monthly"
	user.StripeSubscriptionID = "sub_test_123"
	assert.NoError(t, database.DB.Save(&user).Error)

	// Save original environment variables
	originalAppEnv := os.Getenv("APP_ENV")

	// Set environment variables for testing
	os.Setenv("APP_ENV", "test")

	// Restore environment variables after the test
	defer func() {
		os.Setenv("APP_ENV", originalAppEnv)
	}()

	// Setup the route
	router.POST("/subscription/cancel", paymentController.CancelSubscription)

	// Create a request
	req, err := http.NewRequest("POST", "/subscription/cancel", nil)
	assert.NoError(t, err)

	// Create a response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code (should be a redirect)
	assert.Equal(t, http.StatusSeeOther, w.Code)

	// Verify the user's subscription was marked as canceled
	var updatedUser models.User
	assert.NoError(t, database.DB.First(&updatedUser, user.ID).Error)
	assert.True(t, updatedUser.SubscriptionCanceled)
}
