package routes

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/config"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/middleware"
	"github.com/hail2skins/the-virtual-armory/internal/services/email"
)

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(router *gin.Engine, auth *auth.Auth, emailService email.EmailService, config *config.Config) {
	// Create the auth controller
	authController := controllers.NewAuthController(auth, emailService, config)

	// Create rate limiters
	loginLimiter := middleware.NewRateLimiter()
	passwordResetLimiter := middleware.NewRateLimiter()

	// Auth routes - all without /auth prefix
	router.GET("/login", authController.Login)
	router.POST("/login", loginLimiter.RateLimit(5, time.Minute), authController.ProcessLogin)
	router.GET("/register", authController.Register)
	router.POST("/register", authController.ProcessRegister)
	router.GET("/recover", authController.Recover)
	router.POST("/recover", passwordResetLimiter.RateLimit(3, time.Hour), authController.ProcessRecover)
	router.GET("/logout", authController.Logout)
	router.GET("/verification-pending", authController.VerificationPending)
	router.POST("/resend-verification", authController.ResendVerification)
	router.GET("/verify/:token", authController.VerifyEmail)
	router.GET("/reset-password/:token", authController.ResetPassword)
	router.POST("/reset-password/:token", authController.ProcessResetPassword)

	// Account reactivation routes
	router.GET("/reactivate", authController.ReactivateAccount)
	router.POST("/reactivate", authController.ProcessReactivation)

	// Protected routes (require authentication)
	protected := router.Group("/")
	protected.Use(auth.RequireAuth())
	{
		protected.GET("/owner", authController.Profile)
		protected.GET("/delete-account", authController.DeleteAccount)
		protected.POST("/delete-account", authController.ProcessDeleteAccount)
	}
}
