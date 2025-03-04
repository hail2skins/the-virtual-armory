package auth

import (
	"time"

	"github.com/hail2skins/the-virtual-armory/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// UserWrapper wraps the User model to implement the authboss.User interface
type UserWrapper struct {
	*models.User
}

// NewUserWrapper creates a new UserWrapper
func NewUserWrapper(user *models.User) *UserWrapper {
	return &UserWrapper{User: user}
}

// GetPID gets the user's primary ID
func (u *UserWrapper) GetPID() string {
	return u.Email
}

// PutPID sets the user's primary ID
func (u *UserWrapper) PutPID(pid string) {
	u.Email = pid
}

// GetPassword gets the user's password
func (u *UserWrapper) GetPassword() string {
	return u.Password
}

// GetEmail gets the user's email
func (u *UserWrapper) GetEmail() string {
	return u.Email
}

// GetConfirmed gets the user's confirmed status
func (u *UserWrapper) GetConfirmed() bool {
	return u.Confirmed
}

// GetConfirmSelector gets the user's confirm token
func (u *UserWrapper) GetConfirmSelector() string {
	return u.ConfirmToken
}

// GetConfirmVerifier gets the user's confirm token
func (u *UserWrapper) GetConfirmVerifier() string {
	return u.ConfirmToken
}

// GetLocked gets the user's locked status
func (u *UserWrapper) GetLocked() time.Time {
	return u.Locked
}

// GetAttemptCount gets the user's attempt count
func (u *UserWrapper) GetAttemptCount() int {
	return u.AttemptCount
}

// GetLastAttempt gets the user's last attempt time
func (u *UserWrapper) GetLastAttempt() time.Time {
	return u.LastAttempt
}

// GetRecoverSelector gets the user's recover token
func (u *UserWrapper) GetRecoverSelector() string {
	return u.RecoverToken
}

// GetRecoverVerifier gets the user's recover token
func (u *UserWrapper) GetRecoverVerifier() string {
	return u.RecoverToken
}

// GetRecoverExpiry gets the user's recover token expiry
func (u *UserWrapper) GetRecoverExpiry() time.Time {
	return u.RecoverTokenExpiry
}

// GetRememberToken gets the user's remember token
func (u *UserWrapper) GetRememberToken() string {
	return u.RememberToken
}

// IsAdmin checks if the user is an admin
func (u *UserWrapper) IsAdmin() bool {
	return u.IsAdminUser()
}

// PutPassword sets the user's password
func (u *UserWrapper) PutPassword(password string) {
	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		// Log the error and store an empty password
		// In a production environment, you might want to handle this differently
		u.Password = ""
		return
	}

	// Store the hashed password
	u.Password = string(hashedPassword)
}

// PutEmail sets the user's email
func (u *UserWrapper) PutEmail(email string) {
	u.Email = email
}

// PutConfirmed sets the user's confirmed status
func (u *UserWrapper) PutConfirmed(confirmed bool) {
	u.Confirmed = confirmed
}

// PutConfirmSelector sets the user's confirm token
func (u *UserWrapper) PutConfirmSelector(selector string) {
	u.ConfirmToken = selector
}

// PutConfirmVerifier sets the user's confirm token
func (u *UserWrapper) PutConfirmVerifier(verifier string) {
	u.ConfirmToken = verifier
}

// PutLocked sets the user's locked status
func (u *UserWrapper) PutLocked(locked time.Time) {
	u.Locked = locked
}

// PutAttemptCount sets the user's attempt count
func (u *UserWrapper) PutAttemptCount(count int) {
	u.AttemptCount = count
}

// PutLastAttempt sets the user's last attempt time
func (u *UserWrapper) PutLastAttempt(last time.Time) {
	u.LastAttempt = last
}

// PutRecoverSelector sets the user's recover token
func (u *UserWrapper) PutRecoverSelector(selector string) {
	u.RecoverToken = selector
}

// PutRecoverVerifier sets the user's recover token
func (u *UserWrapper) PutRecoverVerifier(verifier string) {
	u.RecoverToken = verifier
}

// PutRecoverExpiry sets the user's recover token expiry
func (u *UserWrapper) PutRecoverExpiry(expiry time.Time) {
	u.RecoverTokenExpiry = expiry
}

// PutRememberToken sets the user's remember token
func (u *UserWrapper) PutRememberToken(token string) {
	u.RememberToken = token
}

// PutAdminStatus sets the user's admin status
func (u *UserWrapper) PutAdminStatus(isAdmin bool) {
	u.User.IsAdmin = isAdmin
}
