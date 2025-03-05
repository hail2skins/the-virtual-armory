package email

// EmailService is an interface for email services
type EmailService interface {
	// IsConfigured returns whether the email service is configured
	IsConfigured() bool

	// SendVerificationEmail sends a verification email
	SendVerificationEmail(email, token string) error
}
