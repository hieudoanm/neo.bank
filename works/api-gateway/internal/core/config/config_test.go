package config

import (
	"testing"
	"time"
)

func loadWith(t *testing.T, env map[string]string) *Config {
	t.Helper()
	t.Setenv("JWT_PUBLIC_KEY_PEM", "test-key")
	for k, v := range env {
		t.Setenv(k, v)
	}
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return cfg
}

func TestLoadRateLimitDefaults(t *testing.T) {
	cfg := loadWith(t, map[string]string{
		"RATE_LIMIT_AUTH_PER_MIN":      "",
		"RATE_LIMIT_USER_PER_MIN":      "",
		"LOGIN_LOCKOUT_FAILURES":       "",
		"LOGIN_LOCKOUT_WINDOW_MINUTES": "",
	})

	if cfg.RateLimitAuthPerMin != 100 {
		t.Fatalf("expected auth limit 100, got %d", cfg.RateLimitAuthPerMin)
	}
	if cfg.RateLimitUserPerMin != 1000 {
		t.Fatalf("expected user limit 1000, got %d", cfg.RateLimitUserPerMin)
	}
	if cfg.LoginLockoutFailures != 5 {
		t.Fatalf("expected 5 lockout failures, got %d", cfg.LoginLockoutFailures)
	}
	if cfg.LoginLockoutWindow != 15*time.Minute {
		t.Fatalf("expected 15m lockout window, got %v", cfg.LoginLockoutWindow)
	}
}

func TestLoadRateLimitOverrides(t *testing.T) {
	cfg := loadWith(t, map[string]string{
		"RATE_LIMIT_AUTH_PER_MIN":      "7",
		"RATE_LIMIT_USER_PER_MIN":      "9",
		"LOGIN_LOCKOUT_FAILURES":       "3",
		"LOGIN_LOCKOUT_WINDOW_MINUTES": "30",
	})

	if cfg.RateLimitAuthPerMin != 7 || cfg.RateLimitUserPerMin != 9 {
		t.Fatalf("unexpected limits: auth=%d user=%d", cfg.RateLimitAuthPerMin, cfg.RateLimitUserPerMin)
	}
	if cfg.LoginLockoutFailures != 3 || cfg.LoginLockoutWindow != 30*time.Minute {
		t.Fatalf("unexpected lockout: failures=%d window=%v", cfg.LoginLockoutFailures, cfg.LoginLockoutWindow)
	}
}

func TestLoadInvalidIntReturnsError(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY_PEM", "test-key")
	t.Setenv("RATE_LIMIT_AUTH_PER_MIN", "not-a-number")
	if _, err := Load(); err == nil {
		t.Fatal("expected an error for a malformed integer env var")
	}
}

func TestLoadRequiresJWTKey(t *testing.T) {
	t.Setenv("JWT_PUBLIC_KEY_PEM", "")
	if _, err := Load(); err == nil {
		t.Fatal("expected an error when JWT_PUBLIC_KEY_PEM is missing")
	}
}
