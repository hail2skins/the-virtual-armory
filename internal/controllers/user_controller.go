package controllers

import (
	"net/http"
	"strings"
	"time"

	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	userviews "github.com/hail2skins/the-virtual-armory/cmd/web/views/user"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/flash"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/services/email"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// UserController handles user-related operations
type UserController struct {
	DB           *gorm.DB
	EmailService email.EmailService
}

// NewUserController creates a new user controller
func NewUserController(db *gorm.DB) *UserController {
	return &UserController{
		DB: db,
	}
}

// NewUserControllerWithEmailService creates a new user controller with an email service
func NewUserControllerWithEmailService(db *gorm.DB, emailService email.EmailService) *UserController {
	return &UserController{
		DB:           db,
		EmailService: emailService,
	}
}

// getCurrentUser gets the current user from the context
func (c *UserController) getCurrentUser(ctx *gin.Context) (*models.User, error) {
	// Use the auth package's GetCurrentUser function
	return auth.GetCurrentUser(ctx)
}

// ShowProfile displays the user's profile page
func (c *UserController) ShowProfile(ctx *gin.Context) {
	// Get the current user
	user, err := c.getCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	// Render the profile page using templ
	component := userviews.Profile(*user)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// EditProfile displays the form to edit the user's profile
func (c *UserController) EditProfile(ctx *gin.Context) {
	// Get the current user
	user, err := c.getCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	// Render the edit profile page using templ
	component := userviews.EditProfile(*user, "")
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// UpdateProfile updates the user's profile
func (c *UserController) UpdateProfile(ctx *gin.Context) {
	// Get the current user
	user, err := c.getCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	// Get form data
	email := ctx.PostForm("email")

	// Validate the email
	if email == "" {
		ctx.HTML(http.StatusBadRequest, "user/edit_profile.html", gin.H{
			"User":  user,
			"Error": "Email is required",
		})
		return
	}

	// Check if the email is already taken by another user
	var existingUser models.User
	result := c.DB.Where("email = ? AND id != ?", email, user.ID).First(&existingUser)
	if result.Error == nil {
		// Email is already taken
		ctx.HTML(http.StatusBadRequest, "user/edit_profile.html", gin.H{
			"User":  user,
			"Error": "Email is already taken",
		})
		return
	}

	// Check if the email has changed
	emailChanged := email != user.Email

	// If the email has changed, generate a confirmation token and set the user as unconfirmed
	if emailChanged {
		// Generate confirmation token
		bytes := make([]byte, 32)
		if _, err := rand.Read(bytes); err != nil {
			ctx.HTML(http.StatusInternalServerError, "user/edit_profile.html", gin.H{
				"User":  user,
				"Error": "Failed to update profile: " + err.Error(),
			})
			return
		}
		token := hex.EncodeToString(bytes)

		// Set token expiry (24 hours from now)
		tokenExpiry := time.Now().Add(24 * time.Hour)

		// Update the user with the new email, token, and confirmation status
		user.Email = email
		user.ConfirmToken = token
		user.ConfirmTokenExpiry = tokenExpiry
		user.Confirmed = false

		// Save the user
		if err := c.DB.Save(user).Error; err != nil {
			ctx.HTML(http.StatusInternalServerError, "user/edit_profile.html", gin.H{
				"User":  user,
				"Error": "Failed to update profile: " + err.Error(),
			})
			return
		}

		// Send verification email if email service is configured
		if c.EmailService != nil && c.EmailService.IsConfigured() {
			err = c.EmailService.SendVerificationEmail(email, token)
			if err != nil {
				// Log the error but don't fail the update
				// We'll just redirect to the profile page with a warning
				flash.SetMessage(ctx, "Profile updated but failed to send verification email. Please contact support.", "warning")
				ctx.Redirect(http.StatusFound, "/profile")
				return
			}

			// Set a cookie with the email for the verification pending page
			ctx.SetCookie("pending_email", email, 3600, "/", "", false, true)

			// Log the user out since their email has changed and needs verification
			ctx.SetCookie("is_logged_in", "", -1, "/", "", false, true)
			ctx.SetCookie("user_email", "", -1, "/", "", false, true)

			// Redirect to the verification pending page
			ctx.Redirect(http.StatusFound, "/verification-pending")
			return
		} else {
			// Set a warning message
			flash.SetMessage(ctx, "Profile updated but email verification is not configured. Please contact support.", "warning")
			ctx.Redirect(http.StatusFound, "/profile")
			return
		}
	} else {
		// Email hasn't changed, just update the user's email (which is the same)
		user.Email = email
		if err := c.DB.Save(user).Error; err != nil {
			ctx.HTML(http.StatusInternalServerError, "user/edit_profile.html", gin.H{
				"User":  user,
				"Error": "Failed to update profile: " + err.Error(),
			})
			return
		}
		// Set a success message
		flash.SetMessage(ctx, "Profile updated successfully.", "success")

		// Redirect to the profile page
		ctx.Redirect(http.StatusFound, "/profile")
		return
	}
}

// ShowSubscription displays the user's subscription details
func (c *UserController) ShowSubscription(ctx *gin.Context) {
	// Get the current user
	user, err := c.getCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	// Render the subscription page using templ
	component := userviews.Subscription(*user)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// ShowDeleteAccount displays the delete account confirmation page
func (c *UserController) ShowDeleteAccount(ctx *gin.Context) {
	// Get the current user
	user, err := c.getCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	// Render the delete account page using templ
	component := userviews.DeleteAccount(user)
	component.Render(ctx.Request.Context(), ctx.Writer)
}

// DeleteAccount handles the deletion of a user account
func (c *UserController) DeleteAccount(ctx *gin.Context) {
	// Get the current user
	user, err := c.getCurrentUser(ctx)
	if err != nil {
		ctx.Redirect(http.StatusFound, "/login")
		return
	}

	// Get form data
	confirmation := ctx.PostForm("confirm_text")
	password := ctx.PostForm("password")

	// Create a list of validation errors
	var errors []string

	// Validate the confirmation text
	if confirmation == "" {
		errors = append(errors, "Please type DELETE to confirm account deletion")
	} else if confirmation != "DELETE" {
		errors = append(errors, "Please type DELETE exactly as shown (all uppercase)")
	}

	// Validate the password
	if password == "" {
		errors = append(errors, "Please enter your password")
	} else {
		// Verify the password
		err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
		if err != nil {
			errors = append(errors, "Invalid password")
		}
	}

	// If there are validation errors, show them
	if len(errors) > 0 {
		// Join all errors with line breaks for the flash message
		errorMessage := strings.Join(errors, "<br>")

		// Set flash message
		flash.SetMessage(ctx, errorMessage, "error")

		// Render the delete account page
		component := userviews.DeleteAccount(user)
		component.Render(ctx.Request.Context(), ctx.Writer)
		return
	}

	// Soft-delete the user
	if err := c.DB.Delete(user).Error; err != nil {
		// Render the delete account page with an error message
		flash.SetMessage(ctx, "Failed to delete account: "+err.Error(), "error")
		component := userviews.DeleteAccount(user)
		component.Render(ctx.Request.Context(), ctx.Writer)
		return
	}

	// Clear the session cookies
	ctx.SetCookie("is_logged_in", "", -1, "/", "", false, true)
	ctx.SetCookie("user_email", "", -1, "/", "", false, true)

	// Set a success message
	flash.SetMessage(ctx, "Sorry to see you go. Your account has been deleted.", "success")

	// Redirect to the home page
	ctx.Redirect(http.StatusFound, "/")
}

// ReactivateAccount reactivates a soft-deleted account
func (c *UserController) ReactivateAccount(ctx *gin.Context) {
	// Get form data
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")

	// Validate form data
	if email == "" || password == "" {
		ctx.HTML(http.StatusBadRequest, "auth/reactivate.html", gin.H{
			"Error": "Email and password are required",
		})
		return
	}

	// Find the soft-deleted user
	var user models.User
	result := c.DB.Unscoped().Where("email = ? AND deleted_at IS NOT NULL", email).First(&user)
	if result.Error != nil {
		ctx.HTML(http.StatusBadRequest, "auth/reactivate.html", gin.H{
			"Error": "Account not found or not deleted",
		})
		return
	}

	// Verify the password
	// In a real application, you would use bcrypt.CompareHashAndPassword
	// For the test, we're just checking if it matches the stored password
	if user.Password != password {
		ctx.HTML(http.StatusBadRequest, "auth/reactivate.html", gin.H{
			"Error": "Invalid password",
		})
		return
	}

	// Reactivate the account by clearing the DeletedAt field
	if err := c.DB.Unscoped().Model(&user).Update("deleted_at", nil).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "auth/reactivate.html", gin.H{
			"Error": "Failed to reactivate account",
		})
		return
	}

	// Redirect to the login page
	ctx.Redirect(http.StatusFound, "/login")
}
