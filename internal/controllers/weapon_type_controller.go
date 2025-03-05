package controllers

import (
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/weapontype"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"gorm.io/gorm"
)

// WeaponTypeController handles requests related to weapon types
type WeaponTypeController struct {
	DB *gorm.DB
}

// NewWeaponTypeController creates a new WeaponTypeController
func NewWeaponTypeController(db *gorm.DB) *WeaponTypeController {
	return &WeaponTypeController{
		DB: db,
	}
}

// Index displays all weapon types
func (c *WeaponTypeController) Index(ctx *gin.Context) {
	var weaponTypes []models.WeaponType
	if err := c.DB.Order("type").Find(&weaponTypes).Error; err != nil {
		log.Printf("Error fetching weapon types: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch weapon types"})
		return
	}

	component := weapontype.Index(weaponTypes)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// Show displays a specific weapon type
func (c *WeaponTypeController) Show(ctx *gin.Context) {
	id := ctx.Param("id")
	var weaponType models.WeaponType
	if err := c.DB.First(&weaponType, id).Error; err != nil {
		log.Printf("Error fetching weapon type: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Weapon type not found"})
		return
	}

	component := weapontype.Show(weaponType)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// New displays the form to create a new weapon type
func (c *WeaponTypeController) New(ctx *gin.Context) {
	component := weapontype.New()
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// Create creates a new weapon type
func (c *WeaponTypeController) Create(ctx *gin.Context) {
	typeStr := ctx.PostForm("type")
	nickname := ctx.PostForm("nickname")

	weaponType := models.WeaponType{
		Type:     typeStr,
		Nickname: nickname,
	}

	if err := c.DB.Create(&weaponType).Error; err != nil {
		log.Printf("Error creating weapon type: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create weapon type"})
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/admin/weapon-types")
}

// Edit displays the form to edit a weapon type
func (c *WeaponTypeController) Edit(ctx *gin.Context) {
	id := ctx.Param("id")
	var weaponType models.WeaponType
	if err := c.DB.First(&weaponType, id).Error; err != nil {
		log.Printf("Error fetching weapon type: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Weapon type not found"})
		return
	}

	component := weapontype.Edit(weaponType)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// Update updates a weapon type
func (c *WeaponTypeController) Update(ctx *gin.Context) {
	id := ctx.Param("id")
	typeStr := ctx.PostForm("type")
	nickname := ctx.PostForm("nickname")

	var weaponType models.WeaponType
	if err := c.DB.First(&weaponType, id).Error; err != nil {
		log.Printf("Error fetching weapon type: %v", err)
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Weapon type not found"})
		return
	}

	weaponType.Type = typeStr
	weaponType.Nickname = nickname

	if err := c.DB.Save(&weaponType).Error; err != nil {
		log.Printf("Error updating weapon type: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update weapon type"})
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/admin/weapon-types/"+id)
}

// Delete deletes a weapon type
func (c *WeaponTypeController) Delete(ctx *gin.Context) {
	id := ctx.Param("id")

	idUint, err := strconv.ParseUint(id, 10, 64)
	if err != nil {
		log.Printf("Error parsing ID: %v", err)
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	if err := c.DB.Delete(&models.WeaponType{}, idUint).Error; err != nil {
		log.Printf("Error deleting weapon type: %v", err)
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete weapon type"})
		return
	}

	ctx.Redirect(http.StatusSeeOther, "/admin/weapon-types")
}
