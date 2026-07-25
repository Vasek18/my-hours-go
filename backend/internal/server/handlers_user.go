package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Vasek18/my_hours/internal/auth"
)

func (s *Server) health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// me returns the currently authenticated user. requireAuth guarantees the
// context holds a valid user ID.
func (s *Server) me(c *gin.Context) {
	id := c.MustGet(contextUserIDKey).(uuid.UUID)

	user, err := s.q.GetUserByID(c, id)
	if err != nil {
		// Session points at a user that no longer exists; clear it.
		_ = auth.Logout(c)
		respondError(c, http.StatusUnauthorized, "Unauthenticated.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": toUserDTO(user)})
}
