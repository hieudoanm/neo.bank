# NEO Bank — Go Best Practices

_Backend · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [backend index](./README.md)._

## 1. Documents

| Document                                | Covers                                                       |
| --------------------------------------- | ------------------------------------------------------------ |
| [`structure.md`](./structure.md)         | Project structure, code organization, dependency injection   |
| [`error-handling.md`](./error-handling.md) | Error types, error response envelope                         |
| [`cassandra.md`](./cassandra.md)         | Parameterised queries, lightweight transactions (LWTs)       |
| [`configuration.md`](./configuration.md) | Configuration structure, required env vars                   |
| [`testing.md`](./testing.md)             | Unit tests, integration tests with real Cassandra/Kafka      |
| [`observability.md`](./observability.md) | Structured logging, Prometheus metrics                       |
| [`code-style.md`](./code-style.md)       | Go code style, linting rules                                 |

## 2. Project Structure

| Path                          | Purpose                              |
|-------------------------------|--------------------------------------|
| `cmd/api/`                    | API server entry point               |
| `cmd/consumer/`               | Kafka consumer entry point           |
| `cmd/migrate/`                | Migration runner                     |
| `internal/api/`               | HTTP handlers, middleware, routes    |
| `internal/core/`              | Config, logging, errors, security    |
| `internal/db/`                | Cassandra connection, migrations     |
| `internal/repositories/`      | CQL queries, row mapping             |
| `internal/services/`          | Business logic, orchestration        |
| `internal/models/`            | Domain entities, DTOs                |
| `internal/consumer/`          | Kafka consumer logic                 |
| `internal/validation/`        | Input validation                     |

## 3. Code Organization

### Handler → Service → Repository Pattern

Handlers bind input, delegate to services, write responses. Business logic lives in services. Repositories own CQL queries.

### Dependency Injection

Repositories → Services → Handlers → Router (`gin.New()`).

## 4. Error Handling

All errors use the single nested envelope `{"error":{"code","message","details","requestId"}}` (details omitted when empty). `404` preferred over `403` for unauthorised access to another account's resource.

## 5. Database

Parameterised queries only; lightweight transactions for balance updates. No ORM — gocql exposes CQL directly.

## 6. Configuration

Loaded from env vars; validated at startup; fail fast on missing required config. `.env.example` holds placeholders only.

## 7. Testing

Unit tests with mocks; integration tests with real Cassandra/Kafka (provided by CI service containers). Tests assert observable behaviour, not internal implementation.

## 8. Observability

Structured JSON logs with correlation IDs; Prometheus metrics for latency, error rate, throughput, consumer lag.

## 9. Code Style

golangci-lint, go vet, staticcheck; interfaces defined in consumer packages; table-driven tests; error wrapping with `%w`.
