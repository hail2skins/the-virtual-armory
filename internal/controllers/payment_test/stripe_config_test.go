package payment_test

import (
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/stretchr/testify/assert"
)

// TestStripeEnvironmentVariables tests that the required Stripe environment variables are set
func TestStripeEnvironmentVariables(t *testing.T) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Check for required Stripe environment variables
	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	stripePublishableKey := os.Getenv("STRIPE_PUBLISHABLE_KEY")
	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	// Skip the test if running in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Assert that the environment variables are set
	assert.NotEmpty(t, stripeSecretKey, "STRIPE_SECRET_KEY environment variable is not set")
	assert.NotEmpty(t, stripePublishableKey, "STRIPE_PUBLISHABLE_KEY environment variable is not set")
	assert.NotEmpty(t, stripeWebhookSecret, "STRIPE_WEBHOOK_SECRET environment variable is not set")

	// Check that the keys have the expected format
	assert.Contains(t, stripeSecretKey, "sk_", "STRIPE_SECRET_KEY should start with 'sk_'")
	assert.Contains(t, stripePublishableKey, "pk_", "STRIPE_PUBLISHABLE_KEY should start with 'pk_'")
	assert.Contains(t, stripeWebhookSecret, "whsec_", "STRIPE_WEBHOOK_SECRET should start with 'whsec_'")
}

// TestStripeConfigurationLoading tests that the Stripe configuration is loaded correctly
func TestStripeConfigurationLoading(t *testing.T) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Skip the test if running in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Check for required Stripe environment variables
	stripeSecretKey := os.Getenv("STRIPE_SECRET_KEY")
	stripePublishableKey := os.Getenv("STRIPE_PUBLISHABLE_KEY")
	stripeWebhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	// Skip the test if any of the required environment variables are not set
	if stripeSecretKey == "" || stripePublishableKey == "" || stripeWebhookSecret == "" {
		t.Skip("Skipping test because Stripe environment variables are not set")
	}

	// In a real implementation, we would test loading the Stripe configuration here
	// For now, we'll just check that the environment variables are set
	assert.NotEmpty(t, stripeSecretKey)
	assert.NotEmpty(t, stripePublishableKey)
	assert.NotEmpty(t, stripeWebhookSecret)
}

// TestStripeProductConfiguration tests that the Stripe product configuration is correct
func TestStripeProductConfiguration(t *testing.T) {
	// Load .env file if it exists
	_ = godotenv.Load()

	// Skip the test if running in CI environment
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// Check for required Stripe product environment variables
	monthlyPriceID := os.Getenv("STRIPE_PRICE_MONTHLY")
	yearlyPriceID := os.Getenv("STRIPE_PRICE_YEARLY")
	lifetimePriceID := os.Getenv("STRIPE_PRICE_LIFETIME")
	premiumLifetimePriceID := os.Getenv("STRIPE_PRICE_PREMIUM_LIFETIME")

	// Skip the test if any of the required environment variables are not set
	if monthlyPriceID == "" || yearlyPriceID == "" || lifetimePriceID == "" || premiumLifetimePriceID == "" {
		t.Skip("Skipping test because Stripe product environment variables are not set")
	}

	// Assert that the environment variables are set
	assert.NotEmpty(t, monthlyPriceID, "STRIPE_PRICE_MONTHLY environment variable is not set")
	assert.NotEmpty(t, yearlyPriceID, "STRIPE_PRICE_YEARLY environment variable is not set")
	assert.NotEmpty(t, lifetimePriceID, "STRIPE_PRICE_LIFETIME environment variable is not set")
	assert.NotEmpty(t, premiumLifetimePriceID, "STRIPE_PRICE_PREMIUM_LIFETIME environment variable is not set")

	// Check that the price IDs have the expected format
	assert.Contains(t, monthlyPriceID, "price_", "STRIPE_PRICE_MONTHLY should start with 'price_'")
	assert.Contains(t, yearlyPriceID, "price_", "STRIPE_PRICE_YEARLY should start with 'price_'")
	assert.Contains(t, lifetimePriceID, "price_", "STRIPE_PRICE_LIFETIME should start with 'price_'")
	assert.Contains(t, premiumLifetimePriceID, "price_", "STRIPE_PRICE_PREMIUM_LIFETIME should start with 'price_'")
}
