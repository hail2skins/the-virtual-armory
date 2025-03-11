package controllers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/middleware"
	"github.com/stretchr/testify/assert"
)

func TestAdminController_ErrorMetrics(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.New()

	// Create a new admin controller
	controller := NewAdminController()

	// Register the route
	router.GET("/admin/error-metrics", controller.ErrorMetrics)

	// Create a test request
	req, _ := http.NewRequest("GET", "/admin/error-metrics", nil)
	req.Header.Set("Accept", "application/json")

	// Create a test response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that the response contains the expected fields
	assert.Contains(t, response, "error_rates")
	assert.Contains(t, response, "latency_percentiles")
	assert.Contains(t, response, "recent_errors")
	assert.Contains(t, response, "stats")
	assert.Contains(t, response, "time_range")
}

func TestAdminController_ErrorMetrics_WithData(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a new router
	router := gin.New()

	// Create a new admin controller
	controller := NewAdminController()

	// Register the route
	router.GET("/admin/error-metrics", controller.ErrorMetrics)

	// Add some test data to the error metrics
	metrics := middleware.GetErrorMetrics()
	metrics.Record("test_error", http.StatusInternalServerError, 0.5, "/test")
	metrics.Record("test_error", http.StatusInternalServerError, 0.7, "/test")
	metrics.Record("another_error", http.StatusBadRequest, 0.3, "/another")

	// Create a test request
	req, _ := http.NewRequest("GET", "/admin/error-metrics?range=1h", nil)
	req.Header.Set("Accept", "application/json")

	// Create a test response recorder
	w := httptest.NewRecorder()

	// Perform the request
	router.ServeHTTP(w, req)

	// Check the status code
	assert.Equal(t, http.StatusOK, w.Code)

	// Parse the response
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)

	// Check that the response contains the expected fields
	assert.Contains(t, response, "error_rates")
	assert.Contains(t, response, "latency_percentiles")
	assert.Contains(t, response, "recent_errors")
	assert.Contains(t, response, "stats")
	assert.Equal(t, "1h", response["time_range"])

	// Check that the error rates contain our test data
	errorRates, ok := response["error_rates"].(map[string]interface{})
	assert.True(t, ok)
	assert.Contains(t, errorRates, "test_error")
	assert.Contains(t, errorRates, "another_error")
}
