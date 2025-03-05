package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/config"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/services/email"
)

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(router *gin.Engine, auth *auth.Auth, emailService email.EmailService, config *config.Config) {
	// Create the auth controller
	authController := controllers.NewAuthController(auth, emailService, config)

	// Auth routes - all without /auth prefix
	router.GET("/login", authController.Login)
	router.POST("/login", authController.ProcessLogin)
	router.GET("/register", authController.Register)
	router.POST("/register", authController.ProcessRegister)
	router.GET("/recover", authController.Recover)
	router.POST("/recover", authController.ProcessRecover)
	router.GET("/logout", authController.Logout)
	router.GET("/verification-pending", authController.VerificationPending)
	router.POST("/resend-verification", authController.ResendVerification)
	router.GET("/verify/:token", authController.VerifyEmail)

	// Protected routes (require authentication)
	protected := router.Group("/")
	protected.Use(auth.RequireAuth())
	{
		protected.GET("/owner", authController.Profile)
	}

	// Admin routes (require admin privileges)
	admin := router.Group("/admin")
	admin.Use(auth.RequireAdmin())
	{
		admin.GET("/dashboard", authController.AdminDashboard)
	}
}
