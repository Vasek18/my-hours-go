package server

import (
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Vasek18/my_hours/internal/db"
)

// userDTO is the public representation of a user (never exposes the password hash).
type userDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

func toUserDTO(u db.User) userDTO {
	return userDTO{ID: u.ID.String(), Name: u.Name, Email: u.Email}
}

// activityTypeDTO is the public representation of an activity type.
type activityTypeDTO struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
	Sort  int32  `json:"sort"`
}

func toActivityTypeDTO(a db.ActivityType) activityTypeDTO {
	return activityTypeDTO{ID: a.ID.String(), Name: a.Name, Color: a.Color, Sort: a.Sort}
}

// activityDTO is the public representation of an activity. type_id is null when
// the activity has no type; the frontend resolves its color from the type list.
type activityDTO struct {
	ID          string    `json:"id"`
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	TypeID      *string   `json:"type_id"`
}

// dayNoteDTO is the public representation of a diary note for one day.
type dayNoteDTO struct {
	Date    string `json:"date"`
	Content string `json:"content"`
}

func toDayNoteDTO(n db.DayNote) dayNoteDTO {
	return dayNoteDTO{Date: n.NoteDate.Format(isoDate), Content: n.Content}
}

func toActivityDTO(a db.Activity) activityDTO {
	var typeID *string
	if a.TypeID.Valid {
		s := a.TypeID.UUID.String()
		typeID = &s
	}
	return activityDTO{
		ID:          a.ID.String(),
		Description: a.Description,
		StartTime:   a.StartTime,
		EndTime:     a.EndTime,
		TypeID:      typeID,
	}
}

// respondError writes a uniform error body: {"error": "..."}.
func respondError(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{"error": message})
}

// parseIDParam reads the :id path parameter as a UUID. An unparseable id can
// never match an owned row, so it is reported as a 404 (notFound) — returning
// ok=false once the response has been written.
func parseIDParam(c *gin.Context, notFound string) (uuid.UUID, bool) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		respondError(c, http.StatusNotFound, notFound)
		return uuid.Nil, false
	}
	return id, true
}

// respondValidation writes field-level validation errors:
// {"error": "...", "errors": {"field": "message"}}.
func respondValidation(c *gin.Context, fields map[string]string) {
	c.JSON(422, gin.H{"error": "The given data was invalid.", "errors": fields})
}

// validator accumulates per-field validation messages.
type validator struct {
	errors map[string]string
}

func newValidator() *validator {
	return &validator{errors: map[string]string{}}
}

func (v *validator) required(field, value, message string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		v.fail(field, message)
	}
	return value
}

func (v *validator) email(field, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		v.fail(field, "Enter your email address.")
		return value
	}
	if _, err := mail.ParseAddress(value); err != nil {
		v.fail(field, "Enter a valid email address.")
	}
	return value
}

func (v *validator) minLen(field, value string, n int, message string) {
	if len(value) < n {
		v.fail(field, message)
	}
}

func (v *validator) maxLen(field, value string, n int, message string) {
	if len(value) > n {
		v.fail(field, message)
	}
}

// hexColorPattern matches a 6-digit hex color like #4f46e5 (what a native
// <input type="color"> always produces).
var hexColorPattern = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)

func (v *validator) hexColor(field, value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		v.fail(field, "Choose a color.")
		return value
	}
	if !hexColorPattern.MatchString(value) {
		v.fail(field, "Enter a valid hex color.")
	}
	return value
}

func (v *validator) match(field, a, b, message string) {
	if a != b {
		v.fail(field, message)
	}
}

func (v *validator) fail(field, message string) {
	if _, exists := v.errors[field]; !exists {
		v.errors[field] = message
	}
}

func (v *validator) ok() bool {
	return len(v.errors) == 0
}
