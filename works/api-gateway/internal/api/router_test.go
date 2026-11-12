package api

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/json"
	"encoding/pem"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hieudoanm/neo.bank/works/api-gateway/internal/api/middleware"
	"github.com/hieudoanm/neo.bank/works/api-gateway/internal/core/auth"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type testSigner struct {
	key        *rsa.PrivateKey
	keyManager *auth.KeyManager
}

func newTestSigner(t *testing.T) *testSigner {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate key: %v", err)
	}
	der, err := x509.MarshalPKIXPublicKey(&key.PublicKey)
	if err != nil {
		t.Fatalf("marshal public key: %v", err)
	}
	pemBytes := pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
	km, err := auth.NewKeyManager(pemBytes, "kid")
	if err != nil {
		t.Fatalf("new key manager: %v", err)
	}
	return &testSigner{key: key, keyManager: km}
}

func (s *testSigner) token(t *testing.T) string {
	t.Helper()
	claims := &auth.AccessClaims{
		AccountID: "a1",
		Email:     "user@example.com",
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   "u1",
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	signed, err := jwt.NewWithClaims(jwt.SigningMethodRS256, claims).SignedString(s.key)
	if err != nil {
		t.Fatalf("sign token: %v", err)
	}
	return signed
}

// upstream records the last request it received and replies with the configured
// status/body.
type upstream struct {
	server     *httptest.Server
	lastPath   string
	lastMethod string
	status     int
	body       string
}

func newUpstream(t *testing.T) *upstream {
	t.Helper()
	u := &upstream{status: http.StatusOK, body: `{"ok":true}`}
	u.server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u.lastPath = r.URL.Path
		u.lastMethod = r.Method
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(u.status)
		_, _ = w.Write([]byte(u.body))
	}))
	t.Cleanup(u.server.Close)
	return u
}

func newRouter(t *testing.T, limits middleware.RateLimits) (*gin.Engine, *upstream, string) {
	t.Helper()
	signer := newTestSigner(t)
	up := newUpstream(t)
	am := middleware.NewAuthMiddleware(signer.keyManager)
	r := NewRouter(am, up.server.URL, up.server.URL, up.server.URL, up.server.URL, limits, []string{"*"})
	return r, up, signer.token(t)
}

func do(r *gin.Engine, method, path string, headers map[string]string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	// A real server request always carries a cancelable context; without one the
	// reverse proxy probes the deprecated http.CloseNotifier, which the
	// httptest.ResponseRecorder does not implement.
	ctx, cancel := context.WithCancel(req.Context())
	defer cancel()
	req = req.WithContext(ctx)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func TestHealthEndpoints(t *testing.T) {
	r, _, _ := newRouter(t, middleware.RateLimits{})

	for _, path := range []string{"/healthz", "/readyz"} {
		w := do(r, http.MethodGet, path, nil)
		if w.Code != http.StatusOK {
			t.Fatalf("%s status = %d, want 200", path, w.Code)
		}
		var body map[string]string
		if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
			t.Fatalf("%s body: %v", path, err)
		}
		if body["status"] != "ok" {
			t.Fatalf("%s status body = %q, want ok", path, body["status"])
		}
	}
}

func TestRequestIDGeneratedAndEchoed(t *testing.T) {
	r, _, _ := newRouter(t, middleware.RateLimits{})

	generated := do(r, http.MethodGet, "/healthz", nil)
	if generated.Header().Get("X-Request-ID") == "" {
		t.Fatal("expected a generated X-Request-ID header")
	}

	const rid = "abc-123"
	echoed := do(r, http.MethodGet, "/healthz", map[string]string{"X-Request-ID": rid})
	if got := echoed.Header().Get("X-Request-ID"); got != rid {
		t.Fatalf("X-Request-ID = %q, want %q", got, rid)
	}
}

func TestCORSPreflight(t *testing.T) {
	r, _, _ := newRouter(t, middleware.RateLimits{})

	w := do(r, http.MethodOptions, "/api/v1/login", nil)
	if w.Code != http.StatusNoContent {
		t.Fatalf("preflight status = %d, want 204", w.Code)
	}
	if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("Allow-Origin = %q, want *", got)
	}
}

