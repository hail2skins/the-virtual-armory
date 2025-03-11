package routes

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/gorm"
)

// setupAdminDashboardTestRouter creates a test router with the admin dashboard route registered
func setupAdminDashboardTestRouter(t *testing.T) (*gin.Engine, *gorm.DB, *testutils.TestUsers) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	r := gin.New()
	r.Use(gin.Recovery())

	// Setup test database
	db, err := testutils.SetupTestDB()
	require.NoError(t, err, "Failed to setup test database")

	// Create test users
	testUsers := testutils.CreateTestUsers()

	// Create real auth instance
	authInstance, err := auth.New()
	require.NoError(t, err, "Failed to create auth instance")

	// Create admin controller
	adminController := controllers.NewAdminController()

	// Register the admin dashboard route with real middleware
	adminGroup := r.Group("/admin")
	adminGroup.Use(authInstance.RequireAuth())
	adminGroup.Use(authInstance.RequireAdmin())

	// Register admin dashboard route
	adminGroup.GET("/dashboard", adminController.Dashboard)

	// Save users to database
	err = db.Create(&testUsers.Admin).Error
	require.NoError(t, err, "Failed to create admin user")
	err = db.Create(&testUsers.Unsubscribed).Error
	require.NoError(t, err, "Failed to create regular user")

	return r, db, testUsers
}

// TestAdminDashboardEndpointWithGuestUser tests that a guest user is redirected to login with a flash message
func TestAdminDashboardEndpointWithGuestUser(t *testing.T) {
	r, db, _ := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create a test HTTP request without authentication
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err, "Failed to create request")

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Check that the guest user is redirected to login
	assert.Equal(t, http.StatusFound, rr.Code, "Guest user should be redirected")
	assert.Equal(t, "/login", rr.Header().Get("Location"), "Guest user should be redirected to login")

	// Check for flash message cookies
	cookies := rr.Result().Cookies()
	var flashMessage, flashType string
	for _, cookie := range cookies {
		if cookie.Name == "flash_message" {
			flashMessage = cookie.Value
			flashMessage = strings.ReplaceAll(flashMessage, "+", " ")
			t.Logf("Flash message: %q", flashMessage)
		} else if cookie.Name == "flash_type" {
			flashType = cookie.Value
			t.Logf("Flash type: %q", flashType)
		}
	}

	// Assert the exact flash message and type
	assert.NotEmpty(t, flashMessage, "Flash message should be set")
	assert.Equal(t, "You do not have permission to access that page", flashMessage, "Flash message should match exactly")
	assert.Equal(t, "error", flashType, "Flash type should be error")
}

// TestAdminDashboardEndpointWithRegularUser tests that a regular user is redirected to owner page with an admin required message
func TestAdminDashboardEndpointWithRegularUser(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err, "Failed to create request")

	// Set up session for the regular user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Unsubscribed.Email,
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Check that the regular user is redirected to owner page
	assert.Equal(t, http.StatusFound, rr.Code, "Regular user should be redirected")
	assert.Equal(t, "/owner", rr.Header().Get("Location"), "Regular user should be redirected to owner page")

	// Check for flash message cookies
	cookies := rr.Result().Cookies()
	var flashMessage, flashType string
	for _, cookie := range cookies {
		if cookie.Name == "flash_message" {
			flashMessage = cookie.Value
			flashMessage = strings.ReplaceAll(flashMessage, "+", " ")
			t.Logf("Flash message: %q", flashMessage)
		} else if cookie.Name == "flash_type" {
			flashType = cookie.Value
			t.Logf("Flash type: %q", flashType)
		}
	}

	// Assert the exact flash message and type
	assert.NotEmpty(t, flashMessage, "Flash message should be set")
	assert.Equal(t, "You must be an administrator to access this page", flashMessage, "Flash message should match exactly")
	assert.Equal(t, "error", flashType, "Flash type should be error")
}

// TestAdminDashboardEndpointWithAdminUser tests that an admin user can access the endpoint
func TestAdminDashboardEndpointWithAdminUser(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create a test HTTP request
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err, "Failed to create request")

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Check that the admin user can access the endpoint
	assert.Equal(t, http.StatusOK, rr.Code, "Admin user should be able to access the endpoint")
	assert.Contains(t, rr.Body.String(), "Admin Dashboard", "Response should contain dashboard title")
	assert.Contains(t, rr.Body.String(), "Total Users", "Response should contain total users section")
	assert.Contains(t, rr.Body.String(), "Recent Users", "Response should contain recent users section")
}

