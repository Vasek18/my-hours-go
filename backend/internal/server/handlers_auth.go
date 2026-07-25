package server

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/Vasek18/my_hours/internal/auth"
	"github.com/Vasek18/my_hours/internal/db"
)

const resetTokenTTL = time.Hour

type registerRequest struct {
	Name                 string `json:"name"`
	Email                string `json:"email"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

func (s *Server) register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	v := newValidator()
	name := v.required("name", req.Name, "Enter your name.")
	email := v.email("email", req.Email)
	v.minLen("password", req.Password, 8, "Password must be at least 8 characters.")
	v.match("password_confirmation", req.Password, req.PasswordConfirmation, "Passwords do not match.")
	if !v.ok() {
		respondValidation(c, v.errors)
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not create account.")
		return
	}

	user, err := s.q.CreateUser(c, db.CreateUserParams{
		Name:         name,
		Email:        email,
		PasswordHash: hash,
	})
	if err != nil {
		if isUniqueViolation(err) {
			respondValidation(c, map[string]string{"email": "This email is already registered."})
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not create account.")
		return
	}

	if err := auth.Login(c, user.ID); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not start session.")
		return
	}

	c.JSON(http.StatusCreated, gin.H{"user": toUserDTO(user)})
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (s *Server) login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	const invalid = "These credentials do not match our records."

	user, err := s.q.GetUserByEmail(c, req.Email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusUnauthorized, invalid)
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not sign in.")
		return
	}

	if !auth.CheckPassword(user.PasswordHash, req.Password) {
		respondError(c, http.StatusUnauthorized, invalid)
		return
	}

	if err := auth.Login(c, user.ID); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not start session.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"user": toUserDTO(user)})
}

func (s *Server) logout(c *gin.Context) {
	if err := auth.Logout(c); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not sign out.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Signed out."})
}

type forgotPasswordRequest struct {
	Email string `json:"email"`
}

func (s *Server) forgotPassword(c *gin.Context) {
	var req forgotPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	v := newValidator()
	email := v.email("email", req.Email)
	if !v.ok() {
		respondValidation(c, v.errors)
		return
	}

	// Generic response regardless of whether the email exists, to avoid leaking
	// which addresses are registered.
	const generic = "If that email address is in our system, we've sent a reset link."

	user, err := s.q.GetUserByEmail(c, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusOK, gin.H{"message": generic})
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not process request.")
		return
	}

	// Invalidate any previous tokens, then issue a fresh one.
	if err := s.q.DeletePasswordResetTokensForUser(c, user.ID); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not process request.")
		return
	}

	token, hash, err := auth.GenerateToken()
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not process request.")
		return
	}

	if _, err := s.q.CreatePasswordResetToken(c, db.CreatePasswordResetTokenParams{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: time.Now().Add(resetTokenTTL),
	}); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not process request.")
		return
	}

	resetURL := s.cfg.AppURL + "/reset-password?token=" + token
	if err := s.mailer.SendPasswordReset(c, user.Email, resetURL); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not send reset email.")
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": generic})
}

type resetPasswordRequest struct {
	Token                string `json:"token"`
	Password             string `json:"password"`
	PasswordConfirmation string `json:"password_confirmation"`
}

func (s *Server) resetPassword(c *gin.Context) {
	var req resetPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	v := newValidator()
	v.required("token", req.Token, "This reset link is invalid.")
	v.minLen("password", req.Password, 8, "Password must be at least 8 characters.")
	v.match("password_confirmation", req.Password, req.PasswordConfirmation, "Passwords do not match.")
	if !v.ok() {
		respondValidation(c, v.errors)
		return
	}

	const invalidToken = "This password reset link is invalid or has expired."

	row, err := s.q.GetPasswordResetTokenByHash(c, auth.HashToken(req.Token))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondValidation(c, map[string]string{"token": invalidToken})
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not reset password.")
		return
	}

	if time.Now().After(row.ExpiresAt) {
		_ = s.q.DeletePasswordResetToken(c, row.ID)
		respondValidation(c, map[string]string{"token": invalidToken})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not reset password.")
		return
	}

	if err := s.q.UpdateUserPassword(c, db.UpdateUserPasswordParams{
		ID:           row.UserID,
		PasswordHash: hash,
	}); err != nil {
		respondError(c, http.StatusInternalServerError, "Could not reset password.")
		return
	}

	_ = s.q.DeletePasswordResetToken(c, row.ID)

	c.JSON(http.StatusOK, gin.H{"message": "Your password has been reset. You can now sign in."})
}

// isUniqueViolation reports whether err is a Postgres unique-constraint error.
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
