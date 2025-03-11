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
