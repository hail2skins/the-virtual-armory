package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Check if we have the required arguments
	if len(os.Args) != 3 {
		fmt.Println("Usage: go run cmd/scripts/create_admin.go <email> <password>")
		os.Exit(1)
	}

	email := os.Args[1]
	password := os.Args[2]

	// Initialize GORM
	db, err := database.InitGORM()
	if err != nil {
		log.Fatalf("Failed to initialize GORM: %v", err)
	}

	// Check if the user already exists
	var existingUser models.User
	result := db.Where("email = ?", email).First(&existingUser)
	if result.Error == nil {
		fmt.Printf("User with email %s already exists\n", email)
		os.Exit(1)
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create the admin user
	user := models.User{
		Email:     email,
		Password:  string(hashedPassword),
		IsAdmin:   true,
		Confirmed: true,
	}

	// Save the user
	result = db.Create(&user)
	if result.Error != nil {
		log.Fatalf("Failed to create admin user: %v", result.Error)
	}

	fmt.Printf("Admin user %s created successfully\n", email)
}
