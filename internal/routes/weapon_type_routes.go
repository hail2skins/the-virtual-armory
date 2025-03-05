package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/database"
)

// RegisterWeaponTypeRoutes registers all routes related to weapon types
func RegisterWeaponTypeRoutes(r *gin.Engine, authInstance *auth.Auth) {
	// Get database connection
	db := database.GetDB()

	// Create weapon type controller
	weaponTypeController := controllers.NewWeaponTypeController(db)

	// Create a group for admin routes that require authentication
	adminRoutes := r.Group("/admin")
	adminRoutes.Use(authInstance.RequireAdmin())

	// Weapon Type routes
	adminRoutes.GET("/weapon-types", weaponTypeController.Index)
	adminRoutes.GET("/weapon-types/new", weaponTypeController.New)
	adminRoutes.POST("/weapon-types", weaponTypeController.Create)
	adminRoutes.GET("/weapon-types/:id", weaponTypeController.Show)
	adminRoutes.GET("/weapon-types/:id/edit", weaponTypeController.Edit)
	adminRoutes.POST("/weapon-types/:id", weaponTypeController.Update)
	adminRoutes.POST("/weapon-types/:id/delete", weaponTypeController.Delete)
}
