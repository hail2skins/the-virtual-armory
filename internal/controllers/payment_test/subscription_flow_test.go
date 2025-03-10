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

// TestMonthlySubscriptionFlow tests the complete flow for subscribing to the monthly plan
func TestMonthlySubscriptionFlow(t *testing.T) {
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
	router := gin.New()
	paymentController := controllers.NewPaymentController(db)

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

	router.GET("/owner", func(c *gin.Context) {
		// Mock owner page for testing
		c.String(http.StatusOK, "Owner Page")
	})

	router.GET("/pricing", func(c *gin.Context) {
		// Mock pricing page for testing
		c.String(http.StatusOK, "Pricing Page")
	})

	router.GET("/owner/guns", func(c *gin.Context) {
		// Mock owner/guns page for testing
		c.String(http.StatusOK, "Owner Guns Page")
	})

	// STEP 1: Test creating a checkout session for monthly subscription
	t.Log("STEP 1: Creating checkout session for monthly subscription")
	form := url.Values{}
	form.Add("tier", "monthly")
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the success page with a test session ID
	assert.Equal(t, http.StatusSeeOther, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "/payment/success?session_id=cs_test_")

	// Extract the session ID from the redirect URL and add tier information
	sessionID := strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_monthly" // Add tier to session ID for test mode

	// STEP 2: Test handling the payment success
	t.Log("STEP 2: Handling payment success")
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the owner page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/owner")

	// STEP 3: Verify the user's subscription was updated correctly
	t.Log("STEP 3: Verifying user subscription")
	var updatedUser models.User
	err := db.First(&updatedUser, user.ID).Error
	assert.NoError(t, err)

	// Check subscription tier
	assert.Equal(t, "monthly", updatedUser.SubscriptionTier)

	// Check expiration date (should be approximately 30 days from now)
	expectedExpiry := time.Now().AddDate(0, 1, 0)
	assert.WithinDuration(t, expectedExpiry, updatedUser.SubscriptionExpiresAt, 5*time.Minute)

	// STEP 4: Verify only ONE payment record was created
	t.Log("STEP 4: Verifying payment records")
	var payments []models.Payment
	err = db.Where("user_id = ?", user.ID).Find(&payments).Error
	assert.NoError(t, err)

	// Should only have one payment record
	assert.Equal(t, 1, len(payments))

	// Verify payment details
	assert.Equal(t, "Monthly Subscription", payments[0].Description)
	assert.Equal(t, int64(500), payments[0].Amount) // $5.00
	assert.Equal(t, "subscription", payments[0].PaymentType)
	assert.Equal(t, "succeeded", payments[0].Status)
}

// TestYearlySubscriptionFlow tests the complete flow for subscribing to the yearly plan
func TestYearlySubscriptionFlow(t *testing.T) {
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
	router := gin.New()
	paymentController := controllers.NewPaymentController(db)

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

	router.GET("/owner", func(c *gin.Context) {
		// Mock owner page for testing
		c.String(http.StatusOK, "Owner Page")
	})

	router.GET("/pricing", func(c *gin.Context) {
		// Mock pricing page for testing
		c.String(http.StatusOK, "Pricing Page")
	})

	router.GET("/owner/guns", func(c *gin.Context) {
		// Mock owner/guns page for testing
		c.String(http.StatusOK, "Owner Guns Page")
	})

	// STEP 1: Test creating a checkout session for yearly subscription
	t.Log("STEP 1: Creating checkout session for yearly subscription")
	form := url.Values{}
	form.Add("tier", "yearly")
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the success page with a test session ID
	assert.Equal(t, http.StatusSeeOther, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "/payment/success?session_id=cs_test_")

	// Extract the session ID from the redirect URL and add tier information
	sessionID := strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_yearly" // Add tier to session ID for test mode

	// STEP 2: Test handling the payment success
	t.Log("STEP 2: Handling payment success")
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the owner page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/owner")

	// STEP 3: Verify the user's subscription was updated correctly
	t.Log("STEP 3: Verifying user subscription")
	var updatedUser models.User
	err := db.First(&updatedUser, user.ID).Error
	assert.NoError(t, err)

	// Check subscription tier
	assert.Equal(t, "yearly", updatedUser.SubscriptionTier)

	// Check expiration date (should be approximately 1 year from now)
	expectedExpiry := time.Now().AddDate(1, 0, 0)
	assert.WithinDuration(t, expectedExpiry, updatedUser.SubscriptionExpiresAt, 5*time.Minute)

	// STEP 4: Verify only ONE payment record was created
	t.Log("STEP 4: Verifying payment records")
	var payments []models.Payment
	err = db.Where("user_id = ?", user.ID).Find(&payments).Error
	assert.NoError(t, err)

	// Should only have one payment record
	assert.Equal(t, 1, len(payments))

	// Verify payment details
	assert.Equal(t, "Yearly Subscription", payments[0].Description)
	assert.Equal(t, int64(3000), payments[0].Amount) // $30.00
	assert.Equal(t, "subscription", payments[0].PaymentType)
	assert.Equal(t, "succeeded", payments[0].Status)
}

