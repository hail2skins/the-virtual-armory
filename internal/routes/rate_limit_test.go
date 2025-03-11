package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/auth"
	"github.com/hail2skins/the-virtual-armory/internal/config"
	"github.com/hail2skins/the-virtual-armory/internal/database"
	"github.com/hail2skins/the-virtual-armory/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter(t *testing.T) (*gin.Engine, *auth.Auth) {
	gin.SetMode(gin.TestMode)

	// Set required environment variables for testing
	os.Setenv("APP_ENV", "test")
	os.Setenv("STRIPE_WEBHOOK_SECRET", "test_webhook_secret")

	// Setup test database
	db, err := testutils.SetupTestDB()
	if err != nil {
		t.Fatalf("Failed to setup test database: %v", err)
	}

	// Setup config
	cfg := &config.Config{
		Environment: "test",
		Port:        8080,
	}

	// Setup auth
	authInstance, err := auth.New()
	if err != nil {
		t.Fatalf("Failed to create auth: %v", err)
	}

	// Create router
	router := gin.New()

	// Register routes (this should include rate limiting)
	RegisterRoutes(router, authInstance, db, cfg)

	return router, authInstance
}

func TestRouteRateLimiting(t *testing.T) {
	router, _ := setupTestRouter(t)
	defer testutils.CleanupTestDB(database.TestDB)

	// Create a test webhook event
	testEvent := map[string]interface{}{
		"id":      "evt_test_123",
		"type":    "checkout.session.completed",
		"created": time.Now().Unix(),
		"data": map[string]interface{}{
			"object": map[string]interface{}{
				"id":     "cs_test_123",
				"object": "checkout.session",
			},
		},
	}
	eventJSON, _ := json.Marshal(testEvent)

	tests := []struct {
		name          string
		path          string
		method        string
		requestCount  int
		expectedCodes []int
		userAgent     string
		body          []byte
		headers       map[string]string
	}{
		{
			name:          "Login Rate Limit",
			path:          "/login",
			method:        "POST",
			requestCount:  6,
			expectedCodes: []int{200, 200, 200, 200, 200, 429}, // 5 successful, 1 rate limited
		},
		{
			name:          "Password Reset Rate Limit",
			path:          "/recover",
			method:        "POST",
			requestCount:  4,
			expectedCodes: []int{200, 200, 200, 429}, // 3 successful, 1 rate limited
		},
		{
			name:          "Webhook Rate Limit - Non-Stripe",
			path:          "/webhook",
			method:        "POST",
			requestCount:  11,
			expectedCodes: []int{200, 200, 200, 200, 200, 200, 200, 200, 200, 200, 429}, // 10 successful, 1 rate limited
			userAgent:     "TestAgent/1.0",
			body:          eventJSON,
			headers: map[string]string{
				"Stripe-Signature": "test_signature",
			},
		},
		{
			name:          "Webhook No Rate Limit - Stripe",
			path:          "/webhook",
			method:        "POST",
			requestCount:  15,
			expectedCodes: make([]int, 15), // All should be 200
			userAgent:     "Stripe/1.0 (+https://stripe.com/docs/webhooks)",
			body:          eventJSON,
			headers: map[string]string{
				"Stripe-Signature": "test_signature",
			},
		},
	}

	// Initialize all expected codes for Stripe test case
	for i := range tests[3].expectedCodes {
		tests[3].expectedCodes[i] = 200
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i := 0; i < tt.requestCount; i++ {
				w := httptest.NewRecorder()
				var req *http.Request
				if tt.body != nil {
					req, _ = http.NewRequest(tt.method, tt.path, bytes.NewBuffer(tt.body))
				} else {
					req, _ = http.NewRequest(tt.method, tt.path, nil)
				}
				req.RemoteAddr = "192.168.1.1:12345" // Same IP for all requests

				if tt.userAgent != "" {
					req.Header.Set("User-Agent", tt.userAgent)
				}

				// Set additional headers if provided
				if tt.headers != nil {
					for key, value := range tt.headers {
						req.Header.Set(key, value)
					}
				}

				router.ServeHTTP(w, req)
				assert.Equal(t, tt.expectedCodes[i], w.Code,
					"Request %d: expected status %d but got %d",
					i+1, tt.expectedCodes[i], w.Code)

				// Small delay to ensure request timing is distinct
				time.Sleep(time.Millisecond)
			}
		})
	}
}
