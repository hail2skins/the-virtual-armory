package controllers

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/cmd/web/views/gun"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/flash"
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
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get current user"})
		return
	}

	// Get all guns for the current user
	var guns []models.Gun
	if err := c.DB.Preload("WeaponType").Preload("Caliber").Preload("Manufacturer").Where("owner_id = ?", user.ID).Find(&guns).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to get guns"})
		return
	}

	// Check if the user has more guns than they can view
	if !user.HasActiveSubscription() && len(guns) > 2 {
		// Set the flag to show the limited view message
		if len(guns) > 0 {
			guns[0].HasMoreGuns = true
			guns[0].TotalGuns = len(guns)
		}
		// Limit to only 2 guns
		guns = guns[:2]
	}

	// Get flash messages from cookies
	flashMessage, _ := ctx.Cookie("flash_message")
	flashType, _ := ctx.Cookie("flash_type")

	// Clear flash cookies if they exist
	flash.ClearMessage(ctx)

	// Render the index template
	component := gun.Index(guns, user, flashMessage, flashType)
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

	// Get flash messages from cookies
	flashMessage, _ := ctx.Cookie("flash_message")
	flashType, _ := ctx.Cookie("flash_type")

	// Clear flash cookies if they exist
	flash.ClearMessage(ctx)

	// Render the show template with empty flash messages if none exist
	component := gun.Show(*gunItem, flashMessage, flashType)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// New displays the form to create a new gun
func (c *GunController) New(ctx *gin.Context) {
	// Get all weapon types, calibers, and manufacturers for the dropdown lists
	var weaponTypes []models.WeaponType
	var calibers []models.Caliber
	var manufacturers []models.Manufacturer

	// Get weapon types sorted by popularity (descending) and then alphabetically
	if err := c.DB.Order("popularity DESC, type").Find(&weaponTypes).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve weapon types"})
		return
	}

	// Get calibers sorted by popularity (descending) and then alphabetically
	if err := c.DB.Order("popularity DESC, caliber").Find(&calibers).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve calibers"})
		return
	}

	// Get manufacturers sorted by popularity (descending) and then alphabetically
	if err := c.DB.Order("popularity DESC, name").Find(&manufacturers).Error; err != nil {
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
	weaponTypeIDStr := ctx.PostForm("weapon_type_id")
	caliberIDStr := ctx.PostForm("caliber_id")
	manufacturerIDStr := ctx.PostForm("manufacturer_id")
	acquiredStr := ctx.PostForm("acquired")

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

	// Check if the user is on the free tier and already has 2 guns
	if user.SubscriptionTier == "free" {
		var count int64
		c.DB.Model(&models.Gun{}).Where("owner_id = ?", user.ID).Count(&count)

		// If the user already has 2 guns, redirect to the pricing page
		if count >= 2 {
			flash.SetMessage(ctx, "You've reached the limit of 2 guns for the free tier. Please upgrade your subscription to add more guns.", "error")
			ctx.Redirect(http.StatusSeeOther, "/pricing")
			return
		}
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

	// Get weapon types sorted by popularity (descending) and then alphabetically
	if err := c.DB.Order("popularity DESC, type").Find(&weaponTypes).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve weapon types"})
		return
	}

	// Get calibers sorted by popularity (descending) and then alphabetically
	if err := c.DB.Order("popularity DESC, caliber").Find(&calibers).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve calibers"})
		return
	}

	// Get manufacturers sorted by popularity (descending) and then alphabetically
	if err := c.DB.Order("popularity DESC, name").Find(&manufacturers).Error; err != nil {
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
	weaponTypeIDStr := ctx.PostForm("weapon_type_id")
	caliberIDStr := ctx.PostForm("caliber_id")
	manufacturerIDStr := ctx.PostForm("manufacturer_id")
	acquiredStr := ctx.PostForm("acquired")

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

func (c *GunController) SearchCalibers(ctx *gin.Context) {
	query := ctx.Query("q")
	var calibers []models.Caliber

	// If query is empty, return calibers sorted by popularity
	if query == "" {
		// Get calibers sorted by popularity (descending) and then alphabetically
		if err := c.DB.Order("popularity DESC, caliber").Limit(15).Find(&calibers).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		ctx.JSON(http.StatusOK, gin.H{"calibers": calibers})
		return
	}

	// First try exact match on caliber or nickname
	if err := c.DB.Where("caliber = ? OR nickname = ?", query, query).Find(&calibers).Error; err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// If no exact matches, try more specific matching
	if len(calibers) == 0 {
		// Try to match with a more flexible search
		if err := c.DB.Where("caliber LIKE ? OR nickname = ?", query+"%", query).Find(&calibers).Error; err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// If still no matches, try even more flexible matching
		if len(calibers) == 0 {
			// Special case for common calibers
			if query == "45" || query == ".45" {
				// For "45" or ".45", specifically match "45 ACP"
				if err := c.DB.Where("caliber = ?", "45 ACP").Find(&calibers).Error; err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			} else if query == "9" {
				// For "9", specifically match "9mm Parabellum"
				if err := c.DB.Where("caliber = ?", "9mm Parabellum").Find(&calibers).Error; err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			} else {
				// For other cases, use more general matching
				if err := c.DB.Where("caliber LIKE ? OR nickname LIKE ?", "%"+query+"%", "%"+query+"%").
					Order("popularity DESC, caliber").Limit(10).Find(&calibers).Error; err != nil {
					ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
					return
				}
			}
		}
	}

	ctx.JSON(http.StatusOK, gin.H{"calibers": calibers})
}
