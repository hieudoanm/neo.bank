package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port                   string
	Environment            string
	JWTPublicKeyPEM        []byte
	JWTKeyID               string
	CassandraContactPoints []string
	CassandraKeyspace      string
}

func Load() (*Config, error) {
	cfg := &Config{
		Port:                   env("PORT", "8082"),
		Environment:            env("APP_ENV", "development"),
		JWTPublicKeyPEM:        []byte(env("JWT_PUBLIC_KEY_PEM", "")),
		JWTKeyID:               env("JWT_KEY_ID", ""),
		CassandraContactPoints: splitCSV(env("CASSANDRA_CONTACT_POINTS", "localhost:9042")),
		CassandraKeyspace:      env("CASSANDRA_KEYSPACE", "account_keyspace"),
	}
	if cfg.Environment == "production" && len(cfg.JWTPublicKeyPEM) == 0 {
		return nil, fmt.Errorf("missing JWT_PUBLIC_KEY_PEM")
	}
	return cfg, nil
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func splitCSV(s string) []string {
	var out []string
	for _, p := range strings.Split(s, ",") {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	if len(out) == 0 {
		out = append(out, s)
	}
	return out
}

var _ = time.Second
