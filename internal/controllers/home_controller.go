package controllers

import (
	"log"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/home"
	"github.com/hail2skins/the-virtual-armory/internal/flash"
	"github.com/hail2skins/the-virtual-armory/internal/services/email"
)

// HomeController handles all home-related routes
type HomeController struct {
	emailService email.EmailService
}

// NewHomeController creates a new instance of HomeController
func NewHomeController(emailService email.EmailService) *HomeController {
	return &HomeController{
		emailService: emailService,
	}
}

// Index handles the home page request
func (h *HomeController) Index(c *gin.Context) {
	// Check if user is logged in
	isLoggedIn := false
	if cookie, err := c.Cookie("is_logged_in"); err == nil && cookie == "true" {
		isLoggedIn = true
	}

	// Get flash message from cookie
	flashMessage, _ := c.Cookie("flash_message")
	flashType, _ := c.Cookie("flash_type")

	// Log the flash message for debugging
	if flashMessage != "" {
		log.Printf("Home page flash message found: %s (type: %s)", flashMessage, flashType)

		// URL decode the flash message
		flashMessage = strings.Replace(flashMessage, "+", " ", -1)
	}

	// Render the home page
	component := home.Index(isLoggedIn)
	if flashMessage != "" {
		// If we have a flash message, use the version of the template that accepts flash messages
		component = home.IndexWithFlash(isLoggedIn, flashMessage, flashType)
	}

	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Error rendering index page: %v", err)
		return
	}

	// Clear flash cookies after rendering
	if flashMessage != "" {
		flash.ClearMessage(c)
	}
}

// About handles the about page request
func (h *HomeController) About(c *gin.Context) {
	// Check if user is logged in
	isLoggedIn := false
	cookie, err := c.Cookie("is_logged_in")
	if err == nil && cookie == "true" {
		isLoggedIn = true
	}

	component := home.About(isLoggedIn)
	err = component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Error rendering about page: %v", err)
		return
	}
}

// Contact handles the contact page request
func (h *HomeController) Contact(c *gin.Context) {
	// Check if user is logged in
	isLoggedIn := false
	cookie, err := c.Cookie("is_logged_in")
	if err == nil && cookie == "true" {
		isLoggedIn = true
	}

	component := home.Contact(isLoggedIn)
	err = component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Error rendering contact page: %v", err)
		return
	}
}

// HandleHelloForm handles the hello form submission
func (h *HomeController) HandleHelloForm(c *gin.Context) {
	name := c.PostForm("name")
	if name == "" {
		// If no name is provided, render the index page
		h.Index(c)
		return
	}

	// Otherwise render the hello response
	component := home.HelloResponse(name)
	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Error rendering hello response: %v", err)
		return
	}
}

// HandleContactForm handles the contact form submission
func (h *HomeController) HandleContactForm(c *gin.Context) {
	// Check if user is logged in
	isLoggedIn := false
	cookie, err := c.Cookie("is_logged_in")
	if err == nil && cookie == "true" {
		isLoggedIn = true
	}

	// Get form data
	name := c.PostForm("name")
	email := c.PostForm("email")
	subject := c.PostForm("subject")
	message := c.PostForm("message")

	// Validate form data
	if name == "" || email == "" || subject == "" || message == "" {
		component := home.ContactWithMessage(isLoggedIn, "All fields are required", "error")
		err := component.Render(c.Request.Context(), c.Writer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Error rendering contact page: %v", err)
		}
		return
	}

	// If email service is not configured, log the message and show success
	if h.emailService == nil || !h.emailService.IsConfigured() {
		log.Printf("Contact form submission (email service not configured):\nFrom: %s <%s>\nSubject: %s\nMessage: %s",
			name, email, subject, message)

		component := home.ContactWithMessage(isLoggedIn, "Your message has been received. We'll get back to you soon!", "success")
		err := component.Render(c.Request.Context(), c.Writer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Error rendering contact page: %v", err)
		}
		return
	}

	// Send email using the email service
	err = h.emailService.SendContactFormEmail(name, email, subject, message)
	if err != nil {
		log.Printf("Error sending contact form email: %v", err)
		component := home.ContactWithMessage(isLoggedIn, "There was an error sending your message. Please try again later.", "error")
		err := component.Render(c.Request.Context(), c.Writer)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			log.Printf("Error rendering contact page: %v", err)
		}
		return
	}

	// Show success message
	component := home.ContactWithMessage(isLoggedIn, "Your message has been sent. We'll get back to you soon!", "success")
	err = component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Error rendering contact page: %v", err)
	}
}