// TestAdminDashboardTotalUsersMetrics tests that the total users count is accurate
func TestAdminDashboardTotalUsersMetrics(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create additional test users
	additionalUsers := []models.User{
		{Email: "test1@test.com", Password: "password"},
		{Email: "test2@test.com", Password: "password"},
		{Email: "test3@test.com", Password: "password"},
	}

	// Create the users in the database
	for _, user := range additionalUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Count total users in database (3 additional + admin + regular from setup = 5)
	var totalUsers int64
	err := db.Model(&models.User{}).Count(&totalUsers).Error
	require.NoError(t, err)

	// Create request as admin
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain the actual total number of users
	body := rr.Body.String()
	assert.Contains(t, body, fmt.Sprintf("%d", totalUsers)) // Should show actual total users
	assert.NotContains(t, body, "1,234")                    // Should not contain the mock data
}

// TestAdminDashboardUserGrowthPercentage tests that the dashboard displays a growth percentage
func TestAdminDashboardUserGrowthPercentage(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create some users to ensure there's data
	additionalUsers := []models.User{
		{Email: "test1@test.com", Password: "password"},
		{Email: "test2@test.com", Password: "password"},
	}

	for _, user := range additionalUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Create request as admin
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain a percentage sign for growth rate
	body := rr.Body.String()
	assert.Regexp(t, `[+-]?\d+%`, body, "Should show a growth percentage")
	assert.NotContains(t, body, "+12%", "Should not contain mock growth data")
}

// TestAdminDashboardSubscribedUsersMetrics tests that the subscribed users count is displayed
func TestAdminDashboardSubscribedUsersMetrics(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create additional test users with different subscription tiers
	additionalUsers := []models.User{
		{Email: "monthly@test.com", Password: "password", SubscriptionTier: "monthly"},
		{Email: "yearly@test.com", Password: "password", SubscriptionTier: "yearly"},
		{Email: "lifetime@test.com", Password: "password", SubscriptionTier: "lifetime"},
		{Email: "premium@test.com", Password: "password", SubscriptionTier: "premium"},
		{Email: "free@test.com", Password: "password", SubscriptionTier: "free"},
	}

	// Create the users in the database
	for _, user := range additionalUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Create request as admin
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain "Subscribed Users" and a percentage sign for growth rate
	body := rr.Body.String()
	assert.Contains(t, body, "Subscribed Users", "Should show Subscribed Users section")
	assert.NotContains(t, body, "Active Users", "Should not contain the old Active Users label")
	assert.Regexp(t, `[+-]?\d+%`, body, "Should show a growth percentage for subscribed users")
	assert.NotContains(t, body, "987", "Should not contain the mock data for active users")
}

// TestAdminDashboardNewRegistrationsMetrics tests that the new registrations count is displayed
func TestAdminDashboardNewRegistrationsMetrics(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create additional test users
	additionalUsers := []models.User{
		{Email: "newuser1@test.com", Password: "password"},
		{Email: "newuser2@test.com", Password: "password"},
	}

	// Create the users in the database
	for _, user := range additionalUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Create request as admin
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain "New Registrations" and a percentage sign for growth rate
	body := rr.Body.String()
	assert.Contains(t, body, "New Registrations", "Should show New Registrations section")
	assert.Regexp(t, `[+-]?\d+%`, body, "Should show a growth percentage for new registrations")
	assert.NotContains(t, body, "56", "Should not contain the mock data for new registrations")
}

// TestAdminDashboardNewSubscriptionsMetrics tests that the new subscriptions count is displayed
func TestAdminDashboardNewSubscriptionsMetrics(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create additional test users with subscriptions
	additionalUsers := []models.User{
		{Email: "newsub1@test.com", Password: "password", SubscriptionTier: "monthly"},
		{Email: "newsub2@test.com", Password: "password", SubscriptionTier: "yearly"},
	}

	// Create the users in the database
	for _, user := range additionalUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Create request as admin
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain "New Subscriptions" and a percentage sign for growth rate
	body := rr.Body.String()
	assert.Contains(t, body, "New Subscriptions", "Should show New Subscriptions section")
	assert.Regexp(t, `[+-]?\d+%`, body, "Should show a growth percentage for new subscriptions")
	assert.NotContains(t, body, "+15%", "Should not contain the mock growth data")
}