// TestLifetimeSubscriptionFlow tests the complete flow for subscribing to the lifetime plan
func TestLifetimeSubscriptionFlow(t *testing.T) {
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
	router := gin.New()
	paymentController := controllers.NewPaymentController(db)

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

	router.GET("/owner", func(c *gin.Context) {
		// Mock owner page for testing
		c.String(http.StatusOK, "Owner Page")
	})

	router.GET("/pricing", func(c *gin.Context) {
		// Mock pricing page for testing
		c.String(http.StatusOK, "Pricing Page")
	})

	router.GET("/owner/guns", func(c *gin.Context) {
		// Mock owner/guns page for testing
		c.String(http.StatusOK, "Owner Guns Page")
	})

	// STEP 1: Test creating a checkout session for lifetime subscription
	t.Log("STEP 1: Creating checkout session for lifetime subscription")
	form := url.Values{}
	form.Add("tier", "lifetime")
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the success page with a test session ID
	assert.Equal(t, http.StatusSeeOther, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "/payment/success?session_id=cs_test_")

	// Extract the session ID from the redirect URL and add tier information
	sessionID := strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_lifetime" // Add tier to session ID for test mode

	// STEP 2: Test handling the payment success
	t.Log("STEP 2: Handling payment success")
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the owner page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/owner")

	// STEP 3: Verify the user's subscription was updated correctly
	t.Log("STEP 3: Verifying user subscription")
	var updatedUser models.User
	err := db.First(&updatedUser, user.ID).Error
	assert.NoError(t, err)

	// Check subscription tier
	assert.Equal(t, "lifetime", updatedUser.SubscriptionTier)

	// Check expiration date (should be far in the future, effectively never expiring)
	// We use 99 years as a reasonable approximation for "never"
	farFuture := time.Now().AddDate(99, 0, 0)
	assert.True(t, updatedUser.SubscriptionExpiresAt.After(farFuture),
		"Lifetime subscription should have an expiration date far in the future")

	// STEP 4: Verify only ONE payment record was created
	t.Log("STEP 4: Verifying payment records")
	var payments []models.Payment
	err = db.Where("user_id = ?", user.ID).Find(&payments).Error
	assert.NoError(t, err)

	// Should only have one payment record
	assert.Equal(t, 1, len(payments))

	// Verify payment details
	assert.Equal(t, "Lifetime Subscription", payments[0].Description)
	assert.Equal(t, int64(15000), payments[0].Amount) // $150.00
	assert.Equal(t, "subscription", payments[0].PaymentType)
	assert.Equal(t, "succeeded", payments[0].Status)
}

