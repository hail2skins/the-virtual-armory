package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
)

// RegisterCaliberRoutes registers all routes for calibers
func RegisterCaliberRoutes(router *gin.Engine, auth *auth.Auth) {
	// Create a new caliber controller
	caliberController := controllers.NewCaliberController()

	// Admin-only routes for calibers
	admin := router.Group("/admin")
	admin.Use(auth.RequireAdmin())
	{
		// Calibers index
		admin.GET("/calibers", caliberController.Index)

		// New caliber form
		admin.GET("/calibers/new", caliberController.New)

		// Create caliber
		admin.POST("/calibers", caliberController.Create)

		// Show caliber details
		admin.GET("/calibers/:id", caliberController.Show)

		// Edit caliber form
		admin.GET("/calibers/:id/edit", caliberController.Edit)

		// Update caliber
		admin.POST("/calibers/:id", caliberController.Update)

		// Delete caliber
		admin.DELETE("/calibers/:id", caliberController.Delete)

		// Alternative route for delete (for HTML forms without JavaScript)
		admin.POST("/calibers/:id/delete", caliberController.Delete)
	}
}
