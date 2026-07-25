package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"

	"github.com/Vasek18/my_hours/internal/config"
	"github.com/Vasek18/my_hours/internal/mailer"
)

// newTestRouter builds a router backed by an in-memory session store and a nil
// DB pool. Only handlers that return before touching the database are exercised
// here (validation, health, auth gating); the full DB-backed flow is covered by
// the docker-compose end-to-end check.
func newTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	store := memstore.NewStore([]byte("test-secret"))
	srv := New(config.Config{AppEnv: "dev", AppURL: "http://localhost"}, nil, mailer.ConsoleMailer{})
	return srv.Router(store)
}

func TestHealth(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	newTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	var body map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	if body["status"] != "ok" {
		t.Fatalf("body = %v, want status ok", body)
	}
}

func TestRegisterValidation(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/auth/register",
		strings.NewReader(`{"name":"","email":"not-an-email","password":"short","password_confirmation":"nope"}`))
	req.Header.Set("Content-Type", "application/json")
	newTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want 422", w.Code)
	}
	var body struct {
		Errors map[string]string `json:"errors"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	for _, field := range []string{"name", "email", "password", "password_confirmation"} {
		if _, ok := body.Errors[field]; !ok {
			t.Errorf("expected validation error for %q, got %v", field, body.Errors)
		}
	}
}

func TestMeRequiresAuth(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/auth/me", nil)
	newTestRouter().ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestProfileRoutesRequireAuth(t *testing.T) {
	cases := []struct {
		method string
		path   string
	}{
		{http.MethodPost, "/api/auth/profile"},
		{http.MethodPost, "/api/auth/password/change"},
		{http.MethodPost, "/api/auth/email/change"},
	}
	for _, tc := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(tc.method, tc.path, strings.NewReader(`{}`))
		req.Header.Set("Content-Type", "application/json")
		newTestRouter().ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Errorf("%s %s: status = %d, want 401", tc.method, tc.path, w.Code)
		}
	}
}
