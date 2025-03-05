package config

import (
	"log"
	"os"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
)

// Config holds all configuration for the application
type Config struct {
	Port         int
	DatabaseURL  string
	Environment  string
	CookieSecret string
	SessionName  string
	// MailJet configuration
	MailJetAPIKey      string
	MailJetSecretKey   string
	MailJetSenderEmail string
	MailJetSenderName  string
	AppBaseURL         string
}

// New creates a new Config instance with values from environment variables
func New() *Config {
	port, err := strconv.Atoi(getEnv("PORT", "8080"))
	if err != nil {
		log.Printf("Invalid PORT, using default: %v", err)
		port = 8080
	}

	return &Config{
		Port:               port,
		DatabaseURL:        getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/the_virtual_armory?sslmode=disable"),
		Environment:        getEnv("ENVIRONMENT", "development"),
		CookieSecret:       getEnv("COOKIE_SECRET", "something-very-secret"),
		SessionName:        getEnv("SESSION_NAME", "the-virtual-armory-session"),
		MailJetAPIKey:      getEnv("MAILJET_API_KEY", ""),
		MailJetSecretKey:   getEnv("MAILJET_SECRET_KEY", ""),
		MailJetSenderEmail: getEnv("MAILJET_SENDER_EMAIL", "noreply@thevirtualarmory.com"),
		MailJetSenderName:  getEnv("MAILJET_SENDER_NAME", "The Virtual Armory"),
		AppBaseURL:         getEnv("APP_BASE_URL", "http://localhost:8080"),
	}
}

// IsDevelopment returns true if the environment is set to development
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction returns true if the environment is set to production
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// IsTest returns true if the environment is set to test
func (c *Config) IsTest() bool {
	return c.Environment == "test"
}

// Helper function to get environment variables with a default value
func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