// TestAdminDashboardMonthlySubscribersMetrics tests that the monthly subscribers count is displayed
func TestAdminDashboardMonthlySubscribersMetrics(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create additional test users with monthly subscription tier
	additionalUsers := []models.User{
		{Email: "monthly1@test.com", Password: "password", SubscriptionTier: "monthly"},
		{Email: "monthly2@test.com", Password: "password", SubscriptionTier: "monthly"},
	}

	// Create the users in the database
	for _, user := range additionalUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Create request as admin
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain "Monthly Subscribers" and a percentage sign for growth rate
	body := rr.Body.String()
	assert.Contains(t, body, "Monthly Subscribers", "Should show Monthly Subscribers section")
	assert.Regexp(t, `[+-]?\d+%`, body, "Should show a growth percentage for monthly subscribers")
}

// TestAdminDashboardYearlySubscribersMetrics tests that the yearly subscribers count is displayed
func TestAdminDashboardYearlySubscribersMetrics(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create additional test users with yearly subscription tier
	additionalUsers := []models.User{
		{Email: "yearly1@test.com", Password: "password", SubscriptionTier: "yearly"},
		{Email: "yearly2@test.com", Password: "password", SubscriptionTier: "yearly"},
	}

	// Create the users in the database
	for _, user := range additionalUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Create request as admin
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain "Yearly Subscribers" and a percentage sign for growth rate
	body := rr.Body.String()
	assert.Contains(t, body, "Yearly Subscribers", "Should show Yearly Subscribers section")
	assert.Regexp(t, `[+-]?\d+%`, body, "Should show a growth percentage for yearly subscribers")
}

// TestAdminDashboardLifetimeSubscribersMetrics tests that the lifetime subscribers count is displayed
func TestAdminDashboardLifetimeSubscribersMetrics(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create additional test users with lifetime subscription tier
	additionalUsers := []models.User{
		{Email: "lifetime1@test.com", Password: "password", SubscriptionTier: "lifetime"},
		{Email: "lifetime2@test.com", Password: "password", SubscriptionTier: "lifetime"},
	}

	// Create the users in the database
	for _, user := range additionalUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Create request as admin
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain "Lifetime Subscribers" and a percentage sign for growth rate
	body := rr.Body.String()
	assert.Contains(t, body, "Lifetime Subscribers", "Should show Lifetime Subscribers section")
	assert.Regexp(t, `[+-]?\d+%`, body, "Should show a growth percentage for lifetime subscribers")
}

// TestAdminDashboardPremiumSubscribersMetrics tests that the premium subscribers count is displayed
func TestAdminDashboardPremiumSubscribersMetrics(t *testing.T) {
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create additional test users with premium subscription tier
	additionalUsers := []models.User{
		{Email: "premium1@test.com", Password: "password", SubscriptionTier: "premium"},
		{Email: "premium2@test.com", Password: "password", SubscriptionTier: "premium"},
	}

	// Create the users in the database
	for _, user := range additionalUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)
	}

	// Create request as admin
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain "Premium Subscribers" and a percentage sign for growth rate
	body := rr.Body.String()
	assert.Contains(t, body, "Premium Subscribers", "Should show Premium Subscribers section")
	assert.Regexp(t, `[+-]?\d+%`, body, "Should show a growth percentage for premium subscribers")
}