// TestPremiumLifetimeSubscriptionFlow tests the complete flow for subscribing to the premium lifetime plan
func TestPremiumLifetimeSubscriptionFlow(t *testing.T) {
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
	router := gin.New()
	paymentController := controllers.NewPaymentController(db)

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

	router.GET("/owner", func(c *gin.Context) {
		// Mock owner page for testing
		c.String(http.StatusOK, "Owner Page")
	})

	router.GET("/pricing", func(c *gin.Context) {
		// Mock pricing page for testing
		c.String(http.StatusOK, "Pricing Page")
	})

	router.GET("/owner/guns", func(c *gin.Context) {
		// Mock owner/guns page for testing
		c.String(http.StatusOK, "Owner Guns Page")
	})

	// STEP 1: Test creating a checkout session for premium lifetime subscription
	t.Log("STEP 1: Creating checkout session for premium lifetime subscription")
	form := url.Values{}
	form.Add("tier", "premium_lifetime")
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the success page with a test session ID
	assert.Equal(t, http.StatusSeeOther, w.Code)
	redirectURL := w.Header().Get("Location")
	assert.Contains(t, redirectURL, "/payment/success?session_id=cs_test_")

	// Extract the session ID from the redirect URL and add tier information
	sessionID := strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_premium" // Add tier to session ID for test mode

	// STEP 2: Test handling the payment success
	t.Log("STEP 2: Handling payment success")
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to the owner page
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "/owner")

	// STEP 3: Verify the user's subscription was updated correctly
	t.Log("STEP 3: Verifying user subscription")
	var updatedUser models.User
	err := db.First(&updatedUser, user.ID).Error
	assert.NoError(t, err)

	// Check subscription tier
	assert.Equal(t, "premium_lifetime", updatedUser.SubscriptionTier)

	// Check expiration date (should be far in the future, effectively never expiring)
	// We use 99 years as a reasonable approximation for "never"
	farFuture := time.Now().AddDate(99, 0, 0)
	assert.True(t, updatedUser.SubscriptionExpiresAt.After(farFuture),
		"Premium lifetime subscription should have an expiration date far in the future")

	// STEP 4: Verify only ONE payment record was created
	t.Log("STEP 4: Verifying payment records")
	var payments []models.Payment
	err = db.Where("user_id = ?", user.ID).Find(&payments).Error
	assert.NoError(t, err)

	// Should only have one payment record
	assert.Equal(t, 1, len(payments))

	// Verify payment details
	assert.Equal(t, "Premium_lifetime Subscription", payments[0].Description)
	assert.Equal(t, int64(30000), payments[0].Amount) // $300.00
	assert.Equal(t, "subscription", payments[0].PaymentType)
	assert.Equal(t, "succeeded", payments[0].Status)
}

