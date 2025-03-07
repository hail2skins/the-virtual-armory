package main

import (
	"fmt"
	"log"
	"os"

	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using environment variables")
	}

	// Connect to the database
	db := database.GetDB()
	if db == nil {
		log.Fatalf("Failed to get database connection")
	}

	// Query the user
	var user models.User
	if err := db.Where("id = ?", 15).First(&user).Error; err != nil {
		log.Fatalf("Failed to find user: %v", err)
	}

	// Print the user's subscription information
	fmt.Println("User ID:", user.ID)
	fmt.Println("Email:", user.Email)
	fmt.Println("Subscription Tier:", user.SubscriptionTier)
	fmt.Println("Subscription Expires At:", user.SubscriptionExpiresAt.Format("January 2, 2006"))
	fmt.Println("Subscription Canceled:", user.SubscriptionCanceled)
	fmt.Println("Stripe Customer ID:", user.StripeCustomerID)
	fmt.Println("Stripe Subscription ID:", user.StripeSubscriptionID)

	// Query the user's payments
	var payments []models.Payment
	if err := db.Where("user_id = ?", user.ID).Order("created_at desc").Find(&payments).Error; err != nil {
		log.Fatalf("Failed to find payments: %v", err)
	}

	// Print the user's payments
	fmt.Println("\nPayments:")
	for i, payment := range payments {
		fmt.Printf("%d. %s - %s - $%.2f - %s\n", i+1, payment.CreatedAt.Format("January 2, 2006"), payment.Description, float64(payment.Amount)/100.0, payment.Status)
	}

	os.Exit(0)
}
