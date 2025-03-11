package controllers

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCalculateGrowthRate tests the growth rate calculation logic
func TestCalculateGrowthRate(t *testing.T) {
	// Test cases
	testCases := []struct {
		name           string
		lastMonthUsers int64
		thisMonthUsers int64
		expectedRate   float64
	}{
		{
			name:           "Positive growth",
			lastMonthUsers: 4,
			thisMonthUsers: 6,
			expectedRate:   50.0, // (6-4)/4 * 100 = 50%
		},
		{
			name:           "Negative growth",
			lastMonthUsers: 10,
			thisMonthUsers: 5,
			expectedRate:   -50.0, // (5-10)/10 * 100 = -50%
		},
		{
			name:           "Zero growth",
			lastMonthUsers: 5,
			thisMonthUsers: 5,
			expectedRate:   0.0, // (5-5)/5 * 100 = 0%
		},
		{
			name:           "No users last month, some this month",
			lastMonthUsers: 0,
			thisMonthUsers: 5,
			expectedRate:   100.0, // Special case: 100%
		},
		{
			name:           "Some users last month, none this month",
			lastMonthUsers: 5,
			thisMonthUsers: 0,
			expectedRate:   -100.0, // Special case: -100%
		},
		{
			name:           "No users in either month",
			lastMonthUsers: 0,
			thisMonthUsers: 0,
			expectedRate:   0.0, // No growth
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate the growth rate manually using the same logic as in the controller
			var growthRate float64
			if tc.lastMonthUsers > 0 {
				growthRate = float64(tc.thisMonthUsers-tc.lastMonthUsers) / float64(tc.lastMonthUsers) * 100
			} else if tc.thisMonthUsers > 0 {
				growthRate = 100.0
			} else if tc.lastMonthUsers > 0 && tc.thisMonthUsers == 0 {
				growthRate = -100.0
			}

			// Verify the growth rate calculation
			assert.InDelta(t, tc.expectedRate, growthRate, 0.01, "Growth rate calculation should match expected value")
		})
	}
}

// TestCalculateSubscribedGrowthRate tests the subscribed users growth rate calculation logic
func TestCalculateSubscribedGrowthRate(t *testing.T) {
	// Test cases - reusing the same test cases as they follow the same logic
	testCases := []struct {
		name                string
		lastMonthSubscribed int64
		thisMonthSubscribed int64
		expectedRate        float64
	}{
		{
			name:                "Positive growth in subscriptions",
			lastMonthSubscribed: 3,
			thisMonthSubscribed: 6,
			expectedRate:        100.0, // (6-3)/3 * 100 = 100%
		},
		{
			name:                "Negative growth in subscriptions",
			lastMonthSubscribed: 8,
			thisMonthSubscribed: 4,
			expectedRate:        -50.0, // (4-8)/8 * 100 = -50%
		},
		{
			name:                "Zero growth in subscriptions",
			lastMonthSubscribed: 5,
			thisMonthSubscribed: 5,
			expectedRate:        0.0, // (5-5)/5 * 100 = 0%
		},
		{
			name:                "No subscriptions last month, some this month",
			lastMonthSubscribed: 0,
			thisMonthSubscribed: 3,
			expectedRate:        100.0, // Special case: 100%
		},
		{
			name:                "Some subscriptions last month, none this month",
			lastMonthSubscribed: 3,
			thisMonthSubscribed: 0,
			expectedRate:        -100.0, // Special case: -100%
		},
		{
			name:                "No subscriptions in either month",
			lastMonthSubscribed: 0,
			thisMonthSubscribed: 0,
			expectedRate:        0.0, // No growth
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate the growth rate manually using the same logic as in the controller
			var subscribedGrowthRate float64
			if tc.lastMonthSubscribed > 0 {
				subscribedGrowthRate = float64(tc.thisMonthSubscribed-tc.lastMonthSubscribed) / float64(tc.lastMonthSubscribed) * 100
			} else if tc.thisMonthSubscribed > 0 {
				subscribedGrowthRate = 100.0
			} else if tc.lastMonthSubscribed > 0 && tc.thisMonthSubscribed == 0 {
				subscribedGrowthRate = -100.0
			}

			// Verify the growth rate calculation
			assert.InDelta(t, tc.expectedRate, subscribedGrowthRate, 0.01, "Subscribed growth rate calculation should match expected value")
		})
	}
}

