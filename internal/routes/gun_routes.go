package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"gorm.io/gorm"
)

// RegisterGunRoutes registers all gun-related routes
func RegisterGunRoutes(router *gin.Engine, db *gorm.DB, auth *auth.Auth) {
	// Create the gun controller
	gunController := controllers.NewGunController(db)

	// Owner routes (require authentication)
	ownerGroup := router.Group("/owner")
	ownerGroup.Use(auth.RequireAuth())
	{
		// Gun routes nested under owner
		gunGroup := ownerGroup.Group("/guns")
		{
			// List all guns for the owner
			gunGroup.GET("", gunController.Index)

			// Create a new gun
			gunGroup.GET("/new", gunController.New)
			gunGroup.POST("", gunController.Create)

			// Show a specific gun
			gunGroup.GET("/:id", gunController.Show)

			// Edit a gun
			gunGroup.GET("/:id/edit", gunController.Edit)
			gunGroup.POST("/:id", gunController.Update)

			// Delete a gun
			gunGroup.POST("/:id/delete", gunController.Delete)
		}
	}
}
