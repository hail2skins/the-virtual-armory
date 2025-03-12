package controllers

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	paymentViews "github.com/hail2skins/the-virtual-armory/cmd/web/views/payment"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/flash"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
	"github.com/stripe/stripe-go/v72/sub"
	"github.com/stripe/stripe-go/v72/webhook"
	"gorm.io/gorm"
)

// PaymentController handles payment-related routes
type PaymentController struct {
	DB *gorm.DB
}

// NewPaymentController creates a new PaymentController
func NewPaymentController(db *gorm.DB) *PaymentController {
	return &PaymentController{
		DB: db,
	}
}

// ShowPricingPage displays the pricing page
func (c *PaymentController) ShowPricingPage(ctx *gin.Context) {
	// Get the current user if logged in
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		// If there's an error, just show the pricing page without user info
		user = nil
	}

	// Render the pricing page
	component := paymentViews.Pricing(user)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// CreateCheckoutSession creates a new Stripe checkout session
func (c *PaymentController) CreateCheckoutSession(ctx *gin.Context) {
	// Get the subscription tier from the form
	tier := ctx.PostForm("tier")

	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.JSON(http.StatusUnauthorized, gin.H{"error": "You must be logged in to subscribe"})
		return
	}

	// Check if we're in test mode
	if os.Getenv("APP_ENV") == "test" {
		// In test mode, redirect to a test URL
		baseURL := os.Getenv("APP_BASE_URL")
		if baseURL == "" {
			baseURL = "http://localhost:3000" // Default to port 3000 if not set
		}
		testURL := baseURL + "/payment/success?session_id=cs_test_" + strconv.FormatUint(uint64(user.ID), 10)
		ctx.Redirect(http.StatusSeeOther, testURL)
		return
	}

	// Set Stripe API key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Check if automatic tax is enabled in the Stripe account
	// If not, we'll still create the checkout session but without automatic tax
	automaticTaxEnabled := true
	if os.Getenv("STRIPE_TAX_ENABLED") == "false" {
		automaticTaxEnabled = false
		log.Printf("Automatic tax is disabled in the Stripe account")
	}

	// Get the product ID based on the tier
	var productID string
	switch tier {
	case "monthly":
		productID = os.Getenv("STRIPE_PRICE_MONTHLY")
	case "yearly":
		productID = os.Getenv("STRIPE_PRICE_YEARLY")
	case "lifetime":
		productID = os.Getenv("STRIPE_PRICE_LIFETIME")
	case "premium_lifetime":
		productID = os.Getenv("STRIPE_PRICE_PREMIUM")
	default:
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subscription tier"})
		return
	}

	// Define prices based on the tier
	var unitAmount int64
	var currency = "usd"
	var productName string
	var interval string
	var intervalCount int64 = 1

	switch tier {
	case "monthly":
		unitAmount = 500 // $5.00
		productName = "Liking It Plan"
		interval = "month"
	case "yearly":
		unitAmount = 3000 // $30.00
		productName = "Loving It Plan"
		interval = "year"
	case "lifetime":
		unitAmount = 10000 // $100.00
		productName = "Supporter Plan"
		interval = ""
	case "premium_lifetime":
		unitAmount = 100000 // $1000.00
		productName = "Big Baller Plan"
		interval = ""
	}

	// Create checkout session parameters
	params := &stripe.CheckoutSessionParams{
		PaymentMethodTypes: stripe.StringSlice([]string{
			"card",
		}),
		LineItems: []*stripe.CheckoutSessionLineItemParams{
			{
				PriceData: &stripe.CheckoutSessionLineItemPriceDataParams{
					Currency: stripe.String(currency),
					ProductData: &stripe.CheckoutSessionLineItemPriceDataProductDataParams{
						Name:    stripe.String(productName),
						TaxCode: stripe.String("txcd_10000000"), // Standard tax code
					},
					UnitAmount:  stripe.Int64(unitAmount),
					TaxBehavior: stripe.String("exclusive"), // Add tax on top of the price
					Recurring: func() *stripe.CheckoutSessionLineItemPriceDataRecurringParams {
						if interval == "" {
							return nil
						}
						return &stripe.CheckoutSessionLineItemPriceDataRecurringParams{
							Interval:      stripe.String(interval),
							IntervalCount: stripe.Int64(intervalCount),
						}
					}(),
				},
				Quantity: stripe.Int64(1),
			},
		},
		Mode: func() *string {
			if interval == "" {
				return stripe.String(string(stripe.CheckoutSessionModePayment))
			}
			return stripe.String(string(stripe.CheckoutSessionModeSubscription))
		}(),
		SuccessURL:               stripe.String(os.Getenv("APP_BASE_URL") + "/payment/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:                stripe.String(os.Getenv("APP_BASE_URL") + "/payment/cancel"),
		CustomerEmail:            stripe.String(user.Email),
		AutomaticTax:             &stripe.CheckoutSessionAutomaticTaxParams{Enabled: stripe.Bool(automaticTaxEnabled)},
		BillingAddressCollection: stripe.String(string(stripe.CheckoutSessionBillingAddressCollectionRequired)),
	}

	// Add metadata
	params.AddMetadata("user_id", strconv.FormatUint(uint64(user.ID), 10))
	params.AddMetadata("subscription_tier", tier)
	params.AddMetadata("product_id", productID)

	// Log the success URL for debugging
	successURL := os.Getenv("APP_BASE_URL") + "/payment/success?session_id={CHECKOUT_SESSION_ID}"
	log.Printf("Success URL for checkout: %s", successURL)
	log.Printf("APP_BASE_URL: %s", os.Getenv("APP_BASE_URL"))
	log.Printf("Tax settings: AutomaticTax=%v, BillingAddressCollection=Required",
		*params.AutomaticTax.Enabled)

	// Create the checkout session
	s, err := session.New(params)
	if err != nil {
		log.Printf("Error creating checkout session: %v", err)

		// If the error is related to tax, try again without automatic tax
		if strings.Contains(err.Error(), "tax") {
			log.Printf("Retrying checkout session creation without automatic tax")
			params.AutomaticTax.Enabled = stripe.Bool(false)
			s, err = session.New(params)
			if err != nil {
				log.Printf("Error creating checkout session without tax: %v", err)
				ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
				return
			}
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
			return
		}
	}

	log.Printf("Created checkout session for user %d, tier %s, session ID: %s", user.ID, tier, s.ID)
	log.Printf("Checkout URL: %s", s.URL)
	log.Printf("Success URL in session: %s", *params.SuccessURL)

	// Redirect to the checkout page
	ctx.Redirect(http.StatusSeeOther, s.URL)
}