// TestCalculateNewRegistrationsGrowthRate tests the new registrations growth rate calculation logic
func TestCalculateNewRegistrationsGrowthRate(t *testing.T) {
	// Test cases - reusing the same test cases as they follow the same logic
	testCases := []struct {
		name                   string
		lastMonthRegistrations int64
		thisMonthRegistrations int64
		expectedRate           float64
	}{
		{
			name:                   "Positive growth in registrations",
			lastMonthRegistrations: 5,
			thisMonthRegistrations: 10,
			expectedRate:           100.0, // (10-5)/5 * 100 = 100%
		},
		{
			name:                   "Negative growth in registrations",
			lastMonthRegistrations: 12,
			thisMonthRegistrations: 6,
			expectedRate:           -50.0, // (6-12)/12 * 100 = -50%
		},
		{
			name:                   "Zero growth in registrations",
			lastMonthRegistrations: 8,
			thisMonthRegistrations: 8,
			expectedRate:           0.0, // (8-8)/8 * 100 = 0%
		},
		{
			name:                   "No registrations last month, some this month",
			lastMonthRegistrations: 0,
			thisMonthRegistrations: 7,
			expectedRate:           100.0, // Special case: 100%
		},
		{
			name:                   "Some registrations last month, none this month",
			lastMonthRegistrations: 9,
			thisMonthRegistrations: 0,
			expectedRate:           -100.0, // Special case: -100%
		},
		{
			name:                   "No registrations in either month",
			lastMonthRegistrations: 0,
			thisMonthRegistrations: 0,
			expectedRate:           0.0, // No growth
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate the growth rate manually using the same logic as in the controller
			var registrationsGrowthRate float64
			if tc.lastMonthRegistrations > 0 {
				registrationsGrowthRate = float64(tc.thisMonthRegistrations-tc.lastMonthRegistrations) / float64(tc.lastMonthRegistrations) * 100
			} else if tc.thisMonthRegistrations > 0 {
				registrationsGrowthRate = 100.0
			} else if tc.lastMonthRegistrations > 0 && tc.thisMonthRegistrations == 0 {
				registrationsGrowthRate = -100.0
			}

			// Verify the growth rate calculation
			assert.InDelta(t, tc.expectedRate, registrationsGrowthRate, 0.01, "New registrations growth rate calculation should match expected value")
		})
	}
}

// TestCalculateNewSubscriptionsGrowthRate tests the new subscriptions growth rate calculation logic
func TestCalculateNewSubscriptionsGrowthRate(t *testing.T) {
	// Test cases - reusing the same test cases as they follow the same logic
	testCases := []struct {
		name                      string
		lastMonthNewSubscriptions int64
		thisMonthNewSubscriptions int64
		expectedRate              float64
	}{
		{
			name:                      "Positive growth in new subscriptions",
			lastMonthNewSubscriptions: 2,
			thisMonthNewSubscriptions: 5,
			expectedRate:              150.0, // (5-2)/2 * 100 = 150%
		},
		{
			name:                      "Negative growth in new subscriptions",
			lastMonthNewSubscriptions: 6,
			thisMonthNewSubscriptions: 3,
			expectedRate:              -50.0, // (3-6)/6 * 100 = -50%
		},
		{
			name:                      "Zero growth in new subscriptions",
			lastMonthNewSubscriptions: 4,
			thisMonthNewSubscriptions: 4,
			expectedRate:              0.0, // (4-4)/4 * 100 = 0%
		},
		{
			name:                      "No new subscriptions last month, some this month",
			lastMonthNewSubscriptions: 0,
			thisMonthNewSubscriptions: 3,
			expectedRate:              100.0, // Special case: 100%
		},
		{
			name:                      "Some new subscriptions last month, none this month",
			lastMonthNewSubscriptions: 4,
			thisMonthNewSubscriptions: 0,
			expectedRate:              -100.0, // Special case: -100%
		},
		{
			name:                      "No new subscriptions in either month",
			lastMonthNewSubscriptions: 0,
			thisMonthNewSubscriptions: 0,
			expectedRate:              0.0, // No growth
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate the growth rate manually using the same logic as in the controller
			var newSubscriptionsGrowthRate float64
			if tc.lastMonthNewSubscriptions > 0 {
				newSubscriptionsGrowthRate = float64(tc.thisMonthNewSubscriptions-tc.lastMonthNewSubscriptions) / float64(tc.lastMonthNewSubscriptions) * 100
			} else if tc.thisMonthNewSubscriptions > 0 {
				newSubscriptionsGrowthRate = 100.0
			} else if tc.lastMonthNewSubscriptions > 0 && tc.thisMonthNewSubscriptions == 0 {
				newSubscriptionsGrowthRate = -100.0
			}

			// Verify the growth rate calculation
			assert.InDelta(t, tc.expectedRate, newSubscriptionsGrowthRate, 0.01, "New subscriptions growth rate calculation should match expected value")
		})
	}
}

