# NEO Bank — Go Best Practices — Structure

_Part of [go.md](./go.md)._

## Purpose

Defines Go project structure, code organization, and dependency injection patterns used across all NEO Bank backend services. Each service is independently buildable and deployable.

## Project Structure

```text
packages/backend/<service>/
├── cmd/
│   ├── api/           # API server entry point (main.go)
│   ├── consumer/      # Kafka consumer entry point (main.go)
│   └── migrate/       # Migration runner (main.go)
├── internal/
│   ├── api/           # HTTP handlers, middleware, routes, Gin router
│   │   ├── handlers/  # HTTP request handlers (thin)
│   │   ├── middleware/ # Auth, CORS, request ID, error handling
│   │   └── router.go  # Gin route registration
│   ├── core/          # Config, logging, errors, security
│   │   ├── config/    # Configuration loading and validation
│   │   ├── logging/   # Structured logging (slog)
│   │   ├── errors/    # AppError type and sentinel errors
│   │   └── security/  # JWT validation, Argon2id hashing
│   ├── db/            # Cassandra connection, migrations
│   │   ├── cassandra.go # gocql session management
│   │   └── migrations/  # Versioned CQL migration files
│   ├── repositories/  # CQL queries, row mapping only
│   │   ├── account.go   # Account queries
│   │   ├── transfer.go  # Transfer queries
│   │   └── idempotency.go # Idempotency key queries
│   ├── services/      # Business logic, orchestration
│   │   ├── auth.go      # Auth service (login, register, tokens)
│   │   ├── transfer.go  # Transfer service (saga orchestration)
│   │   └── account.go   # Account service (balance, debit, credit)
│   ├── models/        # Domain entities, DTOs
│   │   ├── account.go   # Account domain model
│   │   ├── transfer.go  # Transfer domain model
│   │   └── user.go      # User domain model
│   ├── consumer/      # Kafka consumer logic
│   │   └── transfer_consumer.go # Transfer event consumer
│   └── validation/    # Input validation
│       └── validate.go  # Request validation rules
├── go.mod
├── go.sum
└── Makefile
```

## Code Organization

### Handler → Service → Repository Pattern

All business flows follow this layered pattern:

```
HTTP Request → Gin Handler → Service → Repository → Cassandra/Kafka
```

**Handlers stay thin** — they bind input, delegate to services, and write responses. They contain no business logic.

```go
// internal/api/handlers/transfer.go
func (h *TransferHandler) CreateTransfer(c *gin.Context) {
    var req CreateTransferRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        writeGinError(c, errors.ErrValidation, getRequestID(c), nil)
        return
    }

    ctx := c.Request.Context()
    transfer, err := h.transferService.InitiateTransfer(ctx, req, c.GetHeader("Idempotency-Key"))
    if err != nil {
        writeGinError(c, err, getRequestID(c))
        return
    }

    c.JSON(http.StatusAccepted, transfer)
}
```

**Services own business logic** — validation, coordination, and decision-making:

```go
// internal/services/transfer.go
func (s *TransferService) InitiateTransfer(ctx context.Context, req CreateTransferRequest, idempotencyKey string) (*TransferResponse, error) {
    // 1. Validate inputs
    // 2. Check account ownership
    // 3. Validate sufficient funds
    // 4. Check idempotency key
    // 5. Create transfer record (pending)
    // 6. Publish to Kafka
    // 7. Return response
}
```

**Repositories own CQL queries** — explicit parameterised queries and row mapping only:

```go
// internal/repositories/account.go
func (r *AccountRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Account, error) {
    query := `SELECT account_id, balance_cents, currency_code, updated_at FROM accounts WHERE user_id = ?`
    // parameterised CQL — never interpolate user input
}
```

### Dependency Injection

Repositories → Services → Handlers → Router (`gin.New()`). Dependencies are constructed in `cmd/api/main.go` and injected downward:

```go
// cmd/api/main.go
func main() {
    cfg := config.Load()
    logger := logging.NewLogger(cfg.LogLevel)

    // Repositories
    cassandra := db.NewCassandra(cfg.Cassandra)
    accountRepo := repositories.NewAccountRepository(cassandra.Session(), logger)
    transferRepo := repositories.NewTransferRepository(cassandra.Session(), logger)

    // Services
    authService := services.NewAuthService(accountRepo, cfg.JWT, cfg.Argon2, logger)
    transferService := services.NewTransferService(transferRepo, cfg.Kafka, logger)

    // Handlers
    authHandler := handlers.NewAuthHandler(authService, logger)
    transferHandler := handlers.NewTransferHandler(transferService, logger)

    // Router
    router := gin.New()
    router.Use(middleware.RequestID())
    router.Use(middleware.Logger(logger))
    router.Use(middleware.Auth(authService))
    router.POST("/api/v1/register", authHandler.Register)
    router.POST("/api/v1/login", authHandler.Login)
    router.POST("/api/v1/transfers", transferHandler.CreateTransfer)

    // Start server
    router.Run(cfg.HTTPAddr)
}
```

## Rules

- **Handlers stay thin** — business logic lives in services.
- **Never import route handlers into business logic** — services know nothing about HTTP.
- **Never expose database rows directly as API responses** — map to DTOs first.
- **Context passed through all call chains** — errors wrapped with `%w`.
- **No global state except configuration** — connections are pooled and process-scoped.
- **Each service owns its migrations and schema** — no shared tables between services.
- **Kubernetes manifests for each service** live under `infrastructure/kubernetes/<service>/`.

## Cross-references

- Go best practices index: [`go.md`](./go.md)
- Code style: [`code-style.md`](./code-style.md)
- Error handling: [`error-handling.md`](./error-handling.md)
- Configuration: [`configuration.md`](./configuration.md)
- Architecture: [`docs/04-architecture.md`](../../04-architecture.md#5-repository-and-module-boundaries)
- Planned test: `test_structure` (verifies package dependencies flow in correct direction)