// TestUpgradeMonthlyToYearly tests upgrading from a monthly to a yearly subscription
func TestUpgradeMonthlyToYearly(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up test router and controller
	router := gin.New()
	paymentController := controllers.NewPaymentController(db)

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

	router.GET("/owner", func(c *gin.Context) {
		// Mock owner page for testing
		c.String(http.StatusOK, "Owner Page")
	})

	router.GET("/pricing", func(c *gin.Context) {
		// Mock pricing page for testing
		c.String(http.StatusOK, "Pricing Page")
	})

	router.GET("/owner/guns", func(c *gin.Context) {
		// Mock owner/guns page for testing
		c.String(http.StatusOK, "Owner Guns Page")
	})

	// STEP 1: Subscribe to monthly plan first
	t.Log("STEP 1: Subscribing to monthly plan")
	form := url.Values{}
	form.Add("tier", "monthly")
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract the session ID from the redirect URL and add tier information
	redirectURL := w.Header().Get("Location")
	sessionID := strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_monthly" // Add tier to session ID for test mode

	// Complete the monthly subscription
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify monthly subscription
	var monthlyUser models.User
	err := db.First(&monthlyUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "monthly", monthlyUser.SubscriptionTier)

	// Record the expiration date of the monthly subscription
	monthlyExpiryDate := monthlyUser.SubscriptionExpiresAt

	// STEP 2: Upgrade to yearly plan
	t.Log("STEP 2: Upgrading to yearly plan")
	form = url.Values{}
	form.Add("tier", "yearly")
	req, _ = http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract the session ID from the redirect URL and add tier information
	redirectURL = w.Header().Get("Location")
	sessionID = strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_yearly" // Add tier to session ID for test mode

	// Complete the yearly subscription
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// STEP 3: Verify the user's subscription was updated correctly
	t.Log("STEP 3: Verifying user subscription after upgrade")
	var yearlyUser models.User
	err = db.First(&yearlyUser, user.ID).Error
	assert.NoError(t, err)

	// Check subscription tier
	assert.Equal(t, "yearly", yearlyUser.SubscriptionTier)

	// Check expiration date (should be 1 year from the previous expiration date)
	expectedExpiry := monthlyExpiryDate.AddDate(1, 0, 0)
	assert.WithinDuration(t, expectedExpiry, yearlyUser.SubscriptionExpiresAt, 5*time.Minute,
		"Yearly subscription should extend 1 year from the previous monthly expiration date")

	// STEP 4: Verify payment records
	t.Log("STEP 4: Verifying payment records")
	var payments []models.Payment
	err = db.Where("user_id = ?", user.ID).Find(&payments).Error
	assert.NoError(t, err)

	// Should have two payment records (one for monthly, one for yearly)
	assert.Equal(t, 2, len(payments))

	// Check that we have one payment for each tier
	monthlyPayments := 0
	yearlyPayments := 0

	for _, payment := range payments {
		if payment.Description == "Monthly Subscription" {
			monthlyPayments++
			assert.Equal(t, int64(500), payment.Amount) // $5.00
		} else if payment.Description == "Yearly Subscription" {
			yearlyPayments++
			assert.Equal(t, int64(3000), payment.Amount) // $30.00
		}
	}

	assert.Equal(t, 1, monthlyPayments, "Should have exactly one monthly payment")
	assert.Equal(t, 1, yearlyPayments, "Should have exactly one yearly payment")
}

// TestUpgradeMonthlyToLifetime tests upgrading from a monthly to a lifetime subscription
func TestUpgradeMonthlyToLifetime(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up test router and controller
	router := gin.New()
	paymentController := controllers.NewPaymentController(db)

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

	router.GET("/owner", func(c *gin.Context) {
		// Mock owner page for testing
		c.String(http.StatusOK, "Owner Page")
	})

	router.GET("/pricing", func(c *gin.Context) {
		// Mock pricing page for testing
		c.String(http.StatusOK, "Pricing Page")
	})

	router.GET("/owner/guns", func(c *gin.Context) {
		// Mock owner/guns page for testing
		c.String(http.StatusOK, "Owner Guns Page")
	})

	// STEP 1: Subscribe to monthly plan first
	t.Log("STEP 1: Subscribing to monthly plan")
	form := url.Values{}
	form.Add("tier", "monthly")
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract the session ID from the redirect URL and add tier information
	redirectURL := w.Header().Get("Location")
	sessionID := strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_monthly" // Add tier to session ID for test mode

	// Complete the monthly subscription
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify monthly subscription
	var monthlyUser models.User
	err := db.First(&monthlyUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "monthly", monthlyUser.SubscriptionTier)

	// STEP 2: Upgrade to lifetime plan
	t.Log("STEP 2: Upgrading to lifetime plan")
	form = url.Values{}
	form.Add("tier", "lifetime")
	req, _ = http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract the session ID from the redirect URL and add tier information
	redirectURL = w.Header().Get("Location")
	sessionID = strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_lifetime" // Add tier to session ID for test mode

	// Complete the lifetime subscription
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// STEP 3: Verify the user's subscription was updated correctly
	t.Log("STEP 3: Verifying user subscription after upgrade")
	var lifetimeUser models.User
	err = db.First(&lifetimeUser, user.ID).Error
	assert.NoError(t, err)

	// Check subscription tier
	assert.Equal(t, "lifetime", lifetimeUser.SubscriptionTier)

	// Check expiration date (should be far in the future, effectively never expiring)
	// We use 99 years as a reasonable approximation for "never"
	farFuture := time.Now().AddDate(99, 0, 0)
	assert.True(t, lifetimeUser.SubscriptionExpiresAt.After(farFuture),
		"Lifetime subscription should have an expiration date far in the future")

	// STEP 4: Verify payment records
	t.Log("STEP 4: Verifying payment records")
	var payments []models.Payment
	err = db.Where("user_id = ?", user.ID).Find(&payments).Error
	assert.NoError(t, err)

	// Should have two payment records (one for monthly, one for lifetime)
	assert.Equal(t, 2, len(payments))

	// Check that we have one payment for each tier
	monthlyPayments := 0
	lifetimePayments := 0

	for _, payment := range payments {
		if payment.Description == "Monthly Subscription" {
			monthlyPayments++
			assert.Equal(t, int64(500), payment.Amount) // $5.00
		} else if payment.Description == "Lifetime Subscription" {
			lifetimePayments++
			assert.Equal(t, int64(15000), payment.Amount) // $150.00
		}
	}

	assert.Equal(t, 1, monthlyPayments, "Should have exactly one monthly payment")
	assert.Equal(t, 1, lifetimePayments, "Should have exactly one lifetime payment")
}

