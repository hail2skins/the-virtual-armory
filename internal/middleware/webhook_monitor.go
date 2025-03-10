package middleware

import (
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// WebhookStats tracks statistics about webhook calls
type WebhookStats struct {
	TotalRequests      int64
	SuccessfulRequests int64
	FailedRequests     int64
	LastRequestTime    time.Time
	LastErrorTime      time.Time
	LastError          string
	mu                 sync.Mutex
}

var webhookStats = WebhookStats{}

// GetWebhookStats returns the current webhook statistics
func GetWebhookStats() WebhookStats {
	webhookStats.mu.Lock()
	defer webhookStats.mu.Unlock()
	return webhookStats
}

// ResetWebhookStats resets all webhook statistics to zero
func ResetWebhookStats() {
	webhookStats.mu.Lock()
	defer webhookStats.mu.Unlock()

	webhookStats.TotalRequests = 0
	webhookStats.SuccessfulRequests = 0
	webhookStats.FailedRequests = 0
	webhookStats.LastError = ""
	// Don't reset time fields to zero as they would be invalid
}

// WebhookMonitor middleware tracks webhook health and metrics
func WebhookMonitor() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record start time
		startTime := time.Now()

		// Create a response writer that captures the status code
		blw := &bodyLogWriter{body: []byte{}, ResponseWriter: c.Writer}
		c.Writer = blw

		// Process request
		c.Next()

		// Update stats after request is processed
		webhookStats.mu.Lock()
		defer webhookStats.mu.Unlock()

		webhookStats.TotalRequests++
		webhookStats.LastRequestTime = startTime

		// Check if request was successful
		if blw.Status() >= 200 && blw.Status() < 300 {
			webhookStats.SuccessfulRequests++
		} else {
			webhookStats.FailedRequests++
			webhookStats.LastErrorTime = startTime
			webhookStats.LastError = string(blw.body)

			// Log webhook errors
			log.Printf("[WEBHOOK ERROR] Status: %d, Error: %s", blw.Status(), string(blw.body))
		}

		// Log request duration for monitoring
		duration := time.Since(startTime)
		log.Printf("[WEBHOOK] Path: %s, Method: %s, Status: %d, Duration: %s",
			c.Request.URL.Path, c.Request.Method, blw.Status(), duration)
	}
}

// bodyLogWriter captures the response body and status code
type bodyLogWriter struct {
	gin.ResponseWriter
	body []byte
}

// Write captures the response body
func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body = append(w.body, b...)
	return w.ResponseWriter.Write(b)
}

// WebhookHealthCheck returns a handler that checks webhook health
func WebhookHealthCheck() gin.HandlerFunc {
	return func(c *gin.Context) {
		stats := GetWebhookStats()

		// Calculate success rate
		var successRate float64 = 0
		if stats.TotalRequests > 0 {
			successRate = float64(stats.SuccessfulRequests) / float64(stats.TotalRequests) * 100
		}

		// Check if webhook is healthy
		isHealthy := true
		healthStatus := "healthy"

		// If we've had requests and the success rate is below 80%, consider unhealthy
		if stats.TotalRequests > 0 && successRate < 80 {
			isHealthy = false
			healthStatus = "unhealthy"
		}

		// If it's been more than 24 hours since the last request, consider degraded
		if stats.LastRequestTime.IsZero() || time.Since(stats.LastRequestTime) > 24*time.Hour {
			if isHealthy {
				healthStatus = "degraded"
			}
		}

		// Return health status
		c.JSON(http.StatusOK, gin.H{
			"status":            healthStatus,
			"total_requests":    stats.TotalRequests,
			"successful":        stats.SuccessfulRequests,
			"failed":            stats.FailedRequests,
			"success_rate":      successRate,
			"last_request":      stats.LastRequestTime,
			"last_error":        stats.LastErrorTime,
			"last_error_detail": stats.LastError,
		})
	}
}
