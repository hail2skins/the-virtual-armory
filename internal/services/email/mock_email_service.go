package email

// MockEmailService is a mock implementation of the EmailService interface for testing
type MockEmailService struct {
	SendVerificationEmailCalled bool
	SendVerificationEmailEmail  string
	SendVerificationEmailToken  string
	SendVerificationEmailError  error

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

// IsConfigured is a mock implementation that returns a predefined result
func (m *MockEmailService) IsConfigured() bool {
	m.IsConfiguredCalled = true
	return m.IsConfiguredResult
}