// Note: All subscription tiers (monthly, yearly, lifetime, premium_lifetime) now use Checkout Sessions
// instead of direct Stripe payment links. This ensures consistent handling of payments and proper
// redirection back to the site after checkout completion.

// logWebhookEvent logs webhook event details for monitoring and debugging
func logWebhookEvent(eventType string, eventID string, userID string, status string, details string) {
	log.Printf("[WEBHOOK] Type: %s, ID: %s, User: %s, Status: %s, Details: %s",
		eventType, eventID, userID, status, details)
}

// HandleStripeWebhook handles Stripe webhook events
func (c *PaymentController) HandleStripeWebhook(ctx *gin.Context) {
	// Get the webhook secret from environment variables
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	// Read the request body
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		log.Printf("Error reading webhook request body: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Get the signature from the headers
	signature := ctx.GetHeader("Stripe-Signature")

	var event stripe.Event

	// Check if we're in test mode
	if os.Getenv("APP_ENV") == "test" && signature == "test_signature" {
		// In test mode with test signature, parse the event without verification
		if err := json.Unmarshal(body, &event); err != nil {
			log.Printf("Error parsing webhook event in test mode: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook payload"})
			return
		}
	} else {
		// In production mode, verify the webhook signature
		var err error
		event, err = webhook.ConstructEvent(body, signature, webhookSecret)
		if err != nil {
			log.Printf("Error verifying webhook signature: %v", err)
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook signature"})
			return
		}
	}

	// Log the event for monitoring
	log.Printf("Received Stripe webhook event: %s (ID: %s)", event.Type, event.ID)

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		handleCheckoutSessionCompleted(c, ctx, event)
	case "customer.subscription.created":
		handleSubscriptionCreated(c, ctx, event)
	case "customer.subscription.updated":
		handleSubscriptionUpdated(c, ctx, event)
	case "customer.subscription.deleted":
		handleSubscriptionDeleted(c, ctx, event)
	case "invoice.paid":
		handleInvoicePaid(c, ctx, event)
	case "invoice.payment_failed":
		handleInvoicePaymentFailed(c, ctx, event)
	default:
		// Log unhandled event types
		log.Printf("Unhandled webhook event type: %s", event.Type)
	}

	// Always return a 200 OK to Stripe
	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// handleCheckoutSessionCompleted processes a completed checkout session