// TestUpgradeYearlyToLifetime tests upgrading from a yearly to a lifetime subscription
func TestUpgradeYearlyToLifetime(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up test router and controller
	router := gin.New()
	paymentController := controllers.NewPaymentController(db)

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

	router.GET("/owner", func(c *gin.Context) {
		// Mock owner page for testing
		c.String(http.StatusOK, "Owner Page")
	})

	router.GET("/pricing", func(c *gin.Context) {
		// Mock pricing page for testing
		c.String(http.StatusOK, "Pricing Page")
	})

	router.GET("/owner/guns", func(c *gin.Context) {
		// Mock owner/guns page for testing
		c.String(http.StatusOK, "Owner Guns Page")
	})

	// STEP 1: Subscribe to yearly plan first
	t.Log("STEP 1: Subscribing to yearly plan")
	form := url.Values{}
	form.Add("tier", "yearly")
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract the session ID from the redirect URL and add tier information
	redirectURL := w.Header().Get("Location")
	sessionID := strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_yearly" // Add tier to session ID for test mode

	// Complete the yearly subscription
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify yearly subscription
	var yearlyUser models.User
	err := db.First(&yearlyUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "yearly", yearlyUser.SubscriptionTier)

	// STEP 2: Upgrade to lifetime plan
	t.Log("STEP 2: Upgrading to lifetime plan")
	form = url.Values{}
	form.Add("tier", "lifetime")
	req, _ = http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract the session ID from the redirect URL and add tier information
	redirectURL = w.Header().Get("Location")
	sessionID = strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_lifetime" // Add tier to session ID for test mode

	// Complete the lifetime subscription
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// STEP 3: Verify the user's subscription was updated correctly
	t.Log("STEP 3: Verifying user subscription after upgrade")
	var lifetimeUser models.User
	err = db.First(&lifetimeUser, user.ID).Error
	assert.NoError(t, err)

	// Check subscription tier
	assert.Equal(t, "lifetime", lifetimeUser.SubscriptionTier)

	// Check expiration date (should be far in the future, effectively never expiring)
	// We use 99 years as a reasonable approximation for "never"
	farFuture := time.Now().AddDate(99, 0, 0)
	assert.True(t, lifetimeUser.SubscriptionExpiresAt.After(farFuture),
		"Lifetime subscription should have an expiration date far in the future")

	// STEP 4: Verify payment records
	t.Log("STEP 4: Verifying payment records")
	var payments []models.Payment
	err = db.Where("user_id = ?", user.ID).Find(&payments).Error
	assert.NoError(t, err)

	// Should have two payment records (one for yearly, one for lifetime)
	assert.Equal(t, 2, len(payments))

	// Check that we have one payment for each tier
	yearlyPayments := 0
	lifetimePayments := 0

	for _, payment := range payments {
		if payment.Description == "Yearly Subscription" {
			yearlyPayments++
			assert.Equal(t, int64(3000), payment.Amount) // $30.00
		} else if payment.Description == "Lifetime Subscription" {
			lifetimePayments++
			assert.Equal(t, int64(15000), payment.Amount) // $150.00
		}
	}

	assert.Equal(t, 1, yearlyPayments, "Should have exactly one yearly payment")
	assert.Equal(t, 1, lifetimePayments, "Should have exactly one lifetime payment")
}

