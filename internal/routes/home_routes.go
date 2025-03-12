package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/services/email"
)

// RegisterHomeRoutes registers all home-related routes
func RegisterHomeRoutes(router *gin.Engine, emailService email.EmailService) {
	homeController := controllers.NewHomeController(emailService)

	// Create a home routes group
	homeGroup := router.Group("/")
	{
		// Register routes
		homeGroup.GET("/", homeController.Index)
		homeGroup.POST("/hello", homeController.HandleHelloForm) // For form submission
		homeGroup.GET("/about", homeController.About)
		homeGroup.GET("/contact", homeController.Contact)
		homeGroup.POST("/contact", homeController.HandleContactForm) // For contact form submission
	}
}
