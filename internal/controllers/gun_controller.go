package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/gun"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"gorm.io/gorm"
)

// GunController handles requests related to guns
type GunController struct {
	DB *gorm.DB
}

// NewGunController creates a new GunController
func NewGunController(db *gorm.DB) *GunController {
	return &GunController{
		DB: db,
	}
}

// Index displays a list of all guns belonging to the current user
func (c *GunController) Index(ctx *gin.Context) {
	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to get current user"})
		return
	}

	// Get all guns for the current user
	guns, err := models.FindGunsByOwner(c.DB, user.ID)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve guns"})
		return
	}

	// Render the index template
	component := gun.Index(guns)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// Show displays details for a specific gun
func (c *GunController) Show(ctx *gin.Context) {
	// Get the gun ID from the URL
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid gun ID"})
		return
	}

	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to get current user"})
		return
	}

	// Get the gun
	gunItem, err := models.FindGunByID(c.DB, uint(id), user.ID)
	if err != nil {
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Gun not found"})
		return
	}

	// Render the show template
	component := gun.Show(*gunItem)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// New displays the form to create a new gun
func (c *GunController) New(ctx *gin.Context) {
	// Get all weapon types, calibers, and manufacturers for the dropdown lists
	var weaponTypes []models.WeaponType
	var calibers []models.Caliber
	var manufacturers []models.Manufacturer

	if err := c.DB.Order("type").Find(&weaponTypes).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve weapon types"})
		return
	}

	if err := c.DB.Order("caliber").Find(&calibers).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve calibers"})
		return
	}

	if err := c.DB.Order("name").Find(&manufacturers).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve manufacturers"})
		return
	}

	// Render the new template
	component := gun.New(weaponTypes, calibers, manufacturers)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// Create creates a new gun
func (c *GunController) Create(ctx *gin.Context) {
	// Get form values
	name := ctx.PostForm("name")
	acquiredStr := ctx.PostForm("acquired")
	weaponTypeIDStr := ctx.PostForm("weapon_type_id")
	caliberIDStr := ctx.PostForm("caliber_id")
	manufacturerIDStr := ctx.PostForm("manufacturer_id")

	// Parse IDs
	weaponTypeID, err := strconv.ParseUint(weaponTypeIDStr, 10, 64)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid weapon type ID"})
		return
	}

	caliberID, err := strconv.ParseUint(caliberIDStr, 10, 64)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid caliber ID"})
		return
	}

	manufacturerID, err := strconv.ParseUint(manufacturerIDStr, 10, 64)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid manufacturer ID"})
		return
	}

	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to get current user"})
		return
	}

	// Create the gun object
	gun := models.Gun{
		Name:           name,
		WeaponTypeID:   uint(weaponTypeID),
		CaliberID:      uint(caliberID),
		ManufacturerID: uint(manufacturerID),
		OwnerID:        user.ID,
	}

	// Parse and set the acquired date if provided
	if acquiredStr != "" {
		acquired, err := time.Parse("2006-01-02", acquiredStr)
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid acquired date format"})
			return
		}
		gun.Acquired = &acquired
	}

	// Save the gun to the database
	if err := models.CreateGun(c.DB, &gun); err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to create gun"})
		return
	}

	// Redirect to the guns index page
	ctx.Redirect(http.StatusSeeOther, "/owner/guns")
}

// Edit displays the form to edit a gun
func (c *GunController) Edit(ctx *gin.Context) {
	// Get the gun ID from the URL
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid gun ID"})
		return
	}

	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to get current user"})
		return
	}

	// Get the gun
	gunItem, err := models.FindGunByID(c.DB, uint(id), user.ID)
	if err != nil {
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Gun not found"})
		return
	}

	// Get all weapon types, calibers, and manufacturers for the dropdown lists
	var weaponTypes []models.WeaponType
	var calibers []models.Caliber
	var manufacturers []models.Manufacturer

	if err := c.DB.Order("type").Find(&weaponTypes).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve weapon types"})
		return
	}

	if err := c.DB.Order("caliber").Find(&calibers).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve calibers"})
		return
	}

	if err := c.DB.Order("name").Find(&manufacturers).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve manufacturers"})
		return
	}

	// Render the edit template
	component := gun.Edit(*gunItem, weaponTypes, calibers, manufacturers)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// Update updates a gun
func (c *GunController) Update(ctx *gin.Context) {
	// Get the gun ID from the URL
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid gun ID"})
		return
	}

	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to get current user"})
		return
	}

	// Get the gun
	gunItem, err := models.FindGunByID(c.DB, uint(id), user.ID)
	if err != nil {
		ctx.HTML(http.StatusNotFound, "error.html", gin.H{"error": "Gun not found"})
		return
	}

	// Get form values
	name := ctx.PostForm("name")
	acquiredStr := ctx.PostForm("acquired")
	weaponTypeIDStr := ctx.PostForm("weapon_type_id")
	caliberIDStr := ctx.PostForm("caliber_id")
	manufacturerIDStr := ctx.PostForm("manufacturer_id")

	// Parse IDs
	weaponTypeID, err := strconv.ParseUint(weaponTypeIDStr, 10, 64)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid weapon type ID"})
		return
	}

	caliberID, err := strconv.ParseUint(caliberIDStr, 10, 64)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid caliber ID"})
		return
	}

	manufacturerID, err := strconv.ParseUint(manufacturerIDStr, 10, 64)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid manufacturer ID"})
		return
	}

	// Update the gun
	gunItem.Name = name
	gunItem.WeaponTypeID = uint(weaponTypeID)
	gunItem.CaliberID = uint(caliberID)
	gunItem.ManufacturerID = uint(manufacturerID)

	// Parse and set the acquired date if provided
	if acquiredStr != "" {
		acquired, err := time.Parse("2006-01-02", acquiredStr)
		if err != nil {
			ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid acquired date format"})
			return
		}
		gunItem.Acquired = &acquired
	} else {
		gunItem.Acquired = nil
	}

	// Save the gun to the database
	if err := models.UpdateGun(c.DB, gunItem); err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to update gun"})
		return
	}

	// Redirect to the gun details page
	ctx.Redirect(http.StatusSeeOther, fmt.Sprintf("/owner/guns/%d", id))
}

// Delete deletes a gun
func (c *GunController) Delete(ctx *gin.Context) {
	// Get the gun ID from the URL
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		ctx.HTML(http.StatusBadRequest, "error.html", gin.H{"error": "Invalid gun ID"})
		return
	}

	// Get the current user
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to get current user"})
		return
	}

	// Delete the gun
	if err := models.DeleteGun(c.DB, uint(id), user.ID); err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to delete gun"})
		return
	}

	// Redirect to the guns index page
	ctx.Redirect(http.StatusSeeOther, "/owner/guns")
}
