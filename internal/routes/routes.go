package routes

import (
	"io/fs"
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
)

// RegisterRoutes registers all routes for the application
func RegisterRoutes(r *gin.Engine, authInstance *auth.Auth) {
	// Add CORS middleware
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	// Add Authboss middleware
	r.Use(authInstance.Middleware())

	// Register home routes
	RegisterHomeRoutes(r)

	// Health check route
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Static files
	staticFiles, _ := fs.Sub(web.Files, "assets")
	r.StaticFS("/assets", http.FS(staticFiles))

	// Auth routes - these should match the form action URLs in the templates
	r.GET("/auth/login", func(c *gin.Context) {
		web.AuthLoginHandler(c.Writer, c.Request)
	})

	r.GET("/auth/register", func(c *gin.Context) {
		web.AuthRegisterHandler(c.Writer, c.Request)
	})

	r.GET("/auth/recover", func(c *gin.Context) {
		web.AuthRecoverHandler(c.Writer, c.Request)
	})

	// Also add routes without the /auth prefix for convenience
	r.GET("/login", func(c *gin.Context) {
		web.AuthLoginHandler(c.Writer, c.Request)
	})

	r.GET("/register", func(c *gin.Context) {
		web.AuthRegisterHandler(c.Writer, c.Request)
	})

	r.POST("/register", func(c *gin.Context) {
		// This should be moved to an auth controller in the future
		email := c.PostForm("email")
		password := c.PostForm("password")
		confirmPassword := c.PostForm("confirm_password")

		// Validate form data
		if email == "" || password == "" || confirmPassword == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
			return
		}

		if password != confirmPassword {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
			return
		}

		// Get the storer from Authboss
		storer := authInstance.Config.Storage.Server.(*auth.GORMStorer)

		// Create a new user
		user := storer.New(c.Request.Context())
		userWrapper := user.(*auth.UserWrapper)
		userWrapper.PutEmail(email)
		userWrapper.PutPassword(password)
		userWrapper.PutConfirmed(true) // Auto-confirm for now

		// Save the user
		err := storer.Create(c.Request.Context(), user)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
			return
		}

		c.Redirect(http.StatusFound, "/auth/login")
	})

	r.GET("/recover", func(c *gin.Context) {
		web.AuthRecoverHandler(c.Writer, c.Request)
	})

	// Protected routes
	protected := r.Group("/protected")
	protected.Use(authInstance.RequireAuth())
	{
		protected.GET("/profile", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "This is a protected profile page"})
		})
	}

	// Admin routes
	admin := r.Group("/admin")
	admin.Use(authInstance.RequireAdmin())
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "This is the admin dashboard"})
		})
	}
}
