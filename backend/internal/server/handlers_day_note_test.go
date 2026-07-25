package server

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/Vasek18/my_hours/internal/db"
)

func (f *fakeQuerier) UpsertDayNote(_ context.Context, a db.UpsertDayNoteParams) (db.DayNote, error) {
	return f.upsertDayNote(a)
}
func (f *fakeQuerier) DeleteDayNote(_ context.Context, a db.DeleteDayNoteParams) (int64, error) {
	return f.deleteDayNote(a)
}

func dateParam(v string) gin.Params {
	return gin.Params{{Key: "date", Value: v}}
}

func TestUpsertDayNote_SavesForCurrentUser(t *testing.T) {
	var captured db.UpsertDayNoteParams
	q := &fakeQuerier{
		upsertDayNote: func(a db.UpsertDayNoteParams) (db.DayNote, error) {
			captured = a
			return db.DayNote{UserID: a.UserID, NoteDate: a.NoteDate, Content: a.Content}, nil
		},
	}
	c, w := authedCtx(http.MethodPut, `{"content":"  slept well  "}`, dateParam("2026-07-08"))
	newServer(q).upsertDayNote(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if captured.UserID != testUserID {
		t.Errorf("owner = %v, want session user", captured.UserID)
	}
	if captured.Content != "slept well" {
		t.Errorf("content = %q, want trimmed", captured.Content)
	}
	if !captured.NoteDate.Equal(time.Date(2026, 7, 8, 0, 0, 0, 0, time.UTC)) {
		t.Errorf("date = %v, want 2026-07-08", captured.NoteDate)
	}
}

func TestUpsertDayNote_EmptyDeletes(t *testing.T) {
	deleted := false
	q := &fakeQuerier{
		deleteDayNote: func(db.DeleteDayNoteParams) (int64, error) {
			deleted = true
			return 1, nil
		},
	}
	c, w := authedCtx(http.MethodPut, `{"content":"   "}`, dateParam("2026-07-08"))
	newServer(q).upsertDayNote(c)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !deleted {
		t.Error("expected empty content to delete the note")
	}
}

func TestUpsertDayNote_BadDate(t *testing.T) {
	c, w := authedCtx(http.MethodPut, `{"content":"x"}`, dateParam("08-07-2026"))
	newServer(&fakeQuerier{}).upsertDayNote(c)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}
