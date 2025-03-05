package auth

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/authboss/v3"
	"github.com/volatiletech/authboss/v3/defaults"
)

// Auth is a wrapper around Authboss
type Auth struct {
	*authboss.Authboss
}

// New creates a new Auth instance
func New() (*Auth, error) {
	ab := authboss.New()

	// Set up the storer
	ab.Config.Storage.Server = NewGORMStorer()
	ab.Config.Storage.SessionState = NewSessionStorer()
	ab.Config.Storage.CookieState = NewCookieStorer()

	// Set up the renderers
	ab.Config.Core.ViewRenderer = defaults.JSONRenderer{}
	ab.Config.Core.MailRenderer = defaults.JSONRenderer{}

	// Set up the root URL
	ab.Config.Paths.RootURL = fmt.Sprintf("http://%s:%s", os.Getenv("HOST"), os.Getenv("PORT"))

	// Set up the mount path
	ab.Config.Paths.Mount = "/auth"

	// Set up the session settings
	ab.Config.Modules.RegisterPreserveFields = []string{"email"}

	// Set up the password settings
	ab.Config.Modules.TwoFactorEmailAuthRequired = false // No 2FA for now
	ab.Config.Modules.TOTP2FAIssuer = "TheVirtualArmory"

	// Set up the mailer
	ab.Config.Mail.From = "no-reply@example.com"
	ab.Config.Mail.FromName = "The Virtual Armory"
	ab.Config.Mail.SubjectPrefix = "[The Virtual Armory] "

	// Set up the modules - enable all core modules
	defaults.SetCore(&ab.Config, true, false)

	// Explicitly ensure register module is enabled
	log.Println("Ensuring register module is enabled")

	// Initialize Authboss
	if err := ab.Init(); err != nil {
		return nil, err
	}

	// Log initialization with more details
	log.Println("Authboss initialized with core modules")
	log.Printf("Mount path: %s", ab.Config.Paths.Mount)

	return &Auth{ab}, nil
}

// LoadAndSave is a middleware that loads and saves the session
func (a *Auth) LoadAndSave(next http.Handler) http.Handler {
	return a.LoadClientStateMiddleware(next)
}

// Middleware returns a Gin middleware that handles Authboss requests
func (a *Auth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip processing for static files and favicon
		if path == "/favicon.ico" ||
			strings.HasPrefix(path, "/assets/") ||
			strings.HasPrefix(path, "/styles/") ||
			strings.HasPrefix(path, "/static/") {
			c.Next()
			return
		}

		// Debug log with more details
		log.Printf("Authboss middleware handling path: %s %s (User-Agent: %s)",
			c.Request.Method,
			path,
			c.Request.UserAgent())

		// Skip Authboss middleware for non-Authboss paths
		if len(path) >= 5 && path[:5] == "/auth" {
			log.Printf("Handling Authboss path: %s %s", c.Request.Method, path)
			a.LoadClientStateMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Handle the request with Authboss
				log.Printf("Forwarding to Authboss router: %s %s", r.Method, r.URL.Path)
				a.Config.Core.Router.ServeHTTP(w, r)
			})).ServeHTTP(c.Writer, c.Request)
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAuth is a middleware that requires authentication
func (a *Auth) RequireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for the is_logged_in cookie first (our simplified auth)
		cookie, err := c.Cookie("is_logged_in")
		if err == nil && cookie == "true" {
			// User is logged in via cookie
			c.Next()
			return
		}

		// Fall back to Authboss authentication
		user, err := a.CurrentUser(c.Request)
		if err != nil || user == nil {
			// Redirect to login page
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}
		c.Next()
	}
}

// RequireAdmin is a middleware that requires admin privileges
func (a *Auth) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check for the is_logged_in cookie first (our simplified auth)
		cookie, err := c.Cookie("is_logged_in")
		if err == nil && cookie == "true" {
			// For now, assume all logged-in users are admins
			c.Next()
			return
		}

		// Fall back to Authboss authentication
		user, err := a.CurrentUser(c.Request)
		if err != nil || user == nil {
			// Redirect to login page
			c.Redirect(http.StatusFound, "/login")
			c.Abort()
			return
		}

		// Check if the user is an admin
		if !user.(*UserWrapper).IsAdmin() {
			// Return forbidden
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	}
}
