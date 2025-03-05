package server

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/config"
	"github.com/hail2skins/the-virtual-armory/internal/routes"
	"gorm.io/gorm"
)

// Server represents the HTTP server
type Server struct {
	router *gin.Engine
	config *config.Config
	auth   *auth.Auth
	db     *gorm.DB
}

// New creates a new server instance
func New(cfg *config.Config, auth *auth.Auth, db *gorm.DB) *Server {
	router := gin.Default()

	// Create a new server
	s := &Server{
		router: router,
		config: cfg,
		auth:   auth,
		db:     db,
	}

	// Set up routes
	s.setupRoutes()

	return s
}

// setupRoutes configures all the routes for our application
func (s *Server) setupRoutes() {
	// Register static file handlers
	s.router.Static("/assets", "./cmd/web/assets")
	s.router.Static("/styles", "./cmd/web/styles")
	s.router.Static("/static", "./static")

	// Load HTML templates
	// s.router.LoadHTMLGlob("cmd/web/views/**/*.html")

	// Add a specific handler for favicon.ico
	faviconHandler := func(c *gin.Context) {
		log.Println("Favicon request received")
		c.File("./cmd/web/assets/favicon.ico")
	}
	s.router.GET("/favicon.ico", faviconHandler)
	s.router.HEAD("/favicon.ico", faviconHandler)

	// Register all routes
	routes.RegisterRoutes(s.router, s.auth, s.db, s.config)
}

// Start starts the HTTP server
func (s *Server) Start() error {
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", s.config.Port),
		Handler:      s.router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	log.Printf("Starting server on port %d", s.config.Port)
	return srv.ListenAndServe()
}

// Router returns the router instance
func (s *Server) Router() *gin.Engine {
	return s.router
}