func handleCheckoutSessionCompleted(c *PaymentController, ctx *gin.Context, event stripe.Event) {
	// Parse the event data
	var session stripe.CheckoutSession
	err := json.Unmarshal(event.Data.Raw, &session)
	if err != nil {
		log.Printf("Error parsing checkout.session.completed webhook: %v", err)
		return
	}

	// Log the event for monitoring
	logWebhookEvent("checkout.session.completed", event.ID, session.ClientReferenceID, "processing", "")

	// Get the user ID from the metadata or client reference ID
	var userID string
	if session.Metadata != nil {
		userID = session.Metadata["user_id"]
	}

	// If user_id is not in metadata, try to get it from client_reference_id
	if userID == "" && session.ClientReferenceID != "" {
		userID = session.ClientReferenceID
		log.Printf("Using client_reference_id as user_id: %s", userID)
	}

	if userID == "" {
		log.Printf("Missing user_id in metadata and client_reference_id for session %s", session.ID)
		return
	}

	// Convert userID string to uint
	userIDUint, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		log.Printf("Invalid user_id format: %v", err)
		return
	}

	// Find the user in the database
	var user models.User
	if err := c.DB.First(&user, userIDUint).Error; err != nil {
		log.Printf("Failed to find user %s for session %s: %v", userID, session.ID, err)
		return
	}

	// Get subscription tier from metadata or from the line items
	var subscriptionTier string
	if session.Metadata != nil && session.Metadata["subscription_tier"] != "" {
		subscriptionTier = session.Metadata["subscription_tier"]
	} else {
		// Try to determine the tier from the amount
		if session.AmountTotal >= 100000 {
			// $1000 is premium lifetime (Big Baller)
			subscriptionTier = "premium_lifetime"
			log.Printf("Determined premium_lifetime subscription from amount: %d", session.AmountTotal)
		} else if session.AmountTotal >= 10000 {
			// $100 is lifetime (Supporter)
			subscriptionTier = "lifetime"
			log.Printf("Determined lifetime subscription from amount: %d", session.AmountTotal)
		} else if session.AmountTotal >= 3000 {
			// $30 is yearly (Loving It)
			subscriptionTier = "yearly"
			log.Printf("Determined yearly subscription from amount: %d", session.AmountTotal)
		} else {
			// $5 is monthly (Liking It)
			subscriptionTier = "monthly"
			log.Printf("Determined monthly subscription from amount: %d", session.AmountTotal)
		}
		log.Printf("Determined subscription tier from session: %s", subscriptionTier)
	}

	// Set expiration date based on subscription tier
	var expirationDate time.Time

	// Check if user has an existing subscription that hasn't expired yet
	hasExistingTime := user.HasActiveSubscription() && !user.IsLifetimeSubscriber()

	// If the user has an existing subscription, add time to it instead of replacing
	if hasExistingTime && user.SubscriptionTier == subscriptionTier {
		// If it's the same tier, just keep the existing expiration date
		expirationDate = user.SubscriptionExpiresAt
		log.Printf("Keeping existing expiration date for user %d: %s", user.ID, expirationDate)
	} else if hasExistingTime && subscriptionTier == "yearly" && user.SubscriptionTier == "monthly" {
		// If upgrading from monthly to yearly, add 1 year to the existing expiration date
		expirationDate = user.SubscriptionExpiresAt.AddDate(1, 0, 0)
		log.Printf("Upgrading from monthly to yearly for user %d, new expiration: %s", user.ID, expirationDate)
	} else if subscriptionTier == "premium_lifetime" && user.SubscriptionTier == "lifetime" {
		// If upgrading from lifetime to premium lifetime, keep the same expiration date (already set to 100 years)
		expirationDate = user.SubscriptionExpiresAt
		log.Printf("Upgrading from lifetime to premium lifetime for user %d, keeping expiration: %s", user.ID, expirationDate)
	} else if subscriptionTier == "lifetime" || subscriptionTier == "premium_lifetime" {
		// For lifetime subscriptions, always set to 100 years from now regardless of previous tier
		expirationDate = time.Now().AddDate(100, 0, 0)
		log.Printf("Setting lifetime subscription expiration for user %d to %s", user.ID, expirationDate)
	} else {
		// Start from now for new subscriptions or downgrades
		expirationDate = time.Now()

		// Add time based on the new subscription tier
		switch subscriptionTier {
		case "monthly":
			expirationDate = expirationDate.AddDate(0, 1, 0) // Add 1 month
			log.Printf("Setting monthly subscription expiration for user %d to %s", user.ID, expirationDate)
		case "yearly":
			expirationDate = expirationDate.AddDate(1, 0, 0) // Add 1 year
			log.Printf("Setting yearly subscription expiration for user %d to %s", user.ID, expirationDate)
		case "lifetime", "premium_lifetime":
			expirationDate = time.Now().AddDate(100, 0, 0) // 100 years from now (effectively lifetime)
			log.Printf("Setting lifetime subscription expiration for user %d to %s", user.ID, expirationDate)
		default:
			log.Printf("Unknown subscription tier: %s, defaulting to monthly", subscriptionTier)
			expirationDate = expirationDate.AddDate(0, 1, 0) // Default to 1 month
		}
	}

	// Get or create Stripe customer ID
	var stripeCustomerID string
	if session.Customer != nil {
		stripeCustomerID = session.Customer.ID
	}

	if stripeCustomerID == "" {
		// If no customer ID in the session, use the one from the user if available
		if user.StripeCustomerID != "" {
			stripeCustomerID = user.StripeCustomerID
		} else if os.Getenv("APP_ENV") == "test" {
			// In test mode, use a dummy customer ID
			stripeCustomerID = "cus_test_" + userID
		}
	}

	// Update the user's subscription
	if err := c.DB.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       subscriptionTier,
		"subscription_expires_at": expirationDate,
		"stripe_customer_id":      stripeCustomerID,
		"subscription_canceled":   false, // Reset the canceled flag when resubscribing
	}).Error; err != nil {
		log.Printf("Failed to update user %s for session %s: %v", userID, session.ID, err)
		return
	}

	// Calculate the amount based on the subscription tier
	var amount int64
	if session.AmountTotal > 0 {
		amount = session.AmountTotal
	} else if os.Getenv("APP_ENV") == "test" {
		// In test mode, use dummy amounts
		switch subscriptionTier {
		case "monthly":
			amount = 500 // $5.00
		case "yearly":
			amount = 3000 // $30.00
		case "lifetime":
			amount = 15000 // $150.00
		case "premium_lifetime":
			amount = 30000 // $300.00
		}
	}

	// Create a payment record
	payment := models.Payment{
		UserID:      uint(userIDUint),
		Amount:      amount,
		Currency:    string(session.Currency),
		PaymentType: "subscription",
		Status:      "succeeded",
		Description: strings.Title(subscriptionTier) + " Subscription",
		StripeID:    session.ID,
	}

	if err := models.CreatePayment(c.DB, &payment); err != nil {
		log.Printf("Failed to create payment record for session %s: %v", session.ID, err)
		return
	}

	log.Printf("Created payment record for subscription: %s", subscriptionTier)
	logWebhookEvent("checkout.session.completed", session.ID, userID, "success",
		fmt.Sprintf("User %s subscribed to %s tier", userID, subscriptionTier))
}

