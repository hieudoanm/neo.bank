# NEO Bank — Go Best Practices — Observability

_Part of [go.md](./go.md)._

## Purpose

Defines structured logging, metrics collection, tracing, and alerting patterns for all NEO Bank backend services. Observability is required for diagnosing cross-service flows in a distributed microservices architecture.

## Structured Logging

NEO Bank uses `log/slog` (Go 1.22+) with JSON output for structured logging.

```go
// internal/core/logging/logging.go
package logging

import (
    "log/slog"
    "os"
)

func NewLogger(level string) *slog.Logger {
    var lvl slog.Level
    switch level {
    case "debug":
        lvl = slog.LevelDebug
    case "warn":
        lvl = slog.LevelWarn
    case "error":
        lvl = slog.LevelError
    default:
        lvl = slog.LevelInfo
    }
    return slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
        Level:       lvl,
        AddSource:   false, // avoid leaking source paths in production
    }))
}

// Helper to extract correlation ID from context
func CorrelationID(ctx context.Context) string {
    if id, ok := ctx.Value(correlationIDKey{}).(string); ok {
        return id
    }
    return ""
}
```

### Usage with correlation IDs

```go
logger.InfoContext(ctx, "transfer initiated",
    "transferId", transferID,
    "sourceAccountId", sourceAccountID,
    "destinationAccountId", destinationAccountID,
    "amountCents", amountCents,
    "correlationId", correlationID,
)

logger.ErrorContext(ctx, "transfer failed",
    "transferId", transferID,
    "error", err,
    "correlationId", correlationID,
)
```

### Logging rules

- **All logs are structured JSON** — queryable by log aggregation systems (ELK, CloudWatch, Stackdriver).
- **Log field names are camelCase** (`requestId`, `correlationId`, `latencyMs`, `statusCode`, `errorType`, `errorMessage`, `amountCents`) as defined in [`US-013`](../../03-user-stories/US-013.md).
- **Every request log includes a correlation ID** — enables end-to-end tracing across service boundaries.
- **Never log secrets** — passwords, tokens, API keys, or PII are excluded from logs.
- **Log levels**: `debug` (development), `info` (normal operations), `warn` (recoverable issues), `error` (failures).
- **Log at service boundaries** — HTTP requests, Kafka publishes/consumes, external API calls.

## Metrics (Prometheus)

```go
// internal/core/metrics/metrics.go
package metrics

import "github.com/prometheus/client_golang/prometheus"

var (
    HTTPRequestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{Name: "http_requests_total"},
        []string{"method", "path", "status"},
    )
    HTTPRequestDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{
            Name:    "http_request_duration_seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "path"},
    )
    TransferThroughput = prometheus.NewCounter(
        prometheus.CounterOpts{Name: "transfers_completed_total"},
    )
    TransferFailures = prometheus.NewCounter(
        prometheus.CounterOpts{Name: "transfers_failed_total"},
    )
    KafkaConsumerLag = prometheus.NewGauge(
        prometheus.GaugeOpts{Name: "kafka_consumer_lag"},
    )
    CassandraQueryDuration = prometheus.NewHistogramVec(
        prometheus.HistogramOpts{Name: "cassandra_query_duration_seconds"},
        []string{"query", "consistency"},
    )
)

func init() {
    prometheus.MustRegister(
        HTTPRequestsTotal,
        HTTPRequestDuration,
        TransferThroughput,
        TransferFailures,
        KafkaConsumerLag,
        CassandraQueryDuration,
    )
}
```

### Required metrics

| Metric | Type | Labels | Purpose |
|--------|------|--------|---------|
| `http_requests_total` | Counter | method, path, status | Request volume and error rates |
| `http_request_duration_seconds` | Histogram | method, path | Latency distribution |
| `transfers_completed_total` | Counter | — | Successful transfer count |
| `transfers_failed_total` | Counter | — | Failed transfer count |
| `kafka_consumer_lag` | Gauge | — | Consumer lag behind latest offset |
| `cassandra_query_duration_seconds` | Histogram | query, consistency | DB query latency |

## Distributed Tracing (OpenTelemetry)

- **Correlation ID** (`X-Request-ID`, `X-Correlation-ID`) propagates across service boundaries.
- **OpenTelemetry** traces span HTTP → Kafka → Cassandra boundaries.
- **Trace ID** included in all structured logs for correlation.
- **Required for**: transfer flow (initiation → Kafka → consumer → DB).

## Alerts

| Alert | Condition | Severity |
|-------|-----------|----------|
| High error rate | error rate > 1% | Critical |
| High p99 latency | p99 latency > 2s | Warning |
| Kafka consumer lag | lag > 1000 messages | Critical |
| Disk usage | > 80% | Warning |
| Service unavailable | 5xx responses > 5% | Critical |

## Cross-references

- Go best practices index: [`go.md`](./go.md)
- Non-functional requirements: [`NFR-050..053`](../../01-requirements/non-functional-requirements.md#observability)
- Architecture: [`docs/04-architecture.md`](../../04-architecture.md#10-scalability-considerations)
- Operations: [`docs/12-operations/monitoring-alerting.md`](../../12-operations/monitoring-alerting.md)
- Planned test: `test_observability` (verifies logs include correlation IDs, metrics are registered)