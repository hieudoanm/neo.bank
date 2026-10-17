package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Port                   string
	Environment            string
	JWTPublicKeyPEM        []byte
	JWTKeyID               string
	CassandraContactPoints []string
	CassandraKeyspace      string
	KafkaBrokers           []string
	KafkaTopicTransfers    string
	KafkaTopicTransfersDLQ string
	KafkaConsumerGroup     string
	InitialBalanceCents    int64
	CurrencyCode           string
}

// Load returns configuration for the HTTP API, which requires the JWT public key.
func Load() (*Config, error) {
	cfg := load()
	if len(cfg.JWTPublicKeyPEM) == 0 {
		return nil, fmt.Errorf("missing required configuration: [JWT_PUBLIC_KEY_PEM]")
	}
	return cfg, nil
}

// LoadConsumer returns configuration for the Kafka consumer, which performs no
// HTTP authentication and therefore does not require the JWT public key.
func LoadConsumer() *Config {
	return load()
}

func load() *Config {
	return &Config{
		Port:                   env("PORT", "8082"),
		Environment:            env("APP_ENV", "development"),
		JWTPublicKeyPEM:        []byte(env("JWT_PUBLIC_KEY_PEM", "")),
		JWTKeyID:               env("JWT_KEY_ID", ""),
		CassandraContactPoints: splitCSV(env("CASSANDRA_HOST", "localhost")),
		CassandraKeyspace:      env("CASSANDRA_KEYSPACE", "account_keyspace"),
		KafkaBrokers:           splitCSV(env("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopicTransfers:    env("KAFKA_TOPIC_TRANSFERS", "transfers"),
		KafkaTopicTransfersDLQ: env("KAFKA_TOPIC_TRANSFERS_DLQ", "transfers.dlq"),
		KafkaConsumerGroup:     env("KAFKA_CONSUMER_GROUP", "account-service-saga"),
		InitialBalanceCents:    envInt64("INITIAL_BALANCE_CENTS", 0),
		CurrencyCode:           env("CURRENCY_CODE", "USD"),
	}
}

func envInt64(key string, def int64) int64 {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	n, err := strconv.ParseInt(strings.TrimSpace(v), 10, 64)
	if err != nil {
		return def
	}
	return n
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