// handleSubscriptionCreated processes a new subscription
func handleSubscriptionCreated(c *PaymentController, ctx *gin.Context, event stripe.Event) {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		log.Printf("Error parsing subscription.created event: %v", err)
		return
	}

	// Try to find the user by Stripe customer ID
	var user models.User
	if err := c.DB.Where("stripe_customer_id = ?", subscription.Customer.ID).First(&user).Error; err != nil {
		log.Printf("Failed to find user for customer %s: %v", subscription.Customer.ID, err)
		return
	}

	// Log the event
	logWebhookEvent("customer.subscription.created", subscription.ID, fmt.Sprintf("%d", user.ID), "success",
		fmt.Sprintf("Subscription created for user %d", user.ID))
}

// handleSubscriptionUpdated processes an updated subscription
func handleSubscriptionUpdated(c *PaymentController, ctx *gin.Context, event stripe.Event) {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		log.Printf("Error parsing subscription.updated event: %v", err)
		return
	}

	// Try to find the user by Stripe customer ID
	var user models.User
	if err := c.DB.Where("stripe_customer_id = ?", subscription.Customer.ID).First(&user).Error; err != nil {
		log.Printf("Failed to find user for customer %s: %v", subscription.Customer.ID, err)
		return
	}

	// We'll skip updating the subscription expiration date here
	// since we already set it correctly in handleCheckoutSessionCompleted
	log.Printf("Skipping subscription expiration update for subscription.updated event (user %d)", user.ID)

	// Log the event
	logWebhookEvent("customer.subscription.updated", subscription.ID, fmt.Sprintf("%d", user.ID), "success",
		fmt.Sprintf("Subscription updated for user %d (expiration already set by checkout.session.completed)", user.ID))
}

