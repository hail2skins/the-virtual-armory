package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/services/email"
	"gorm.io/gorm"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(router *gin.Engine, db *gorm.DB, auth *auth.Auth, emailService email.EmailService) {
	// Create user controller with email service
	userController := controllers.NewUserControllerWithEmailService(db, emailService)

	// Protected routes (require authentication)
	protected := router.Group("/")
	protected.Use(auth.RequireAuth())
	{
		// Profile routes
		protected.GET("/profile", userController.ShowProfile)
		protected.GET("/profile/edit", userController.EditProfile)
		protected.POST("/profile/update", userController.UpdateProfile)
		protected.GET("/profile/subscription", userController.ShowSubscription)
		protected.GET("/profile/delete", userController.ShowDeleteAccount)
		protected.POST("/profile/delete", userController.DeleteAccount)
		protected.POST("/profile/reactivate", userController.ReactivateAccount)
	}
}
