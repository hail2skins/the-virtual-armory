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
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupPricingTestRouter sets up a test router with the payment controller for pricing tests
func setupPricingTestRouter(t *testing.T, db *gorm.DB) (*gin.Engine, *controllers.PaymentController) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.Default()

	// For tests, we don't need to load actual templates
	// Just mock the HTML renderer to prevent panics
	router.HTMLRender = &testRenderer{}

	// Create the payment controller
	paymentController := controllers.NewPaymentController(db)

	return router, paymentController
}

// TestPricingPageContent tests that the pricing page displays the correct content
func TestPricingPageContent(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up test router and controller
	router, paymentController := setupPricingTestRouter(t, db)

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
	assert.Contains(t, w.Body.String(), "Subscription Plans")
	assert.Contains(t, w.Body.String(), "Free Tier")
	assert.Contains(t, w.Body.String(), "Monthly Subscription")
	assert.Contains(t, w.Body.String(), "Yearly Subscription")
	assert.Contains(t, w.Body.String(), "Lifetime Subscription")
}

// TestPricingPageWithLoggedInUser tests that the pricing page displays correctly for a logged-in user
func TestPricingPageWithLoggedInUser(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up test router and controller
	router, paymentController := setupPricingTestRouter(t, db)

	// Create a test user
	user := createTestUser(t, db)

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

	// Check that the pricing page is displayed with the user's information
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Subscription Plans")
	assert.Contains(t, w.Body.String(), "Current Plan: Free Tier")
}

// TestPricingPageWithSubscribedUser tests that the pricing page shows the current subscription information
func TestPricingPageWithSubscribedUser(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up test router and controller
	router, paymentController := setupPricingTestRouter(t, db)

	// Create a test user with a subscription
	user := createTestUser(t, db)
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": time.Now().Add(30 * 24 * time.Hour),
	})

	// Set up the route for the pricing page
	router.GET("/pricing", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.ShowPricingPage(c)
	})

	// Test accessing the pricing page as a subscribed user
	req, _ := http.NewRequest("GET", "/pricing", nil)
	// Add authentication cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the pricing page is displayed with the user's subscription information
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Current Plan: Monthly Subscription")
	assert.Contains(t, w.Body.String(), "Expires:")
}

// TestStripeCheckoutRedirect tests that selecting a subscription option redirects to Stripe checkout
func TestStripeCheckoutRedirect(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up test router and controller
	router, paymentController := setupPricingTestRouter(t, db)

	// Create a test user
	user := createTestUser(t, db)

	// Set up the route for creating a checkout session
	router.POST("/checkout", func(c *gin.Context) {
		// Set authentication cookies for the test
		c.SetCookie("is_logged_in", "true", 3600, "/", "localhost", false, true)
		c.SetCookie("user_email", user.Email, 3600, "/", "localhost", false, true)

		paymentController.CreateCheckoutSession(c)
	})

	// Test creating a checkout session
	formData := url.Values{
		"tier": {"monthly"},
	}
	req, _ := http.NewRequest("POST", "/checkout", strings.NewReader(formData.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// Add authentication cookies to the request
	req.AddCookie(&http.Cookie{Name: "is_logged_in", Value: "true"})
	req.AddCookie(&http.Cookie{Name: "user_email", Value: user.Email})

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the user is redirected to Stripe checkout
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Contains(t, w.Header().Get("Location"), "checkout.stripe.com")
}
