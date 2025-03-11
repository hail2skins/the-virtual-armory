package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hail2skins/the-virtual-armory/internal/errors"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
	}
}

// isStripeWebhook checks if the request is from Stripe
func isStripeWebhook(c *gin.Context) bool {
	userAgent := c.GetHeader("User-Agent")
	return strings.HasPrefix(userAgent, "Stripe/")
}

// RateLimit creates middleware that limits requests per client
// limit: number of requests allowed
// duration: time window for the limit (e.g., 1 minute)
func (rl *RateLimiter) RateLimit(limit int, duration time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip rate limiting for Stripe webhooks
		if c.FullPath() == "/webhook" && isStripeWebhook(c) {
			c.Next()
			return
		}

		// Use IP address as client identifier
		clientIP := c.ClientIP()

		// Add path to identifier for separate limits per endpoint
		identifier := fmt.Sprintf("%s:%s", clientIP, c.FullPath())

		rl.mu.Lock()
		defer rl.mu.Unlock()

		// Clean old requests
		now := time.Now()
		windowStart := now.Add(-duration)

		// Get existing requests for this client
		times := rl.requests[identifier]
		valid := make([]time.Time, 0)

		// Keep only requests within our time window
		for _, t := range times {
			if t.After(windowStart) {
				valid = append(valid, t)
			}
		}

		// Update requests for this client
		rl.requests[identifier] = valid

		// Check if limit exceeded
		if len(valid) >= limit {
			err := errors.NewValidationError(fmt.Sprintf("Rate limit exceeded. Try again in %v", duration))
			c.AbortWithStatus(http.StatusTooManyRequests)
			c.Error(err)
			return
		}

		// Add current request
		rl.requests[identifier] = append(rl.requests[identifier], now)

		c.Next()
	}
}

// LoginRateLimit creates middleware specifically for login attempts
func (rl *RateLimiter) LoginRateLimit() gin.HandlerFunc {
	return rl.RateLimit(5, time.Minute)
}

// PasswordResetRateLimit creates middleware specifically for password reset attempts
func (rl *RateLimiter) PasswordResetRateLimit() gin.HandlerFunc {
	return rl.RateLimit(3, time.Hour)
}

// WebhookRateLimit creates middleware specifically for webhook endpoints
func (rl *RateLimiter) WebhookRateLimit() gin.HandlerFunc {
	return rl.RateLimit(10, time.Minute)
}
