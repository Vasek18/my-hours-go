package server

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Vasek18/my_hours/internal/db"
)

const isoDate = "2006-01-02"

type dayNoteRequest struct {
	Content string `json:"content"`
}

// listDayNotes returns the user's notes whose date falls in [from, to).
func (s *Server) listDayNotes(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)

	from, errFrom := time.Parse(isoDate, c.Query("from"))
	to, errTo := time.Parse(isoDate, c.Query("to"))
	if errFrom != nil || errTo != nil {
		respondError(c, http.StatusBadRequest, "Invalid or missing date range.")
		return
	}

	rows, err := s.q.ListDayNotesInRange(c, db.ListDayNotesInRangeParams{
		UserID:   userID,
		FromDate: from,
		ToDate:   to,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not load your notes.")
		return
	}

	dtos := make([]dayNoteDTO, 0, len(rows))
	for _, row := range rows {
		dtos = append(dtos, toDayNoteDTO(row))
	}
	c.JSON(http.StatusOK, gin.H{"day_notes": dtos})
}

// getDayNote returns the note for one day, or day_note: null when there is none.
func (s *Server) getDayNote(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)
	date, ok := parseDateParam(c)
	if !ok {
		return
	}

	note, err := s.q.GetDayNote(c, db.GetDayNoteParams{UserID: userID, NoteDate: date})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			c.JSON(http.StatusOK, gin.H{"day_note": nil})
			return
		}
		respondError(c, http.StatusInternalServerError, "Could not load the note.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"day_note": toDayNoteDTO(note)})
}

// upsertDayNote creates or replaces a day's note. Clearing the text removes it,
// and the response is day_note: null in that case.
func (s *Server) upsertDayNote(c *gin.Context) {
	userID := c.MustGet(contextUserIDKey).(uuid.UUID)
	date, ok := parseDateParam(c)
	if !ok {
		return
	}

	var req dayNoteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		respondError(c, http.StatusBadRequest, "Invalid request body.")
		return
	}

	content := strings.TrimSpace(req.Content)
	if len(content) > 10000 {
		respondValidation(c, map[string]string{"content": "Note is too long."})
		return
	}

	if content == "" {
		if _, err := s.q.DeleteDayNote(c, db.DeleteDayNoteParams{UserID: userID, NoteDate: date}); err != nil {
			respondError(c, http.StatusInternalServerError, "Could not save the note.")
			return
		}
		c.JSON(http.StatusOK, gin.H{"day_note": nil})
		return
	}

	note, err := s.q.UpsertDayNote(c, db.UpsertDayNoteParams{
		UserID:   userID,
		NoteDate: date,
		Content:  content,
	})
	if err != nil {
		respondError(c, http.StatusInternalServerError, "Could not save the note.")
		return
	}
	c.JSON(http.StatusOK, gin.H{"day_note": toDayNoteDTO(note)})
}

// parseDateParam reads the :date path parameter as a calendar date (YYYY-MM-DD).
func parseDateParam(c *gin.Context) (time.Time, bool) {
	date, err := time.Parse(isoDate, c.Param("date"))
	if err != nil {
		respondError(c, http.StatusBadRequest, "Invalid date.")
		return time.Time{}, false
	}
	return date, true
}
