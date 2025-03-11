package controllers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/admin"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/middleware"
	"github.com/hail2skins/the-virtual-armory/internal/models"
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
	// Get current time and normalize it to the start of the current month
	now := time.Now()
	startOfMonth := time.Date(now.Year(), now.Month(), 1, 0, 0, 0, 0, now.Location())
	startOfLastMonth := startOfMonth.AddDate(0, -1, 0)

	// Count total users
	var totalUsers int64
	if err := database.GetDB().Model(&models.User{}).Count(&totalUsers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching total users")
		return
	}

	// Count users created in different time periods
	var thisMonthUsers int64
	var lastMonthUsers int64

	// Count users created this month (from start of this month to now)
	if err := database.GetDB().Model(&models.User{}).
		Where("created_at >= ?", startOfMonth).
		Count(&thisMonthUsers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching this month's users")
		return
	}

	// Count users created last month (from start of last month to start of this month)
	if err := database.GetDB().Model(&models.User{}).
		Where("created_at >= ? AND created_at < ?", startOfLastMonth, startOfMonth).
		Count(&lastMonthUsers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching last month's users")
		return
	}

	// Calculate growth rate
	var growthRate float64
	if lastMonthUsers > 0 {
		growthRate = float64(thisMonthUsers-lastMonthUsers) / float64(lastMonthUsers) * 100
	} else if thisMonthUsers > 0 {
		growthRate = 100 // If there were no users last month but there are this month, that's 100% growth
	} else if lastMonthUsers > 0 && thisMonthUsers == 0 {
		growthRate = -100 // If there were users last month but none this month, that's -100% growth
	}

	data := admin.DashboardData{
		TotalUsers:     totalUsers,
		UserGrowthRate: growthRate,
	}

	component := admin.Dashboard(data)
	if err := component.Render(ctx.Request.Context(), ctx.Writer); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}
