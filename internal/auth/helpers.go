package auth

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
)

// GetCurrentUser retrieves the current user from the context
func GetCurrentUser(ctx *gin.Context) (*models.User, error) {
	// Check for the is_logged_in cookie first (our simplified auth)
	cookie, err := ctx.Cookie("is_logged_in")
	if err == nil && cookie == "true" {
		// Get the database connection
		db := database.GetDB()

		// Get the email from the session
		email, _ := ctx.Cookie("user_email")
		if email == "" {
			// For backward compatibility, use a default email if not set
			email = "admin@example.com"
		}

		// Retrieve the user from the database
		var user models.User
		result := db.Where("email = ?", email).First(&user)
		if result.Error != nil {
			return nil, result.Error
		}

		return &user, nil
	}

	// If we're using the simplified auth and not logged in, return an error
	return nil, errors.New("user not authenticated")
}
