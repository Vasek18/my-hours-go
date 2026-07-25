package server

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Vasek18/my_hours/internal/auth"
	"github.com/Vasek18/my_hours/internal/db"
)

const emailChangeTokenTTL = time.Hour

type changeEmailRequest struct {
	NewEmail        string `json:"new_email"`
	CurrentPassword string `json:"current_password"`
}

// changeEmail begins an email change: it verifies the current password and sends
// a confirmation link to the new address. The email is only updated once that
// link is confirmed (proving the user controls the new inbox).
//
// To avoid disclosing whether an address is already registered, the response is
// always the same generic message regardless of whether the new address is free.
func (s *Server) changeEmail(c *gin.Context) {
	id := c.MustGet(contextUserIDKey).(uuid.UUID)

	var req changeEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	v := newValidator()
	newEmail := v.email("new_email", req.NewEmail)
	v.required("current_password", req.CurrentPassword, "Enter your current password.")
	if !v.ok() {
		respondValidation(c, v.errors)
		return
	}

	user, err := s.q.GetUserByID(c, id)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not update your email.")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.CurrentPassword) {
		respondValidation(c, map[string]string{"current_password": "That password is incorrect."})
		return
	}

	// Comparing against the user's own current address is not a disclosure.
	if strings.EqualFold(newEmail, user.Email) {
		respondValidation(c, map[string]string{"new_email": "This is already your email address."})
		return
	}

	const generic = "If that address is available, we've sent a confirmation link to it."

	// Anti-enumeration: if the address belongs to another account, respond with
	// the same generic message without issuing a usable token.
	switch _, err := s.q.GetUserByEmail(c, newEmail); {
	case err == nil:
		c.JSON(http.StatusOK, gin.H{"message": generic})
		return
	case !errors.Is(err, pgx.ErrNoRows):
		respondError(c, http.StatusInternalServerError, "Could not update your email.")
		return
	}

	// Address is free: invalidate any prior pending change, then issue a fresh one.
	if err := s.q.DeleteEmailChangeTokensForUser(c, id); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not update your email.")
		return
	}

	token, hash, err := auth.GenerateToken()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not update your email.")
		return
	}

	if _, err := s.q.CreateEmailChangeToken(c, db.CreateEmailChangeTokenParams{
		UserID:    id,
		NewEmail:  newEmail,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(emailChangeTokenTTL),
	}); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not update your email.")
		return
	}

	confirmURL := s.cfg.AppURL + "/confirm-email?token=" + token
	if err := s.mailer.SendEmailChangeConfirmation(c, newEmail, confirmURL); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not send confirmation email.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": generic})
}

type confirmEmailRequest struct {
	Token string `json:"token"`
}

// confirmEmailChange validates an email-change token and applies the new address.
// It is unauthenticated so the link works from any browser/session.
func (s *Server) confirmEmailChange(c *gin.Context) {
	var req confirmEmailRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	v := newValidator()
	v.required("token", req.Token, "This confirmation link is invalid.")
	if !v.ok() {
		respondValidation(c, v.errors)
		return
	}

	const invalid = "This email confirmation link is invalid or has expired."

	row, err := s.q.GetEmailChangeTokenByHash(c, auth.HashToken(req.Token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondValidation(c, map[string]string{"token": invalid})
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not confirm your email.")
		return
	}

	if time.Now().After(row.ExpiresAt) {
		_ = s.q.DeleteEmailChangeToken(c, row.ID)
		respondValidation(c, map[string]string{"token": invalid})
		return
	}

	// Re-check availability at confirm time in case the address was taken since.
	switch _, err := s.q.GetUserByEmail(c, row.NewEmail); {
	case err == nil:
		_ = s.q.DeleteEmailChangeToken(c, row.ID)
		respondValidation(c, map[string]string{"token": invalid})
		return
	case !errors.Is(err, pgx.ErrNoRows):
		respondError(c, http.StatusInternalServerError, "Could not confirm your email.")
		return
	}

	updated, err := s.q.UpdateUserEmail(c, db.UpdateUserEmailParams{ID: row.UserID, Email: row.NewEmail})
	if err != nil {
		if isUniqueViolation(err) {
			_ = s.q.DeleteEmailChangeToken(c, row.ID)
			respondValidation(c, map[string]string{"token": invalid})
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not confirm your email.")
		return
	}

	_ = s.q.DeleteEmailChangeTokensForUser(c, row.UserID)

	c.JSON(http.StatusOK, gin.H{
		"message": "Your email address has been updated.",
		"user":    toUserDTO(updated),
	})
}
