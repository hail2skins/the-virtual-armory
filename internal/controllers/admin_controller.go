package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"
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

// Dashboard renders the admin dashboard
func (c *AdminController) Dashboard(ctx *gin.Context) {
	// Get current time
	now := time.Now()

	// Get pagination parameters
	page, err := strconv.Atoi(ctx.DefaultQuery("page", "1"))
	if err != nil || page < 1 {
		page = 1
	}

	perPage, err := strconv.Atoi(ctx.DefaultQuery("perPage", "10"))
	if err != nil {
		perPage = 10
	}
	// Limit perPage to valid options
	validPerPage := map[int]bool{10: true, 25: true, 50: true, 100: true}
	if !validPerPage[perPage] {
		perPage = 10
	}

	// Get sorting parameters
	sortBy := ctx.DefaultQuery("sortBy", "created_at")
	sortOrder := ctx.DefaultQuery("sortOrder", "desc")

	// Validate sortBy
	validSortFields := map[string]string{
		"email":             "email",
		"created_at":        "created_at",
		"last_login":        "last_attempt", // Map last_login to last_attempt
		"subscription_tier": "subscription_tier",
		"deleted":           "deleted_at", // Map deleted to deleted_at
	}

	dbSortField, ok := validSortFields[sortBy]
	if !ok {
		dbSortField = "created_at"
		sortBy = "created_at"
	}

	// Validate sortOrder
	if sortOrder != "asc" && sortOrder != "desc" {
		sortOrder = "desc"
	}

	// Calculate start of current month and previous month
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

	// Calculate growth rate for total users
	var userGrowthRate float64
	if lastMonthUsers > 0 {
		userGrowthRate = float64(thisMonthUsers-lastMonthUsers) / float64(lastMonthUsers) * 100
	} else if thisMonthUsers > 0 {
		userGrowthRate = 100 // If there were no users last month but there are this month, that's 100% growth
	} else if lastMonthUsers > 0 && thisMonthUsers == 0 {
		userGrowthRate = -100 // If there were users last month but none this month, that's -100% growth
	}

	// Count total subscribed users (non-free tier)
	var subscribedUsers int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier != 'free'").
		Count(&subscribedUsers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching subscribed users")
		return
	}

	// Count subscribed users created in different time periods
	var thisMonthSubscribed int64
	var lastMonthSubscribed int64

	// Count subscribed users as of now
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier != 'free' AND created_at < ?", now).
		Count(&thisMonthSubscribed).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching this month's subscribed users")
		return
	}

	// Count subscribed users as of the start of this month
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier != 'free' AND created_at < ?", startOfMonth).
		Count(&lastMonthSubscribed).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching last month's subscribed users")
		return
	}

	// Calculate growth rate for subscribed users
	var subscribedGrowthRate float64
	if lastMonthSubscribed > 0 {
		subscribedGrowthRate = float64(thisMonthSubscribed-lastMonthSubscribed) / float64(lastMonthSubscribed) * 100
	} else if thisMonthSubscribed > 0 {
		subscribedGrowthRate = 100 // If there were no subscribed users last month but there are this month, that's 100% growth
	} else if lastMonthSubscribed > 0 && thisMonthSubscribed == 0 {
		subscribedGrowthRate = -100 // If there were subscribed users last month but none this month, that's -100% growth
	}

	// Count new registrations (users created this month)
	var newRegistrations int64
	if err := database.GetDB().Model(&models.User{}).
		Where("created_at >= ? AND created_at < ?", startOfMonth, now).
		Count(&newRegistrations).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching new registrations")
		return
	}

	// Count new registrations last month
	var lastMonthRegistrations int64
	if err := database.GetDB().Model(&models.User{}).
		Where("created_at >= ? AND created_at < ?", startOfLastMonth, startOfMonth).
		Count(&lastMonthRegistrations).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching last month's registrations")
		return
	}

	// Calculate growth rate for new registrations
	var newRegistrationsGrowthRate float64
	if lastMonthRegistrations > 0 {
		newRegistrationsGrowthRate = float64(newRegistrations-lastMonthRegistrations) / float64(lastMonthRegistrations) * 100
	} else if newRegistrations > 0 {
		newRegistrationsGrowthRate = 100 // If there were no registrations last month but there are this month, that's 100% growth
	} else if lastMonthRegistrations > 0 && newRegistrations == 0 {
		newRegistrationsGrowthRate = -100 // If there were registrations last month but none this month, that's -100% growth
	}

	// Count new subscriptions (users who subscribed this month)
	var newSubscriptions int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier != 'free' AND created_at >= ? AND created_at < ?", startOfMonth, now).
		Count(&newSubscriptions).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching new subscriptions")
		return
	}

	// Count new subscriptions last month
	var lastMonthNewSubscriptions int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier != 'free' AND created_at >= ? AND created_at < ?", startOfLastMonth, startOfMonth).
		Count(&lastMonthNewSubscriptions).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching last month's new subscriptions")
		return
	}

	// Calculate growth rate for new subscriptions
	var newSubscriptionsGrowthRate float64
	if lastMonthNewSubscriptions > 0 {
		newSubscriptionsGrowthRate = float64(newSubscriptions-lastMonthNewSubscriptions) / float64(lastMonthNewSubscriptions) * 100
	} else if newSubscriptions > 0 {
		newSubscriptionsGrowthRate = 100 // If there were no new subscriptions last month but there are this month, that's 100% growth
	} else if lastMonthNewSubscriptions > 0 && newSubscriptions == 0 {
		newSubscriptionsGrowthRate = -100 // If there were new subscriptions last month but none this month, that's -100% growth
	}

	// Count monthly subscribers
	var monthlySubscribers int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier = 'monthly'").
		Count(&monthlySubscribers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching monthly subscribers")
		return
	}

	// Count monthly subscribers as of the start of this month
	var lastMonthMonthlySubscribers int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier = 'monthly' AND created_at < ?", startOfMonth).
		Count(&lastMonthMonthlySubscribers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching last month's monthly subscribers")
		return
	}

	// Calculate growth rate for monthly subscribers
	var monthlyGrowthRate float64
	if lastMonthMonthlySubscribers > 0 {
		monthlyGrowthRate = float64(monthlySubscribers-lastMonthMonthlySubscribers) / float64(lastMonthMonthlySubscribers) * 100
	} else if monthlySubscribers > 0 {
		monthlyGrowthRate = 100 // If there were no monthly subscribers last month but there are this month, that's 100% growth
	} else if lastMonthMonthlySubscribers > 0 && monthlySubscribers == 0 {
		monthlyGrowthRate = -100 // If there were monthly subscribers last month but none this month, that's -100% growth
	}

	// Count yearly subscribers
	var yearlySubscribers int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier = 'yearly'").
		Count(&yearlySubscribers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching yearly subscribers")
		return
	}

	// Count yearly subscribers as of the start of this month
	var lastMonthYearlySubscribers int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier = 'yearly' AND created_at < ?", startOfMonth).
		Count(&lastMonthYearlySubscribers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching last month's yearly subscribers")
		return
	}

	// Calculate growth rate for yearly subscribers
	var yearlyGrowthRate float64
	if lastMonthYearlySubscribers > 0 {
		yearlyGrowthRate = float64(yearlySubscribers-lastMonthYearlySubscribers) / float64(lastMonthYearlySubscribers) * 100
	} else if yearlySubscribers > 0 {
		yearlyGrowthRate = 100 // If there were no yearly subscribers last month but there are this month, that's 100% growth
	} else if lastMonthYearlySubscribers > 0 && yearlySubscribers == 0 {
		yearlyGrowthRate = -100 // If there were yearly subscribers last month but none this month, that's -100% growth
	}

	// Count lifetime subscribers
	var lifetimeSubscribers int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier = 'lifetime'").
		Count(&lifetimeSubscribers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching lifetime subscribers")
		return
	}

	// Count lifetime subscribers as of the start of this month
	var lastMonthLifetimeSubscribers int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier = 'lifetime' AND created_at < ?", startOfMonth).
		Count(&lastMonthLifetimeSubscribers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching last month's lifetime subscribers")
		return
	}

	// Calculate growth rate for lifetime subscribers
	var lifetimeGrowthRate float64
	if lastMonthLifetimeSubscribers > 0 {
		lifetimeGrowthRate = float64(lifetimeSubscribers-lastMonthLifetimeSubscribers) / float64(lastMonthLifetimeSubscribers) * 100
	} else if lifetimeSubscribers > 0 {
		lifetimeGrowthRate = 100 // If there were no lifetime subscribers last month but there are this month, that's 100% growth
	} else if lastMonthLifetimeSubscribers > 0 && lifetimeSubscribers == 0 {
		lifetimeGrowthRate = -100 // If there were lifetime subscribers last month but none this month, that's -100% growth
	}

	// Count premium subscribers
	var premiumSubscribers int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier = 'premium'").
		Count(&premiumSubscribers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching premium subscribers")
		return
	}

	// Count premium subscribers as of the start of this month
	var lastMonthPremiumSubscribers int64
	if err := database.GetDB().Model(&models.User{}).
		Where("subscription_tier = 'premium' AND created_at < ?", startOfMonth).
		Count(&lastMonthPremiumSubscribers).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching last month's premium subscribers")
		return
	}

	// Calculate growth rate for premium subscribers
	var premiumGrowthRate float64
	if lastMonthPremiumSubscribers > 0 {
		premiumGrowthRate = float64(premiumSubscribers-lastMonthPremiumSubscribers) / float64(lastMonthPremiumSubscribers) * 100
	} else if premiumSubscribers > 0 {
		premiumGrowthRate = 100 // If there were no premium subscribers last month but there are this month, that's 100% growth
	} else if lastMonthPremiumSubscribers > 0 && premiumSubscribers == 0 {
		premiumGrowthRate = -100 // If there were premium subscribers last month but none this month, that's -100% growth
	}

	// Get recent users with pagination and sorting
	var users []models.User
	offset := (page - 1) * perPage

	// Create the base query
	query := database.GetDB().Model(&models.User{}).Unscoped() // Unscoped to include soft-deleted users

	// Apply sorting
	query = query.Order(fmt.Sprintf("%s %s", dbSortField, sortOrder))

	// Get total count for pagination
	var totalCount int64
	if err := query.Count(&totalCount).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error counting users")
		return
	}

	// Apply pagination
	if err := query.Limit(perPage).Offset(offset).Find(&users).Error; err != nil {
		ctx.String(http.StatusInternalServerError, "Error fetching users")
		return
	}

	// Convert users to UserData for the template
	recentUsers := make([]admin.UserData, len(users))
	for i, user := range users {
		recentUsers[i] = admin.UserData{
			ID:               user.ID,
			Email:            user.Email,
			CreatedAt:        user.CreatedAt,
			LastLogin:        user.LastAttempt,
			SubscriptionTier: user.SubscriptionTier,
			IsDeleted:        !user.DeletedAt.Time.IsZero(),
		}
	}

	// Calculate total pages
	totalPages := int(math.Ceil(float64(totalCount) / float64(perPage)))
	if totalPages < 1 {
		totalPages = 1
	}

	data := admin.DashboardData{
		TotalUsers:                 totalUsers,
		UserGrowthRate:             userGrowthRate,
		SubscribedUsers:            subscribedUsers,
		SubscribedGrowthRate:       subscribedGrowthRate,
		NewRegistrations:           newRegistrations,
		NewRegistrationsGrowthRate: newRegistrationsGrowthRate,
		NewSubscriptions:           newSubscriptions,
		NewSubscriptionsGrowthRate: newSubscriptionsGrowthRate,
		MonthlySubscribers:         monthlySubscribers,
		MonthlyGrowthRate:          monthlyGrowthRate,
		YearlySubscribers:          yearlySubscribers,
		YearlyGrowthRate:           yearlyGrowthRate,
		LifetimeSubscribers:        lifetimeSubscribers,
		LifetimeGrowthRate:         lifetimeGrowthRate,
		PremiumSubscribers:         premiumSubscribers,
		PremiumGrowthRate:          premiumGrowthRate,
		RecentUsers:                recentUsers,
		CurrentPage:                page,
		TotalPages:                 totalPages,
		PerPage:                    perPage,
		SortBy:                     sortBy,
		SortOrder:                  sortOrder,
	}

	component := admin.Dashboard(data)
	if err := component.Render(ctx.Request.Context(), ctx.Writer); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}
