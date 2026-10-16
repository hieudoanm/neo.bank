package config

import (
	"fmt"
	"os"
	"time"
)

type Config struct {
	Port                   string
	Environment            string
	CassandraContactPoints []string
	CassandraKeyspace      string
	CassandraTimeout       time.Duration
	JWTPrivateKeyPEM       []byte
	JWTPublicKeyPEM        []byte
	JWTKeyID               string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                   getEnv("PORT", "8081"),
		Environment:            getEnv("APP_ENV", "development"),
		CassandraContactPoints: splitEnv("CASSANDRA_CONTACT_POINTS", "localhost:9042"),
		CassandraKeyspace:      getEnv("CASSANDRA_KEYSPACE", "auth_keyspace"),
		CassandraTimeout:       5 * time.Second,
		JWTPrivateKeyPEM:       []byte(getEnv("JWT_PRIVATE_KEY_PEM", "")),
		JWTPublicKeyPEM:        []byte(getEnv("JWT_PUBLIC_KEY_PEM", "")),
		JWTKeyID:               getEnv("JWT_KEY_ID", ""),
	}

	var missing []string
	if cfg.Environment == "production" {
		if len(cfg.JWTPrivateKeyPEM) == 0 {
			missing = append(missing, "JWT_PRIVATE_KEY_PEM")
		}
		if len(cfg.JWTPublicKeyPEM) == 0 {
			missing = append(missing, "JWT_PUBLIC_KEY_PEM")
		}
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
		parts := splitComma(v)
		if len(parts) > 0 {
			return parts
		}
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
