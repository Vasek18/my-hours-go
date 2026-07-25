package server

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/Vasek18/my_hours/internal/auth"
)

// contextUserIDKey is the gin context key holding the authenticated user's ID.
const contextUserIDKey = "userID"

// requireAuth aborts the request with 401 unless a valid session exists.
func (s *Server) requireAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		id, ok := auth.CurrentUserID(c)
		if !ok {
			respondError(c, http.StatusUnauthorized, "Unauthenticated.")
			c.Abort()
			return
		}
		c.Set(contextUserIDKey, id)
		c.Next()
	}
}
