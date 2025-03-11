package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/middleware"
	"gorm.io/gorm"
)

// RegisterPaymentRoutes registers all payment related routes
func RegisterPaymentRoutes(r *gin.Engine, db *gorm.DB, authInstance *auth.Auth) {
	// Create payment controller
	paymentController := controllers.NewPaymentController(db)

	// Create rate limiter for webhooks
	webhookLimiter := middleware.NewRateLimiter()

	// Public routes
	r.GET("/pricing", paymentController.ShowPricingPage)

	// Webhook route for Stripe (public) with monitoring and rate limiting middleware
	r.POST("/webhook",
		middleware.WebhookMonitor(),
		webhookLimiter.RateLimit(10, time.Minute),
		paymentController.HandleStripeWebhook,
	)

	// Webhook health check endpoint (admin only)
	admin := r.Group("/admin")
	admin.Use(authInstance.RequireAuth())
	admin.Use(authInstance.RequireAdmin()) // Use the existing RequireAdmin middleware
	{
		admin.GET("/webhook-health", middleware.WebhookHealthCheck())
	}

	// Payment success/cancel routes (public)
	r.GET("/payment/success", paymentController.HandlePaymentSuccess)
	r.GET("/payment/cancel", paymentController.HandlePaymentCancellation)

	// Protected routes for checkout and subscription management
	authorized := r.Group("/")
	authorized.Use(authInstance.RequireAuth())
	{
		// Checkout routes
		authorized.GET("/checkout", paymentController.HandleCheckoutRedirect)
		authorized.POST("/checkout", paymentController.CreateCheckoutSession)

		// Payment history route
		authorized.GET("/owner/payment-history", paymentController.ShowPaymentHistory)

		// Subscription cancellation routes
		authorized.GET("/subscription/cancel/confirm", paymentController.ShowCancelConfirmation)
		authorized.POST("/subscription/cancel", paymentController.CancelSubscription)
	}
}
