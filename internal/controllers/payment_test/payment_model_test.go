package payment_test

import (
	"testing"
	"time"

	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestPaymentModel tests the Payment model
func TestPaymentModel(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := createTestUser(t, db)

	// Create a test payment
	payment := models.Payment{
		UserID:           user.ID,
		Amount:           500, // $5.00
		Currency:         "usd",
		StripePaymentID:  "pi_test123",
		Status:           "succeeded",
		SubscriptionTier: "monthly",
	}

	// Save the payment to the database
	err := db.Create(&payment).Error
	assert.NoError(t, err)
	assert.NotZero(t, payment.ID)

	// Retrieve the payment from the database
	var retrievedPayment models.Payment
	err = db.First(&retrievedPayment, payment.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, payment.UserID, retrievedPayment.UserID)
	assert.Equal(t, payment.Amount, retrievedPayment.Amount)
	assert.Equal(t, payment.Currency, retrievedPayment.Currency)
	assert.Equal(t, payment.StripePaymentID, retrievedPayment.StripePaymentID)
	assert.Equal(t, payment.Status, retrievedPayment.Status)
	assert.Equal(t, payment.SubscriptionTier, retrievedPayment.SubscriptionTier)
}

// TestUserSubscriptionFields tests the subscription-related fields in the User model
func TestUserSubscriptionFields(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := createTestUser(t, db)

	// Set subscription fields
	futureTime := time.Now().Add(30 * 24 * time.Hour) // 30 days in the future
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": futureTime,
		"stripe_customer_id":      "cus_test123",
	})

	// Retrieve the user from the database
	var retrievedUser models.User
	err := db.First(&retrievedUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "monthly", retrievedUser.SubscriptionTier)
	assert.True(t, retrievedUser.SubscriptionExpiresAt.After(time.Now()))
	assert.Equal(t, "cus_test123", retrievedUser.StripeCustomerID)

	// Test subscription status methods
	assert.True(t, retrievedUser.HasActiveSubscription())
	assert.False(t, retrievedUser.IsLifetimeSubscriber())

	// Test with a lifetime subscription
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "lifetime",
		"subscription_expires_at": time.Time{}, // No expiration
	})

	// Retrieve the updated user
	err = db.First(&retrievedUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "lifetime", retrievedUser.SubscriptionTier)
	assert.True(t, retrievedUser.SubscriptionExpiresAt.IsZero())

	// Test subscription status methods for lifetime subscription
	assert.True(t, retrievedUser.HasActiveSubscription())
	assert.True(t, retrievedUser.IsLifetimeSubscriber())

	// Test with an expired subscription
	expiredTime := time.Now().Add(-24 * time.Hour) // 1 day ago
	db.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "monthly",
		"subscription_expires_at": expiredTime,
	})

	// Retrieve the updated user
	err = db.First(&retrievedUser, user.ID).Error
	assert.NoError(t, err)
	assert.Equal(t, "monthly", retrievedUser.SubscriptionTier)
	assert.True(t, retrievedUser.SubscriptionExpiresAt.Before(time.Now()))

	// Test subscription status methods for expired subscription
	assert.False(t, retrievedUser.HasActiveSubscription())
	assert.False(t, retrievedUser.IsLifetimeSubscriber())
}

// TestPaymentHistory tests retrieving a user's payment history
func TestPaymentHistory(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a test user
	user := createTestUser(t, db)

	// Create multiple payments for the user
	payments := []models.Payment{
		{
			UserID:           user.ID,
			Amount:           500, // $5.00
			Currency:         "usd",
			StripePaymentID:  "pi_test1",
			Status:           "succeeded",
			SubscriptionTier: "monthly",
		},
		{
			UserID:           user.ID,
			Amount:           3000, // $30.00
			Currency:         "usd",
			StripePaymentID:  "pi_test2",
			Status:           "succeeded",
			SubscriptionTier: "yearly",
		},
		{
			UserID:           user.ID,
			Amount:           10000, // $100.00
			Currency:         "usd",
			StripePaymentID:  "pi_test3",
			Status:           "succeeded",
			SubscriptionTier: "lifetime",
		},
	}

	// Save the payments to the database
	for _, payment := range payments {
		err := db.Create(&payment).Error
		assert.NoError(t, err)
	}

	// Retrieve the user's payment history
	var retrievedPayments []models.Payment
	err := db.Where("user_id = ?", user.ID).Order("created_at desc").Find(&retrievedPayments).Error
	assert.NoError(t, err)
	assert.Equal(t, len(payments), len(retrievedPayments))

	// The payments should be in reverse chronological order
	assert.Equal(t, "lifetime", retrievedPayments[0].SubscriptionTier)
	assert.Equal(t, "yearly", retrievedPayments[1].SubscriptionTier)
	assert.Equal(t, "monthly", retrievedPayments[2].SubscriptionTier)
}
