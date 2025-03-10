package email

// MockEmailService is a mock implementation of the EmailService interface for testing
type MockEmailService struct {
	SendVerificationEmailCalled bool
	SendVerificationEmailEmail  string
	SendVerificationEmailToken  string
	SendVerificationEmailError  error

	SendPasswordResetEmailCalled bool
	SendPasswordResetEmailEmail  string
	SendPasswordResetEmailLink   string
	SendPasswordResetEmailError  error

	IsConfiguredCalled bool
	IsConfiguredResult bool
}

// SendVerificationEmail is a mock implementation that records the call
func (m *MockEmailService) SendVerificationEmail(email, token string) error {
	m.SendVerificationEmailCalled = true
	m.SendVerificationEmailEmail = email
	m.SendVerificationEmailToken = token
	return m.SendVerificationEmailError
}

// SendPasswordResetEmail is a mock implementation that records the call
func (m *MockEmailService) SendPasswordResetEmail(email, resetLink string) error {
	m.SendPasswordResetEmailCalled = true
	m.SendPasswordResetEmailEmail = email
	m.SendPasswordResetEmailLink = resetLink
	return m.SendPasswordResetEmailError
}

// IsConfigured is a mock implementation that returns a predefined result
func (m *MockEmailService) IsConfigured() bool {
	m.IsConfiguredCalled = true
	return m.IsConfiguredResult
}
