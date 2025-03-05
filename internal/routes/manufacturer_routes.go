package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
)

// RegisterManufacturerRoutes registers all manufacturer-related routes
func RegisterManufacturerRoutes(router *gin.Engine, auth *auth.Auth) {
	// Create the manufacturer controller
	manufacturerController := controllers.NewManufacturerController()

	// Admin-only routes for manufacturers
	admin := router.Group("/admin")
	admin.Use(auth.RequireAdmin())
	{
		// Manufacturers index
		admin.GET("/manufacturers", manufacturerController.Index)

		// New manufacturer form
		admin.GET("/manufacturers/new", manufacturerController.New)

		// Create manufacturer
		admin.POST("/manufacturers", manufacturerController.Create)

		// Show manufacturer details
		admin.GET("/manufacturers/:id", manufacturerController.Show)

		// Edit manufacturer form
		admin.GET("/manufacturers/:id/edit", manufacturerController.Edit)

		// Update manufacturer
		admin.POST("/manufacturers/:id", manufacturerController.Update)

		// Delete manufacturer
		admin.DELETE("/manufacturers/:id", manufacturerController.Delete)

		// Alternative route for delete (for HTML forms without JavaScript)
		admin.POST("/manufacturers/:id/delete", manufacturerController.Delete)
	}
}
