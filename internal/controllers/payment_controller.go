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
						Name: stripe.String(productName),
					},
					UnitAmount: stripe.Int64(unitAmount),
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
		SuccessURL:    stripe.String(os.Getenv("APP_BASE_URL") + "/payment/success?session_id={CHECKOUT_SESSION_ID}"),
		CancelURL:     stripe.String(os.Getenv("APP_BASE_URL") + "/payment/cancel"),
		CustomerEmail: stripe.String(user.Email),
	}

	// Add metadata
	params.AddMetadata("user_id", strconv.FormatUint(uint64(user.ID), 10))
	params.AddMetadata("subscription_tier", tier)
	params.AddMetadata("product_id", productID)

	// Create the checkout session
	s, err := session.New(params)
	if err != nil {
		log.Printf("Error creating checkout session: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
		return
	}

	log.Printf("Created checkout session for user %d, tier %s, session ID: %s", user.ID, tier, s.ID)

	// Redirect to the checkout page
	ctx.Redirect(http.StatusSeeOther, s.URL)
}

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
		log.Printf("Error parsing checkout.session.completed event: %v", err)
		return
	}

	// Get the user ID and subscription tier from the metadata
	userIDStr, ok := session.Metadata["user_id"]
	if !ok {
		log.Printf("Missing user_id in metadata for session %s", session.ID)
		return
	}

	subscriptionTier, ok := session.Metadata["subscription_tier"]
	if !ok {
		log.Printf("Missing subscription_tier in metadata for session %s", session.ID)
		return
	}

	// Convert user ID to uint
	userID, err := strconv.ParseUint(userIDStr, 10, 64)
	if err != nil {
		log.Printf("Invalid user_id in metadata for session %s: %v", session.ID, err)
		return
	}

	// Find the user in the database
	var user models.User
	if err := c.DB.First(&user, userID).Error; err != nil {
		log.Printf("Failed to find user %s for session %s: %v", userIDStr, session.ID, err)
		return
	}

	// Set expiration date based on subscription tier
	var expirationDate time.Time
	switch subscriptionTier {
	case "monthly":
		expirationDate = time.Now().AddDate(0, 1, 0) // 1 month from now
	case "yearly":
		expirationDate = time.Now().AddDate(1, 0, 0) // 1 year from now
	case "lifetime", "premium_lifetime":
		expirationDate = time.Now().AddDate(100, 0, 0) // 100 years from now (effectively lifetime)
	default:
		expirationDate = time.Now() // No subscription
	}

	// If this is a subscription (not a one-time payment), store the subscription ID
	var stripeCustomerID string
	if session.Subscription != nil && session.Customer != nil {
		// Update the user's Stripe customer ID if available
		stripeCustomerID = session.Customer.ID
	} else if os.Getenv("APP_ENV") == "test" {
		// In test mode, use a dummy customer ID
		stripeCustomerID = "cus_test_" + userIDStr
	}

	// Update the user's subscription
	if err := c.DB.Model(&user).Updates(map[string]interface{}{
		"subscription_tier":       subscriptionTier,
		"subscription_expires_at": expirationDate,
		"stripe_customer_id":      stripeCustomerID,
	}).Error; err != nil {
		log.Printf("Failed to update user %s for session %s: %v", userIDStr, session.ID, err)
		return
	}

	// Create a payment record
	var amount int64 = 0
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
			amount = 10000 // $100.00
		case "premium_lifetime":
			amount = 100000 // $1000.00
		}
	}

	payment := models.Payment{
		UserID:      uint(userID),
		Amount:      amount,
		Currency:    string(session.Currency),
		PaymentType: "subscription",
		Status:      "succeeded",
		Description: subscriptionTier + " Subscription",
		StripeID:    session.ID,
	}

	if err := models.CreatePayment(c.DB, &payment); err != nil {
		log.Printf("Failed to create payment record for session %s: %v", session.ID, err)
		return
	}

	logWebhookEvent("checkout.session.completed", session.ID, userIDStr, "success",
		fmt.Sprintf("User %s subscribed to %s tier", userIDStr, subscriptionTier))
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

	// Update the subscription expiration date based on the current period end
	expirationDate := time.Unix(subscription.CurrentPeriodEnd, 0)

	// Update the user's subscription expiration
	if err := c.DB.Model(&user).Update("subscription_expires_at", expirationDate).Error; err != nil {
		log.Printf("Failed to update subscription expiration for user %d: %v", user.ID, err)
		return
	}

	// Log the event
	logWebhookEvent("customer.subscription.updated", subscription.ID, fmt.Sprintf("%d", user.ID), "success",
		fmt.Sprintf("Subscription updated for user %d, new expiration: %s", user.ID, expirationDate))
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

	// If this is a subscription invoice, update the expiration date
	if invoice.Subscription != nil && invoice.PeriodEnd > 0 {
		// Use the period end from the invoice directly
		expirationDate := time.Unix(invoice.PeriodEnd, 0)
		if err := c.DB.Model(&user).Update("subscription_expires_at", expirationDate).Error; err != nil {
			log.Printf("Failed to update subscription expiration for user %d: %v", user.ID, err)
		} else {
			log.Printf("Updated subscription expiration for user %d to %s", user.ID, expirationDate)
		}
	}

	// Create a payment record
	payment := models.Payment{
		UserID:      user.ID,
		Amount:      invoice.AmountPaid,
		Currency:    string(invoice.Currency),
		PaymentType: "invoice",
		Status:      "succeeded",
		Description: "Invoice Payment",
		StripeID:    invoice.ID,
	}

	if err := models.CreatePayment(c.DB, &payment); err != nil {
		log.Printf("Failed to create payment record for invoice %s: %v", invoice.ID, err)
	}

	// Log the event
	logWebhookEvent("invoice.paid", invoice.ID, fmt.Sprintf("%d", user.ID), "success",
		fmt.Sprintf("Invoice paid for user %d, amount: %d %s", user.ID, invoice.AmountPaid, invoice.Currency))
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

		// Extract the tier from the session ID or use a default
		tier := "monthly" // Default to monthly in test mode

		// Set expiration date based on subscription tier
		var expirationDate time.Time

		// Check if user has an existing subscription that hasn't expired yet
		hasExistingTime := user.HasActiveSubscription() && !user.IsLifetimeSubscriber()

		// If the user has an existing subscription, add time to it instead of replacing
		if hasExistingTime {
			// Start from the current expiration date
			expirationDate = user.SubscriptionExpiresAt
		} else {
			// Start from now
			expirationDate = time.Now()
		}

		// Add time based on the new subscription tier
		switch tier {
		case "monthly":
			expirationDate = expirationDate.AddDate(0, 1, 0) // Add 1 month
		case "yearly":
			expirationDate = expirationDate.AddDate(1, 0, 0) // Add 1 year
		case "lifetime", "premium_lifetime":
			expirationDate = time.Now().AddDate(100, 0, 0) // 100 years from now (effectively lifetime)
		}

		// Update the user's subscription
		if err := c.DB.Model(user).Updates(map[string]interface{}{
			"subscription_tier":       tier,
			"subscription_expires_at": expirationDate,
			"stripe_customer_id":      "cus_test_" + strconv.FormatUint(uint64(user.ID), 10),
			"subscription_canceled":   false, // Reset the canceled flag when resubscribing
		}).Error; err != nil {
			log.Printf("Failed to update user %d subscription in test mode: %v", user.ID, err)
		}

		// Create a payment record
		payment := models.Payment{
			UserID:      user.ID,
			Amount:      500, // $5.00 for monthly
			Currency:    "usd",
			PaymentType: "subscription",
			Status:      "succeeded",
			Description: tier + " Subscription",
			StripeID:    sessionID,
		}

		if err := models.CreatePayment(c.DB, &payment); err != nil {
			log.Printf("Failed to create payment record in test mode: %v", err)
		}

		// Set a success message with cookies that are accessible to JavaScript
		// MaxAge: 60 seconds, Path: /, Secure: false, HttpOnly: false
		ctx.SetCookie("flash_message", "Your payment was successful! Thank you for your subscription.", 60, "/", "", false, false)
		ctx.SetCookie("flash_type", "success", 60, "/", "", false, false)

		// NEVER CHANGE THIS REDIRECT - IT MUST ALWAYS GO TO /owner
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

	// Get the subscription tier from the metadata
	tier, ok := s.Metadata["subscription_tier"]
	if !ok {
		log.Printf("Missing subscription_tier in metadata for session %s", s.ID)
		tier = "monthly" // Default to monthly if not specified
	}

	// Set expiration date based on subscription tier
	var expirationDate time.Time

	// Check if user has an existing subscription that hasn't expired yet
	hasExistingTime := user.HasActiveSubscription() && !user.IsLifetimeSubscriber()

	// If the user has an existing subscription, add time to it instead of replacing
	if hasExistingTime {
		// Start from the current expiration date
		expirationDate = user.SubscriptionExpiresAt
	} else {
		// Start from now
		expirationDate = time.Now()
	}

	// Add time based on the new subscription tier
	switch tier {
	case "monthly":
		expirationDate = expirationDate.AddDate(0, 1, 0) // Add 1 month
	case "yearly":
		expirationDate = expirationDate.AddDate(1, 0, 0) // Add 1 year
	case "lifetime", "premium_lifetime":
		expirationDate = time.Now().AddDate(100, 0, 0) // 100 years from now (effectively lifetime)
	}

	// Update the user's subscription
	if err := c.DB.Model(user).Updates(map[string]interface{}{
		"subscription_tier":       tier,
		"subscription_expires_at": expirationDate,
		"stripe_customer_id":      s.Customer.ID,
		"subscription_canceled":   false, // Reset the canceled flag when resubscribing
	}).Error; err != nil {
		log.Printf("Failed to update user %d subscription: %v", user.ID, err)
	}

	// Create a payment record
	var amount int64 = 0
	if s.AmountTotal > 0 {
		amount = s.AmountTotal
	} else {
		// Use default amounts if not available
		switch tier {
		case "monthly":
			amount = 500 // $5.00
		case "yearly":
			amount = 3000 // $30.00
		case "lifetime":
			amount = 10000 // $100.00
		case "premium_lifetime":
			amount = 100000 // $1000.00
		}
	}

	payment := models.Payment{
		UserID:      user.ID,
		Amount:      amount,
		Currency:    string(s.Currency),
		PaymentType: "subscription",
		Status:      "succeeded",
		Description: tier + " Subscription",
		StripeID:    s.ID,
	}

	if err := models.CreatePayment(c.DB, &payment); err != nil {
		log.Printf("Failed to create payment record: %v", err)
	}

	// Log the successful payment
	log.Printf("Payment success for user %d, session %s", user.ID, sessionID)

	// Set a success message with cookies that are accessible to JavaScript
	// MaxAge: 60 seconds, Path: /, Secure: false, HttpOnly: false
	ctx.SetCookie("flash_message", "Your payment was successful! Thank you for your subscription.", 60, "/", "", false, false)
	ctx.SetCookie("flash_type", "success", 60, "/", "", false, false)

	// NEVER CHANGE THIS REDIRECT - IT MUST ALWAYS GO TO /owner
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
	ctx.SetCookie("flash_message", "Your payment was cancelled. If you have any questions, please contact support.", 5, "/", "", false, true)
	ctx.SetCookie("flash_type", "info", 5, "/", "", false, true)

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
		ctx.SetCookie("flash_message", "", -1, "/", "", false, false)
		ctx.SetCookie("flash_type", "", -1, "/", "", false, false)
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
		ctx.SetCookie("flash_message", "You don't have an active recurring subscription to cancel.", 5, "/", "", false, true)
		ctx.SetCookie("flash_type", "error", 5, "/", "", false, true)
		ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
		return
	}

	// Check if the subscription is already canceled
	if user.SubscriptionCanceled {
		ctx.SetCookie("flash_message", "Your subscription is already scheduled for cancellation.", 5, "/", "", false, true)
		ctx.SetCookie("flash_type", "info", 5, "/", "", false, true)
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
		ctx.SetCookie("flash_message", "You don't have an active recurring subscription to cancel.", 5, "/", "", false, true)
		ctx.SetCookie("flash_type", "error", 5, "/", "", false, true)
		ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
		return
	}

	// Check if the subscription is already canceled
	if user.SubscriptionCanceled {
		ctx.SetCookie("flash_message", "Your subscription is already scheduled for cancellation.", 5, "/", "", false, true)
		ctx.SetCookie("flash_type", "info", 5, "/", "", false, true)
		ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
		return
	}

	// Check if we're in test mode
	if os.Getenv("APP_ENV") == "test" {
		// In test mode, just mark the subscription as canceled
		if err := c.DB.Model(user).Update("subscription_canceled", true).Error; err != nil {
			log.Printf("Failed to cancel subscription for user %d in test mode: %v", user.ID, err)
			ctx.SetCookie("flash_message", "Failed to cancel subscription. Please try again.", 60, "/", "", false, false)
			ctx.SetCookie("flash_type", "error", 60, "/", "", false, false)
			ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
			return
		}

		// Set success message
		ctx.SetCookie("flash_message", "Your subscription has been canceled. You will continue to have access until "+user.SubscriptionExpiresAt.Format("January 2, 2006")+".", 60, "/", "", false, false)
		ctx.SetCookie("flash_type", "success", 60, "/", "", false, false)
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
			ctx.SetCookie("flash_message", "Failed to find your subscription. Please contact support.", 60, "/", "", false, false)
			ctx.SetCookie("flash_type", "error", 60, "/", "", false, false)
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
		ctx.SetCookie("flash_message", "Failed to cancel subscription with Stripe. Please try again.", 60, "/", "", false, false)
		ctx.SetCookie("flash_type", "error", 60, "/", "", false, false)
		ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
		return
	}

	// Mark the subscription as canceled in our database
	if err := c.DB.Model(user).Update("subscription_canceled", true).Error; err != nil {
		log.Printf("Failed to mark subscription as canceled for user %d: %v", user.ID, err)
	}

	// Set success message
	ctx.SetCookie("flash_message", "Your subscription has been canceled. You will continue to have access until "+user.SubscriptionExpiresAt.Format("January 2, 2006")+".", 60, "/", "", false, false)
	ctx.SetCookie("flash_type", "success", 60, "/", "", false, false)
	ctx.Redirect(http.StatusSeeOther, "/owner/payment-history")
}
