package config

import (
	"os"
	"strings"
	"time"
)

type Config struct {
	Port                   string
	CassandraContactPoints []string
	CassandraKeyspace      string
	CassandraTimeout       time.Duration
	KafkaBrokers           []string
	KafkaTopicTransfers    string
	JWTKeyID               string
	JWTPrivateKeyPEM       string
	JWTPublicKeyPEM        string
	LogLevel               string
}

func Load() Config {
	cfg := Config{
		Port:                   getEnv("PORT", "8081"),
		CassandraContactPoints: getEnvList("CASSANDRA_HOST", "localhost"),
		CassandraKeyspace:      getEnv("CASSANDRA_KEYSPACE", "auth_keyspace"),
		CassandraTimeout:       5 * time.Second,
		KafkaBrokers:           getEnvList("KAFKA_BROKERS", "localhost:9092"),
		KafkaTopicTransfers:    getEnv("KAFKA_TOPIC_TRANSFERS", "transfers"),
		JWTKeyID:               getEnv("JWT_KEY_ID", "dev-key-1"),
		JWTPrivateKeyPEM:       getEnv("JWT_PRIVATE_KEY_PEM", ""),
		JWTPublicKeyPEM:        getEnv("JWT_PUBLIC_KEY_PEM", ""),
		LogLevel:               getEnv("LOG_LEVEL", "info"),
	}
	return cfg
}

func getEnv(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func getEnvList(k, d string) []string {
	var out []string
	for _, part := range strings.Split(getEnv(k, d), ",") {
		if p := strings.TrimSpace(part); p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 {
		out = append(out, d)
	}
	return out
}
