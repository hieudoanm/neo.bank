package handlers

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	applierrors "github.com/hieudoanm/neo.bank/works/auth/internal/core/errors"
	"github.com/hieudoanm/neo.bank/works/auth/internal/services"
)

type fakeService struct {
	registerRes *services.RegisterResult
	registerErr error
	loginRes    *services.RegisterResult
	loginErr    error
	refreshRes  *services.RefreshResult
	refreshErr  error
	logoutErr   error
}

func (f *fakeService) Register(_ context.Context, _, _ string) (*services.RegisterResult, error) {
	return f.registerRes, f.registerErr
}

func (f *fakeService) Login(_ context.Context, _, _ string) (*services.RegisterResult, error) {
	return f.loginRes, f.loginErr
}

func (f *fakeService) Refresh(_ context.Context, _ string) (*services.RefreshResult, error) {
	return f.refreshRes, f.refreshErr
}

func (f *fakeService) Logout(_ context.Context, _ string) error {
	return f.logoutErr
}

func newRouter(svc AuthService) *gin.Engine {
	gin.SetMode(gin.TestMode)
	h := NewAuthHandler(svc)
	r := gin.New()
	r.POST("/register", h.Register)
	r.POST("/login", h.Login)
	r.POST("/refresh", h.Refresh)
	r.POST("/logout", h.Logout)
	return r
}

func perform(r *gin.Engine, method, path, body string, headers map[string]string) *httptest.ResponseRecorder {
	var reader *strings.Reader
	if body == "" {
		reader = strings.NewReader("")
	} else {
		reader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestRegisterSuccess(t *testing.T) {
	svc := &fakeService{registerRes: &services.RegisterResult{UserID: "u1", AccountID: "a1", AccessToken: "at", RefreshToken: "rt"}}
	w := perform(newRouter(svc), http.MethodPost, "/register", `{"email":"u@example.com","password":"Str0ng!Passw0rd"}`, nil)
	if w.Code != http.StatusCreated {
		t.Fatalf("status = %d, want 201", w.Code)
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if resp["userId"] != "u1" || resp["accountId"] != "a1" || resp["accessToken"] != "at" || resp["tokenType"] != "Bearer" {
		t.Fatalf("unexpected body: %v", resp)
	}
}

func TestRegisterInvalidBody(t *testing.T) {
	w := perform(newRouter(&fakeService{}), http.MethodPost, "/register", `{`, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want 400", w.Code)
	}
}

func TestRegisterMapsAppErrors(t *testing.T) {
	t.Run("with details", func(t *testing.T) {
		svc := &fakeService{registerErr: applierrors.ValidationError(map[string]string{"email": "required"})}
		w := perform(newRouter(svc), http.MethodPost, "/register", `{}`, map[string]string{"X-Request-ID": "rid-1"})
		if w.Code != http.StatusBadRequest {
			t.Fatalf("status = %d, want 400", w.Code)
		}
		if !strings.Contains(w.Body.String(), `"requestId":"rid-1"`) || !strings.Contains(w.Body.String(), `"details"`) {
			t.Fatalf("unexpected body: %s", w.Body.String())
		}
	})

	t.Run("without details", func(t *testing.T) {
		svc := &fakeService{registerErr: applierrors.ErrEmailAlreadyRegistered}
		w := perform(newRouter(svc), http.MethodPost, "/register", `{}`, nil)
		if w.Code != http.StatusConflict {
			t.Fatalf("status = %d, want 409", w.Code)
		}
		if strings.Contains(w.Body.String(), `"details"`) {
			t.Fatalf("did not expect details: %s", w.Body.String())
		}
	})

	t.Run("non app error becomes internal", func(t *testing.T) {
		svc := &fakeService{registerErr: errors.New("boom")}
		w := perform(newRouter(svc), http.MethodPost, "/register", `{}`, nil)
		if w.Code != http.StatusInternalServerError {
			t.Fatalf("status = %d, want 500", w.Code)
		}
		if !strings.Contains(w.Body.String(), "INTERNAL_ERROR") {
			t.Fatalf("unexpected body: %s", w.Body.String())
		}
	})
}

func TestLogin(t *testing.T) {
	svc := &fakeService{loginRes: &services.RegisterResult{UserID: "u1", AccountID: "a1", AccessToken: "at", RefreshToken: "rt"}}
	w := perform(newRouter(svc), http.MethodPost, "/login", `{"email":"u@example.com","password":"Str0ng!Passw0rd"}`, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}

	w = perform(newRouter(&fakeService{}), http.MethodPost, "/login", `{`, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid body status = %d, want 400", w.Code)
	}

	svc = &fakeService{loginErr: applierrors.ErrInvalidCredentials}
	w = perform(newRouter(svc), http.MethodPost, "/login", `{}`, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("error status = %d, want 401", w.Code)
	}
}

func TestRefresh(t *testing.T) {
	svc := &fakeService{refreshRes: &services.RefreshResult{AccessToken: "at2", RefreshToken: "rt2"}}
	w := perform(newRouter(svc), http.MethodPost, "/refresh", `{"refreshToken":"rt"}`, nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if !strings.Contains(w.Body.String(), `"accessToken":"at2"`) {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}

	w = perform(newRouter(&fakeService{}), http.MethodPost, "/refresh", `{`, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid body status = %d, want 400", w.Code)
	}

	svc = &fakeService{refreshErr: applierrors.ErrInvalidRefreshToken}
	w = perform(newRouter(svc), http.MethodPost, "/refresh", `{"refreshToken":"rt"}`, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("error status = %d, want 401", w.Code)
	}
}

func TestLogout(t *testing.T) {
	w := perform(newRouter(&fakeService{}), http.MethodPost, "/logout", `{"refreshToken":"rt"}`, nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want 204", w.Code)
	}

	w = perform(newRouter(&fakeService{}), http.MethodPost, "/logout", `{`, nil)
	if w.Code != http.StatusBadRequest {
		t.Fatalf("invalid body status = %d, want 400", w.Code)
	}

	svc := &fakeService{logoutErr: applierrors.ErrInvalidRefreshToken}
	w = perform(newRouter(svc), http.MethodPost, "/logout", `{"refreshToken":"rt"}`, nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("error status = %d, want 401", w.Code)
	}
}

func TestGetRequestID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	if got := getRequestID(c); got != "" {
		t.Fatalf("getRequestID = %q, want empty", got)
	}
	c.Request.Header.Set("X-Request-ID", "rid")
	if got := getRequestID(c); got != "rid" {
		t.Fatalf("getRequestID = %q, want rid", got)
	}
}
