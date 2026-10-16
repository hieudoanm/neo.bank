package config

import (
	"reflect"
	"testing"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"PORT", "APP_ENV", "JWT_PUBLIC_KEY_PEM", "JWT_KEY_ID", "CASSANDRA_HOST", "CASSANDRA_KEYSPACE", "KAFKA_BROKERS", "KAFKA_TOPIC_TRANSFERS", "KAFKA_TOPIC_TRANSFERS_DLQ", "KAFKA_CONSUMER_GROUP"} {
		t.Setenv(k, "")
	}
}

func TestLoadRequiresPublicKey(t *testing.T) {
	clearEnv(t)
	if _, err := Load(); err == nil {
		t.Fatal("expected Load to fail without JWT_PUBLIC_KEY_PEM")
	}
}

func TestLoadDefaultsWithKey(t *testing.T) {
	clearEnv(t)
	t.Setenv("JWT_PUBLIC_KEY_PEM", "key")
	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.Port != "8082" || cfg.Environment != "development" {
		t.Fatalf("unexpected cfg: %+v", cfg)
	}
	if cfg.CassandraKeyspace != "account_keyspace" {
		t.Fatalf("unexpected keyspace: %s", cfg.CassandraKeyspace)
	}
	if !reflect.DeepEqual(cfg.CassandraContactPoints, []string{"localhost"}) {
		t.Fatalf("unexpected contact points: %v", cfg.CassandraContactPoints)
	}
	if cfg.KafkaTopicTransfers != "transfers" || cfg.KafkaConsumerGroup != "account-service-saga" {
		t.Fatalf("unexpected kafka cfg: %+v", cfg)
	}
}

func TestLoadConsumerDoesNotRequireKey(t *testing.T) {
	clearEnv(t)
	t.Setenv("KAFKA_BROKERS", "a:9092, b:9092")
	cfg := LoadConsumer()
	if len(cfg.JWTPublicKeyPEM) != 0 {
		t.Fatalf("expected empty public key, got %q", cfg.JWTPublicKeyPEM)
	}
	if !reflect.DeepEqual(cfg.KafkaBrokers, []string{"a:9092", "b:9092"}) {
		t.Fatalf("unexpected brokers: %v", cfg.KafkaBrokers)
	}
}

func TestSplitCSVFallsBackWhenEmpty(t *testing.T) {
	if got := splitCSV(" , "); !reflect.DeepEqual(got, []string{" , "}) {
		t.Fatalf("splitCSV fallback = %v", got)
	}
}
