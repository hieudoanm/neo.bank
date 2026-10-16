package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port            string
	Environment     string
	JWTPublicKeyPEM []byte
	JWTKeyID        string
	AuthServiceURL  string
	CORSOrigins     []string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:            getEnv("PORT", "8080"),
		Environment:     getEnv("APP_ENV", "development"),
		JWTPublicKeyPEM: []byte(getEnv("JWT_PUBLIC_KEY_PEM", "")),
		JWTKeyID:        getEnv("JWT_KEY_ID", ""),
		AuthServiceURL:  getEnv("AUTH_SERVICE_URL", "http://localhost:8081"),
		CORSOrigins:     splitEnv("CORS_ORIGINS", "*"),
		ReadTimeout:     10 * time.Second,
		WriteTimeout:    10 * time.Second,
	}

	var missing []string
	if cfg.Environment == "production" && len(cfg.JWTPublicKeyPEM) == 0 {
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
