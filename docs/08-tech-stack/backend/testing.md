# NEO Bank — Go Best Practices — Testing

_Part of [go.md](./go.md)._

## Purpose

Defines unit and integration testing patterns for the Go backend. Tests assert observable behaviour, not internal implementation. Cassandra-dependent tests use a real database; Kafka-dependent tests use a real Kafka instance.

## Test Pyramid

| Layer | Tool | Target Coverage | Speed |
|-------|------|-----------------|-------|
| Unit | Go testing + testify | 80% of backend code | Fast (< 2 min) |
| Integration | Go testing + testify | Critical paths (auth, transfer, balance) | Medium |
| Contract | API response validation | All endpoints | Medium |
| E2E | Playwright (frontend) | Core user journeys | Slow |

## Unit Tests

Unit tests use Go's standard `testing` package with `testify` for assertions. External dependencies (repositories, Kafka producers) are mocked using `gomock` or hand-written mocks.

```go
// internal/services/transfer_service_test.go
package services

import (
    "context"
    "testing"

    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
)

func TestTransferService_InitiateTransfer_Success(t *testing.T) {
    mockRepo := &MockAccountRepository{}
    mockRepo.On("GetByUserID", mock.Anything, "user-1").
        Return(&Account{ID: "acc-1", BalanceCents: 10000}, nil)

    svc := NewTransferService(mockRepo, nil, testLogger)

    req := CreateTransferRequest{
        UserID:               "user-1",
        DestinationAccountID: "acc-2",
        AmountCents:          5000,
    }

    resp, err := svc.InitiateTransfer(context.Background(), req, "idem-123")

    assert.NoError(t, err)
    assert.Equal(t, "pending", resp.Status)
    assert.NotEmpty(t, resp.TransferID)
    mockRepo.AssertExpectations(t)
}

func TestTransferService_InitiateTransfer_InsufficientFunds(t *testing.T) {
    mockRepo := &MockAccountRepository{}
    mockRepo.On("GetByUserID", mock.Anything, "user-1").
        Return(&Account{ID: "acc-1", BalanceCents: 100}, nil)

    svc := NewTransferService(mockRepo, nil, testLogger)

    req := CreateTransferRequest{
        UserID:               "user-1",
        DestinationAccountID: "acc-2",
        AmountCents:          5000,
    }

    resp, err := svc.InitiateTransfer(context.Background(), req, "idem-123")

    assert.Error(t, err)
    assert.Nil(t, resp)
    assert.Contains(t, err.Error(), "Insufficient funds")
}
```

### Unit test rules

- **Table-driven tests** for multiple cases (see `code-style.md`).
- **Mock external dependencies** — repositories, Kafka producers, email services.
- **Test both success and failure paths** — including validation errors, insufficient funds, duplicate idempotency keys.
- **Test file naming**: `*_test.go` in the same package as the code under test.
- **Tests are deterministic** — no randomness, no time-dependent assertions without mocking time.

## Integration Tests with Real Cassandra/Kafka

Integration tests run against real Cassandra and Kafka instances — the CI service containers (see [`github-actions.md`](../infrastructure/github-actions.md)) or the local Docker Compose stack. They are marked with `testing.Short()` so they can be skipped in fast CI runs.

```go
// internal/repositories/account_integration_test.go
package repositories

import (
    "context"
    "os"
    "strings"
    "testing"
    "time"

    "github.com/stretchr/testify/require"
    "github.com/gocql/gocql"
)

func TestAccountRepository_DebitPreventsNegativeBalance(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test - requires Cassandra")
    }

    ctx := context.Background()

    // Connect to the real Cassandra instance provided by CI (service container)
    // or the local Docker Compose stack.
    hosts := strings.Split(os.Getenv("CASSANDRA_HOST"), ",")
    cluster := gocql.ClusterConfig{
        Hosts:    hosts,
        Keyspace: "account_keyspace_test",  // each service uses its own *_test keyspace
        Timeout:  10 * time.Second,
    }
    session, err := cluster.CreateSession()
    require.NoError(t, err)
    defer session.Close()

    repo := NewAccountRepository(session, testLogger)

    // Create account with 1000 cents
    err = repo.CreateAccount(ctx, "acc-1", "user-1", 1000, "USD")
    require.NoError(t, err)

    // Debit 500 cents — should succeed
    err = repo.Debit(ctx, "acc-1", 500)
    require.NoError(t, err)

    // Debit 600 cents — should fail (only 500 remaining)
    err = repo.Debit(ctx, "acc-1", 600)
    require.Error(t, err)
    require.Contains(t, err.Error(), "insufficient")
}
```

