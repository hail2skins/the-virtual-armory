package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/manufacturer"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
)

// ManufacturerController handles all manufacturer-related routes
type ManufacturerController struct{}

// NewManufacturerController creates a new instance of ManufacturerController
func NewManufacturerController() *ManufacturerController {
	return &ManufacturerController{}
}

// Index handles the manufacturers list page request
func (m *ManufacturerController) Index(c *gin.Context) {
	// Get all manufacturers from the database
	var manufacturers []models.Manufacturer
	result := database.GetDB().Find(&manufacturers)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		log.Printf("Error fetching manufacturers: %v", result.Error)
		return
	}

	// Render the manufacturers index page using templ component
	component := manufacturer.Index(manufacturers)
	component.Render(c, c.Writer)
}

// New handles the new manufacturer form request
func (m *ManufacturerController) New(c *gin.Context) {
	// Render the new manufacturer form using templ component
	component := manufacturer.New()
	component.Render(c, c.Writer)
}

// Create handles the creation of a new manufacturer
func (m *ManufacturerController) Create(c *gin.Context) {
	// Get form data
	name := c.PostForm("name")
	nickname := c.PostForm("nickname")
	country := c.PostForm("country")

	// Validate required fields
	if name == "" || country == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and country are required"})
		return
	}

	// Create a new manufacturer
	manufacturer := models.Manufacturer{
		Name:     name,
		Nickname: nickname,
		Country:  country,
	}

	// Save to database
	result := database.GetDB().Create(&manufacturer)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		log.Printf("Error creating manufacturer: %v", result.Error)
		return
	}

	// Redirect to manufacturers index
	c.Redirect(http.StatusFound, "/admin/manufacturers")
}

// Show handles the manufacturer details page request
func (m *ManufacturerController) Show(c *gin.Context) {
	// Get manufacturer ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manufacturer ID"})
		return
	}

	// Get manufacturer from database
	var mfr models.Manufacturer
	result := database.GetDB().First(&mfr, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manufacturer not found"})
		return
	}

	// Render the manufacturer details page using templ component
	component := manufacturer.Show(mfr)
	component.Render(c, c.Writer)
}

// Edit handles the edit manufacturer form request
func (m *ManufacturerController) Edit(c *gin.Context) {
	// Get manufacturer ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manufacturer ID"})
		return
	}

	// Get manufacturer from database
	var mfr models.Manufacturer
	result := database.GetDB().First(&mfr, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manufacturer not found"})
		return
	}

	// Render the edit manufacturer form using templ component
	component := manufacturer.Edit(mfr)
	component.Render(c, c.Writer)
}

// Update handles the update of an existing manufacturer
func (m *ManufacturerController) Update(c *gin.Context) {
	// Get manufacturer ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manufacturer ID"})
		return
	}

	// Get manufacturer from database
	var manufacturer models.Manufacturer
	result := database.GetDB().First(&manufacturer, id)
	if result.Error != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Manufacturer not found"})
		return
	}

	// Get form data
	name := c.PostForm("name")
	nickname := c.PostForm("nickname")
	country := c.PostForm("country")

	// Validate required fields
	if name == "" || country == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Name and country are required"})
		return
	}

	// Update manufacturer
	manufacturer.Name = name
	manufacturer.Nickname = nickname
	manufacturer.Country = country

	// Save to database
	result = database.GetDB().Save(&manufacturer)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		log.Printf("Error updating manufacturer: %v", result.Error)
		return
	}

	// Redirect to manufacturer details
	c.Redirect(http.StatusFound, "/admin/manufacturers/"+strconv.FormatUint(id, 10))
}

// Delete handles the deletion of a manufacturer
func (m *ManufacturerController) Delete(c *gin.Context) {
	// Get manufacturer ID from URL
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid manufacturer ID"})
		return
	}

	// Delete manufacturer from database
	result := database.GetDB().Delete(&models.Manufacturer{}, id)
	if result.Error != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": result.Error.Error()})
		log.Printf("Error deleting manufacturer: %v", result.Error)
		return
	}

	// Redirect to manufacturers index
	c.Redirect(http.StatusFound, "/admin/manufacturers")
}
