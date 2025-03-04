package auth

import (
	"net/http"
	"time"

	"github.com/volatiletech/authboss/v3"
)

// CookieStorer is a cookie storer for Authboss
type CookieStorer struct{}

// NewCookieStorer creates a new CookieStorer
func NewCookieStorer() *CookieStorer {
	return &CookieStorer{}
}

// ReadState reads the cookie state
func (c *CookieStorer) ReadState(r *http.Request) (authboss.ClientState, error) {
	// In a real application, you would read the cookie from the request
	// For now, we'll use a simple in-memory store
	cookie := &Cookie{
		data: make(map[string]string),
	}

	// Get the remember me cookie
	rememberCookie, err := r.Cookie("remember")
	if err == nil {
		cookie.data["remember"] = rememberCookie.Value
	}

	return cookie, nil
}

// WriteState writes the cookie state
func (c *CookieStorer) WriteState(w http.ResponseWriter, state authboss.ClientState, events []authboss.ClientStateEvent) error {
	// In a real application, you would write the cookie to the response
	// For now, we'll use a simple in-memory store
	cookie := state.(*Cookie)

	// Set the remember me cookie
	if remember, ok := cookie.Get("remember"); ok {
		http.SetCookie(w, &http.Cookie{
			Name:     "remember",
			Value:    remember,
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(30 * 24 * time.Hour), // 30 days
		})
	}

	return nil
}

// Cookie is a cookie for Authboss
type Cookie struct {
	data map[string]string
}

// Get gets a value from the cookie
func (c *Cookie) Get(key string) (string, bool) {
	val, ok := c.data[key]
	return val, ok
}

// Set sets a value in the cookie
func (c *Cookie) Set(key, value string) {
	c.data[key] = value
}

// Delete deletes a value from the cookie
func (c *Cookie) Delete(key string) {
	delete(c.data, key)
}
