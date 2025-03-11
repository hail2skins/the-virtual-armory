package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/errors"
	"github.com/hail2skins/the-virtual-armory/internal/logger"
)

// ErrorHandler returns a middleware that handles errors using our custom error types and logger
func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request
		c.Next()

		// Check if there were any errors during processing
		if len(c.Errors) > 0 {
			// Get the last error
			err := c.Errors.Last()

			// Log the error with our custom logger
			logger.Error("Request error", err.Err, map[string]interface{}{
				"path":    c.Request.URL.Path,
				"method":  c.Request.Method,
				"user_id": getUserID(c),
			})

			// Handle the error with our custom error handler
			errors.HandleError(c, err.Err)

			// Stop further handlers from executing
			c.Abort()
		}
	}
}

// getUserID attempts to get the user ID from the context
// Returns 0 if no user ID is found
func getUserID(c *gin.Context) uint {
	if user, exists := c.Get("user"); exists {
		if u, ok := user.(interface{ GetID() uint }); ok {
			return u.GetID()
		}
	}
	return 0
}