// handleSubscriptionDeleted processes a cancelled subscription
func handleSubscriptionDeleted(c *PaymentController, ctx *gin.Context, event stripe.Event) {
	var subscription stripe.Subscription
	err := json.Unmarshal(event.Data.Raw, &subscription)
	if err != nil {
		log.Printf("Error parsing subscription.deleted event: %v", err)
		return
	}

	// Try to find the user by Stripe customer ID
	var user models.User
	if err := c.DB.Where("stripe_customer_id = ?", subscription.Customer.ID).First(&user).Error; err != nil {
		log.Printf("Failed to find user for customer %s: %v", subscription.Customer.ID, err)
		return
	}

	// Update the user's subscription to free tier
	if err := c.DB.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       "free",
		"subscription_expires_at": time.Now(),
	}).Error; err != nil {
		log.Printf("Failed to downgrade subscription for user %d: %v", user.ID, err)
		return
	}

	// Log the event
	logWebhookEvent("customer.subscription.deleted", subscription.ID, fmt.Sprintf("%d", user.ID), "success",
		fmt.Sprintf("Subscription cancelled for user %d", user.ID))
}

// handleInvoicePaid processes a paid invoice
func handleInvoicePaid(c *PaymentController, ctx *gin.Context, event stripe.Event) {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		log.Printf("Error parsing invoice.paid event: %v", err)
		return
	}

	// Try to find the user by Stripe customer ID
	var user models.User
	if err := c.DB.Where("stripe_customer_id = ?", invoice.Customer.ID).First(&user).Error; err != nil {
		log.Printf("Failed to find user for customer %s: %v", invoice.Customer.ID, err)
		return
	}

	// Determine the subscription tier from the invoice amount
	var subscriptionTier string

	// Check the invoice amount to determine the tier
	if invoice.AmountPaid >= 100000 {
		// $1000 is premium lifetime (Big Baller)
		subscriptionTier = "premium_lifetime"
		log.Printf("Determined premium_lifetime subscription from invoice amount: %d", invoice.AmountPaid)
	} else if invoice.AmountPaid >= 10000 {
		// $100 is lifetime (Supporter)
		subscriptionTier = "lifetime"
		log.Printf("Determined lifetime subscription from invoice amount: %d", invoice.AmountPaid)
	} else if invoice.AmountPaid >= 3000 {
		// $30 is yearly (Loving It)
		subscriptionTier = "yearly"
		log.Printf("Determined yearly subscription from invoice amount: %d", invoice.AmountPaid)
	} else {
		// $5 is monthly (Liking It)
		subscriptionTier = "monthly"
		log.Printf("Determined monthly subscription from invoice amount: %d", invoice.AmountPaid)
	}

	// For invoice.paid events, we'll skip creating a payment record
	// since we already created one in handleCheckoutSessionCompleted
	log.Printf("Skipping payment record creation for invoice.paid event (user %d, tier %s)", user.ID, subscriptionTier)

	// Log the event
	logWebhookEvent("invoice.paid", invoice.ID, fmt.Sprintf("%d", user.ID), "success",
		fmt.Sprintf("Invoice paid for user %d, amount: %d %s, tier: %s (payment record already created by checkout.session.completed)",
			user.ID, invoice.AmountPaid, string(invoice.Currency), subscriptionTier))
}

