package auth

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// SessionName is the cookie name used for the session.
const SessionName = "my_hours_session"

const sessionUserIDKey = "user_id"

// Login stores the authenticated user's ID in the session.
func Login(c *gin.Context, userID uuid.UUID) error {
	s := sessions.Default(c)
	s.Set(sessionUserIDKey, userID.String())
	return s.Save()
}

// Logout clears the session.
func Logout(c *gin.Context) error {
	s := sessions.Default(c)
	s.Clear()
	return s.Save()
}

// CurrentUserID returns the logged-in user's ID, or false if there is no valid session.
func CurrentUserID(c *gin.Context) (uuid.UUID, bool) {
	s := sessions.Default(c)
	raw, ok := s.Get(sessionUserIDKey).(string)
	if !ok || raw == "" {
		return uuid.Nil, false
	}
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