// TestAdminDashboardRecentUsersTable tests that the admin dashboard shows the recent users table
func TestAdminDashboardRecentUsersTable(t *testing.T) {
	// Setup test router
	r, db, testUsers := setupAdminDashboardTestRouter(t)
	defer db.Migrator().DropTable(&models.User{})

	// Create additional test users with different subscription tiers
	// Create 15 users to test pagination (default is 10 per page)
	additionalUsers := []models.User{
		{Email: "user1@test.com", Password: "password", SubscriptionTier: "free"},
		{Email: "user2@test.com", Password: "password", SubscriptionTier: "monthly"},
		{Email: "user3@test.com", Password: "password", SubscriptionTier: "yearly"},
		{Email: "user4@test.com", Password: "password", SubscriptionTier: "lifetime"},
		{Email: "user5@test.com", Password: "password", SubscriptionTier: "premium"},
		{Email: "user6@test.com", Password: "password", SubscriptionTier: "free"},
		{Email: "user7@test.com", Password: "password", SubscriptionTier: "monthly"},
		{Email: "user8@test.com", Password: "password", SubscriptionTier: "yearly"},
		{Email: "user9@test.com", Password: "password", SubscriptionTier: "lifetime"},
		{Email: "user10@test.com", Password: "password", SubscriptionTier: "premium"},
		{Email: "user11@test.com", Password: "password", SubscriptionTier: "free"},
		{Email: "user12@test.com", Password: "password", SubscriptionTier: "monthly"},
		{Email: "user13@test.com", Password: "password", SubscriptionTier: "yearly"},
		{Email: "user14@test.com", Password: "password", SubscriptionTier: "lifetime"},
		{Email: "user15@test.com", Password: "password", SubscriptionTier: "premium"},
	}

	// Create a soft-deleted user
	deletedUser := models.User{
		Email:            "deleted@test.com",
		Password:         "password",
		SubscriptionTier: "free",
	}
	err := db.Create(&deletedUser).Error
	require.NoError(t, err)

	// Soft delete the user
	err = db.Delete(&deletedUser).Error
	require.NoError(t, err)

	// Create the users in the database with staggered creation times
	for i, user := range additionalUsers {
		err := db.Create(&user).Error
		require.NoError(t, err)

		// Update the CreatedAt field to simulate users created at different times
		createdAt := time.Now().Add(time.Duration(-(i + 1)) * time.Hour)
		err = db.Model(&user).Update("created_at", createdAt).Error
		require.NoError(t, err)
	}

	// Create request as admin
	req, err := http.NewRequest("GET", "/admin/dashboard", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr := httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain the Recent Users table
	body := rr.Body.String()

	// Check for table headers
	assert.Contains(t, body, "Email", "Should show Email column")
	assert.Contains(t, body, "Registered", "Should show Registered column")
	assert.Contains(t, body, "Last Login", "Should show Last Login column")
	assert.Contains(t, body, "Subscribed", "Should show Subscribed column")
	assert.Contains(t, body, "Deleted", "Should show Deleted column")

	// Check for pagination controls
	assert.Contains(t, body, "Show:", "Should show pagination controls")
	assert.Contains(t, body, "value=\"10\"", "Should have option for 10 users per page")
	assert.Contains(t, body, "value=\"25\"", "Should have option for 25 users per page")
	assert.Contains(t, body, "value=\"50\"", "Should have option for 50 users per page")
	assert.Contains(t, body, "value=\"100\"", "Should have option for 100 users per page")

	// Check for View All Users link
	assert.Contains(t, body, "View All Users", "Should show View All Users link")
	assert.Contains(t, body, "href=\"/admin/users\"", "Should link to /admin/users")

	// Check for user data
	assert.Contains(t, body, "user1@test.com", "Should show user1@test.com")
	assert.Contains(t, body, "Free", "Should show Free subscription tier")
	assert.Contains(t, body, "Monthly", "Should show Monthly subscription tier")
	assert.Contains(t, body, "Yearly", "Should show Yearly subscription tier")
	assert.Contains(t, body, "Lifetime", "Should show Lifetime subscription tier")
	assert.Contains(t, body, "Premium", "Should show Premium subscription tier")

	// Test pagination by requesting page 2
	req, err = http.NewRequest("GET", "/admin/dashboard?page=2&perPage=10", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr = httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain users from page 2
	body = rr.Body.String()
	assert.Contains(t, body, "user11@test.com", "Should show user11@test.com on page 2")

	// Test sorting by subscription tier
	req, err = http.NewRequest("GET", "/admin/dashboard?sortBy=subscription_tier&sortOrder=asc", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr = httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// Test for deleted users
	req, err = http.NewRequest("GET", "/admin/dashboard?sortBy=deleted&sortOrder=desc", nil)
	require.NoError(t, err)

	// Set up session for the admin user
	req.AddCookie(&http.Cookie{
		Name:  "is_logged_in",
		Value: "true",
	})
	req.AddCookie(&http.Cookie{
		Name:  "user_email",
		Value: testUsers.Admin.Email,
	})
	req.AddCookie(&http.Cookie{
		Name:  "is_admin",
		Value: "true",
	})

	// Create a ResponseRecorder to record the response
	rr = httptest.NewRecorder()

	// Serve the HTTP request
	r.ServeHTTP(rr, req)

	// Verify response
	assert.Equal(t, http.StatusOK, rr.Code)

	// The response should contain the deleted user
	body = rr.Body.String()
	assert.Contains(t, body, "deleted@test.com", "Should show deleted@test.com")
	assert.Contains(t, body, "Yes", "Should show Yes for deleted status")
}
