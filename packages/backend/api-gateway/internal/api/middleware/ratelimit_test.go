package middleware

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type fakeClock struct {
	mu sync.Mutex
	t  time.Time
}

func newFakeClock() *fakeClock {
	return &fakeClock{t: time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)}
}

func (c *fakeClock) now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.t
}

func (c *fakeClock) advance(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.t = c.t.Add(d)
}

func TestRateLimiterAllowsUpToMaxThenDenies(t *testing.T) {
	limiter := NewRateLimiter(3, time.Minute)
	clock := newFakeClock()
	limiter.now = clock.now

	for i := 0; i < 3; i++ {
		if allowed, _ := limiter.Allow("k"); !allowed {
			t.Fatalf("request %d should be allowed", i+1)
		}
	}
	allowed, retry := limiter.Allow("k")
	if allowed {
		t.Fatal("4th request should be denied")
	}
	if retry < time.Second || retry > time.Minute {
		t.Fatalf("retry-after out of range: %v", retry)
	}
}

func TestRateLimiterAllowsAgainAfterWindow(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	clock := newFakeClock()
	limiter.now = clock.now

	if allowed, _ := limiter.Allow("k"); !allowed {
		t.Fatal("first request should be allowed")
	}
	if allowed, _ := limiter.Allow("k"); allowed {
		t.Fatal("second request should be denied")
	}

	clock.advance(time.Minute + time.Second)
	if allowed, _ := limiter.Allow("k"); !allowed {
		t.Fatal("request after the window should be allowed")
	}
}

func TestRateLimiterKeysAreIndependent(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	clock := newFakeClock()
	limiter.now = clock.now

	if allowed, _ := limiter.Allow("a"); !allowed {
		t.Fatal("a's first request should be allowed")
	}
	if allowed, _ := limiter.Allow("a"); allowed {
		t.Fatal("a's second request should be denied")
	}
	if allowed, _ := limiter.Allow("b"); !allowed {
		t.Fatal("b should not be affected by a")
	}
}

func TestFailureTrackerLocksAfterThreshold(t *testing.T) {
	tracker := NewFailureTracker(5, 15*time.Minute)
	clock := newFakeClock()
	tracker.now = clock.now

	for i := 0; i < 4; i++ {
		tracker.Record("ip")
	}
	if locked, _ := tracker.Locked("ip"); locked {
		t.Fatal("should not be locked before the threshold")
	}
	tracker.Record("ip")
	if locked, retry := tracker.Locked("ip"); !locked {
		t.Fatal("should be locked at the threshold")
	} else if retry <= 0 || retry > 15*time.Minute {
		t.Fatalf("retry-after out of range: %v", retry)
	}
}

func TestFailureTrackerUnlocksAfterWindow(t *testing.T) {
	tracker := NewFailureTracker(2, 15*time.Minute)
	clock := newFakeClock()
	tracker.now = clock.now

	tracker.Record("ip")
	tracker.Record("ip")
	if locked, _ := tracker.Locked("ip"); !locked {
		t.Fatal("should be locked")
	}

	clock.advance(15*time.Minute + time.Second)
	if locked, _ := tracker.Locked("ip"); locked {
		t.Fatal("should be unlocked after the window")
	}
}

func newTestEngine(setup func(r *gin.Engine)) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	setup(r)
	return r
}

func perform(r *gin.Engine, method, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(method, path, nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	return rec
}

func TestRateLimitMiddlewareAllowsUnderLimit(t *testing.T) {
	limiter := NewRateLimiter(2, time.Minute)
	r := newTestEngine(func(r *gin.Engine) {
		r.GET("/x", RateLimitMiddleware(limiter, ClientIPKey), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
	})

	if rec := perform(r, http.MethodGet, "/x"); rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestRateLimitMiddlewareReturns429WithRetryAfter(t *testing.T) {
	limiter := NewRateLimiter(1, time.Minute)
	r := newTestEngine(func(r *gin.Engine) {
		r.GET("/x", RateLimitMiddleware(limiter, ClientIPKey), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
	})

	if rec := perform(r, http.MethodGet, "/x"); rec.Code != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", rec.Code)
	}
	rec := perform(r, http.MethodGet, "/x")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429, got %d", rec.Code)
	}
	retry, err := strconv.Atoi(rec.Header().Get("Retry-After"))
	if err != nil || retry < 1 {
		t.Fatalf("expected a positive Retry-After header, got %q", rec.Header().Get("Retry-After"))
	}
	if body := rec.Body.String(); !strings.Contains(body, "RATE_LIMITED") {
		t.Fatalf("expected RATE_LIMITED error envelope, got %s", body)
	}
}

func TestAuthenticatedUserKeyUsesUserID(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("userId", "user-123")
	if got := AuthenticatedUserKey(c); got != "user-123" {
		t.Fatalf("expected user-123, got %q", got)
	}
}

func TestLoginLockoutMiddlewareBlocksAfterFailures(t *testing.T) {
	tracker := NewFailureTracker(3, 15*time.Minute)
	clock := newFakeClock()
	tracker.now = clock.now

	r := newTestEngine(func(r *gin.Engine) {
		r.POST("/login", LoginLockoutMiddleware(tracker), func(c *gin.Context) {
			c.Status(http.StatusUnauthorized)
		})
	})

	for i := 0; i < 3; i++ {
		if rec := perform(r, http.MethodPost, "/login"); rec.Code != http.StatusUnauthorized {
			t.Fatalf("attempt %d: expected 401, got %d", i+1, rec.Code)
		}
	}
	rec := perform(r, http.MethodPost, "/login")
	if rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected 429 after lockout, got %d", rec.Code)
	}
	if rec.Header().Get("Retry-After") == "" {
		t.Fatal("expected a Retry-After header on lockout")
	}
}

func TestLoginLockoutMiddlewareIgnoresSuccessfulLogins(t *testing.T) {
	tracker := NewFailureTracker(2, 15*time.Minute)
	r := newTestEngine(func(r *gin.Engine) {
		r.POST("/login", LoginLockoutMiddleware(tracker), func(c *gin.Context) {
			c.Status(http.StatusOK)
		})
	})

	for i := 0; i < 5; i++ {
		if rec := perform(r, http.MethodPost, "/login"); rec.Code != http.StatusOK {
			t.Fatalf("attempt %d: successful logins must not be throttled, got %d", i+1, rec.Code)
		}
	}
}

func TestLoginLockoutMiddlewareRecoversAfterWindow(t *testing.T) {
	tracker := NewFailureTracker(2, 15*time.Minute)
	clock := newFakeClock()
	tracker.now = clock.now

	r := newTestEngine(func(r *gin.Engine) {
		r.POST("/login", LoginLockoutMiddleware(tracker), func(c *gin.Context) {
			c.Status(http.StatusUnauthorized)
		})
	})

	for i := 0; i < 2; i++ {
		perform(r, http.MethodPost, "/login")
	}
	if rec := perform(r, http.MethodPost, "/login"); rec.Code != http.StatusTooManyRequests {
		t.Fatalf("expected lockout, got %d", rec.Code)
	}

	clock.advance(15*time.Minute + time.Second)
	if rec := perform(r, http.MethodPost, "/login"); rec.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after the lockout window, got %d", rec.Code)
	}
}
