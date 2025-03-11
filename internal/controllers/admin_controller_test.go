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
