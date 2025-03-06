package controllers

import (
	"crypto/rand"
	"encoding/hex"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	authviews "github.com/hail2skins/the-virtual-armory/cmd/web/views/auth"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/config"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/services/email"
	"golang.org/x/crypto/bcrypt"
)

// AuthController handles authentication-related routes
type AuthController struct {
	Auth         *auth.Auth
	EmailService email.EmailService
	config       *config.Config
}

// NewAuthController creates a new AuthController
func NewAuthController(auth *auth.Auth, emailService email.EmailService, config *config.Config) *AuthController {
	return &AuthController{
		Auth:         auth,
		EmailService: emailService,
		config:       config,
	}
}

// Login handles the login page
func (c *AuthController) Login(ctx *gin.Context) {
	// Check if the verified query parameter is present
	verified := ctx.Query("verified")
	if verified == "true" {
		// Use the LoginFormWithVerified template
		component := authviews.LoginFormWithVerified("", "")
		component.Render(ctx, ctx.Writer)
		return
	}

	component := authviews.LoginForm("", "")
	component.Render(ctx, ctx.Writer)
}

// Register handles the registration page
func (c *AuthController) Register(ctx *gin.Context) {
	component := authviews.RegisterForm("", "")
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
	password := ctx.PostForm("password")

	// Validate form data
	if email == "" || password == "" {
		component := authviews.LoginForm("Email and password are required", email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Get the database connection
	db := database.GetDB()

	// Find the user by email
	var user models.User
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		component := authviews.LoginForm("Invalid email or password", email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Check if the user is confirmed
	if !user.Confirmed {
		component := authviews.LoginForm("Please verify your email before logging in", email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Compare the hashed password
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		component := authviews.LoginForm("Invalid email or password", email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// This would normally be handled by Authboss
	// For now, we'll simulate a successful login by setting session cookies
	ctx.SetCookie("is_logged_in", "true", 3600, "/", "", false, true)
	ctx.SetCookie("user_email", email, 3600, "/", "", false, true)

	// Redirect to the owner page
	ctx.Redirect(http.StatusSeeOther, "/owner")
}

// ProcessRegister handles the registration form submission
func (c *AuthController) ProcessRegister(ctx *gin.Context) {
	// Get form data
	email := ctx.PostForm("email")
	password := ctx.PostForm("password")
	confirmPassword := ctx.PostForm("confirm_password")

	// Validate form data
	if email == "" || password == "" || confirmPassword == "" {
		component := authviews.RegisterForm("All fields are required", email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Validate email format
	if !strings.Contains(email, "@") {
		component := authviews.RegisterForm("Invalid email format", email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Validate password match
	if password != confirmPassword {
		component := authviews.RegisterForm("Passwords do not match", email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Validate password strength
	if len(password) < 8 {
		component := authviews.RegisterForm("Password must be at least 8 characters long", email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Get the database connection
	// In tests, we'll use the test database
	db := database.GetDB()

	// Check if email already exists
	var existingUser models.User
	result := db.Where("email = ?", email).First(&existingUser)
	if result.Error == nil {
		component := authviews.RegisterForm("Email already registered", email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Generate confirmation token
	token, err := generateToken(32)
	if err != nil {
		component := authviews.RegisterForm("Error creating user: "+err.Error(), email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Set token expiry (24 hours from now)
	tokenExpiry := time.Now().Add(24 * time.Hour)

	// Hash the password using bcrypt
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		component := authviews.RegisterForm("Error creating user: "+err.Error(), email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Create new user
	user := models.User{
		Email:              email,
		Password:           string(hashedPassword),
		ConfirmToken:       token,
		ConfirmTokenExpiry: tokenExpiry,
		Confirmed:          false,
	}

	result = db.Create(&user)
	if result.Error != nil {
		component := authviews.RegisterForm("Error creating user: "+result.Error.Error(), email)
		component.Render(ctx, ctx.Writer)
		return
	}

	// Send verification email
	if c.EmailService != nil && c.EmailService.IsConfigured() {
		err = c.EmailService.SendVerificationEmail(email, token)
		if err != nil {
			// Log the error but don't fail the registration
			log.Printf("Error sending verification email: %v", err)
		} else {
			log.Printf("Verification email sent to %s", email)
		}
	}

	// Redirect to a verification pending page
	ctx.Redirect(http.StatusSeeOther, "/verification-pending")
}

// VerifyEmail handles email verification
func (c *AuthController) VerifyEmail(ctx *gin.Context) {
	// Get token from URL parameter
	token := ctx.Param("token")
	if token == "" {
		component := authviews.Error("Invalid verification token")
		component.Render(ctx, ctx.Writer)
		return
	}

	// Find the user with this verification token
	var user models.User
	db := database.GetDB()
	result := db.Where("confirm_token = ?", token).First(&user)
	if result.Error != nil {
		log.Printf("Error finding user by verification token: %v", result.Error)
		component := authviews.Error("Invalid or expired verification token. Please request a new verification email.")
		component.Render(ctx, ctx.Writer)
		return
	}

	// Check if token is expired
	if user.ConfirmTokenExpiry.Before(time.Now()) {
		log.Printf("Verification token expired for user %s", user.Email)
		component := authviews.Error("Your verification token has expired. Please request a new verification email.")
		component.Render(ctx, ctx.Writer)
		return
	}

	// Mark user as verified
	user.Confirmed = true
	user.ConfirmToken = ""
	user.ConfirmTokenExpiry = time.Time{}

	err := db.Save(&user).Error
	if err != nil {
		log.Printf("Error updating user verification status: %v", err)
		component := authviews.Error("An error occurred while verifying your email. Please try again later.")
		component.Render(ctx, ctx.Writer)
		return
	}

	// Redirect to login page with success message
	ctx.Redirect(http.StatusSeeOther, "/login?verified=true")
}

// VerificationPending shows a page indicating that verification is pending
func (c *AuthController) VerificationPending(ctx *gin.Context) {
	component := authviews.VerificationPending()
	component.Render(ctx, ctx.Writer)
}

// ResendVerification handles resending the verification email
func (c *AuthController) ResendVerification(ctx *gin.Context) {
	// Get email from form
	email := ctx.PostForm("email")
	if email == "" {
		component := authviews.Error("Email is required")
		component.Render(ctx, ctx.Writer)
		return
	}

	// Find the user by email
	var user models.User
	db := database.GetDB()
	result := db.Where("email = ?", email).First(&user)
	if result.Error != nil {
		log.Printf("Error finding user by email: %v", result.Error)
		component := authviews.Error("User not found")
		component.Render(ctx, ctx.Writer)
		return
	}

	// Check if the user is already confirmed
	if user.Confirmed {
		component := authviews.Error("Your email is already verified")
		component.Render(ctx, ctx.Writer)
		return
	}

	// Generate a new token
	token, err := generateToken(32)
	if err != nil {
		log.Printf("Error generating token: %v", err)
		component := authviews.Error("An error occurred. Please try again later.")
		component.Render(ctx, ctx.Writer)
		return
	}

	// Update the user with the new token
	user.ConfirmToken = token
	user.ConfirmTokenExpiry = time.Now().Add(24 * time.Hour)
	err = db.Save(&user).Error
	if err != nil {
		log.Printf("Error updating user with new token: %v", err)
		component := authviews.Error("An error occurred. Please try again later.")
		component.Render(ctx, ctx.Writer)
		return
	}

	// Send the verification email
	err = c.EmailService.SendVerificationEmail(user.Email, token)
	if err != nil {
		log.Printf("Error sending verification email: %v", err)
		component := authviews.Error("An error occurred while sending the verification email. Please try again later.")
		component.Render(ctx, ctx.Writer)
		return
	}

	// Redirect to verification pending page
	ctx.Redirect(http.StatusSeeOther, "/verification-pending")
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
	// For now, we'll just clear the cookies and redirect to the home page with a flash message
	ctx.SetCookie("is_logged_in", "", -1, "/", "", false, true)
	ctx.SetCookie("user_email", "", -1, "/", "", false, true)

	// Set a flash message cookie
	ctx.SetCookie("flash_message", "You have been successfully logged out", 5, "/", "", false, false)
	ctx.SetCookie("flash_type", "success", 5, "/", "", false, false)

	ctx.Redirect(http.StatusSeeOther, "/")
}

// Helper function to generate a random token
func generateToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}
