package models

import (
	"time"

	"gorm.io/gorm"
)

// Payment represents a payment made by a user
type Payment struct {
	gorm.Model
	UserID           uint   `gorm:"not null"`
	Amount           int    `gorm:"not null"` // Amount in cents
	Currency         string `gorm:"not null;default:'usd'"`
	Status           string `gorm:"not null;default:'pending'"`
	StripePaymentID  string
	StripeCustomerID string
	Description      string
	SubscriptionTier string `gorm:"not null"`
	PeriodStart      time.Time
	PeriodEnd        time.Time
}

// CreatePayment creates a new payment in the database
func CreatePayment(db *gorm.DB, payment *Payment) error {
	return db.Create(payment).Error
}

// FindPaymentByID retrieves a payment by its ID
func FindPaymentByID(db *gorm.DB, id uint) (*Payment, error) {
	var payment Payment
	if err := db.First(&payment, id).Error; err != nil {
		return nil, err
	}
	return &payment, nil
}

// FindPaymentsByUserID retrieves all payments for a user
func FindPaymentsByUserID(db *gorm.DB, userID uint) ([]Payment, error) {
	var payments []Payment
	if err := db.Where("user_id = ?", userID).Order("created_at desc").Find(&payments).Error; err != nil {
		return nil, err
	}
	return payments, nil
}

// UpdatePayment updates an existing payment in the database
func UpdatePayment(db *gorm.DB, payment *Payment) error {
	return db.Save(payment).Error
}
