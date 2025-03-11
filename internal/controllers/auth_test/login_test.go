package auth

import (
	"net/http"
	"net/url"
	"testing"
	"time"

	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

func TestLoginLastAttemptUpdate(t *testing.T) {
	// Setup
	router, db, authController, _ := SetupTestRouter(t)
	defer CleanupTest(db)

	// Set the test database for this test
	database.TestDB = db

	// Register the route
	router.POST("/login", authController.ProcessLogin)

	// Create a test user with a confirmed status
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte("password123"), bcrypt.DefaultCost)
	require.NoError(t, err)

	user := models.User{
		Email:     "test-login@example.com",
		Password:  string(hashedPassword),
		Confirmed: true,
	}
	result := db.Create(&user)
	require.NoError(t, result.Error)
	require.NotZero(t, user.ID)

	// Create login form data
	form := url.Values{}
	form.Add("email", "test-login@example.com")
	form.Add("password", "password123")

	// Create a test request
	req, w := CreateFormRequest("POST", "/login", form)

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert response
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/owner", w.Header().Get("Location"))

	// Check that the LastAttempt field was updated
	var updatedUser models.User
	result = db.First(&updatedUser, user.ID)
	require.NoError(t, result.Error)

	assert.False(t, updatedUser.LastAttempt.IsZero(), "LastAttempt should not be zero")
	assert.WithinDuration(t, time.Now(), updatedUser.LastAttempt, 5*time.Second, "LastAttempt should be set to a recent time")
}
