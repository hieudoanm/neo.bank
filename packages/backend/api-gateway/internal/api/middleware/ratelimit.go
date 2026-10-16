package middleware

import (
	"math"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// RateLimits configures gateway rate limiting. A zero value disables the
// corresponding guard.
type RateLimits struct {
	AuthPerMinute int
	UserPerMinute int
	LoginFailures int
	LoginWindow   time.Duration
}

// nowFunc allows tests to control time.
type nowFunc func() time.Time

// RateLimiter is a sliding-window counter keyed by an opaque string.
type RateLimiter struct {
	mu     sync.Mutex
	max    int
	window time.Duration
	events map[string][]time.Time
	now    nowFunc
}

func NewRateLimiter(max int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		max:    max,
		window: window,
		events: make(map[string][]time.Time),
		now:    time.Now,
	}
}

// Allow records a request for key and reports whether it is within the limit.
// When denied it returns the duration until the caller may retry.
func (l *RateLimiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	kept := prune(l.events[key], now.Add(-l.window))
	if len(kept) >= l.max {
		l.events[key] = kept
		return false, retryAfter(kept[0], l.window, now)
	}
	l.events[key] = append(kept, now)
	return true, 0
}

// FailureTracker counts failed attempts (for example failed logins) per key and
// locks the key out once the threshold is reached within the window.
type FailureTracker struct {
	mu        sync.Mutex
	threshold int
	window    time.Duration
	events    map[string][]time.Time
	now       nowFunc
}

func NewFailureTracker(threshold int, window time.Duration) *FailureTracker {
	return &FailureTracker{
		threshold: threshold,
		window:    window,
		events:    make(map[string][]time.Time),
		now:       time.Now,
	}
}

// Record registers a failed attempt for key.
func (t *FailureTracker) Record(key string) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	t.events[key] = append(prune(t.events[key], now.Add(-t.window)), now)
}

// Locked reports whether key has reached the failure threshold within the
// window, and if so how long until the oldest failure ages out.
func (t *FailureTracker) Locked(key string) (bool, time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	now := t.now()
	kept := prune(t.events[key], now.Add(-t.window))
	t.events[key] = kept
	if len(kept) >= t.threshold {
		return true, retryAfter(kept[0], t.window, now)
	}
	return false, 0
}

// prune drops events that occurred at or before cutoff. Events are appended in
// chronological order, so only a prefix is removed.
func prune(events []time.Time, cutoff time.Time) []time.Time {
	i := 0
	for i < len(events) && !events[i].After(cutoff) {
		i++
	}
	if i == 0 {
		return events
	}
	return events[i:]
}

func retryAfter(oldest time.Time, window time.Duration, now time.Time) time.Duration {
	retry := oldest.Add(window).Sub(now)
	if retry < time.Second {
		return time.Second
	}
	return retry
}

// ClientIPKey keys limits by the caller's IP address (public endpoints).
func ClientIPKey(c *gin.Context) string {
	return c.ClientIP()
}

// AuthenticatedUserKey keys limits by the authenticated JWT subject (protected
// endpoints), falling back to the client IP if unset.
func AuthenticatedUserKey(c *gin.Context) string {
	if userID := c.GetString("userId"); userID != "" {
		return userID
	}
	return c.ClientIP()
}

// RateLimitMiddleware rejects requests once the limiter denies the key.
func RateLimitMiddleware(limiter *RateLimiter, keyFunc func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := keyFunc(c)
		if key == "" {
			c.Next()
			return
		}
		if allowed, retry := limiter.Allow(key); !allowed {
			writeRateLimited(c, retry)
			c.Abort()
			return
		}
		c.Next()
	}
}

// LoginLockoutMiddleware blocks a key that has exceeded the failure threshold
// and records a failure whenever the wrapped handler responds 401.
func LoginLockoutMiddleware(tracker *FailureTracker) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.ClientIP()
		if locked, retry := tracker.Locked(key); locked {
			writeRateLimited(c, retry)
			c.Abort()
			return
		}

		recorder := &statusRecorder{ResponseWriter: c.Writer}
		c.Writer = recorder
		c.Next()

		if recorder.status == http.StatusUnauthorized {
			tracker.Record(key)
		}
	}
}

// statusRecorder captures the response status without buffering the body.
type statusRecorder struct {
	gin.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(code int) {
	r.status = code
	r.ResponseWriter.WriteHeader(code)
}

func (r *statusRecorder) Write(data []byte) (int, error) {
	if r.status == 0 {
		r.status = http.StatusOK
	}
	return r.ResponseWriter.Write(data)
}

func writeRateLimited(c *gin.Context, retry time.Duration) {
	seconds := int(math.Ceil(retry.Seconds()))
	if seconds < 1 {
		seconds = 1
	}
	c.Header("Retry-After", strconv.Itoa(seconds))
	c.JSON(http.StatusTooManyRequests, gin.H{
		"error": gin.H{
			"code":      "RATE_LIMITED",
			"message":   "rate limit exceeded",
			"requestId": getRequestID(c),
		},
	})
}
