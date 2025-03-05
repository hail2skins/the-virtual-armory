package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/caliber"
	"gorm.io/gorm"

	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
)

// CaliberController handles caliber-related operations
type CaliberController struct {
	DB *gorm.DB
}

// NewCaliberController creates a new caliber controller
func NewCaliberController() *CaliberController {
	return &CaliberController{
		DB: database.GetDB(),
	}
}

// Index displays all calibers
func (c *CaliberController) Index(ctx *gin.Context) {
	var calibers []models.Caliber
	if err := c.DB.Find(&calibers).Error; err != nil {
		log.Printf("Error fetching calibers: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to retrieve calibers",
		})
		return
	}

	log.Printf("Found %d calibers", len(calibers))
	component := caliber.Index(calibers)
	component.Render(ctx, ctx.Writer)
}

// Show displays a single caliber
func (c *CaliberController) Show(ctx *gin.Context) {
	id := ctx.Param("id")
	var cal models.Caliber

	if err := c.DB.First(&cal, id).Error; err != nil {
		log.Printf("Error finding caliber %s: %v", id, err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Caliber not found",
		})
		return
	}

	component := caliber.Show(cal)
	component.Render(ctx, ctx.Writer)
}

// New displays the form to create a new caliber
func (c *CaliberController) New(ctx *gin.Context) {
	component := caliber.New()
	component.Render(ctx, ctx.Writer)
}

// Create creates a new caliber
func (c *CaliberController) Create(ctx *gin.Context) {
	cal := models.Caliber{
		Caliber:  ctx.PostForm("caliber"),
		Nickname: ctx.PostForm("nickname"),
	}

	if err := c.DB.Create(&cal).Error; err != nil {
		log.Printf("Error creating caliber: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to create caliber",
		})
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/admin/calibers")
}

// Edit displays the form to edit a caliber
func (c *CaliberController) Edit(ctx *gin.Context) {
	id := ctx.Param("id")
	var cal models.Caliber

	if err := c.DB.First(&cal, id).Error; err != nil {
		log.Printf("Error finding caliber %s for edit: %v", id, err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Caliber not found",
		})
		return
	}

	component := caliber.Edit(cal)
	component.Render(ctx, ctx.Writer)
}

// Update updates a caliber
func (c *CaliberController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	var cal models.Caliber

	if err := c.DB.First(&cal, id).Error; err != nil {
		log.Printf("Error finding caliber %s for update: %v", id, err)
		ctx.JSON(http.StatusNotFound, gin.H{
			"error": "Caliber not found",
		})
		return
	}

	cal.Caliber = ctx.PostForm("caliber")
	cal.Nickname = ctx.PostForm("nickname")

	if err := c.DB.Save(&cal).Error; err != nil {
		log.Printf("Error updating caliber %s: %v", id, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to update caliber",
		})
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/admin/calibers")
}

// Delete deletes a caliber
func (c *CaliberController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")
	idInt, err := strconv.Atoi(id)
	if err != nil {
		log.Printf("Invalid caliber ID %s: %v", id, err)
		ctx.JSON(http.StatusBadRequest, gin.H{
			"error": "Invalid caliber ID",
		})
		return
	}

	if err := c.DB.Delete(&models.Caliber{}, idInt).Error; err != nil {
		log.Printf("Error deleting caliber %s: %v", id, err)
		ctx.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete caliber",
		})
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/admin/calibers")
}
