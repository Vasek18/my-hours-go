package server

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Vasek18/my_hours/internal/db"
)

const (
	defaultActivityTypeSort = 500
	maxActivityTypeSort     = 1_000_000 // bounds sort well within int32
)

type activityTypeRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
	Sort  *int   `json:"sort"`
}

func (s *Server) listActivityTypes(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)

	rows, err := s.q.ListActivityTypes(c, userID)
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not load your activity types.")
		return
	}

	dtos := make([]activityTypeDTO, 0, len(rows))
	for _, row := range rows {
		dtos = append(dtos, toActivityTypeDTO(row))
	}
	c.JSON(http.StatusOK, gin.H{"activity_types": dtos})
}

func (s *Server) createActivityType(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)

	var req activityTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	name, color, sort, ok := validateActivityType(c, req)
	if !ok {
		return
	}

	activityType, err := s.q.CreateActivityType(c, db.CreateActivityTypeParams{
		UserID: userID,
		Name:   name,
		Color:  color,
		Sort:   sort,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not create the activity type.")
		return
	}
	c.JSON(http.StatusCreated, gin.H{"activity_type": toActivityTypeDTO(activityType)})
}

func (s *Server) getActivityType(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)
	id, ok := parseIDParam(c, "Activity type not found.")
	if !ok {
		return
	}

	activityType, err := s.q.GetActivityType(c, db.GetActivityTypeParams{ID: id, UserID: userID})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "Activity type not found.")
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not load the activity type.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"activity_type": toActivityTypeDTO(activityType)})
}

func (s *Server) updateActivityType(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)
	id, ok := parseIDParam(c, "Activity type not found.")
	if !ok {
		return
	}

	var req activityTypeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	name, color, sort, ok := validateActivityType(c, req)
	if !ok {
		return
	}

	activityType, err := s.q.UpdateActivityType(c, db.UpdateActivityTypeParams{
		ID:     id,
		UserID: userID,
		Name:   name,
		Color:  color,
		Sort:   sort,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			respondError(c, http.StatusNotFound, "Activity type not found.")
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not update the activity type.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"activity_type": toActivityTypeDTO(activityType)})
}

func (s *Server) deleteActivityType(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)
	id, ok := parseIDParam(c, "Activity type not found.")
	if !ok {
		return
	}

	affected, err := s.q.DeleteActivityType(c, db.DeleteActivityTypeParams{ID: id, UserID: userID})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not delete the activity type.")
		return
	}
	if affected == 0 {
		respondError(c, http.StatusNotFound, "Activity type not found.")
		return
	}
	c.Status(http.StatusNoContent)
}

// validateActivityType returns the cleaned fields, or writes a 422 and ok=false.
func validateActivityType(c *gin.Context, req activityTypeRequest) (name, color string, sort int32, ok bool) {
	v := newValidator()
	name = v.required("name", req.Name, "Enter a name.")
	v.maxLen("name", name, 100, "Name is too long.")
	color = v.hexColor("color", req.Color)

	sort = int32(defaultActivityTypeSort)
	if req.Sort != nil {
		if *req.Sort < 0 || *req.Sort > maxActivityTypeSort {
			v.fail("sort", "Sort must be between 0 and 1,000,000.")
		} else {
			sort = int32(*req.Sort)
		}
	}

	if !v.ok() {
		respondValidation(c, v.errors)
		return "", "", 0, false
	}
	return name, color, sort, true
}