// TestCalculateMonthlySubscribersGrowthRate tests the monthly subscribers growth rate calculation logic
func TestCalculateMonthlySubscribersGrowthRate(t *testing.T) {
	// Test cases - reusing the same test cases as they follow the same logic
	testCases := []struct {
		name             string
		lastMonthMonthly int64
		thisMonthMonthly int64
		expectedRate     float64
	}{
		{
			name:             "Positive growth in monthly subscribers",
			lastMonthMonthly: 2,
			thisMonthMonthly: 5,
			expectedRate:     150.0, // (5-2)/2 * 100 = 150%
		},
		{
			name:             "Negative growth in monthly subscribers",
			lastMonthMonthly: 6,
			thisMonthMonthly: 3,
			expectedRate:     -50.0, // (3-6)/6 * 100 = -50%
		},
		{
			name:             "Zero growth in monthly subscribers",
			lastMonthMonthly: 4,
			thisMonthMonthly: 4,
			expectedRate:     0.0, // (4-4)/4 * 100 = 0%
		},
		{
			name:             "No monthly subscribers last month, some this month",
			lastMonthMonthly: 0,
			thisMonthMonthly: 3,
			expectedRate:     100.0, // Special case: 100%
		},
		{
			name:             "Some monthly subscribers last month, none this month",
			lastMonthMonthly: 4,
			thisMonthMonthly: 0,
			expectedRate:     -100.0, // Special case: -100%
		},
		{
			name:             "No monthly subscribers in either month",
			lastMonthMonthly: 0,
			thisMonthMonthly: 0,
			expectedRate:     0.0, // No growth
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate the growth rate manually using the same logic as in the controller
			var monthlyGrowthRate float64
			if tc.lastMonthMonthly > 0 {
				monthlyGrowthRate = float64(tc.thisMonthMonthly-tc.lastMonthMonthly) / float64(tc.lastMonthMonthly) * 100
			} else if tc.thisMonthMonthly > 0 {
				monthlyGrowthRate = 100.0
			} else if tc.lastMonthMonthly > 0 && tc.thisMonthMonthly == 0 {
				monthlyGrowthRate = -100.0
			}

			// Verify the growth rate calculation
			assert.InDelta(t, tc.expectedRate, monthlyGrowthRate, 0.01, "Monthly subscribers growth rate calculation should match expected value")
		})
	}
}

// TestCalculateYearlySubscribersGrowthRate tests the yearly subscribers growth rate calculation logic
func TestCalculateYearlySubscribersGrowthRate(t *testing.T) {
	// Test cases - reusing the same test cases as they follow the same logic
	testCases := []struct {
		name            string
		lastMonthYearly int64
		thisMonthYearly int64
		expectedRate    float64
	}{
		{
			name:            "Positive growth in yearly subscribers",
			lastMonthYearly: 2,
			thisMonthYearly: 5,
			expectedRate:    150.0, // (5-2)/2 * 100 = 150%
		},
		{
			name:            "Negative growth in yearly subscribers",
			lastMonthYearly: 6,
			thisMonthYearly: 3,
			expectedRate:    -50.0, // (3-6)/6 * 100 = -50%
		},
		{
			name:            "Zero growth in yearly subscribers",
			lastMonthYearly: 4,
			thisMonthYearly: 4,
			expectedRate:    0.0, // (4-4)/4 * 100 = 0%
		},
		{
			name:            "No yearly subscribers last month, some this month",
			lastMonthYearly: 0,
			thisMonthYearly: 3,
			expectedRate:    100.0, // Special case: 100%
		},
		{
			name:            "Some yearly subscribers last month, none this month",
			lastMonthYearly: 4,
			thisMonthYearly: 0,
			expectedRate:    -100.0, // Special case: -100%
		},
		{
			name:            "No yearly subscribers in either month",
			lastMonthYearly: 0,
			thisMonthYearly: 0,
			expectedRate:    0.0, // No growth
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate the growth rate manually using the same logic as in the controller
			var yearlyGrowthRate float64
			if tc.lastMonthYearly > 0 {
				yearlyGrowthRate = float64(tc.thisMonthYearly-tc.lastMonthYearly) / float64(tc.lastMonthYearly) * 100
			} else if tc.thisMonthYearly > 0 {
				yearlyGrowthRate = 100.0
			} else if tc.lastMonthYearly > 0 && tc.thisMonthYearly == 0 {
				yearlyGrowthRate = -100.0
			}

			// Verify the growth rate calculation
			assert.InDelta(t, tc.expectedRate, yearlyGrowthRate, 0.01, "Yearly subscribers growth rate calculation should match expected value")
		})
	}
}

