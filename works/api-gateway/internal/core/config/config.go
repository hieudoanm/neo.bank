package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port                  string
	Environment           string
	JWTPublicKeyPEM       []byte
	JWTKeyID              string
	AuthServiceURL        string
	AccountServiceURL     string
	TransferServiceURL    string
	TransactionServiceURL string
	CORSOrigins           []string
	ReadTimeout           time.Duration
	WriteTimeout          time.Duration
	RateLimitAuthPerMin   int
	RateLimitUserPerMin   int
	LoginLockoutFailures  int
	LoginLockoutWindow    time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                  getEnv("PORT", "8080"),
		Environment:           getEnv("APP_ENV", "development"),
		JWTPublicKeyPEM:       []byte(getEnv("JWT_PUBLIC_KEY_PEM", "")),
		JWTKeyID:              getEnv("JWT_KEY_ID", ""),
		AuthServiceURL:        getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		AccountServiceURL:     getEnv("ACCOUNT_SERVICE_URL", "http://localhost:8082"),
		TransferServiceURL:    getEnv("TRANSFER_SERVICE_URL", "http://localhost:8083"),
		TransactionServiceURL: getEnv("TRANSACTION_SERVICE_URL", "http://localhost:8084"),
		CORSOrigins:           splitEnv("CORS_ORIGINS", "*"),
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
	}

	var err error
	if cfg.RateLimitAuthPerMin, err = getEnvInt("RATE_LIMIT_AUTH_PER_MIN", 100); err != nil {
		return nil, err
	}
	if cfg.RateLimitUserPerMin, err = getEnvInt("RATE_LIMIT_USER_PER_MIN", 1000); err != nil {
		return nil, err
	}
	if cfg.LoginLockoutFailures, err = getEnvInt("LOGIN_LOCKOUT_FAILURES", 5); err != nil {
		return nil, err
	}
	lockoutMinutes, err := getEnvInt("LOGIN_LOCKOUT_WINDOW_MINUTES", 15)
	if err != nil {
		return nil, err
	}
	cfg.LoginLockoutWindow = time.Duration(lockoutMinutes) * time.Minute

	var missing []string
	if len(cfg.JWTPublicKeyPEM) == 0 {
		missing = append(missing, "JWT_PUBLIC_KEY_PEM")
	}
	if len(missing) > 0 {
		return nil, fmt.Errorf("missing required configuration: %v", missing)
	}
	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

// getEnvInt parses an integer env var, returning an error for malformed
// values instead of silently falling back to the default.
func getEnvInt(key string, def int) (int, error) {
	v := os.Getenv(key)
	if v == "" {
		return def, nil
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q: must be an integer", key, v)
	}
	return n, nil
}

func splitEnv(key, def string) []string {
	if v := os.Getenv(key); v != "" {
		return splitComma(v)
	}
	return splitComma(def)
}

func splitComma(s string) []string {
	var out []string
	for _, part := range splitString(s, ",") {
		if p := trimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		out = append(out, s)
	}
	return out
}

func splitString(s, sep string) []string {
	var out []string
	start := 0
	for i := 0; i < len(s); i++ {
		if s[i:i+1] == sep {
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

func trimSpace(s string) string {
	start := 0
	for start < len(s) && (s[start] == ' ' || s[start] == '\t' || s[start] == '\n' || s[start] == '\r') {
		start++
	}
	end := len(s)
	for end > start && (s[end-1] == ' ' || s[end-1] == '\t' || s[end-1] == '\n' || s[end-1] == '\r') {
		end--
	}
	return s[start:end]
}
