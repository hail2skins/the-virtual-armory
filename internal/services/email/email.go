package email

// EmailService is an interface for email services
type EmailService interface {
	// IsConfigured returns whether the email service is configured
	IsConfigured() bool

	// SendVerificationEmail sends a verification email
	SendVerificationEmail(email, token string) error

	// SendPasswordResetEmail sends a password reset email with a custom link
	SendPasswordResetEmail(email, resetLink string) error

	// SendContactFormEmail sends a contact form submission to the admin
	SendContactFormEmail(name, email, subject, message string) error
}