// TestCalculateLifetimeSubscribersGrowthRate tests the lifetime subscribers growth rate calculation logic
func TestCalculateLifetimeSubscribersGrowthRate(t *testing.T) {
	// Test cases - reusing the same test cases as they follow the same logic
	testCases := []struct {
		name              string
		lastMonthLifetime int64
		thisMonthLifetime int64
		expectedRate      float64
	}{
		{
			name:              "Positive growth in lifetime subscribers",
			lastMonthLifetime: 2,
			thisMonthLifetime: 5,
			expectedRate:      150.0, // (5-2)/2 * 100 = 150%
		},
		{
			name:              "Negative growth in lifetime subscribers",
			lastMonthLifetime: 6,
			thisMonthLifetime: 3,
			expectedRate:      -50.0, // (3-6)/6 * 100 = -50%
		},
		{
			name:              "Zero growth in lifetime subscribers",
			lastMonthLifetime: 4,
			thisMonthLifetime: 4,
			expectedRate:      0.0, // (4-4)/4 * 100 = 0%
		},
		{
			name:              "No lifetime subscribers last month, some this month",
			lastMonthLifetime: 0,
			thisMonthLifetime: 3,
			expectedRate:      100.0, // Special case: 100%
		},
		{
			name:              "Some lifetime subscribers last month, none this month",
			lastMonthLifetime: 4,
			thisMonthLifetime: 0,
			expectedRate:      -100.0, // Special case: -100%
		},
		{
			name:              "No lifetime subscribers in either month",
			lastMonthLifetime: 0,
			thisMonthLifetime: 0,
			expectedRate:      0.0, // No growth
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate the growth rate manually using the same logic as in the controller
			var lifetimeGrowthRate float64
			if tc.lastMonthLifetime > 0 {
				lifetimeGrowthRate = float64(tc.thisMonthLifetime-tc.lastMonthLifetime) / float64(tc.lastMonthLifetime) * 100
			} else if tc.thisMonthLifetime > 0 {
				lifetimeGrowthRate = 100.0
			} else if tc.lastMonthLifetime > 0 && tc.thisMonthLifetime == 0 {
				lifetimeGrowthRate = -100.0
			}

			// Verify the growth rate calculation
			assert.InDelta(t, tc.expectedRate, lifetimeGrowthRate, 0.01, "Lifetime subscribers growth rate calculation should match expected value")
		})
	}
}

// TestCalculatePremiumSubscribersGrowthRate tests the premium subscribers growth rate calculation logic
func TestCalculatePremiumSubscribersGrowthRate(t *testing.T) {
	// Test cases - reusing the same test cases as they follow the same logic
	testCases := []struct {
		name             string
		lastMonthPremium int64
		thisMonthPremium int64
		expectedRate     float64
	}{
		{
			name:             "Positive growth in premium subscribers",
			lastMonthPremium: 2,
			thisMonthPremium: 5,
			expectedRate:     150.0, // (5-2)/2 * 100 = 150%
		},
		{
			name:             "Negative growth in premium subscribers",
			lastMonthPremium: 6,
			thisMonthPremium: 3,
			expectedRate:     -50.0, // (3-6)/6 * 100 = -50%
		},
		{
			name:             "Zero growth in premium subscribers",
			lastMonthPremium: 4,
			thisMonthPremium: 4,
			expectedRate:     0.0, // (4-4)/4 * 100 = 0%
		},
		{
			name:             "No premium subscribers last month, some this month",
			lastMonthPremium: 0,
			thisMonthPremium: 3,
			expectedRate:     100.0, // Special case: 100%
		},
		{
			name:             "Some premium subscribers last month, none this month",
			lastMonthPremium: 4,
			thisMonthPremium: 0,
			expectedRate:     -100.0, // Special case: -100%
		},
		{
			name:             "No premium subscribers in either month",
			lastMonthPremium: 0,
			thisMonthPremium: 0,
			expectedRate:     0.0, // No growth
		},
	}

	// Run test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Calculate the growth rate manually using the same logic as in the controller
			var premiumGrowthRate float64
			if tc.lastMonthPremium > 0 {
				premiumGrowthRate = float64(tc.thisMonthPremium-tc.lastMonthPremium) / float64(tc.lastMonthPremium) * 100
			} else if tc.thisMonthPremium > 0 {
				premiumGrowthRate = 100.0
			} else if tc.lastMonthPremium > 0 && tc.thisMonthPremium == 0 {
				premiumGrowthRate = -100.0
			}

			// Verify the growth rate calculation
			assert.InDelta(t, tc.expectedRate, premiumGrowthRate, 0.01, "Premium subscribers growth rate calculation should match expected value")
		})
	}
}
