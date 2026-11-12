package api

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/hieudoanm/neo.bank/works/auth/internal/api/handlers"
	"github.com/hieudoanm/neo.bank/works/auth/internal/services"
)

type stubService struct{}

func (stubService) Register(_ context.Context, _, _ string) (*services.RegisterResult, error) {
	return &services.RegisterResult{UserID: "u", AccountID: "a", AccessToken: "at", RefreshToken: "rt"}, nil
}

func (stubService) Login(_ context.Context, _, _ string) (*services.RegisterResult, error) {
	return &services.RegisterResult{UserID: "u", AccountID: "a", AccessToken: "at", RefreshToken: "rt"}, nil
}

func (stubService) Refresh(_ context.Context, _ string) (*services.RefreshResult, error) {
	return &services.RefreshResult{AccessToken: "at", RefreshToken: "rt"}, nil
}

func (stubService) Logout(_ context.Context, _ string) error { return nil }

func newTestEngine() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return NewRouter(handlers.NewAuthHandler(stubService{}), nil)
}

func TestHealthRoutes(t *testing.T) {
	r := newTestEngine()
	for _, path := range []string{"/healthz", "/readyz"} {
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, path, nil))
		if w.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", path, w.Code)
		}
		if w.Body.String() != `{"status":"ok"}` {
			t.Fatalf("%s body = %s", path, w.Body.String())
		}
	}
}

func TestRequestIDMiddlewareGeneratesAndEchoes(t *testing.T) {
	r := newTestEngine()

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/healthz", nil))
	if w.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected generated X-Request-ID header")
	}

	w = httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("X-Request-ID", "given-id")
	r.ServeHTTP(w, req)
	if w.Header().Get("X-Request-ID") != "given-id" {
		t.Fatalf("X-Request-ID = %q, want given-id", w.Header().Get("X-Request-ID"))
	}
}

func TestCORSPreflight(t *testing.T) {
	r := newTestEngine()
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodOptions, "/api/v1/register", nil)
	r.ServeHTTP(w, req)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}
	if w.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatalf("missing CORS origin header")
	}
}

func TestPublicAuthRoutes(t *testing.T) {
	r := newTestEngine()
	cases := []struct {
		path string
		body string
		want int
	}{
		{"/api/v1/register", `{"email":"u@example.com","password":"Str0ng!Passw0rd"}`, http.StatusCreated},
		{"/api/v1/login", `{"email":"u@example.com","password":"Str0ng!Passw0rd"}`, http.StatusOK},
		{"/api/v1/refresh", `{"refreshToken":"rt"}`, http.StatusOK},
		{"/api/v1/logout", `{"refreshToken":"rt"}`, http.StatusNoContent},
	}
	for _, c := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, c.path, strings.NewReader(c.body))
		req.Header.Set("Content-Type", "application/json")
		r.ServeHTTP(w, req)
		if w.Code != c.want {
			t.Fatalf("%s status = %d, want %d", c.path, w.Code, c.want)
		}
	}
}
