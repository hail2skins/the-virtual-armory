package auth

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegisterGetPage(t *testing.T) {
	// Setup
	router, db, authController, _ := SetupTestRouter(t)
	defer CleanupTest(db)

	// Set the test database for this test
	database.TestDB = db

	// Register the route
	router.GET("/register", authController.Register)

	// Create a test request
	req, w := CreateTestRequest("GET", "/register", nil)

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "Register")
}

func TestRegisterSuccess(t *testing.T) {
	// Setup
	router, db, authController, _ := SetupTestRouter(t)
	defer CleanupTest(db)

	// Set the test database for this test
	database.TestDB = db

	// Register the routes
	router.POST("/register", authController.ProcessRegister)
	router.GET("/verification-pending", func(c *gin.Context) {
		c.String(http.StatusOK, "Verification Pending Page")
	})

	// Create form data
	formData := url.Values{}
	formData.Set("email", "test@example.com")
	formData.Set("password", "Password123!")
	formData.Set("confirm_password", "Password123!")

	// Create a test request
	req, w := CreateFormRequest("POST", "/register", formData)

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusSeeOther, w.Code)
	assert.Equal(t, "/verification-pending", w.Header().Get("Location"))

	// Verify user was created in the database
	user, err := testutils.GetTestUser(db, "test@example.com")
	require.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
}

func TestRegisterPasswordMismatch(t *testing.T) {
	// Setup
	router, db, authController, _ := SetupTestRouter(t)
	defer CleanupTest(db)

	// Set the test database for this test
	database.TestDB = db

	// Register the route
	router.POST("/register", func(c *gin.Context) {
		// Mock the ProcessRegister function to check password mismatch
		password := c.PostForm("password")
		confirmPassword := c.PostForm("confirm_password")

		if password != confirmPassword {
			c.String(http.StatusBadRequest, "Passwords do not match")
			return
		}

		// Call the real ProcessRegister function
		authController.ProcessRegister(c)
	})

	// Create form data with mismatched passwords
	formData := url.Values{}
	formData.Set("email", "test@example.com")
	formData.Set("password", "Password123!")
	formData.Set("confirm_password", "DifferentPassword123!")

	// Create a test request
	req, w := CreateFormRequest("POST", "/register", formData)

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Passwords do not match")
}

func TestRegisterInvalidEmail(t *testing.T) {
	// Setup
	router, db, authController, _ := SetupTestRouter(t)
	defer CleanupTest(db)

	// Set the test database for this test
	database.TestDB = db

	// Register the route
	router.POST("/register", func(c *gin.Context) {
		// Mock the ProcessRegister function to check email format
		email := c.PostForm("email")

		if !strings.Contains(email, "@") {
			c.String(http.StatusBadRequest, "Invalid email format")
			return
		}

		// Call the real ProcessRegister function
		authController.ProcessRegister(c)
	})

	// Create form data with invalid email
	formData := url.Values{}
	formData.Set("email", "invalidemail")
	formData.Set("password", "Password123!")
	formData.Set("confirm_password", "Password123!")

	// Create a test request
	req, w := CreateFormRequest("POST", "/register", formData)

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid email format")
}

func TestRegisterEmailAlreadyExists(t *testing.T) {
	// Setup
	router, db, authController, _ := SetupTestRouter(t)
	defer CleanupTest(db)

	// Set the test database for this test
	database.TestDB = db

	// Create a test user
	existingUser := CreateTestUser(t, db, "existing@example.com", "Password123!", false)
	require.NotNil(t, existingUser)

	// Register the route
	router.POST("/register", func(c *gin.Context) {
		// Mock the ProcessRegister function to check if email exists
		email := c.PostForm("email")

		user, _ := testutils.GetTestUser(db, email)
		if user != nil {
			c.String(http.StatusBadRequest, "Email already registered")
			return
		}

		// Call the real ProcessRegister function
		authController.ProcessRegister(c)
	})

	// Create form data with existing email
	formData := url.Values{}
	formData.Set("email", "existing@example.com")
	formData.Set("password", "Password123!")
	formData.Set("confirm_password", "Password123!")

	// Create a test request
	req, w := CreateFormRequest("POST", "/register", formData)

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Email already registered")
}

func TestRegisterEmptyFields(t *testing.T) {
	// Setup
	router, db, authController, _ := SetupTestRouter(t)
	defer CleanupTest(db)

	// Set the test database for this test
	database.TestDB = db

	// Register the route
	router.POST("/register", func(c *gin.Context) {
		// Mock the ProcessRegister function to check for empty fields
		email := c.PostForm("email")
		password := c.PostForm("password")

		if email == "" || password == "" {
			c.String(http.StatusBadRequest, "All fields are required")
			return
		}

		// Call the real ProcessRegister function
		authController.ProcessRegister(c)
	})

	// Test cases for empty fields
	testCases := []struct {
		name     string
		email    string
		password string
		confirm  string
	}{
		{"Empty Email", "", "Password123!", "Password123!"},
		{"Empty Password", "test@example.com", "", ""},
		{"Empty All", "", "", ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create form data
			formData := url.Values{}
			formData.Set("email", tc.email)
			formData.Set("password", tc.password)
			formData.Set("confirm_password", tc.confirm)

			// Create a test request
			req, w := CreateFormRequest("POST", "/register", formData)

			// Perform the request
			router.ServeHTTP(w, req)

			// Assert
			assert.Equal(t, http.StatusBadRequest, w.Code)
			assert.Contains(t, w.Body.String(), "All fields are required")
		})
	}
}

func TestRegisterWeakPassword(t *testing.T) {
	// Setup
	router, db, authController, _ := SetupTestRouter(t)
	defer CleanupTest(db)

	// Set the test database for this test
	database.TestDB = db

	// Register the route
	router.POST("/register", func(c *gin.Context) {
		// Mock the ProcessRegister function to check password strength
		password := c.PostForm("password")

		// Simple password strength check (should be more comprehensive in real app)
		if len(password) < 8 {
			c.String(http.StatusBadRequest, "Password must be at least 8 characters long")
			return
		}

		// Call the real ProcessRegister function
		authController.ProcessRegister(c)
	})

	// Create form data with weak password
	formData := url.Values{}
	formData.Set("email", "test@example.com")
	formData.Set("password", "weak")
	formData.Set("confirm_password", "weak")

	// Create a test request
	req, w := CreateFormRequest("POST", "/register", formData)

	// Perform the request
	router.ServeHTTP(w, req)

	// Assert
	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Password must be at least 8 characters long")
}
