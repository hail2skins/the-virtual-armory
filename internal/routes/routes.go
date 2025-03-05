package routes

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
)

// HeadMiddleware handles HEAD requests by converting them to GET requests
func HeadMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == "HEAD" {
			// For HEAD requests, we'll just return a 200 OK with no body
			// This is a simple approach that works for most cases
			c.Header("Content-Type", "text/html; charset=utf-8")
			c.Status(http.StatusOK)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RegisterRoutes registers all routes for the application
func RegisterRoutes(r *gin.Engine, authInstance *auth.Auth) {
	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH", "HEAD"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	// Add HEAD middleware
	r.Use(HeadMiddleware())

	// Add Authboss middleware
	r.Use(authInstance.Middleware())

	// Register health check route
	RegisterHealthRoute(r)

	// Register home routes
	RegisterHomeRoutes(r)

	// Register auth routes
	RegisterAuthRoutes(r, authInstance)
}