// handleInvoicePaymentFailed processes a failed invoice payment
func handleInvoicePaymentFailed(c *PaymentController, ctx *gin.Context, event stripe.Event) {
	var invoice stripe.Invoice
	err := json.Unmarshal(event.Data.Raw, &invoice)
	if err != nil {
		log.Printf("Error parsing invoice.payment_failed event: %v", err)
		return
	}

	// Try to find the user by Stripe customer ID
	var user models.User
	if err := c.DB.Where("stripe_customer_id = ?", invoice.Customer.ID).First(&user).Error; err != nil {
		log.Printf("Failed to find user for customer %s: %v", invoice.Customer.ID, err)
		return
	}

	// Create a payment record for the failed payment
	payment := models.Payment{
		UserID:      user.ID,
		Amount:      invoice.AmountDue,
		Currency:    string(invoice.Currency),
		PaymentType: "invoice",
		Status:      "failed",
		Description: "Failed Invoice Payment",
		StripeID:    invoice.ID,
	}

	if err := models.CreatePayment(c.DB, &payment); err != nil {
		log.Printf("Failed to create payment record for failed invoice %s: %v", invoice.ID, err)
	}

	// Log the event
	logWebhookEvent("invoice.payment_failed", invoice.ID, fmt.Sprintf("%d", user.ID), "failed",
		fmt.Sprintf("Invoice payment failed for user %d, amount: %d %s", user.ID, invoice.AmountDue, invoice.Currency))
}

// HandlePaymentSuccess handles successful payments
func (c *PaymentController) HandlePaymentSuccess(ctx *gin.Context) {
	// Get the session ID from the query parameters
	sessionID := ctx.Query("session_id")
	if sessionID == "" {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Missing session ID"})
		return
	}

	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Check if we're in test mode and using a test session ID
	if os.Getenv("APP_ENV") == "test" && strings.HasPrefix(sessionID, "cs_test_") {
		// In test mode, we don't need to verify with Stripe
		// Just log the success and redirect
		log.Printf("Test payment success for user %d, session %s", user.ID, sessionID)

		// Extract the tier from the session ID (for test purposes)
		var tier string
		if strings.Contains(sessionID, "monthly") {
			tier = "monthly"
		} else if strings.Contains(sessionID, "yearly") {
			tier = "yearly"
		} else if strings.Contains(sessionID, "premium") {
			tier = "premium_lifetime"
		} else if strings.Contains(sessionID, "lifetime") {
			tier = "lifetime"
		} else {
			// Default to monthly if not specified
			tier = "monthly"
		}

		// Calculate expiration date based on tier
		var expirationDate time.Time
		hasExistingTime := !user.SubscriptionExpiresAt.IsZero() && user.SubscriptionExpiresAt.After(time.Now())

		if hasExistingTime && tier == user.SubscriptionTier {
			// If the tier hasn't changed, keep the existing expiration date
			expirationDate = user.SubscriptionExpiresAt
		} else if hasExistingTime && tier == "yearly" && user.SubscriptionTier == "monthly" {
			// If upgrading from monthly to yearly, add 1 year to the existing expiration date
			expirationDate = user.SubscriptionExpiresAt.AddDate(1, 0, 0)
		} else if tier == "premium_lifetime" && user.SubscriptionTier == "lifetime" {
			// If upgrading from lifetime to premium lifetime, keep the same expiration date
			expirationDate = user.SubscriptionExpiresAt
		} else if tier == "lifetime" || tier == "premium_lifetime" {
			// For lifetime subscriptions, set to 100 years from now
			expirationDate = time.Now().AddDate(100, 0, 0)
		} else {
			// Start from now for new subscriptions or downgrades
			expirationDate = time.Now()

			// Add time based on the new subscription tier
			switch tier {
			case "monthly":
				expirationDate = expirationDate.AddDate(0, 1, 0) // Add 1 month
			case "yearly":
				expirationDate = expirationDate.AddDate(1, 0, 0) // Add 1 year
			}
		}

		// Update the user's subscription
		if err := c.DB.Model(user).Updates(map[string]interface{}{
			"subscription_tier":       tier,
			"subscription_expires_at": expirationDate,
			"subscription_canceled":   false,
		}).Error; err != nil {
			log.Printf("Failed to update user %d subscription in test mode: %v", user.ID, err)
		}

		// Create a payment record
		var amount int64
		switch tier {
		case "monthly":
			amount = 500 // $5.00
		case "yearly":
			amount = 3000 // $30.00
		case "lifetime":
			amount = 15000 // $150.00
		case "premium_lifetime":
			amount = 30000 // $300.00
		}

		payment := models.Payment{
			UserID:      user.ID,
			Amount:      amount,
			Currency:    "usd",
			PaymentType: "subscription",
			Status:      "succeeded",
			Description: strings.Title(tier) + " Subscription",
			StripeID:    sessionID,
		}

		if err := models.CreatePayment(c.DB, &payment); err != nil {
			log.Printf("Failed to create payment record in test mode: %v", err)
		}

		// Set a success message
		flash.SetMessage(ctx, "Your payment was successful! Thank you for your subscription.", "success")

		// Redirect to the owner page
		ctx.Redirect(http.StatusSeeOther, "/owner")
		return
	}

	// Set Stripe API key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Retrieve the session from Stripe to verify it
	s, err := session.Get(sessionID, nil)
	if err != nil {
		log.Printf("Error retrieving session %s: %v", sessionID, err)
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to verify payment"})
		return
	}

	// Verify that the session belongs to this user
	userIDStr, ok := s.Metadata["user_id"]
	if !ok || userIDStr != strconv.FormatUint(uint64(user.ID), 10) {
		log.Printf("Session %s does not belong to user %d", sessionID, user.ID)
		ctx.HTML(http.StatusForbidden, "error.html", gin.H{"error": "Invalid session"})
		return
	}

	// Log the successful payment
	log.Printf("Payment success for user %d, session %s", user.ID, sessionID)

	// Set a success message
	ctx.SetCookie("flash_message", "Your payment was successful! Thank you for your subscription.", 5, "/", "", false, true)
	ctx.SetCookie("flash_type", "success", 5, "/", "", false, true)

	// Redirect to the owner page
	ctx.Redirect(http.StatusSeeOther, "/owner")
}

