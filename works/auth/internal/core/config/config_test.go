package config

import (
	"reflect"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	for _, k := range []string{"PORT", "CASSANDRA_HOST", "CASSANDRA_KEYSPACE", "JWT_KEY_ID", "JWT_PRIVATE_KEY_PEM", "JWT_PUBLIC_KEY_PEM", "LOG_LEVEL"} {
		t.Setenv(k, "")
	}

	cfg := Load()
	if cfg.Port != "8081" {
		t.Fatalf("Port = %q, want 8081", cfg.Port)
	}
	if cfg.CassandraKeyspace != "auth_keyspace" {
		t.Fatalf("CassandraKeyspace = %q", cfg.CassandraKeyspace)
	}
	if !reflect.DeepEqual(cfg.CassandraContactPoints, []string{"localhost"}) {
		t.Fatalf("CassandraContactPoints = %v", cfg.CassandraContactPoints)
	}
	if cfg.JWTKeyID != "dev-key-1" {
		t.Fatalf("JWTKeyID = %q", cfg.JWTKeyID)
	}
	if cfg.LogLevel != "info" {
		t.Fatalf("LogLevel = %q", cfg.LogLevel)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	t.Setenv("PORT", "9999")
	t.Setenv("CASSANDRA_HOST", " host-a , host-b ,, ")
	t.Setenv("CASSANDRA_KEYSPACE", "custom_keyspace")
	t.Setenv("JWT_KEY_ID", "kid")
	t.Setenv("JWT_PRIVATE_KEY_PEM", "priv")
	t.Setenv("JWT_PUBLIC_KEY_PEM", "pub")
	t.Setenv("LOG_LEVEL", "debug")

	cfg := Load()
	if cfg.Port != "9999" || cfg.CassandraKeyspace != "custom_keyspace" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if !reflect.DeepEqual(cfg.CassandraContactPoints, []string{"host-a", "host-b"}) {
		t.Fatalf("CassandraContactPoints = %v", cfg.CassandraContactPoints)
	}
	if cfg.JWTPrivateKeyPEM != "priv" || cfg.JWTPublicKeyPEM != "pub" || cfg.JWTKeyID != "kid" {
		t.Fatalf("unexpected jwt cfg: %+v", cfg)
	}
	if cfg.LogLevel != "debug" {
		t.Fatalf("LogLevel = %q", cfg.LogLevel)
	}
}

func TestGetEnvListFallsBackWhenEmpty(t *testing.T) {
	t.Setenv("CASSANDRA_HOST", ",,,")
	if got := getEnvList("CASSANDRA_HOST", "localhost"); !reflect.DeepEqual(got, []string{"localhost"}) {
		t.Fatalf("getEnvList = %v", got)
	}
}