// TestUpgradeLifetimeToPremium tests upgrading from a lifetime to a premium lifetime subscription
func TestUpgradeLifetimeToPremium(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set test environment
	t.Setenv("APP_ENV", "test")
	t.Setenv("APP_BASE_URL", "http://localhost:3000")

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

	// Set up test router and controller
	router := gin.New()
	paymentController := controllers.NewPaymentController(db)

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

	router.GET("/owner", func(c *gin.Context) {
		// Mock owner page for testing
		c.String(http.StatusOK, "Owner Page")
	})

	router.GET("/pricing", func(c *gin.Context) {
		// Mock pricing page for testing
		c.String(http.StatusOK, "Pricing Page")
	})

	router.GET("/owner/guns", func(c *gin.Context) {
		// Mock owner/guns page for testing
		c.String(http.StatusOK, "Owner Guns Page")
	})

	// STEP 1: Subscribe to lifetime plan first
	t.Log("STEP 1: Subscribing to lifetime plan")
	form := url.Values{}
	form.Add("tier", "lifetime")
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract the session ID from the redirect URL and add tier information
	redirectURL := w.Header().Get("Location")
	sessionID := strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_lifetime" // Add tier to session ID for test mode

	// Complete the lifetime subscription
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Verify lifetime subscription
	var lifetimeUser models.User
	err := db.First(&lifetimeUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "lifetime", lifetimeUser.SubscriptionTier)

	// Record the expiration date of the lifetime subscription
	lifetimeExpiryDate := lifetimeUser.SubscriptionExpiresAt

	// STEP 2: Upgrade to premium lifetime plan
	t.Log("STEP 2: Upgrading to premium lifetime plan")
	form = url.Values{}
	form.Add("tier", "premium_lifetime")
	req, _ = http.NewRequest("POST", "/checkout", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Extract the session ID from the redirect URL and add tier information
	redirectURL = w.Header().Get("Location")
	sessionID = strings.Split(strings.Split(redirectURL, "session_id=")[1], "&")[0]
	sessionID = sessionID + "_premium" // Add tier to session ID for test mode

	// Complete the premium lifetime subscription
	req, _ = http.NewRequest("GET", "/payment/success?session_id="+sessionID, nil)
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// STEP 3: Verify the user's subscription was updated correctly
	t.Log("STEP 3: Verifying user subscription after upgrade")
	var premiumUser models.User
	err = db.First(&premiumUser, user.ID).Error
	assert.NoError(t, err)

	// Check subscription tier
	assert.Equal(t, "premium_lifetime", premiumUser.SubscriptionTier)

	// Check expiration date (should be the same as the lifetime subscription)
	assert.Equal(t, lifetimeExpiryDate, premiumUser.SubscriptionExpiresAt,
		"Premium lifetime subscription should keep the same expiration date as lifetime")

	// STEP 4: Verify payment records
	t.Log("STEP 4: Verifying payment records")
	var payments []models.Payment
	err = db.Where("user_id = ?", user.ID).Find(&payments).Error
	assert.NoError(t, err)

	// Should have two payment records (one for lifetime, one for premium)
	assert.Equal(t, 2, len(payments))

	// Check that we have one payment for each tier
	lifetimePayments := 0
	premiumPayments := 0

	for _, payment := range payments {
		if payment.Description == "Lifetime Subscription" {
			lifetimePayments++
			assert.Equal(t, int64(15000), payment.Amount) // $150.00
		} else if payment.Description == "Premium_lifetime Subscription" {
			premiumPayments++
			assert.Equal(t, int64(30000), payment.Amount) // $300.00
		}
	}

	assert.Equal(t, 1, lifetimePayments, "Should have exactly one lifetime payment")
	assert.Equal(t, 1, premiumPayments, "Should have exactly one premium lifetime payment")
}
