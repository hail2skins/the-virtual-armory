package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/config"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/server"
)

func main() {
	// Load configuration
	cfg := config.New()

	// Initialize database
	_, err := database.InitGORM()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Initialize Authboss
	authInstance, err := auth.New()
	if err != nil {
		log.Fatalf("Failed to initialize Authboss: %v", err)
	}

	// Create and start the server
	srv := server.New(cfg, authInstance)

	// Start the server in a goroutine
	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	log.Printf("Server started on port %d", cfg.Port)

	// Wait for interrupt signal to gracefully shut down the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}