// HandlePaymentCancellation handles cancelled payments
func (c *PaymentController) HandlePaymentCancellation(ctx *gin.Context) {
	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err == nil {
		// Log the cancellation
		log.Printf("Payment cancelled by user %d", user.ID)
	}

	// Set a message
	flash.SetMessage(ctx, "Your payment was cancelled. If you have any questions, please contact support.", "info")

	// Redirect to the pricing page
	ctx.Redirect(http.StatusSeeOther, "/pricing")
}

// ShowPaymentHistory displays the payment history for the current user
func (c *PaymentController) ShowPaymentHistory(ctx *gin.Context) {
	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Get the user's payment history
	payments, err := models.GetPaymentsByUserID(c.DB, user.ID)
	if err != nil {
		// Handle error
		log.Printf("Failed to retrieve payment history for user %d: %v", user.ID, err)
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to retrieve payment history",
		})
		return
	}

	// Get flash message from cookie
	flashMessage, _ := ctx.Cookie("flash_message")
	flashType, _ := ctx.Cookie("flash_type")

	// Log the flash message for debugging
	if flashMessage != "" {
		log.Printf("Payment history flash message found: %s (type: %s)", flashMessage, flashType)
	}

	// Render the payment history template with flash message
	component := paymentViews.PaymentHistory(user, payments, flashMessage, flashType)
	component.Render(ctx.Request.Context(), ctx.Writer)

	// Clear flash cookies after rendering
	if flashMessage != "" {
		flash.ClearMessage(ctx)
	}
}

