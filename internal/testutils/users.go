package testutils

import (
	"time"

	"github.com/hail2skins/the-virtual-armory/internal/models"
)

// TestUsers holds all test user instances
type TestUsers struct {
	Admin              models.User
	Unsubscribed       models.User
	MonthlySubscriber  models.User
	YearlySubscriber   models.User
	LifetimeSubscriber models.User
	PremiumSubscriber  models.User
	FutureTestUser1    models.User
	FutureTestUser2    models.User
}

// CreateTestUsers creates a set of test users with different subscription levels
func CreateTestUsers() *TestUsers {
	// Create admin user
	admin := models.User{
		Email:     "admin@test.com",
		Password:  "$2a$10$ZZZ.HashedPasswordForTesting123", // We'll set actual hashed passwords as needed
		IsAdmin:   true,
		Confirmed: true,
	}

	// Create unsubscribed user
	unsubscribed := models.User{
		Email:     "unsubscribed@test.com",
		Password:  "$2a$10$ZZZ.HashedPasswordForTesting123",
		IsAdmin:   false,
		Confirmed: true,
	}

	// Create monthly subscriber
	monthlySubscriber := models.User{
		Email:                 "monthly@test.com",
		Password:              "$2a$10$ZZZ.HashedPasswordForTesting123",
		IsAdmin:               false,
		Confirmed:             true,
		SubscriptionTier:      "monthly",
		SubscriptionExpiresAt: time.Now().AddDate(0, 1, 0),
	}

	// Create yearly subscriber
	yearlySubscriber := models.User{
		Email:                 "yearly@test.com",
		Password:              "$2a$10$ZZZ.HashedPasswordForTesting123",
		IsAdmin:               false,
		Confirmed:             true,
		SubscriptionTier:      "yearly",
		SubscriptionExpiresAt: time.Now().AddDate(1, 0, 0),
	}

	// Create lifetime subscriber
	lifetimeSubscriber := models.User{
		Email:                 "lifetime@test.com",
		Password:              "$2a$10$ZZZ.HashedPasswordForTesting123",
		IsAdmin:               false,
		Confirmed:             true,
		SubscriptionTier:      "lifetime",
		SubscriptionExpiresAt: time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	// Create premium subscriber
	premiumSubscriber := models.User{
		Email:                 "premium@test.com",
		Password:              "$2a$10$ZZZ.HashedPasswordForTesting123",
		IsAdmin:               false,
		Confirmed:             true,
		SubscriptionTier:      "premium",
		SubscriptionExpiresAt: time.Date(9999, 12, 31, 23, 59, 59, 0, time.UTC),
	}

	// Create additional users for future tests
	futureUser1 := models.User{
		Email:     "future1@test.com",
		Password:  "$2a$10$ZZZ.HashedPasswordForTesting123",
		IsAdmin:   false,
		Confirmed: true,
	}

	futureUser2 := models.User{
		Email:     "future2@test.com",
		Password:  "$2a$10$ZZZ.HashedPasswordForTesting123",
		IsAdmin:   false,
		Confirmed: true,
	}

	return &TestUsers{
		Admin:              admin,
		Unsubscribed:       unsubscribed,
		MonthlySubscriber:  monthlySubscriber,
		YearlySubscriber:   yearlySubscriber,
		LifetimeSubscriber: lifetimeSubscriber,
		PremiumSubscriber:  premiumSubscriber,
		FutureTestUser1:    futureUser1,
		FutureTestUser2:    futureUser2,
	}
}
