package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	_ "github.com/joho/godotenv/autoload"

	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/database"
)

type Server struct {
	port int
	db   database.Service
	auth *auth.Auth
}

func NewServer() *http.Server {
	port, _ := strconv.Atoi(os.Getenv("PORT"))

	// Initialize GORM
	_, err := database.InitGORM()
	if err != nil {
		log.Fatalf("Failed to initialize GORM: %v", err)
	}

	// Initialize Authboss
	authInstance, err := auth.New()
	if err != nil {
		log.Fatalf("Failed to initialize Authboss: %v", err)
	}

	NewServer := &Server{
		port: port,
		db:   database.New(),
		auth: authInstance,
	}

	// Declare Server config
	server := &http.Server{
		Addr:         fmt.Sprintf(":%d", NewServer.port),
		Handler:      NewServer.RegisterRoutes(),
		IdleTimeout:  time.Minute,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return server
}
