package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/Vasek18/my_hours/internal/db"
)

// fakeQuerier implements db.Querier by embedding it (nil): only the methods a
// test exercises are overridden; any other call panics, keeping tests honest.
type fakeQuerier struct {
	db.Querier
	getActivityType func(db.GetActivityTypeParams) (db.ActivityType, error)
	createActivity  func(db.CreateActivityParams) (db.Activity, error)
	getActivity     func(db.GetActivityParams) (db.Activity, error)
	deleteActivity  func(db.DeleteActivityParams) (int64, error)
	createType      func(db.CreateActivityTypeParams) (db.ActivityType, error)
	deleteType      func(db.DeleteActivityTypeParams) (int64, error)
	upsertDayNote   func(db.UpsertDayNoteParams) (db.DayNote, error)
	deleteDayNote   func(db.DeleteDayNoteParams) (int64, error)
}

func (f *fakeQuerier) GetActivityType(_ context.Context, a db.GetActivityTypeParams) (db.ActivityType, error) {
	return f.getActivityType(a)
}
func (f *fakeQuerier) CreateActivity(_ context.Context, a db.CreateActivityParams) (db.Activity, error) {
	return f.createActivity(a)
}
func (f *fakeQuerier) GetActivity(_ context.Context, a db.GetActivityParams) (db.Activity, error) {
	return f.getActivity(a)
}
func (f *fakeQuerier) DeleteActivity(_ context.Context, a db.DeleteActivityParams) (int64, error) {
	return f.deleteActivity(a)
}
func (f *fakeQuerier) CreateActivityType(_ context.Context, a db.CreateActivityTypeParams) (db.ActivityType, error) {
	return f.createType(a)
}
func (f *fakeQuerier) DeleteActivityType(_ context.Context, a db.DeleteActivityTypeParams) (int64, error) {
	return f.deleteType(a)
}

var testUserID = uuid.MustParse("11111111-1111-1111-1111-111111111111")

func newServer(q db.Querier) *Server {
	return &Server{q: q}
}

// authedCtx builds a gin context with the auth middleware's user id already set.
func authedCtx(method, body string, params gin.Params) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, "/", strings.NewReader(body))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Set(contextUserIDKey, testUserID)
	c.Params = params
	return c, w
}

func fieldErrors(t *testing.T, w *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var body struct {
		Errors map[string]string `json:"errors"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	return body.Errors
}

const validActivityBody = `{"description":"work","start_time":"2026-07-08T09:00:00Z","end_time":"2026-07-08T10:00:00Z"}`

func TestCreateActivity_Validation(t *testing.T) {
	c, w := authedCtx(http.MethodPost, `{"description":"","start_time":"2026-07-08T10:00:00Z","end_time":"2026-07-08T09:00:00Z"}`, nil)
	newServer(&fakeQuerier{}).createActivity(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", w.Code)
	}
	errs := fieldErrors(t, w)
	if errs["description"] == "" {
		t.Errorf("expected description error, got %v", errs)
	}
}

func TestCreateActivity_ForeignTypeRejected(t *testing.T) {
	q := &fakeQuerier{
		getActivityType: func(db.GetActivityTypeParams) (db.ActivityType, error) {
			return db.ActivityType{}, pgx.ErrNoRows // not this user's type
		},
	}
	body := `{"description":"work","start_time":"2026-07-08T09:00:00Z","end_time":"2026-07-08T10:00:00Z","type_id":"22222222-2222-2222-2222-222222222222"}`
	c, w := authedCtx(http.MethodPost, body, nil)
	newServer(q).createActivity(c)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", w.Code)
	}
	if got := fieldErrors(t, w)["type_id"]; got != "Unknown activity type." {
		t.Errorf("type_id error = %q, want rejection", got)
	}
}

func TestCreateActivity_CapitalizesAndScopesToUser(t *testing.T) {
	var captured db.CreateActivityParams
	q := &fakeQuerier{
		createActivity: func(a db.CreateActivityParams) (db.Activity, error) {
			captured = a
			return db.Activity{ID: uuid.New(), UserID: a.UserID, Description: a.Description}, nil
		},
	}
	c, w := authedCtx(http.MethodPost, validActivityBody, nil)
	newServer(q).createActivity(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	if captured.Description != "Work" {
		t.Errorf("description = %q, want capitalized %q", captured.Description, "Work")
	}
	if captured.UserID != testUserID {
		t.Errorf("owner = %v, want session user", captured.UserID)
	}
}

func TestGetActivity_NotFoundForForeignRow(t *testing.T) {
	q := &fakeQuerier{
		getActivity: func(db.GetActivityParams) (db.Activity, error) {
			return db.Activity{}, pgx.ErrNoRows
		},
	}
	id := uuid.New().String()
	c, w := authedCtx(http.MethodGet, "", gin.Params{{Key: "id", Value: id}})
	newServer(q).getActivity(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestGetActivity_BadIDIs404(t *testing.T) {
	c, w := authedCtx(http.MethodGet, "", gin.Params{{Key: "id", Value: "not-a-uuid"}})
	newServer(&fakeQuerier{}).getActivity(c)

	if w.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", w.Code)
	}
}

func TestDeleteActivity(t *testing.T) {
	cases := []struct {
		name     string
		affected int64
		want     int
	}{
		{"missing row", 0, http.StatusNotFound},
		{"deleted", 1, http.StatusNoContent},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			q := &fakeQuerier{
				deleteActivity: func(db.DeleteActivityParams) (int64, error) { return tc.affected, nil },
			}
			id := uuid.New().String()
			c, _ := authedCtx(http.MethodDelete, "", gin.Params{{Key: "id", Value: id}})
			newServer(q).deleteActivity(c)

			if c.Writer.Status() != tc.want {
				t.Fatalf("status = %d, want %d", c.Writer.Status(), tc.want)
			}
		})
	}
}

func TestCapitalizeFirst(t *testing.T) {
	cases := map[string]string{"work": "Work", "Work": "Work", "": "", "é": "É", "a b": "A b"}
	for in, want := range cases {
		if got := capitalizeFirst(in); got != want {
			t.Errorf("capitalizeFirst(%q) = %q, want %q", in, got, want)
		}
	}
}
