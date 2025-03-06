package models

import (
	"time"

	"gorm.io/gorm"
)

// User represents a user in the system
type User struct {
	gorm.Model
	Email    string `gorm:"uniqueIndex;not null"`
	Password string `gorm:"not null"`
	IsAdmin  bool   `gorm:"default:false"`

	// Subscription related fields
	SubscriptionTier      string `gorm:"default:'free'"`
	SubscriptionExpiresAt time.Time
	StripeCustomerID      string

	// Authboss required fields
	RecoverToken       string
	RecoverTokenExpiry time.Time

	// For remember me functionality
	RememberToken string

	// For confirm functionality
	ConfirmToken       string
	ConfirmTokenExpiry time.Time
	Confirmed          bool `gorm:"default:false"`

	// For lock functionality
	AttemptCount int
	LastAttempt  time.Time
	Locked       time.Time
}

// GetPID gets the user's primary ID
func (u *User) GetPID() string {
	return u.Email
}

// PutPID sets the user's primary ID
func (u *User) PutPID(pid string) {
	u.Email = pid
}

// GetPassword gets the user's password
func (u *User) GetPassword() string {
	return u.Password
}

// GetEmail gets the user's email
func (u *User) GetEmail() string {
	return u.Email
}

// GetConfirmed gets the user's confirmed status
func (u *User) GetConfirmed() bool {
	return u.Confirmed
}

// GetConfirmSelector gets the user's confirm token
func (u *User) GetConfirmSelector() string {
	return u.ConfirmToken
}

// GetConfirmVerifier gets the user's confirm token
func (u *User) GetConfirmVerifier() string {
	return u.ConfirmToken
}

// GetLocked gets the user's locked status
func (u *User) GetLocked() time.Time {
	return u.Locked
}

// GetAttemptCount gets the user's attempt count
func (u *User) GetAttemptCount() int {
	return u.AttemptCount
}

// GetLastAttempt gets the user's last attempt time
func (u *User) GetLastAttempt() time.Time {
	return u.LastAttempt
}

// GetRecoverSelector gets the user's recover token
func (u *User) GetRecoverSelector() string {
	return u.RecoverToken
}

// GetRecoverVerifier gets the user's recover token
func (u *User) GetRecoverVerifier() string {
	return u.RecoverToken
}

// GetRecoverExpiry gets the user's recover token expiry
func (u *User) GetRecoverExpiry() time.Time {
	return u.RecoverTokenExpiry
}

// GetRememberToken gets the user's remember token
func (u *User) GetRememberToken() string {
	return u.RememberToken
}

// IsAdminUser checks if the user is an admin
func (u *User) IsAdminUser() bool {
	return u.IsAdmin
}

// PutPassword sets the user's password
func (u *User) PutPassword(password string) {
	u.Password = password
}

// PutEmail sets the user's email
func (u *User) PutEmail(email string) {
	u.Email = email
}

// PutConfirmed sets the user's confirmed status
func (u *User) PutConfirmed(confirmed bool) {
	u.Confirmed = confirmed
}

// PutConfirmSelector sets the user's confirm token
func (u *User) PutConfirmSelector(selector string) {
	u.ConfirmToken = selector
}

// PutConfirmVerifier sets the user's confirm token
func (u *User) PutConfirmVerifier(verifier string) {
	u.ConfirmToken = verifier
}

// PutLocked sets the user's locked status
func (u *User) PutLocked(locked time.Time) {
	u.Locked = locked
}

// PutAttemptCount sets the user's attempt count
func (u *User) PutAttemptCount(count int) {
	u.AttemptCount = count
}

// PutLastAttempt sets the user's last attempt time
func (u *User) PutLastAttempt(last time.Time) {
	u.LastAttempt = last
}

// PutRecoverSelector sets the user's recover token
func (u *User) PutRecoverSelector(selector string) {
	u.RecoverToken = selector
}

// PutRecoverVerifier sets the user's recover token
func (u *User) PutRecoverVerifier(verifier string) {
	u.RecoverToken = verifier
}

// PutRecoverExpiry sets the user's recover token expiry
func (u *User) PutRecoverExpiry(expiry time.Time) {
	u.RecoverTokenExpiry = expiry
}

// PutRememberToken sets the user's remember token
func (u *User) PutRememberToken(token string) {
	u.RememberToken = token
}

// PutAdminStatus sets the user's admin status
func (u *User) PutAdminStatus(isAdmin bool) {
	u.IsAdmin = isAdmin
}

// HasActiveSubscription checks if the user has an active subscription
func (u *User) HasActiveSubscription() bool {
	// If the subscription tier is free, return false
	if u.SubscriptionTier == "free" {
		return false
	}

	// If the subscription is lifetime or premium_lifetime, return true
	if u.IsLifetimeSubscriber() {
		return true
	}

	// Otherwise, check if the subscription is expired
	return time.Now().Before(u.SubscriptionExpiresAt)
}

// IsLifetimeSubscriber checks if the user has a lifetime subscription
func (u *User) IsLifetimeSubscriber() bool {
	return u.SubscriptionTier == "lifetime" || u.SubscriptionTier == "premium_lifetime"
}