// ShowCancelConfirmation displays the subscription cancellation confirmation page
func (c *PaymentController) ShowCancelConfirmation(ctx *gin.Context) {
	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Check if the user has an active subscription
	if user.SubscriptionTier == "free" || user.IsLifetimeSubscriber() {
		// Redirect to payment history if the user doesn't have a recurring subscription
		flash.SetMessage(ctx, "You don't have an active recurring subscription to cancel.", "error")
		ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
		return
	}

	// Check if the subscription is already canceled
	if user.SubscriptionCanceled {
		flash.SetMessage(ctx, "Your subscription is already scheduled for cancellation.", "info")
		ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
		return
	}

	// Render the cancellation confirmation page
	component := paymentViews.CancelConfirmation(user)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// CancelSubscription cancels the user's subscription
func (c *PaymentController) CancelSubscription(ctx *gin.Context) {
	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Check if the user has an active subscription
	if user.SubscriptionTier == "free" || user.IsLifetimeSubscriber() {
		// Redirect to payment history if the user doesn't have a recurring subscription
		flash.SetMessage(ctx, "You don't have an active recurring subscription to cancel.", "error")
		ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
		return
	}

	// Check if the subscription is already canceled
	if user.SubscriptionCanceled {
		flash.SetMessage(ctx, "Your subscription is already scheduled for cancellation.", "info")
		ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
		return
	}

	// Check if we're in test mode
	if os.Getenv("APP_ENV") == "test" {
		// In test mode, just mark the subscription as canceled
		if err := c.DB.Model(user).Update("subscription_canceled", true).Error; err != nil {
			log.Printf("Failed to cancel subscription for user %d in test mode: %v", user.ID, err)
			flash.SetMessage(ctx, "Failed to cancel subscription. Please try again.", "error")
			ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
			return
		}

		// Set success message
		flash.SetMessage(ctx, "Your subscription has been canceled. You will continue to have access until "+user.SubscriptionExpiresAt.Format("January 2, 2006")+".", "success")
		ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
		return
	}

	// Set Stripe API key
	stripe.Key = os.Getenv("STRIPE_SECRET_KEY")

	// Check if the user has a Stripe subscription ID
	if user.StripeSubscriptionID == "" {
		// Try to find the subscription by customer ID
		params := &stripe.SubscriptionListParams{}
		params.Customer = user.StripeCustomerID
		iter := sub.List(params)
		var subscription *stripe.Subscription
		for iter.Next() {
			subscription = iter.Subscription()
			break
		}

		if subscription == nil {
			log.Printf("Failed to find subscription for user %d with customer ID %s", user.ID, user.StripeCustomerID)
			flash.SetMessage(ctx, "Failed to find your subscription. Please contact support.", "error")
			ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
			return
		}

		// Save the subscription ID for future use
		if err := c.DB.Model(user).Update("stripe_subscription_id", subscription.ID).Error; err != nil {
			log.Printf("Failed to save subscription ID for user %d: %v", user.ID, err)
		}

		// Use the found subscription ID
		user.StripeSubscriptionID = subscription.ID
	}

	// Cancel the subscription at period end
	_, err = sub.Update(user.StripeSubscriptionID, &stripe.SubscriptionParams{
		CancelAtPeriodEnd: stripe.Bool(true),
	})

	if err != nil {
		log.Printf("Failed to cancel subscription for user %d: %v", user.ID, err)
		flash.SetMessage(ctx, "Failed to cancel subscription with Stripe. Please try again.", "error")
		ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
		return
	}

	// Mark the subscription as canceled in our database
	if err := c.DB.Model(user).Update("subscription_canceled", true).Error; err != nil {
		log.Printf("Failed to mark subscription as canceled for user %d: %v", user.ID, err)
	}

	// Set success message
	flash.SetMessage(ctx, "Your subscription has been canceled. You will continue to have access until "+user.SubscriptionExpiresAt.Format("January 2, 2006")+".", "success")
	ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
}

// HandleCheckoutRedirect handles GET requests to /checkout and redirects to the appropriate Stripe payment link
func (c *PaymentController) HandleCheckoutRedirect(ctx *gin.Context) {
	// Get the subscription tier from the query parameters
	tier := ctx.Query("tier")
	if tier == "" {
		ctx.Redirect(http.StatusSeeOther, "/pricing")
		return
	}

	// Check if user is logged in
	_, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// Get the appropriate Stripe payment link
	var paymentLink string
	switch tier {
	case "monthly":
		paymentLink = os.Getenv("STRIPE_LINK_MONTHLY")
	case "yearly":
		paymentLink = os.Getenv("STRIPE_LINK_YEARLY")
	case "lifetime":
		paymentLink = os.Getenv("STRIPE_LINK_LIFETIME")
	case "premium_lifetime":
		paymentLink = os.Getenv("STRIPE_LINK_PREMIUM")
	default:
		ctx.Redirect(http.StatusSeeOther, "/pricing")
		return
	}

	// If the payment link is not set, fall back to the POST checkout endpoint
	if paymentLink == "" {
		// Render a form that submits to the POST /checkout endpoint
		html := `
		<!DOCTYPE html>
		<html>
		<head>
			<title>Redirecting to Checkout</title>
			<script>
				document.addEventListener('DOMContentLoaded', function() {
					document.getElementById('checkout-form').submit();
				});
			</script>
		</head>
		<body>
			<form id="checkout-form" method="POST" action="/checkout">
				<input type="hidden" name="tier" value="` + tier + `">
				<p>Redirecting to checkout...</p>
				<button type="submit">Click here if you are not redirected automatically</button>
			</form>
		</body>
		</html>
		`
		ctx.Header("Content-Type", "text/html")
		ctx.String(http.StatusOK, html)
		return
	}

	// Redirect to the Stripe payment link
	ctx.Redirect(http.StatusSeeOther, paymentLink)
}
