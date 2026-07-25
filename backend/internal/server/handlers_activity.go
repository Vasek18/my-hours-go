package server

import (
	"errors"
	"net/http"
	"time"
	"unicode"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Vasek18/my_hours/internal/db"
)

type activityRequest struct {
	Description string    `json:"description"`
	StartTime   time.Time `json:"start_time"`
	EndTime     time.Time `json:"end_time"`
	TypeID      *string   `json:"type_id"`
}

// listActivities returns the user's activities whose start falls in [from, to).
func (s *Server) listActivities(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)

	from, errFrom := time.Parse(time.RFC3339, c.Query("from"))
	to, errTo := time.Parse(time.RFC3339, c.Query("to"))
	if errFrom != nil || errTo != nil {
		respondError(c, http.StatusBadRequest, "Invalid or missing time range.")
		return
	}

	rows, err := s.q.ListActivitiesInRange(c, db.ListActivitiesInRangeParams{
		UserID:   userID,
		FromTime: from,
		ToTime:   to,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not load your activities.")
		return
	}

	dtos := make([]activityDTO, 0, len(rows))
	for _, row := range rows {
		dtos = append(dtos, toActivityDTO(row))
	}
	c.JSON(http.StatusOK, gin.H{"activities": dtos})
}

func (s *Server) createActivity(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)

	var req activityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	fields, ok := s.resolveActivity(c, userID, req)
	if !ok {
		return
	}

	activity, err := s.q.CreateActivity(c, db.CreateActivityParams{
		UserID:      userID,
		TypeID:      fields.typeID,
		Description: fields.description,
		StartTime:   fields.start,
		EndTime:     fields.end,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not create the activity.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"activity": toActivityDTO(activity)})
}

func (s *Server) getActivity(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)
	id, ok := parseIDParam(c, "Activity not found.")
	if !ok {
		return
	}

	activity, err := s.q.GetActivity(c, db.GetActivityParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "Activity not found.")
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not load the activity.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"activity": toActivityDTO(activity)})
}

func (s *Server) updateActivity(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)
	id, ok := parseIDParam(c, "Activity not found.")
	if !ok {
		return
	}

	var req activityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	fields, ok := s.resolveActivity(c, userID, req)
	if !ok {
		return
	}

	activity, err := s.q.UpdateActivity(c, db.UpdateActivityParams{
		ID:          id,
		UserID:      userID,
		TypeID:      fields.typeID,
		Description: fields.description,
		StartTime:   fields.start,
		EndTime:     fields.end,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "Activity not found.")
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not update the activity.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"activity": toActivityDTO(activity)})
}

func (s *Server) deleteActivity(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)
	id, ok := parseIDParam(c, "Activity not found.")
	if !ok {
		return
	}

	affected, err := s.q.DeleteActivity(c, db.DeleteActivityParams{ID: id, UserID: userID})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not delete the activity.")
		return
	}
	if affected == 0 {
		respondError(c, http.StatusNotFound, "Activity not found.")
		return
	}
	c.Status(http.StatusNoContent)
}

// activityFields holds a validated create/update request.
type activityFields struct {
	description string
	start, end  time.Time
	typeID      uuid.NullUUID
}

// resolveActivity validates the request and returns the cleaned fields. Any
// supplied type_id is verified to belong to the current user, so another user's
// type is rejected without disclosing its existence. On failure it writes the
// response and returns ok=false.
func (s *Server) resolveActivity(c *gin.Context, userID uuid.UUID, req activityRequest) (activityFields, bool) {
	v := newValidator()
	description := capitalizeFirst(v.required("description", req.Description, "Enter a description."))
	v.maxLen("description", description, 200, "Description is too long.")

	switch {
	case req.StartTime.IsZero() || req.EndTime.IsZero():
		v.fail("start_time", "Start and end time are required.")
	case !req.EndTime.After(req.StartTime):
		v.fail("end_time", "End time must be after start time.")
	}

	typeID, ok := s.resolveOwnedType(c, userID, req.TypeID, v)
	if !ok {
		return activityFields{}, false
	}

	if !v.ok() {
		respondValidation(c, v.errors)
		return activityFields{}, false
	}
	return activityFields{description, req.StartTime, req.EndTime, typeID}, true
}

// resolveOwnedType parses an optional type_id and confirms it belongs to userID.
// ok=false means an internal error was already reported; an unknown/foreign type
// is recorded as a field error on v instead.
func (s *Server) resolveOwnedType(c *gin.Context, userID uuid.UUID, raw *string, v *validator) (uuid.NullUUID, bool) {
	if raw == nil || *raw == "" {
		return uuid.NullUUID{}, true
	}
	id, err := uuid.Parse(*raw)
	if err != nil {
		v.fail("type_id", "Unknown activity type.")
		return uuid.NullUUID{}, true
	}
	if _, err := s.q.GetActivityType(c, db.GetActivityTypeParams{ID: id, UserID: userID}); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			v.fail("type_id", "Unknown activity type.")
			return uuid.NullUUID{}, true
		}
		respondError(c, http.StatusInternalServerError, "Could not save the activity.")
		return uuid.NullUUID{}, false
	}
	return uuid.NullUUID{UUID: id, Valid: true}, true
}

// capitalizeFirst upper-cases the first rune ("work" -> "Work").
func capitalizeFirst(s string) string {
	if s == "" {
		return s
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}
