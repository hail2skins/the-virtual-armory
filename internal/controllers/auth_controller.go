package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	authviews "github.com/hail2skins/the-virtual-armory/cmd/web/views/auth"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
)

// AuthController handles authentication-related routes
type AuthController struct {
	Auth *auth.Auth
}

// NewAuthController creates a new AuthController
func NewAuthController(auth *auth.Auth) *AuthController {
	return &AuthController{
		Auth: auth,
	}
}

// Login handles the login page
func (c *AuthController) Login(ctx *gin.Context) {
	component := authviews.LoginForm("", "")
	component.Render(ctx, ctx.Writer)
}

// Register handles the registration page
func (c *AuthController) Register(ctx *gin.Context) {
	component := authviews.RegisterForm()
	component.Render(ctx, ctx.Writer)
}

// Recover handles the password recovery page
func (c *AuthController) Recover(ctx *gin.Context) {
	component := authviews.RecoverForm()
	component.Render(ctx, ctx.Writer)
}

// Profile handles the user profile page
func (c *AuthController) Profile(ctx *gin.Context) {
	// Get the current user from the context
	user, err := auth.GetCurrentUser(ctx)
	if err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to get current user"})
		return
	}

	// Get the user's guns from the database
	db := database.GetDB()
	var guns []models.Gun
	if err := db.Where("owner_id = ?", user.ID).
		Preload("WeaponType").
		Preload("Caliber").
		Preload("Manufacturer").
		Find(&guns).Error; err != nil {
		ctx.HTML(http.StatusInternalServerError, "error.html", gin.H{"error": "Failed to retrieve guns"})
		return
	}

	component := authviews.Profile(*user, guns)
	component.Render(ctx, ctx.Writer)
}

// AdminDashboard handles the admin dashboard page
func (c *AuthController) AdminDashboard(ctx *gin.Context) {
	// Check if the user has admin privileges
	// user, err := c.Auth.CurrentUser(ctx.Request)
	// if err != nil || user == nil {
	//     ctx.AbortWithStatus(http.StatusForbidden)
	//     return
	// }

	component := authviews.AdminDashboard()
	component.Render(ctx, ctx.Writer)
}

// ProcessLogin handles the login form submission
func (c *AuthController) ProcessLogin(ctx *gin.Context) {
	// Get the email from the form
	email := ctx.PostForm("email")

	// This would normally be handled by Authboss
	// For now, we'll simulate a successful login by setting session cookies
	ctx.SetCookie("is_logged_in", "true", 3600, "/", "", false, true)
	ctx.SetCookie("user_email", email, 3600, "/", "", false, true)

	// Redirect to the owner page
	ctx.Redirect(http.StatusSeeOther, "/owner")
}

// ProcessRegister handles the registration form submission
func (c *AuthController) ProcessRegister(ctx *gin.Context) {
	// This would normally be handled by Authboss
	// For now, we'll just redirect to the login page
	ctx.Redirect(http.StatusSeeOther, "/login")
}

// ProcessRecover handles the password recovery form submission
func (c *AuthController) ProcessRecover(ctx *gin.Context) {
	// This would normally be handled by Authboss
	// For now, we'll just redirect to the login page
	ctx.Redirect(http.StatusSeeOther, "/login")
}

// Logout handles user logout
func (c *AuthController) Logout(ctx *gin.Context) {
	// This would normally be handled by Authboss
	// For now, we'll just clear the cookie and redirect to the home page
	ctx.SetCookie("is_logged_in", "", -1, "/", "", false, true)
	ctx.Redirect(http.StatusSeeOther, "/")
}
