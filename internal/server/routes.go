package server

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"io/fs"

	"log"

	"context"

	"github.com/hail2skins/the-virtual-armory/cmd/web"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
)

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	// Add Authboss middleware
	r.Use(s.auth.Middleware())

	// Public routes
	r.GET("/", s.HelloWorldHandler)
	r.GET("/health", s.healthHandler)

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
		// Get form data
		email := c.PostForm("email")
		password := c.PostForm("password")
		confirmPassword := c.PostForm("confirm_password")

		log.Printf("Received registration request: email=%s", email)

		// Validate form data
		if email == "" || password == "" || confirmPassword == "" {
			log.Printf("Missing required fields")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Missing required fields"})
			return
		}

		if password != confirmPassword {
			log.Printf("Passwords do not match")
			c.JSON(http.StatusBadRequest, gin.H{"error": "Passwords do not match"})
			return
		}

		// Get the storer from Authboss
		storer := s.auth.Config.Storage.Server.(*auth.GORMStorer)

		// Create a new user
		user := storer.New(context.Background())
		userWrapper := user.(*auth.UserWrapper)
		userWrapper.PutEmail(email)
		userWrapper.PutPassword(password)
		userWrapper.PutConfirmed(true) // Auto-confirm for now

		// Save the user
		err := storer.Create(context.Background(), user)
		if err != nil {
			log.Printf("Error creating user: %v", err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error creating user"})
			return
		}

		log.Printf("User created successfully: %s", email)
		c.Redirect(http.StatusFound, "/auth/login")
	})

	r.GET("/recover", func(c *gin.Context) {
		web.AuthRecoverHandler(c.Writer, c.Request)
	})

	// Protected routes
	protected := r.Group("/protected")
	protected.Use(s.auth.RequireAuth())
	{
		protected.GET("/profile", s.profileHandler)
	}

	// Admin routes
	admin := r.Group("/admin")
	admin.Use(s.auth.RequireAdmin())
	{
		admin.GET("/dashboard", s.adminDashboardHandler)
	}

	return r
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) profileHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "This is a protected profile page"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) adminDashboardHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "This is the admin dashboard"

	c.JSON(http.StatusOK, resp)
}
