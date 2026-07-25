package server

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Vasek18/my_hours/internal/auth"
	"github.com/Vasek18/my_hours/internal/db"
)

type updateProfileRequest struct {
	Name string `json:"name"`
}

// updateProfile changes the authenticated user's name.
func (s *Server) updateProfile(c *gin.Context) {
	id := c.MustGet(contextUserIDKey).(uuid.UUID)

	var req updateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	v := newValidator()
	name := v.required("name", req.Name, "Enter your name.")
	if !v.ok() {
		respondValidation(c, v.errors)
		return
	}

	user, err := s.q.UpdateUserName(c, db.UpdateUserNameParams{ID: id, Name: name})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not update your profile.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": toUserDTO(user)})
}

type changePasswordRequest struct {
	CurrentPassword      string `json:"current_password"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

// changePassword updates the authenticated user's password after verifying their
// current password.
func (s *Server) changePassword(c *gin.Context) {
	id := c.MustGet(contextUserIDKey).(uuid.UUID)

	var req changePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	v := newValidator()
	v.required("current_password", req.CurrentPassword, "Enter your current password.")
	v.minLen("password", req.Password, 8, "Password must be at least 8 characters.")
	v.match("password_confirmation", req.Password, req.PasswordConfirmation, "Passwords do not match.")
	if !v.ok() {
		respondValidation(c, v.errors)
		return
	}

	user, err := s.q.GetUserByID(c, id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not update your password.")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		respondValidation(c, map[string]string{"current_password": "That password is incorrect."})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not update your password.")
		return
	}

	if err := s.q.UpdateUserPassword(c, db.UpdateUserPasswordParams{ID: id, PasswordHash: hash}); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not update your password.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Your password has been updated."})
}
