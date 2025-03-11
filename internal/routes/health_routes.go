package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// RegisterHealthRoute registers the basic health check route
func RegisterHealthRoute(router *gin.Engine) {
	// Basic health check - unprotected for load balancers and basic monitoring
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
		})
	})
}
