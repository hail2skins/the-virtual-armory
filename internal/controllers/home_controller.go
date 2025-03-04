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
	component := home.Index()
	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Error rendering index page: %v", err)
		return
	}
}

// About handles the about page request
func (h *HomeController) About(c *gin.Context) {
	component := home.About()
	err := component.Render(c.Request.Context(), c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		log.Printf("Error rendering about page: %v", err)
		return
	}
}

// Contact handles the contact page request
func (h *HomeController) Contact(c *gin.Context) {
	component := home.Contact()
	err := component.Render(c.Request.Context(), c.Writer)
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
