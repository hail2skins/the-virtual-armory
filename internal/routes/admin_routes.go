package routes

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"gorm.io/gorm"
)

// SetupAdminRoutes sets up the admin routes
func SetupAdminRoutes(router *gin.Engine, db *gorm.DB) {
	// Create controllers
	manufacturerController := controllers.NewManufacturerController()
	caliberController := controllers.NewCaliberController()
	weaponTypeController := controllers.NewWeaponTypeController(db)

	// Admin routes
	adminRoutes := router.Group("/admin")
	// Add authentication middleware here if needed
	// adminRoutes.Use(middleware.RequireAuth())

	// Manufacturer routes
	adminRoutes.GET("/manufacturers", manufacturerController.Index)
	adminRoutes.GET("/manufacturers/new", manufacturerController.New)
	adminRoutes.POST("/manufacturers", manufacturerController.Create)
	adminRoutes.GET("/manufacturers/:id", manufacturerController.Show)
	adminRoutes.GET("/manufacturers/:id/edit", manufacturerController.Edit)
	adminRoutes.POST("/manufacturers/:id", manufacturerController.Update)
	adminRoutes.POST("/manufacturers/:id/delete", manufacturerController.Delete)

	// Caliber routes
	adminRoutes.GET("/calibers", caliberController.Index)
	adminRoutes.GET("/calibers/new", caliberController.New)
	adminRoutes.POST("/calibers", caliberController.Create)
	adminRoutes.GET("/calibers/:id", caliberController.Show)
	adminRoutes.GET("/calibers/:id/edit", caliberController.Edit)
	adminRoutes.POST("/calibers/:id", caliberController.Update)
	adminRoutes.POST("/calibers/:id/delete", caliberController.Delete)

	// Weapon Type routes
	adminRoutes.GET("/weapon-types", weaponTypeController.Index)
	adminRoutes.GET("/weapon-types/new", weaponTypeController.New)
	adminRoutes.POST("/weapon-types", weaponTypeController.Create)
	adminRoutes.GET("/weapon-types/:id", weaponTypeController.Show)
	adminRoutes.GET("/weapon-types/:id/edit", weaponTypeController.Edit)
	adminRoutes.POST("/weapon-types/:id", weaponTypeController.Update)
	adminRoutes.POST("/weapon-types/:id/delete", weaponTypeController.Delete)
}

// RegisterAdminRoutes registers all admin routes
func RegisterAdminRoutes(router *gin.Engine, adminController *controllers.AdminController, authInstance *auth.Auth) {
	// Create an admin group with authentication and admin middleware
	adminGroup := router.Group("/admin")
	adminGroup.Use(authInstance.RequireAuth())
	adminGroup.Use(authInstance.RequireAdmin())

	// Register admin routes
	adminGroup.GET("/error-metrics", adminController.ErrorMetrics)
}

// RegisterAdminHealthRoutes registers admin health-related routes
func RegisterAdminHealthRoutes(router *gin.Engine, adminHealthController *controllers.AdminHealthController, authInstance *auth.Auth) {
	// Create an admin group with authentication and admin middleware
	adminGroup := router.Group("/admin")
	adminGroup.Use(authInstance.RequireAuth())
	adminGroup.Use(authInstance.RequireAdmin())

	// Register admin health routes
	adminGroup.GET("/detailed-health", adminHealthController.DetailedHealth)
}
