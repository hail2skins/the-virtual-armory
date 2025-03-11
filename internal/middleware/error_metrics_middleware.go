package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/metrics"
)

var (
	// Global error metrics instance
	errorMetrics *metrics.ErrorMetrics
)

// Initialize the error metrics
func init() {
	errorMetrics = metrics.NewErrorMetrics()

	// Start a goroutine to periodically clean up old error entries
	go func() {
		for {
			time.Sleep(1 * time.Hour)
			errorMetrics.Cleanup(24 * time.Hour * 7) // Keep errors for 7 days
		}
	}()
}

// GetErrorMetrics returns the global error metrics instance
func GetErrorMetrics() *metrics.ErrorMetrics {
	return errorMetrics
}

// ErrorMetricsMiddleware returns a middleware that records error metrics
func ErrorMetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Record the start time
		startTime := time.Now()

		// Process the request
		c.Next()

		// Calculate the request duration
		duration := time.Since(startTime).Seconds()

		// If there were errors, record them
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Get the error type
			errorType := "internal_error" // Default type

			// Try to determine the error type
			switch err.Err.(type) {
			case interface{ ErrorType() string }:
				// If the error has an ErrorType method, use that
				errorType = err.Err.(interface{ ErrorType() string }).ErrorType()
			default:
				// Otherwise, use the error message
				errorType = err.Error()
			}

			// Record the error metrics
			errorMetrics.Record(
				errorType,
				c.Writer.Status(),
				duration,
				c.Request.URL.Path,
			)
		}
	}
}
