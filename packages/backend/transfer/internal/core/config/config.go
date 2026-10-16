package config

import (
	"fmt"
	"os"
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
}

// Load returns configuration for the HTTP API, which requires the JWT public key.
func Load() (*Config, error) {
	cfg := load()
	if len(cfg.JWTPublicKeyPEM) == 0 {
		return nil, fmt.Errorf("missing required configuration: [JWT_PUBLIC_KEY_PEM]")
	}
	return cfg, nil
}

// LoadConsumer returns configuration for the Kafka saga consumer, which performs
// no HTTP authentication and therefore does not require the JWT public key.
func LoadConsumer() *Config {
	return load()
}

func load() *Config {
	return &Config{
		Port:                   env("PORT", "8083"),
		Environment:            env("APP_ENV", "development"),
		JWTPublicKeyPEM:        []byte(env("JWT_PUBLIC_KEY_PEM", "")),
		JWTKeyID:               env("JWT_KEY_ID", ""),
		CassandraContactPoints: split(env("CASSANDRA_HOST", "localhost")),
		CassandraKeyspace:      env("CASSANDRA_KEYSPACE", "transfer_keyspace"),
		KafkaBrokers:           split(env("KAFKA_BROKERS", "localhost:9092")),
		KafkaTopicTransfers:    env("KAFKA_TOPIC_TRANSFERS", "transfers"),
		KafkaTopicTransfersDLQ: env("KAFKA_TOPIC_TRANSFERS_DLQ", "transfers.dlq"),
		KafkaConsumerGroup:     env("KAFKA_CONSUMER_GROUP", "transfer-service-saga"),
	}
}

func env(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func split(s string) []string {
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
