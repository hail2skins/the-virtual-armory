package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/errors"
)

// AuthRequired returns a middleware that requires authentication
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the user is authenticated
		if user, exists := c.Get("user"); !exists || user == nil {
			// User is not authenticated, return an error
			c.Error(errors.NewAuthError("Authentication required"))
			c.Abort()
			return
		}

		// User is authenticated, continue
		c.Next()
	}
}

// AdminRequired returns a middleware that requires admin privileges
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if the user is authenticated
		user, exists := c.Get("user")
		if !exists || user == nil {
			// User is not authenticated, return an error
			c.Error(errors.NewAuthError("Authentication required"))
			c.Abort()
			return
		}

		// Check if the user is an admin
		if admin, ok := user.(interface{ IsAdmin() bool }); !ok || !admin.IsAdmin() {
			// User is not an admin, return an error
			c.Error(errors.NewAuthError("Admin privileges required"))
			c.Abort()
			return
		}

		// User is an admin, continue
		c.Next()
	}
}
