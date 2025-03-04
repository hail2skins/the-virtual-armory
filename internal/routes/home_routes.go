package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
)

// RegisterHomeRoutes registers all home-related routes
func RegisterHomeRoutes(router *gin.Engine) {
	homeController := controllers.NewHomeController()

	// Create a home routes group
	homeGroup := router.Group("/")
	{
		homeGroup.GET("/", homeController.Index)
		homeGroup.POST("/hello", homeController.HandleHelloForm) // For form submission
		homeGroup.GET("/about", homeController.About)
		homeGroup.GET("/contact", homeController.Contact)
	}
}
