package routes

import (
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

	// Public routes
	r.GET("/pricing", paymentController.ShowPricingPage)

	// Webhook route for Stripe (public) with monitoring middleware
	r.POST("/webhook", middleware.WebhookMonitor(), paymentController.HandleStripeWebhook)

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
		// Checkout route
		authorized.POST("/checkout", paymentController.CreateCheckoutSession)

		// Payment history route
		authorized.GET("/owner/payment-history", paymentController.ShowPaymentHistory)
	}
}