func TestPublicRouteProxiesToAuth(t *testing.T) {
	r, up, _ := newRouter(t, middleware.RateLimits{})

	w := do(r, http.MethodPost, "/api/v1/register", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if up.lastPath != "/api/v1/register" || up.lastMethod != http.MethodPost {
		t.Fatalf("upstream saw %s %s, want POST /api/v1/register", up.lastMethod, up.lastPath)
	}
}

func TestProtectedRouteRequiresAuth(t *testing.T) {
	r, _, _ := newRouter(t, middleware.RateLimits{})

	w := do(r, http.MethodGet, "/api/v1/accounts/a1/balance", nil)
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
	if !strings.Contains(w.Body.String(), "INVALID_CREDENTIALS") {
		t.Fatalf("body = %s, want INVALID_CREDENTIALS", w.Body.String())
	}
}

func TestProtectedRouteForwardsWithValidToken(t *testing.T) {
	r, up, token := newRouter(t, middleware.RateLimits{})

	w := do(r, http.MethodGet, "/api/v1/accounts/a1/balance", map[string]string{
		"Authorization": "Bearer " + token,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", w.Code)
	}
	if up.lastPath != "/api/v1/accounts/a1/balance" {
		t.Fatalf("upstream path = %q", up.lastPath)
	}
}

func TestProtectedRouteRejectsInvalidToken(t *testing.T) {
	r, _, _ := newRouter(t, middleware.RateLimits{})

	w := do(r, http.MethodGet, "/api/v1/accounts/a1/balance", map[string]string{
		"Authorization": "Bearer not-a-token",
	})
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("status = %d, want 401", w.Code)
	}
}

func TestAuthRateLimitEnabled(t *testing.T) {
	r, _, _ := newRouter(t, middleware.RateLimits{AuthPerMinute: 1})

	first := do(r, http.MethodPost, "/api/v1/register", nil)
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d, want 200", first.Code)
	}
	second := do(r, http.MethodPost, "/api/v1/register", nil)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want 429", second.Code)
	}
	if second.Header().Get("Retry-After") == "" {
		t.Fatal("expected Retry-After header")
	}
	if !strings.Contains(second.Body.String(), "RATE_LIMITED") {
		t.Fatalf("body = %s, want RATE_LIMITED", second.Body.String())
	}
}

func TestUserRateLimitEnabled(t *testing.T) {
	r, _, token := newRouter(t, middleware.RateLimits{UserPerMinute: 1})
	headers := map[string]string{"Authorization": "Bearer " + token}

	if got := do(r, http.MethodGet, "/api/v1/accounts/a1/balance", headers).Code; got != http.StatusOK {
		t.Fatalf("first status = %d, want 200", got)
	}
	if got := do(r, http.MethodGet, "/api/v1/accounts/a1/balance", headers).Code; got != http.StatusTooManyRequests {
		t.Fatalf("second status = %d, want 429", got)
	}
}

func TestLoginLockoutAfterRepeatedFailures(t *testing.T) {
	r, up, _ := newRouter(t, middleware.RateLimits{LoginFailures: 1, LoginWindow: time.Minute})
	up.status = http.StatusUnauthorized
	up.body = `{"error":{"code":"INVALID_CREDENTIALS"}}`

	first := do(r, http.MethodPost, "/api/v1/login", nil)
	if first.Code != http.StatusUnauthorized {
		t.Fatalf("first login status = %d, want 401", first.Code)
	}
	second := do(r, http.MethodPost, "/api/v1/login", nil)
	if second.Code != http.StatusTooManyRequests {
		t.Fatalf("second login status = %d, want 429 (locked out)", second.Code)
	}
}

func TestProxyErrorHandlerReturnsBadGateway(t *testing.T) {
	signer := newTestSigner(t)
	am := middleware.NewAuthMiddleware(signer.keyManager)
	// Point at a closed port so the reverse proxy fails to dial upstream.
	r := NewRouter(am, "http://127.0.0.1:1", "http://127.0.0.1:1", "http://127.0.0.1:1", "http://127.0.0.1:1", middleware.RateLimits{}, []string{"*"})

	w := do(r, http.MethodPost, "/api/v1/register", nil)
	if w.Code != http.StatusBadGateway {
		t.Fatalf("status = %d, want 502", w.Code)
	}
	if !strings.Contains(w.Body.String(), "SERVICE_UNAVAILABLE") {
		t.Fatalf("body = %s, want SERVICE_UNAVAILABLE", w.Body.String())
	}
}

func TestNewRouterPanicsOnInvalidServiceURL(t *testing.T) {
	signer := newTestSigner(t)
	am := middleware.NewAuthMiddleware(signer.keyManager)

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for invalid service URL")
		}
	}()
	NewRouter(am, "://invalid", "http://localhost", "http://localhost", "http://localhost", middleware.RateLimits{}, []string{"*"})
}
