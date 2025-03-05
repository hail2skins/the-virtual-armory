package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
)

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(router *gin.Engine, auth *auth.Auth) {
	// Create the auth controller
	authController := controllers.NewAuthController(auth)

	// Auth routes with /auth prefix
	authGroup := router.Group("/auth")
	{
		// Login routes
		authGroup.GET("/login", authController.Login)
		authGroup.POST("/login", authController.ProcessLogin)

		// Password recovery routes
		authGroup.GET("/recover", authController.Recover)
		authGroup.POST("/recover", authController.ProcessRecover)
	}

	// Auth routes without /auth prefix for convenience
	router.GET("/login", authController.Login)
	router.POST("/login", authController.ProcessLogin)
	router.GET("/register", authController.Register)
	router.POST("/register", authController.ProcessRegister)
	router.GET("/recover", authController.Recover)
	router.POST("/recover", authController.ProcessRecover)
	router.GET("/logout", authController.Logout)

	// Protected routes (require authentication)
	protected := router.Group("/")
	protected.Use(auth.RequireAuth())
	{
		protected.GET("/profile", authController.Profile)
	}

	// Admin routes (require admin privileges)
	admin := router.Group("/admin")
	admin.Use(auth.RequireAdmin())
	{
		admin.GET("/dashboard", authController.AdminDashboard)
	}
}
