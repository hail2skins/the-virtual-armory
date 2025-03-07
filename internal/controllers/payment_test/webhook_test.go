package payment_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/controllers/payment_test/payment_test_utils"
	"github.com/hail2skins/the-virtual-armory/internal/middleware"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

// TestWebhookMonitoring tests the webhook monitoring middleware
func TestWebhookMonitoring(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create a router with the webhook monitoring middleware
	router := gin.Default()
	router.POST("/webhook", middleware.WebhookMonitor(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "success"})
	})

	// Create a request
	req, _ := http.NewRequest("POST", "/webhook", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Get the webhook stats
	stats := middleware.GetWebhookStats()

	// Verify that the stats were updated
	assert.Equal(t, int64(1), stats.TotalRequests)
	assert.Equal(t, int64(1), stats.SuccessfulRequests)
	assert.Equal(t, int64(0), stats.FailedRequests)
	assert.False(t, stats.LastRequestTime.IsZero())
}

// TestWebhookHealthCheck tests the webhook health check endpoint
func TestWebhookHealthCheck(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create a router with the webhook health check endpoint
	router := gin.Default()
	router.GET("/webhook-health", middleware.WebhookHealthCheck())

	// Create a request
	req, _ := http.NewRequest("GET", "/webhook-health", nil)

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Verify the response contains the expected fields
	assert.Contains(t, response, "status")
	assert.Contains(t, response, "total_requests")
	assert.Contains(t, response, "successful")
	assert.Contains(t, response, "failed")
	assert.Contains(t, response, "success_rate")
}

// TestWebhookFailureTracking tests that the middleware tracks failed requests
func TestWebhookFailureTracking(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Create a router with the webhook monitoring middleware
	router := gin.Default()
	router.POST("/webhook-fail", middleware.WebhookMonitor(), func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "test error"})
	})

	// Create a request
	req, _ := http.NewRequest("POST", "/webhook-fail", bytes.NewBufferString("{}"))
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Get the webhook stats
	stats := middleware.GetWebhookStats()

	// Verify that the stats were updated to include the failure
	assert.True(t, stats.FailedRequests > 0)
	assert.False(t, stats.LastErrorTime.IsZero())
	assert.NotEmpty(t, stats.LastError)
}

// TestWebhookHandlerInTestMode tests that the webhook handler works in test mode
func TestWebhookHandlerInTestMode(t *testing.T) {
	// Set up test environment
	gin.SetMode(gin.TestMode)

	// Set test environment
	t.Setenv("APP_ENV", "test")

	// Create a test database
	db := payment_test_utils.SetupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Create a payment controller
	paymentController := controllers.NewPaymentController(db)

	// Create a router with the webhook handler
	router := gin.Default()
	router.POST("/webhook", middleware.WebhookMonitor(), paymentController.HandleStripeWebhook)

	// Create a mock checkout.session.completed event
	eventData := map[string]interface{}{
		"id":   "evt_test123",
		"type": "checkout.session.completed",
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id": "cs_test123",
				"metadata": map[string]interface{}{
					"user_id":           "1",
					"subscription_tier": "monthly",
				},
				"customer_email": "test@example.com",
			},
		},
	}

	eventJSON, _ := json.Marshal(eventData)

	// Create a request with the test signature
	req, _ := http.NewRequest("POST", "/webhook", bytes.NewBuffer(eventJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Stripe-Signature", "test_signature")

	// Perform the request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Check the response - should be OK in test mode with test signature
	assert.Equal(t, http.StatusOK, w.Code)
}
