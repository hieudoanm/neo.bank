# NEO Bank — Go Best Practices — Configuration

_Part of [go.md](./go.md)._

## Purpose

Defines application configuration structure, loading rules, required environment variables, and secrets management for all NEO Bank backend services.

## Configuration Structure

```go
// internal/core/config/config.go
package config

type Config struct {
    AppEnv       string        `mapstructure:"app_env"`
    HTTPAddr     string        `mapstructure:"http_addr"`
    ReadTimeout  time.Duration `mapstructure:"read_timeout"`
    WriteTimeout time.Duration `mapstructure:"write_timeout"`
    Cassandra    CassandraConfig `mapstructure:"cassandra"`
    Kafka        KafkaConfig     `mapstructure:"kafka"`
    JWT          JWTConfig       `mapstructure:"jwt"`
    Argon2       Argon2Config    `mapstructure:"argon2"`
    CORSOrigins  []string        `mapstructure:"cors_origins"`
    LogLevel     string          `mapstructure:"log_level"`
}

type CassandraConfig struct {
    ContactPoints []string `mapstructure:"contact_points"`
    Port          int      `mapstructure:"port"`
    Keyspace      string   `mapstructure:"keyspace"`
    Timeout       time.Duration `mapstructure:"timeout"`
}

type KafkaConfig struct {
    Brokers       []string `mapstructure:"brokers"`
    TopicTransfers string  `mapstructure:"topic_transfers"`
    ConsumerGroup string  `mapstructure:"consumer_group"`
}

type JWTConfig struct {
    PrivateKeyPEM string        `mapstructure:"private_key_pem"`
    PublicKeyPEM  string        `mapstructure:"public_key_pem"`
    AccessTTL     time.Duration `mapstructure:"access_ttl"`
    RefreshTTL    time.Duration `mapstructure:"refresh_ttl"`
    Issuer        string        `mapstructure:"issuer"`
    Audience      string        `mapstructure:"audience"`
}

type Argon2Config struct {
    Memory      uint32 `mapstructure:"memory"`
    Iterations  uint32 `mapstructure:"iterations"`
    Parallelism uint8  `mapstructure:"parallelism"`
    SaltLength  uint32 `mapstructure:"salt_length"`
    KeyLength   uint32 `mapstructure:"key_length"`
}
```

## Loading

```go
func Load() (*Config, error) {
    // 1. Load from environment variables
    // 2. Load optional .env file for local development
    // 3. Validate all required fields
    // 4. Fail fast with clear error messages on missing/invalid config
}
```

## Required Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `APP_ENV` | Application environment | `development`, `staging`, `production` |
| `PORT` | HTTP listen port | `8080` |
| `JWT_PRIVATE_KEY_PEM` | RS256 private key (PEM) for signing access tokens | `-----BEGIN PRIVATE KEY-----...` |
| `JWT_PUBLIC_KEY_PEM` | RS256 public key (PEM) for verifying access tokens | `-----BEGIN PUBLIC KEY-----...` |
| `JWT_KEY_ID` | Key ID advertised in the JWT `kid` header | `dev-key-1` |
| `JWT_ACCESS_TTL` | Access token TTL | `15m` |
| `JWT_REFRESH_TTL` | Refresh token TTL | `720h` (7 days) |
| `CASSANDRA_HOST` | Cassandra contact points (comma-separated) | `cassandra:9042` |
| `CASSANDRA_KEYSPACE` | Cassandra keyspace name | `auth_keyspace` |
| `KAFKA_BROKERS` | Kafka broker addresses (comma-separated) | `kafka:9092` |
| `KAFKA_TOPIC_TRANSFERS` | Kafka topic for transfer events | `transfers` |
| `KAFKA_TOPIC_TRANSFERS_DLQ` | Kafka topic for failed transfer/command events | `transfers.dlq` |
| `KAFKA_CONSUMER_GROUP` | Kafka consumer group ID | `account-service-saga` |
| `INITIAL_BALANCE_CENTS` | Account Service: opening balance (minor units) granted when a user registers | `0` (default), `100000` (demo) |
| `CURRENCY_CODE` | Account Service: ISO 4217 currency code assigned to new accounts | `USD` |
| `ARGON2_MEMORY` | Argon2id memory (KiB) | `65536` |
| `ARGON2_ITERATIONS` | Argon2id iterations | `3` |
| `ARGON2_PARALLELISM` | Argon2id parallelism | `4` |
| `CORS_ORIGINS` | Allowed CORS origins (comma-separated) | `https://app.neo.bank` |
| `RATE_LIMIT_AUTH_PER_MIN` | API Gateway per-IP limit on public auth routes (0 disables) | `100` |
| `RATE_LIMIT_USER_PER_MIN` | API Gateway per-user (JWT subject) limit on authenticated routes (0 disables) | `1000` |
| `LOGIN_LOCKOUT_FAILURES` | Failed logins per IP before lockout (0 disables) | `5` |
| `LOGIN_LOCKOUT_WINDOW_MINUTES` | Login lockout window | `15` |
| `LOG_LEVEL` | Log level | `debug`, `info`, `warn`, `error` |

## Rules

- **Validate at startup and fail clearly** on missing required configuration. No silent defaults for secrets.
- **Environment variables locally via `.env`**; managed secrets in production (AWS Secrets Manager, GCP Secret Manager).
- **`.env.example` holds placeholders only.** Never commit real secrets to the repository.
- **No global state except configuration.** Configuration is loaded once and passed via dependency injection.
- **Connections are pooled and process-scoped** — Cassandra sessions and Kafka producers are created once and shared.
- **Time durations use `time.Duration`** — parsed from strings like `15m`, `30s`, `1h`.
- **CORS origins are explicitly listed** — no wildcard origins in production.

## Secrets Management

| Environment | Secret Storage | Notes |
|-------------|---------------|-------|
| Local development | `.env` file (gitignored) | Placeholders in `.env.example` |
| Staging / Production | AWS Secrets Manager or GCP Secret Manager | Injected via Kubernetes Secrets |
| CI/CD | GitHub Actions secrets / repository variables | Never logged |

## Cross-references

- Go best practices index: [`go.md`](./go.md)
- Project structure: [`structure.md`](./structure.md)
- Backend conventions: `AGENTS.md`
- Security rules: `AGENTS.md` (keep secrets out of the repository)
- Planned test: `test_config_load` (verifies config loads from env, fails on missing required vars)