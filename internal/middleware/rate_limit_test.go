package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestLoginRateLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewRateLimiter()

	// Create a test router with rate limiting
	router := gin.New()
	router.Use(limiter.RateLimit(5, time.Minute))
	router.POST("/login", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// Test successful requests within limit
	for i := 0; i < 5; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/login", nil)
		req.RemoteAddr = "192.168.1.1:12345" // Simulate same IP
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Test rate limit exceeded
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/login", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestPasswordResetRateLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewRateLimiter()

	// Create a test router with rate limiting
	router := gin.New()
	router.Use(limiter.RateLimit(3, time.Hour))
	router.POST("/password-reset", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// Test successful requests within limit
	for i := 0; i < 3; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/password-reset", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Test rate limit exceeded
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/password-reset", nil)
	req.RemoteAddr = "192.168.1.1:12345"
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusTooManyRequests, w.Code)
}

func TestWebhookRateLimiting(t *testing.T) {
	gin.SetMode(gin.TestMode)
	limiter := NewRateLimiter()

	// Create a test router with rate limiting
	router := gin.New()
	router.Use(limiter.RateLimit(10, time.Minute)) // Allow more requests for webhooks
	router.POST("/webhook", func(c *gin.Context) {
		c.String(http.StatusOK, "success")
	})

	// Test Stripe webhook requests (should bypass rate limiting)
	for i := 0; i < 15; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/webhook", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("User-Agent", "Stripe/1.0 (+https://stripe.com/docs/webhooks)")
		router.ServeHTTP(w, req)
		assert.Equal(t, http.StatusOK, w.Code)
	}

	// Test non-Stripe webhook requests (should be rate limited)
	for i := 0; i < 11; i++ {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/webhook", nil)
		req.RemoteAddr = "192.168.1.1:12345"
		req.Header.Set("User-Agent", "SomeOtherService/1.0")
		router.ServeHTTP(w, req)
		if i < 10 {
			assert.Equal(t, http.StatusOK, w.Code)
		} else {
			assert.Equal(t, http.StatusTooManyRequests, w.Code)
		}
	}
}