### Kafka integration test

```go
// internal/consumer/transfer_consumer_integration_test.go
func TestTransferConsumer_IdempotentUnderDuplicateDelivery(t *testing.T) {
    if testing.Short() {
        t.Skip("skipping integration test - requires Kafka")
    }
    // Start Kafka + Cassandra containers
    // Publish same transfer event twice
    // Verify account debited only once
}
```

### Integration test rules

- **Cassandra-dependent tests use a real database, not mocks** (CI provides Cassandra service).
- **Kafka-dependent tests use a real Kafka instance, not mocks** (CI provides Kafka service).
- **Account isolation and authorisation are always tested negatively** — user A cannot access user B's data.
- **Consumer tests assume duplicate message delivery** — idempotency is verified under redelivery.
- **Mark tests that need external services** with `testing.Short()` so they can be skipped deliberately, never silently.
- **Migrations apply cleanly** before integration tests run.

## Test Organization

| Directory | Contents |
|-----------|----------|
| `internal/*/*_test.go` | Unit tests alongside source code |
| `internal/*/*_integration_test.go` | Integration tests (real Cassandra/Kafka) |
| `testdata/` | Synthetic test fixtures (never real financial data) |

## Fixtures

Test fixtures are synthetic and deterministic:

```go
// testdata/fixtures.go
var TestAccount = Account{
    ID:           "acc-test-1",
    UserID:       "user-test-1",
    BalanceCents: 10000,
    CurrencyCode: "USD",
}

var TestTransfer = Transfer{
    ID:                   "transfer-test-1",
    SourceAccountID:      "acc-test-1",
    DestinationAccountID: "acc-test-2",
    AmountCents:          5000,
    Status:               "pending",
}
```

**Never commit real financial data or PII in test fixtures.**

## Rules

- **Tests assert observable behaviour, not internal implementation** — test through public APIs.
- **Account isolation and authorisation are always tested negatively** — every protected endpoint has a negative test.
- **Consumer tests assume duplicate message delivery** — idempotency is explicitly verified.
- **Fixtures are synthetic and deterministic** — no random values, no real PII.
- **Mark tests that need external services** — skipped with `testing.Short()`, never silently skipped.
- **All tests pass in CI** — `go test ./...` is a CI gate.

## Running Tests

```bash
# Unit tests only (fast)
go test -short ./...

# Unit + integration (requires Docker)
go test ./...

# With coverage
go test -cover ./...

# Per-module coverage gate (enforced: ≥ 80%, excludes cmd/, internal/db/, internal/repositories/)
bash scripts/coverage/backend.sh packages/backend/auth
bash scripts/coverage/backend.sh --all

# Specific package
go test -v ./internal/services/...
```

## Cross-references

- Go best practices index: [`go.md`](./go.md)
- Code style: [`code-style.md`](./code-style.md) (table-driven tests)
- Non-functional requirements: [`NFR-063`](../../01-requirements/non-functional-requirements.md#compliance-and-operational)
- Testing strategy: [`docs/09-testing-strategy.md`](../../09-testing-strategy.md)
- CI pipeline: [`docs/08-tech-stack/infrastructure/github-actions.md`](../infrastructure/github-actions.md)
- Planned test: `go test ./...` (CI gate)