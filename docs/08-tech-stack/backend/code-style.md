# NEO Bank — Go Best Practices — Code Style

_Part of [go.md](./go.md)._

## Purpose

Defines Go code style, linting, and formatting requirements for all NEO Bank backend services. Consistent code style enables code review, reduces cognitive load, and prevents common bugs.

## Toolchain

| Tool | Purpose | Configuration |
|------|---------|---------------|
| `golangci-lint` | Linting (aggregate linter) | Strict config in `.golangci.yaml` |
| `go vet` | Static analysis | Built-in, runs in CI |
| `staticcheck` | Additional static checks | Integrated via golangci-lint |
| `gofmt` | Formatting | Standard Go formatter |
| `goimports` | Import ordering | Integrated via golangci-lint |

## Formatting

- Run `gofmt -l .` before committing; all Go files must be gofmt-clean.
- Run `goimports -w .` to organise imports (standard library, third-party, local).
- Line length: no hard limit, but prefer lines under 100 characters.
- No trailing whitespace; files end with a newline.

## Naming Conventions

| Element | Convention | Example |
|---------|-----------|---------|
| Package | lowercase, no underscores | `authservice`, `transferrepo` |
| Exported type | PascalCase | `AccountService`, `CreateTransferRequest` |
| Unexported type | PascalCase | `accountRepository`, `jwtValidator` |
| Exported function | PascalCase | `InitiateTransfer`, `ValidatePassword` |
| Unexported function | PascalCase | `hashPassword`, `generateTokens` |
| Exported constant | PascalCase | `MaxTransferAmount`, `DefaultPageSize` |
| Unexported constant | PascalCase | `maxTransferAmount`, `defaultPageSize` |
| Exported variable | PascalCase | `ErrInsufficientFunds`, `ErrAccountNotFound` |
| Unexported variable | PascalCase | `defaultLogger`, `kafkaProducer` |
| Interface | noun, PascalCase | `AccountRepository`, `EventPublisher` |
| Struct field | PascalCase (exported) | `AccountID`, `BalanceCents` |
| Struct field | PascalCase (unexported) | `accountID`, `balanceCents` |
| Receiver name | short, consistent | `s` for service, `r` for repository, `h` for handler |
| Error variable | `Err` prefix | `ErrUnauthorized`, `ErrValidation` |
| Sentinel error | `Err` prefix | `ErrInsufficientFunds` |

## Interfaces

**Interfaces are defined in the consumer package, not the producer package.** This follows Go's "accept interfaces, return structs" principle:

```go
// Consumer defines the interface it needs
type AccountReader interface {
    GetBalance(ctx context.Context, accountID uuid.UUID) (int64, error)
}

// Producer returns a concrete type
type AccountRepository struct { ... }

func (r *AccountRepository) GetBalance(ctx context.Context, accountID uuid.UUID) (int64, error) {
    // ...
}
```

- Keep interfaces small (one or two methods).
- Name interfaces after the capability, not the implementation (`Reader`, `Writer`, `Store`, `Publisher`).
- Do not export interfaces that are only used internally.

## Context

**Context is passed through all call chains.** Every function that performs I/O (database, HTTP, Kafka) accepts `context.Context` as its first parameter:

```go
func (s *TransferService) InitiateTransfer(ctx context.Context, req CreateTransferRequest) (*TransferResponse, error)
func (r *AccountRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*Account, error)
func (h *AuthHandler) Login(c *gin.Context) {
    ctx := c.Request.Context()
    // ...
}
```

- Never store Context in a struct (pass it as a parameter).
- Always use `context.Background()` or `context.WithTimeout/WithCancel` at entry points.
- Propagate correlation IDs through Context values.

## Error Handling

**Error wrapping with `%w`** for unwrapping with `errors.As` and `errors.Is`:

```go
// Good: wrap with %w
return fmt.Errorf("query account: %w", err)

// Good: check with errors.Is
if errors.Is(err, gocql.ErrNotFound) {
    return nil, ErrAccountNotFound
}

// Bad: use %v (loses error chain)
return fmt.Errorf("query account: %v", err)
```

- All errors use the `AppError` type with `Code`, `Message`, `Details`, and `Err` fields.
- `404` is preferred over `403` for unauthorised access to another account's resource.
- Never return stack traces, CQL, or credentials in error responses.
- Sentinel errors (`ErrUnauthorized`, `ErrAccountNotFound`, `ErrTransferNotFound`, etc.) are defined in `internal/core/errors/` and carry canonical codes only.

## Financial Data Rules

- **All amounts are integers in minor units (cents)** — never `float64` or `float32`.
- **Balance updates use lightweight transactions (Paxos)** — `IF balance_cents >= ?` to prevent negative balances.
- **Account-scoped access is verified by the application layer** — never trust a raw client-supplied ID.
- **Every mutating operation enforces an idempotency key** — consumers are idempotent under duplicate delivery.
- **Deny by default** — `404` is preferred over disclosing another account's resource.

## Testing Style

- **Table-driven tests** for multiple cases:
```go
func TestValidatePassword(t *testing.T) {
    tests := []struct {
        name     string
        password string
        wantErr  bool
    }{
        {name: "valid", password: "Str0ng!Pass", wantErr: false},
        {name: "too short", password: "Short1!", wantErr: true},
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := ValidatePassword(tt.password)
            if (err != nil) != tt.wantErr {
                t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
            }
        })
    }
}
```

- Test file naming: `*_test.go` in the same package as the code under test.
- Mock interfaces for unit tests; use real Cassandra/Kafka for integration tests (provided by CI service containers).
- Integration tests are skipped with `testing.Short()` when external services are unavailable.

## Cross-references

- Go best practices index: [`go.md`](./go.md)
- Project structure: [`structure.md`](./structure.md)
- Error handling: [`error-handling.md`](./error-handling.md)
- Configuration: [`configuration.md`](./configuration.md)
- Testing: [`testing.md`](./testing.md)
- Backend conventions: `AGENTS.md`
- Lint: `golangci-lint run ./...` (CI gate)