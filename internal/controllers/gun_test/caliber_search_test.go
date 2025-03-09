package gun_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/controllers"
	"github.com/hail2skins/the-virtual-armory/internal/database/seed"
	"github.com/hail2skins/the-virtual-armory/internal/models"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

// setupTestDB sets up a test database
func setupTestDB(t *testing.T) *gorm.DB {
	// Set up a test database
	db, err := testutils.SetupTestDB()
	assert.NoError(t, err)

	// Seed the database with calibers
	seed.SeedCalibers(db)

	return db
}

// TestCaliberSearch tests that searching for calibers works correctly
func TestCaliberSearch(t *testing.T) {
	// Set up test database
	db := setupTestDB(t)
	defer testutils.CleanupTestDB(db)

	// Set up the router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Set up the gun controller
	gunController := controllers.NewGunController(db)

	// Set up the route for searching calibers
	router.GET("/api/calibers/search", gunController.SearchCalibers)

	// Test cases
	testCases := []struct {
		name           string
		query          string
		expectedStatus int
		expectedCount  int
	}{
		{
			name:           "Search for 45 should find .45 ACP",
			query:          "45",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Search for .45 should find .45 ACP",
			query:          ".45",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
		{
			name:           "Search for 9 should find 9mm Parabellum",
			query:          "9",
			expectedStatus: http.StatusOK,
			expectedCount:  1,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test request
			req, _ := http.NewRequest("GET", "/api/calibers/search?q="+tc.query, nil)
			resp := httptest.NewRecorder()

			// Serve the request
			router.ServeHTTP(resp, req)

			// Assert the response status
			assert.Equal(t, tc.expectedStatus, resp.Code)

			// Parse the response body
			var response struct {
				Calibers []models.Caliber `json:"calibers"`
			}
			err := json.NewDecoder(resp.Body).Decode(&response)
			assert.NoError(t, err)

			// Assert that the response contains the expected number of calibers
			assert.Equal(t, tc.expectedCount, len(response.Calibers))
		})
	}
}
