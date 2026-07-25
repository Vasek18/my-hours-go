package server

import (
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/Vasek18/my_hours/internal/db"
)

func TestCreateActivityType_Validation(t *testing.T) {
	cases := []struct {
		name  string
		body  string
		field string
	}{
		{"missing name", `{"name":"","color":"#4f46e5"}`, "name"},
		{"bad color", `{"name":"Work","color":"blue"}`, "color"},
		{"sort out of range", `{"name":"Work","color":"#4f46e5","sort":-1}`, "sort"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, w := authedCtx(http.MethodPost, tc.body, nil)
			newServer(&fakeQuerier{}).createActivityType(c)

			if w.Code != http.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want 422", w.Code)
			}
			if fieldErrors(t, w)[tc.field] == "" {
				t.Errorf("expected error on %q, got %v", tc.field, fieldErrors(t, w))
			}
		})
	}
}

func TestCreateActivityType_Success(t *testing.T) {
	var captured db.CreateActivityTypeParams
	q := &fakeQuerier{
		createType: func(a db.CreateActivityTypeParams) (db.ActivityType, error) {
			captured = a
			return db.ActivityType{ID: uuid.New(), UserID: a.UserID, Name: a.Name, Color: a.Color, Sort: a.Sort}, nil
		},
	}
	c, w := authedCtx(http.MethodPost, `{"name":"Work","color":"#4f46e5"}`, nil)
	newServer(q).createActivityType(c)

	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	if captured.UserID != testUserID {
		t.Errorf("owner = %v, want session user", captured.UserID)
	}
	if captured.Sort != defaultActivityTypeSort {
		t.Errorf("sort = %d, want default %d", captured.Sort, defaultActivityTypeSort)
	}
}

func TestDeleteActivityType(t *testing.T) {
	cases := []struct {
		affected int64
		want     int
	}{{0, http.StatusNotFound}, {1, http.StatusNoContent}}
	for _, tc := range cases {
		q := &fakeQuerier{
			deleteType: func(db.DeleteActivityTypeParams) (int64, error) { return tc.affected, nil },
		}
		c, _ := authedCtx(http.MethodDelete, "", gin.Params{{Key: "id", Value: uuid.New().String()}})
		newServer(q).deleteActivityType(c)

		if c.Writer.Status() != tc.want {
			t.Fatalf("affected=%d status = %d, want %d", tc.affected, c.Writer.Status(), tc.want)
		}
	}
}
