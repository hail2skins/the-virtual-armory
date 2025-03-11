package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/admin"
	"github.com/hail2skins/the-virtual-armory/internal/middleware"
)

// AdminController handles admin-related routes
type AdminController struct{}

// NewAdminController creates a new admin controller
func NewAdminController() *AdminController {
	return &AdminController{}
}

// ErrorMetrics returns error metrics for the admin dashboard
func (c *AdminController) ErrorMetrics(ctx *gin.Context) {
	// Get the error metrics
	metrics := middleware.GetErrorMetrics()

	// Get the time range from the query parameters
	timeRange := ctx.DefaultQuery("range", "24h")

	// Parse the time range
	var duration time.Duration
	switch timeRange {
	case "1h":
		duration = time.Hour
	case "6h":
		duration = 6 * time.Hour
	case "24h":
		duration = 24 * time.Hour
	case "7d":
		duration = 7 * 24 * time.Hour
	case "30d":
		duration = 30 * 24 * time.Hour
	default:
		duration = 24 * time.Hour
	}

	// Get the error rates for the specified duration
	errorRates := metrics.GetErrorRates(duration)

	// Get the latency percentiles
	latencyPercentiles := metrics.GetLatencyPercentiles()

	// Get recent errors
	recentErrors := metrics.GetRecentErrors(10)

	// Convert recent errors to the template format
	templateRecentErrors := make([]admin.RecentError, len(recentErrors))
	for i, err := range recentErrors {
		templateRecentErrors[i] = admin.RecentError{
			ErrorType:    err.ErrorType,
			Count:        err.Count,
			LastOccurred: err.LastOccurred,
			Path:         err.Path,
		}
	}

	// Get overall stats
	stats := metrics.GetStats()

	// Create the template data
	data := admin.ErrorMetricsData{
		ErrorRates:         errorRates,
		LatencyPercentiles: latencyPercentiles,
		RecentErrors:       templateRecentErrors,
		Stats:              stats,
		TimeRange:          timeRange,
	}

	// Check if the request accepts JSON
	if ctx.GetHeader("Accept") == "application/json" {
		// Return the metrics as JSON
		ctx.JSON(http.StatusOK, gin.H{
			"error_rates":         errorRates,
			"latency_percentiles": latencyPercentiles,
			"recent_errors":       recentErrors,
			"stats":               stats,
			"time_range":          timeRange,
		})
		return
	}

	// Render the template
	admin.ErrorMetrics(data).Render(ctx.Request.Context(), ctx.Writer)
}

// Dashboard renders the admin dashboard page
func (c *AdminController) Dashboard(ctx *gin.Context) {
	component := admin.Dashboard()
	component.Render(ctx.Request.Context(), ctx.Writer)
}
