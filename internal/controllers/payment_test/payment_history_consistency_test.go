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

// TestPaymentHistorySubscriptionInfoConsistency tests that the payment history page
// displays the same subscription information as the pricing page
func TestPaymentHistorySubscriptionInfoConsistency(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer payment_test_utils.CleanupTestDB(t, db)

	// Create test users with different subscription tiers
	users := map[string]*models.User{
		"monthly": {
			Email:                 "monthly@example.com",
			Password:              "password",
			SubscriptionTier:      "monthly",
			SubscriptionExpiresAt: time.Now().AddDate(0, 1, 0), // 1 month from now
		},
		"yearly": {
			Email:                 "yearly@example.com",
			Password:              "password",
			SubscriptionTier:      "yearly",
			SubscriptionExpiresAt: time.Now().AddDate(1, 0, 0), // 1 year from now
		},
		"lifetime": {
			Email:            "lifetime@example.com",
			Password:         "password",
			SubscriptionTier: "lifetime",
		},
		"premium_lifetime": {
			Email:            "premium@example.com",
			Password:         "password",
			SubscriptionTier: "premium_lifetime",
		},
	}

	// Create users in the database
	for _, user := range users {
		result := db.Create(user)
		if result.Error != nil {
			t.Fatalf("Failed to create test user: %v", result.Error)
		}
	}

	// Set up router with the payment controller
	router, paymentController := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the routes
	router.GET("/pricing", paymentController.ShowPricingPage)
	router.GET("/owner/payment-history", func(c *gin.Context) {
		// For testing purposes, we'll manually set the user in the context
		// This simulates the authentication middleware
		userEmail := c.Request.Header.Get("X-User-Email")
		if userEmail != "" {
			var user models.User
			if err := db.Where("email = ?", userEmail).First(&user).Error; err == nil {
				c.Set("user", &user)
				// Use the mock payment history page
				payment_test_utils.MockPaymentHistory(c, &user)
				return
			}
		}
		// If no user is found, redirect to login
		c.Redirect(http.StatusSeeOther, "/login")
	})

	// Test for each subscription tier
	for tierName, user := range users {
		t.Run("Testing "+tierName+" tier", func(t *testing.T) {
			// Get the pricing page content
			pricingReq, _ := http.NewRequest("GET", "/pricing", nil)
			pricingReq.Header.Set("X-User-Email", user.Email)
			pricingW := httptest.NewRecorder()

			// Set the user in the context for the pricing page
			router.ServeHTTP(pricingW, pricingReq)
			pricingContent := pricingW.Body.String()

			// Get the payment history page content
			historyReq, _ := http.NewRequest("GET", "/owner/payment-history", nil)
			historyReq.Header.Set("X-User-Email", user.Email)
			historyW := httptest.NewRecorder()

			// Set the user in the context for the payment history page
			router.ServeHTTP(historyW, historyReq)
			historyContent := historyW.Body.String()

			// Check that both pages show the same tier name
			switch tierName {
			case "monthly":
				assert.Contains(t, pricingContent, "Liking It")
				assert.Contains(t, historyContent, "Liking It")

				// Check for consistent benefits
				assert.Contains(t, pricingContent, "Unlimited guns/ammo")
				assert.Contains(t, historyContent, "Unlimited guns/ammo")

			case "yearly":
				assert.Contains(t, pricingContent, "Loving It")
				assert.Contains(t, historyContent, "Loving It")

				// Check for consistent benefits
				assert.Contains(t, pricingContent, "Unlimited guns/ammo")
				assert.Contains(t, historyContent, "Unlimited guns/ammo")

			case "lifetime":
				assert.Contains(t, pricingContent, "Supporter")
				assert.Contains(t, historyContent, "Supporter")

				// Check for consistent benefits
				assert.Contains(t, pricingContent, "Unlimited guns/ammo")
				assert.Contains(t, historyContent, "Unlimited guns/ammo")

			case "premium_lifetime":
				assert.Contains(t, pricingContent, "Big Baller")
				assert.Contains(t, historyContent, "Big Baller")

				// Check for consistent benefits
				assert.Contains(t, pricingContent, "Everything the site has")
				assert.Contains(t, historyContent, "Everything the site has")
			}
		})
	}
}

// TestPaymentHistoryMockForConsistency tests the payment history page with a mock
// to ensure it displays consistent subscription information
func TestPaymentHistoryMockForConsistency(t *testing.T) {
	// Set up test database
	db := payment_test_utils.SetupTestDB(t)
	defer payment_test_utils.CleanupTestDB(t, db)

	// Create a test user with a subscription
	user := &models.User{
		Email:                 "test@example.com",
		Password:              "password",
		SubscriptionTier:      "monthly",
		SubscriptionExpiresAt: time.Now().AddDate(0, 1, 0), // 1 month from now
	}
	result := db.Create(user)
	if result.Error != nil {
		t.Fatalf("Failed to create test user: %v", result.Error)
	}

	// Set up router
	router, _ := payment_test_utils.SetupPricingTestRouter(t, db)

	// Set up the route for the payment history page with a mock
	router.GET("/owner/payment-history", func(c *gin.Context) {
		// Set the user in the context
		c.Set("user", user)

		// Use the mock payment history page
		payment_test_utils.MockPaymentHistory(c, user)
	})

	// Test accessing the payment history page
	req, _ := http.NewRequest("GET", "/owner/payment-history", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check that the response is OK
	assert.Equal(t, http.StatusOK, w.Code)

	// Check that the payment history page shows the correct subscription information
	// This should match the pricing page's description of the monthly tier
	assert.Contains(t, w.Body.String(), "Liking It")
	assert.Contains(t, w.Body.String(), "Unlimited guns/ammo")
	assert.Contains(t, w.Body.String(), "Unlimited range days")
	assert.Contains(t, w.Body.String(), "Unlimited maintenance records")
	assert.Contains(t, w.Body.String(), "Cancel anytime")
}
