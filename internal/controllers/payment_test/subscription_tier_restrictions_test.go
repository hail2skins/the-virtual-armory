package payment_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestYearlySubscriberCannotSubscribeToMonthly tests that a user with a yearly subscription
// cannot subscribe to the monthly package
func TestYearlySubscriberCannotSubscribeToMonthly(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer payment_test_utils.CleanupTestDB(t, db)

	// Create a test user with a yearly subscription
	expiresAt := time.Now().AddDate(0, 0, 365) // 1 year from now
	user := &models.User{
		Email:                 "yearly@example.com",
		Password:              "password",
		SubscriptionTier:      "yearly",
		SubscriptionExpiresAt: expiresAt,
	}
	result := db.Create(user)
	if result.Error != nil {
		t.Fatalf("Failed to create test user: %v", result.Error)
	}

	// Set up router with the payment controller
	router, _ := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for the pricing page
	router.GET("/pricing", func(c *gin.Context) {
		// Set the user in the context
		c.Set("user", user)

		// Use the mock pricing page
		payment_test_utils.MockPricingPage(c, user)
	})

	// Test accessing the pricing page as a logged-in user with yearly subscription
	req, _ := http.NewRequest("GET", "/pricing", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the response is OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the monthly subscription option is disabled
	// The monthly subscription link should not be present
	assert.NotContains(t, w.Body.String(), `href="/checkout?tier=monthly"`)

	// Instead, it should show a disabled button or message
	assert.Contains(t, w.Body.String(), `class="block w-full bg-gray-400 text-white font-semibold py-2 px-4 rounded cursor-not-allowed text-center"`)
	assert.Contains(t, w.Body.String(), `Already subscribed to a higher tier`)
}

// TestLifetimeSubscriberCannotSubscribeToMonthlyOrYearly tests that a user with a lifetime subscription
// cannot subscribe to the monthly or yearly packages
func TestLifetimeSubscriberCannotSubscribeToMonthlyOrYearly(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer payment_test_utils.CleanupTestDB(t, db)

	// Create a test user with a lifetime subscription
	user := &models.User{
		Email:            "lifetime@example.com",
		Password:         "password",
		SubscriptionTier: "lifetime",
	}
	result := db.Create(user)
	if result.Error != nil {
		t.Fatalf("Failed to create test user: %v", result.Error)
	}

	// Set up router with the payment controller
	router, _ := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for the pricing page
	router.GET("/pricing", func(c *gin.Context) {
		// Set the user in the context
		c.Set("user", user)

		// Use the mock pricing page
		payment_test_utils.MockPricingPage(c, user)
	})

	// Test accessing the pricing page as a logged-in user with lifetime subscription
	req, _ := http.NewRequest("GET", "/pricing", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the response is OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the monthly and yearly subscription options are disabled
	// The subscription links should not be present
	assert.NotContains(t, w.Body.String(), `href="/checkout?tier=monthly"`)
	assert.NotContains(t, w.Body.String(), `href="/checkout?tier=yearly"`)

	// Instead, they should show disabled buttons or messages
	assert.Contains(t, w.Body.String(), `Already subscribed to a higher tier`)
}

// TestPremiumLifetimeSubscriberCannotSubscribeToLowerTiers tests that a user with a premium lifetime subscription
// cannot subscribe to any lower tier packages
func TestPremiumLifetimeSubscriberCannotSubscribeToLowerTiers(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer payment_test_utils.CleanupTestDB(t, db)

	// Create a test user with a premium lifetime subscription
	user := &models.User{
		Email:            "premium@example.com",
		Password:         "password",
		SubscriptionTier: "premium_lifetime",
	}
	result := db.Create(user)
	if result.Error != nil {
		t.Fatalf("Failed to create test user: %v", result.Error)
	}

	// Set up router with the payment controller
	router, _ := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for the pricing page
	router.GET("/pricing", func(c *gin.Context) {
		// Set the user in the context
		c.Set("user", user)

		// Use the mock pricing page
		payment_test_utils.MockPricingPage(c, user)
	})

	// Test accessing the pricing page as a logged-in user with premium lifetime subscription
	req, _ := http.NewRequest("GET", "/pricing", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the response is OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that all lower tier subscription options are disabled
	// The subscription links should not be present
	assert.NotContains(t, w.Body.String(), `href="/checkout?tier=monthly"`)
	assert.NotContains(t, w.Body.String(), `href="/checkout?tier=yearly"`)
	assert.NotContains(t, w.Body.String(), `href="/checkout?tier=lifetime"`)

	// Instead, they should show disabled buttons or messages
	assert.Contains(t, w.Body.String(), `Already subscribed to a higher tier`)
}
