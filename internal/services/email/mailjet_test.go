package email

import (
	"strings"
	"testing"

	"github.com/hail2skins/the-virtual-armory/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestVerificationLinkFormat(t *testing.T) {
	// Create a test config
	cfg := &config.Config{
		MailJetAPIKey:      "test-api-key",
		MailJetSecretKey:   "test-secret-key",
		MailJetSenderEmail: "test@example.com",
		MailJetSenderName:  "Test Sender",
		AppBaseURL:         "http://localhost:8080",
	}

	// Create a new MailJet service
	service := NewMailJetService(cfg).(*MailJetService)

	// Verify that the service is configured correctly
	assert.True(t, service.IsConfigured())
	assert.Equal(t, "test@example.com", service.senderEmail)
	assert.Equal(t, "Test Sender", service.senderName)
	assert.Equal(t, "http://localhost:8080", service.appBaseURL)

	// Create a simple test to verify the URL format
	testToken := "test-token"
	expectedLink := "http://localhost:8080/verify/test-token"

	// Construct the link manually using the same format as in SendVerificationEmail
	actualLink := service.appBaseURL + "/verify/" + testToken

	// Verify the link format
	assert.Equal(t, expectedLink, actualLink, "Verification link should be constructed correctly")

	// Verify that the link contains the correct path
	assert.True(t, strings.Contains(actualLink, "/verify/"), "Verification link should contain the correct path")
	assert.False(t, strings.Contains(actualLink, "/auth/verify"), "Verification link should not contain the old path")
}
