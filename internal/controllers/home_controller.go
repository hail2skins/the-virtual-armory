package controllers

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/home"
)

// HomeController handles all home-related routes
type HomeController struct{}

// NewHomeController creates a new instance of HomeController
func NewHomeController() *HomeController {
	return &HomeController{}
}

// Index handles the home page request
func (h *HomeController) Index(c *gin.Context) {
	// Check if user is logged in
	isLoggedIn := false
	if cookie, err := c.Cookie("is_logged_in"); err == nil && cookie == "true" {
		isLoggedIn = true
	}

	// Check for flash messages and clear them
	if _, err := c.Cookie("flash_message"); err == nil {
		c.SetCookie("flash_message", "", -1, "/", "", false, true)
	}

	if _, err := c.Cookie("flash_type"); err == nil {
		c.SetCookie("flash_type", "", -1, "/", "", false, true)
	}

	// Render the home page
	component := home.Index(isLoggedIn)
	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Error rendering index page: %v", err)
		return
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
