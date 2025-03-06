package payment_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestPricingPageContent tests that the pricing page displays the correct content
func TestPricingPageContent(t *testing.T) {
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
	assert.Contains(t, w.Body.String(), "Choose Your Plan")
	assert.Contains(t, w.Body.String(), "Free")
	assert.Contains(t, w.Body.String(), "Monthly")
	assert.Contains(t, w.Body.String(), "Yearly")
	assert.Contains(t, w.Body.String(), "Lifetime")
	assert.Contains(t, w.Body.String(), "Premium Lifetime")
}

// TestPricingPageWithLoggedInUser tests that the pricing page displays correctly for a logged-in user
func TestPricingPageWithLoggedInUser(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

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
	assert.Contains(t, w.Body.String(), "Choose Your Plan")
	assert.Contains(t, w.Body.String(), "Your Current Plan: Free Tier")
}

// TestPricingPageWithSubscribedUser tests that the pricing page shows the current subscription information
func TestPricingPageWithSubscribedUser(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Create a test user with a subscription
	user := payment_test_utils.CreateTestUser(t, db)
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
	assert.Contains(t, w.Body.String(), "Your Current Plan: Monthly Subscription")
	assert.Contains(t, w.Body.String(), "Expires on")
}

// TestStripeCheckoutRedirect tests that selecting a subscription option redirects to Stripe checkout
func TestStripeCheckoutRedirect(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up test router and controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Create a test user
	user := payment_test_utils.CreateTestUser(t, db)

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
