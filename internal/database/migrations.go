package database

import (
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"gorm.io/gorm"
)

// RunMigrations runs all database migrations
func RunMigrations(db *gorm.DB) error {
	// Auto migrate all models
	if err := db.AutoMigrate(
		&models.User{},
		&models.WeaponType{},
		&models.Caliber{},
		&models.Manufacturer{},
		&models.Gun{},
		&models.Payment{},
	); err != nil {
		return err
	}

	// Add StripeSubscriptionID column to users table if it doesn't exist
	if !db.Migrator().HasColumn(&models.User{}, "stripe_subscription_id") {
		if err := db.Migrator().AddColumn(&models.User{}, "stripe_subscription_id"); err != nil {
			return err
		}
	}

	// Add SubscriptionCanceled column to users table if it doesn't exist
	if !db.Migrator().HasColumn(&models.User{}, "subscription_canceled") {
		if err := db.Migrator().AddColumn(&models.User{}, "subscription_canceled"); err != nil {
			return err
		}
	}

	return nil
}
