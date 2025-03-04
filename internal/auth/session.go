package auth

import (
	"crypto/rand"
	"encoding/base64"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/volatiletech/authboss/v3"
)

// SessionData holds the session data
type SessionData struct {
	SessionID string
	UserID    string
}

// SessionStorer is a session storer for Authboss
type SessionStorer struct {
	sessions map[string]SessionData
}

// NewSessionStorer creates a new SessionStorer
func NewSessionStorer() *SessionStorer {
	return &SessionStorer{
		sessions: make(map[string]SessionData),
	}
}

// ReadState reads the session state
func (s *SessionStorer) ReadState(r *http.Request) (authboss.ClientState, error) {
	// In a real application, you would read the session from a store
	// For now, we'll use a simple in-memory store
	session := &Session{
		data: make(map[string]string),
	}

	// Get the session ID from the cookie
	cookie, err := r.Cookie("session")
	if err != nil {
		return session, nil
	}

	// Get the session data
	if sessionData, ok := s.sessions[cookie.Value]; ok {
		session.data["session_id"] = sessionData.SessionID
		session.data["user_id"] = sessionData.UserID
	}

	return session, nil
}

// WriteState writes the session state
func (s *SessionStorer) WriteState(w http.ResponseWriter, state authboss.ClientState, events []authboss.ClientStateEvent) error {
	// In a real application, you would write the session to a store
	// For now, we'll use a simple in-memory store
	session := state.(*Session)

	// Get the session ID
	sessionID, _ := session.Get("session_id")
	if sessionID == "" {
		// Create a new session ID
		sessionID = randomString(32)
		session.Set("session_id", sessionID)

		// Set the cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "session",
			Value:    sessionID,
			Path:     "/",
			HttpOnly: true,
		})
	}

	// Store the session data
	userID, _ := session.Get("user_id")
	s.sessions[sessionID] = SessionData{
		SessionID: sessionID,
		UserID:    userID,
	}

	return nil
}

// Session is a session for Authboss
type Session struct {
	data map[string]string
}

// Get gets a value from the session
func (s *Session) Get(key string) (string, bool) {
	val, ok := s.data[key]
	return val, ok
}

// Set sets a value in the session
func (s *Session) Set(key, value string) {
	s.data[key] = value
}

// Delete deletes a value from the session
func (s *Session) Delete(key string) {
	delete(s.data, key)
}

// randomString generates a random string of the specified length
func randomString(length int) string {
	b := make([]byte, length)
	rand.Read(b)
	return base64.StdEncoding.EncodeToString(b)
}

// ClientStateMiddleware is a middleware that loads and saves the client state
func ClientStateMiddleware(ab *authboss.Authboss) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Process the request
		c.Next()
	}
}
