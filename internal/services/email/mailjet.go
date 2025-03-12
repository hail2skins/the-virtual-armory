package email

import (
	"fmt"
	"log"

	"github.com/hail2skins/the-virtual-armory/internal/config"
	mailjet "github.com/mailjet/mailjet-apiv3-go/v3"
)

// MailJetService handles email sending via MailJet
type MailJetService struct {
	client       *mailjet.Client
	senderEmail  string
	senderName   string
	appBaseURL   string
	adminEmail   string
	isConfigured bool
}

// Ensure MailJetService implements EmailService
var _ EmailService = (*MailJetService)(nil)

// NewMailJetService creates a new MailJet service
func NewMailJetService(cfg *config.Config) EmailService {
	// Check if MailJet is configured
	if cfg.MailJetAPIKey == "" || cfg.MailJetSecretKey == "" {
		log.Println("MailJet API keys not configured. Email functionality will be disabled.")
		return &MailJetService{
			isConfigured: false,
		}
	}

	// Create MailJet client
	client := mailjet.NewMailjetClient(cfg.MailJetAPIKey, cfg.MailJetSecretKey)

	return &MailJetService{
		client:       client,
		senderEmail:  cfg.MailJetSenderEmail,
		senderName:   cfg.MailJetSenderName,
		appBaseURL:   cfg.AppBaseURL,
		adminEmail:   cfg.AdminEmail,
		isConfigured: true,
	}
}

// IsConfigured returns whether the MailJet service is configured
func (s *MailJetService) IsConfigured() bool {
	return s.isConfigured
}

// SendVerificationEmail sends an email verification link to the user
func (s *MailJetService) SendVerificationEmail(email, token string) error {
	if !s.isConfigured {
		log.Println("MailJet not configured. Skipping email verification.")
		return nil
	}

	verificationLink := fmt.Sprintf("%s/verify/%s", s.appBaseURL, token)

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: s.senderEmail,
				Name:  s.senderName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: email,
				},
			},
			Subject:  "Verify Your Email - The Virtual Armory",
			TextPart: fmt.Sprintf("Please verify your email by clicking on the following link: %s", verificationLink),
			HTMLPart: fmt.Sprintf(`
				<h3>Welcome to The Virtual Armory!</h3>
				<p>Please verify your email address by clicking the link below:</p>
				<p><a href="%s">Verify Email</a></p>
				<p>If you did not create an account, please ignore this email.</p>
				<p>Thank you,<br>The Virtual Armory Team</p>
			`, verificationLink),
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := s.client.SendMailV31(&messages)
	if err != nil {
		log.Printf("Error sending verification email: %v", err)
		return err
	}

	log.Printf("Verification email sent to %s", email)
	return nil
}

// SendPasswordResetEmail sends a password reset email with a custom link
func (s *MailJetService) SendPasswordResetEmail(email, resetLink string) error {
	if !s.isConfigured {
		log.Println("MailJet not configured. Skipping password reset email.")
		return nil
	}

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: s.senderEmail,
				Name:  s.senderName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: email,
				},
			},
			Subject:  "Reset Your Password - The Virtual Armory",
			TextPart: fmt.Sprintf("Click the following link to reset your password: %s", resetLink),
			HTMLPart: fmt.Sprintf(`
				<h3>Reset Your Password</h3>
				<p>Click the following link to reset your password:</p>
				<p><a href="%s">Reset Password</a></p>
				<p>If you did not request a password reset, please ignore this email.</p>
				<p>This link will expire in 24 hours.</p>
			`, resetLink),
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := s.client.SendMailV31(&messages)
	if err != nil {
		log.Printf("Error sending password reset email: %v", err)
		return err
	}

	log.Printf("Password reset email sent to %s", email)
	return nil
}

// SendContactFormEmail sends a contact form submission to the admin
func (s *MailJetService) SendContactFormEmail(name, email, subject, message string) error {
	if !s.isConfigured {
		log.Println("MailJet not configured. Skipping contact form email.")
		return nil
	}

	// Use the admin email from config, or use the sender email as fallback
	adminEmail := s.adminEmail
	if adminEmail == "" {
		adminEmail = s.senderEmail
		log.Println("Admin email not configured. Using sender email as recipient.")
	}

	// Format the email content
	htmlContent := fmt.Sprintf(`
		<h3>New Contact Form Submission</h3>
		<p><strong>From:</strong> %s &lt;%s&gt;</p>
		<p><strong>Subject:</strong> %s</p>
		<p><strong>Message:</strong></p>
		<div style="padding: 10px; border: 1px solid #ccc; border-radius: 5px; background-color: #f9f9f9;">
			%s
		</div>
		<p>To reply, simply respond to this email or send a new email to: %s</p>
	`, name, email, subject, message, email)

	textContent := fmt.Sprintf(`
New Contact Form Submission
--------------------------
From: %s <%s>
Subject: %s

Message:
%s

To reply, simply respond to this email or send a new email to: %s
	`, name, email, subject, message, email)

	messagesInfo := []mailjet.InfoMessagesV31{
		{
			From: &mailjet.RecipientV31{
				Email: s.senderEmail,
				Name:  s.senderName,
			},
			To: &mailjet.RecipientsV31{
				mailjet.RecipientV31{
					Email: adminEmail,
				},
			},
			Subject:  fmt.Sprintf("Contact Form: %s", subject),
			TextPart: textContent,
			HTMLPart: htmlContent,
			ReplyTo: &mailjet.RecipientV31{
				Email: s.senderEmail,
				Name:  s.senderName,
			},
		},
	}

	messages := mailjet.MessagesV31{Info: messagesInfo}
	_, err := s.client.SendMailV31(&messages)
	if err != nil {
		log.Printf("Error sending contact form email: %v", err)
		return err
	}

	log.Printf("Contact form email sent to admin from %s <%s>", name, email)
	return nil
}
