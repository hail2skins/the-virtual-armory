package controllers

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/payment"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/checkout/session"
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
	component := payment.Pricing(user)
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
		productID = os.Getenv("STRIPE_PRICE_PREMIUM_LIFETIME")
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
		unitAmount = 999 // $9.99
		productName = "Monthly Subscription"
		interval = "month"
	case "yearly":
		unitAmount = 9999 // $99.99
		productName = "Yearly Subscription"
		interval = "year"
	case "lifetime":
		unitAmount = 19999 // $199.99
		productName = "Lifetime Access"
		interval = ""
	case "premium_lifetime":
		unitAmount = 29999 // $299.99
		productName = "Premium Lifetime Access"
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create checkout session"})
		return
	}

	// Redirect to the checkout page
	ctx.Redirect(http.StatusSeeOther, s.URL)
}

// HandleStripeWebhook handles Stripe webhook events
func (c *PaymentController) HandleStripeWebhook(ctx *gin.Context) {
	// Get the webhook secret from environment variables
	webhookSecret := os.Getenv("STRIPE_WEBHOOK_SECRET")

	// Read the request body
	body, err := ioutil.ReadAll(ctx.Request.Body)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request body"})
		return
	}

	// Get the signature from the headers
	signature := ctx.GetHeader("Stripe-Signature")

	// Verify the webhook signature
	event, err := webhook.ConstructEvent(body, signature, webhookSecret)
	if err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid webhook signature"})
		return
	}

	// Handle different event types
	switch event.Type {
	case "checkout.session.completed":
		// Parse the event data
		var session stripe.CheckoutSession
		err := json.Unmarshal(event.Data.Raw, &session)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Failed to parse event data"})
			return
		}

		// Get the user ID and subscription tier from the metadata
		userIDStr, ok := session.Metadata["user_id"]
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing user_id in metadata"})
			return
		}

		subscriptionTier, ok := session.Metadata["subscription_tier"]
		if !ok {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Missing subscription_tier in metadata"})
			return
		}

		// Convert user ID to uint
		userID, err := strconv.ParseUint(userIDStr, 10, 64)
		if err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user_id in metadata"})
			return
		}

		// Find the user in the database
		var user models.User
		if err := c.DB.First(&user, userID).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to find user"})
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

		// Update the user's subscription
		if err := c.DB.Model(&user).Updates(map[string]interface{}{
			"subscription_tier":       subscriptionTier,
			"subscription_expires_at": expirationDate,
		}).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}

		// Create a payment record
		payment := models.Payment{
			UserID:      uint(userID),
			Amount:      session.AmountTotal,
			Currency:    string(session.Currency),
			PaymentType: "subscription",
			Status:      "succeeded",
			Description: subscriptionTier + " Subscription",
			StripeID:    session.ID,
		}

		if err := models.CreatePayment(c.DB, &payment); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create payment record"})
			return
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"status": "success"})
}

// HandlePaymentSuccess handles successful payments
func (c *PaymentController) HandlePaymentSuccess(ctx *gin.Context) {
	// Get the session ID from the query parameters
	// We're not using sessionID yet, but we'll need it in a real implementation
	// to verify the payment with Stripe
	_ = ctx.Query("session_id")

	// Get the current user
	// We're not using user yet, but we'll need it in a real implementation
	// to update the user's subscription status
	_, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusSeeOther, "/login")
		return
	}

	// In a real implementation, we would verify the session ID with Stripe
	// For now, we'll just redirect to the guns page
	ctx.Redirect(http.StatusSeeOther, "/owner/guns")
}

// HandlePaymentCancellation handles cancelled payments
func (c *PaymentController) HandlePaymentCancellation(ctx *gin.Context) {
	// In a real implementation, we would log the cancellation
	// For now, we'll just redirect to the pricing page
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
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "Failed to retrieve payment history",
		})
		return
	}

	// Render the payment history template
	component := payment.PaymentHistory(user, payments)
	component.Render(ctx.Request.Context(), ctx.Writer)
}
