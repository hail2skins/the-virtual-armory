package controllers

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/admin"
	"gorm.io/gorm"
)

// SystemMetrics contains system resource metrics
type SystemMetrics struct {
	MemoryUsageMB    float64
	TotalAllocatedMB float64
	SystemMemoryMB   float64
	Goroutines       int
	NumCPU           int
}

// ExternalService represents an external service and its status
type ExternalService struct {
	Name   string
	Status string
}

// AdminHealthController handles admin health-related routes
type AdminHealthController struct {
	DB *gorm.DB
}

// NewAdminHealthController creates a new admin health controller
func NewAdminHealthController(db *gorm.DB) *AdminHealthController {
	return &AdminHealthController{
		DB: db,
	}
}

// DetailedHealth returns detailed health information for the admin dashboard
func (c *AdminHealthController) DetailedHealth(ctx *gin.Context) {
	// Check database connectivity
	dbStatus := "connected"
	sqlDB, err := c.DB.DB()
	if err != nil || sqlDB.Ping() != nil {
		dbStatus = "disconnected"
	}

	// Get system metrics
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	systemMetrics := admin.SystemMetrics{
		MemoryUsageMB:    float64(m.Alloc) / 1024 / 1024,
		TotalAllocatedMB: float64(m.TotalAlloc) / 1024 / 1024,
		SystemMemoryMB:   float64(m.Sys) / 1024 / 1024,
		Goroutines:       runtime.NumGoroutine(),
		NumCPU:           runtime.NumCPU(),
	}

	// Check external services (simplified for now)
	externalServices := []admin.ExternalService{
		{Name: "Stripe", Status: "connected"}, // In a real implementation, we would check Stripe connectivity
		{Name: "Email", Status: "connected"},  // In a real implementation, we would check email service
	}

	// Create the template data
	data := admin.DetailedHealthData{
		Status:           "ok",
		Timestamp:        time.Now(),
		Database:         dbStatus,
		System:           systemMetrics,
		ExternalServices: externalServices,
		Version:          "1.0.0", // Add application version
	}

	// Check if the request accepts JSON
	if ctx.GetHeader("Accept") == "application/json" {
		// Return the metrics as JSON
		ctx.JSON(http.StatusOK, gin.H{
			"status":            data.Status,
			"timestamp":         data.Timestamp.Format(time.RFC3339),
			"database":          data.Database,
			"system":            data.System,
			"external_services": data.ExternalServices,
			"version":           data.Version,
		})
		return
	}

	// Render the template
	admin.DetailedHealth(data).Render(ctx.Request.Context(), ctx.Writer)
}
