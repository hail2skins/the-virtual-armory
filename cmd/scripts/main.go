package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	_ "github.com/joho/godotenv/autoload"
	"golang.org/x/crypto/bcrypt"
)

func main() {
	// Define command flags
	createAdminCmd := flag.NewFlagSet("create-admin", flag.ExitOnError)
	deleteUserCmd := flag.NewFlagSet("delete-user", flag.ExitOnError)

	// create-admin flags
	createAdminEmail := createAdminCmd.String("email", "", "Email for the admin user")
	createAdminPassword := createAdminCmd.String("password", "", "Password for the admin user")

	// delete-user flags
	deleteUserEmail := deleteUserCmd.String("email", "", "Email of the user to delete")

	// Check if a subcommand is provided
	if len(os.Args) < 2 {
		fmt.Println("Expected 'create-admin' or 'delete-user' subcommands")
		os.Exit(1)
	}

	// Parse the appropriate command
	switch os.Args[1] {
	case "create-admin":
		createAdminCmd.Parse(os.Args[2:])
		if *createAdminEmail == "" || *createAdminPassword == "" {
			createAdminCmd.PrintDefaults()
			os.Exit(1)
		}
		createAdmin(*createAdminEmail, *createAdminPassword)
	case "delete-user":
		deleteUserCmd.Parse(os.Args[2:])
		if *deleteUserEmail == "" {
			deleteUserCmd.PrintDefaults()
			os.Exit(1)
		}
		deleteUser(*deleteUserEmail)
	default:
		fmt.Printf("Unknown subcommand: %s\n", os.Args[1])
		fmt.Println("Expected 'create-admin' or 'delete-user' subcommands")
		os.Exit(1)
	}
}

// createAdmin creates a new admin user
func createAdmin(email, password string) {
	// Initialize database
	db, err := database.InitGORM()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Check if user already exists
	var existingUser models.User
	result := db.Where("email = ?", email).First(&existingUser)
	if result.Error == nil {
		// User exists, check if they're already an admin
		if existingUser.IsAdmin {
			log.Printf("User %s is already an admin", email)
			return
		}

		// Update user to be an admin
		existingUser.IsAdmin = true
		if err := db.Save(&existingUser).Error; err != nil {
			log.Fatalf("Failed to update user: %v", err)
		}
		log.Printf("User %s has been updated to admin status", email)
		return
	}

	// Hash the password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		log.Fatalf("Failed to hash password: %v", err)
	}

	// Create new admin user
	user := models.User{
		Email:     email,
		Password:  string(hashedPassword),
		IsAdmin:   true,
		Confirmed: true, // Auto-confirm admin users
	}

	if err := db.Create(&user).Error; err != nil {
		log.Fatalf("Failed to create admin user: %v", err)
	}

	log.Printf("Admin user %s created successfully", email)
}

// deleteUser deletes a user by email
func deleteUser(email string) {
	// Initialize database
	db, err := database.InitGORM()
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer database.CloseDB()

	// Find the user
	var user models.User
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		log.Fatalf("User not found: %v", result.Error)
	}

	// Check if user is an admin
	if user.IsAdmin {
		log.Println("Warning: You are about to delete an admin user.")
		fmt.Print("Are you sure you want to continue? (y/n): ")
		var confirm string
		fmt.Scanln(&confirm)
		if confirm != "y" && confirm != "Y" {
			log.Println("Operation cancelled.")
			return
		}
	}

	// Delete the user
	if err := db.Delete(&user).Error; err != nil {
		log.Fatalf("Failed to delete user: %v", err)
	}

	log.Printf("User %s has been deleted", email)
}
