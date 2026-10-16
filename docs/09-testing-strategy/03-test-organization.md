# NEO Bank — Test Organization

## Purpose

Defines the directory structure for backend and frontend tests so that tests are colocated with code, integration tests are separated from unit tests, and E2E tests are independently runnable.

## 1. Backend Test Structure

Each service is an independent Go module under `packages/backend/<service>/`, so this structure repeats per service:

```text
packages/backend/<service>/
├── cmd/
│   ├── api/
│   │   └── main_test.go          # API entry point tests
│   ├── consumer/
│   │   └── main_test.go          # Consumer entry point tests
│   └── migrate/
│       └── main_test.go          # Migration runner tests
├── internal/
│   ├── api/
│   │   ├── handlers/
│   │   │   ├── auth_handler_test.go
│   │   │   └── transfer_handler_test.go
│   │   └── middleware/
│   │       └── auth_middleware_test.go
│   ├── core/
│   │   ├── config/
│   │   │   └── config_test.go
│   │   ├── logging/
│   │   │   └── logging_test.go
│   │   ├── errors/
│   │   │   └── errors_test.go
│   │   └── security/
│   │       └── jwt_test.go
│   ├── db/
│   │   ├── cassandra_test.go
│   │   └── migrations/
│   │       └── migration_test.go
│   ├── repositories/
│   │   ├── account_repository_test.go
│   │   ├── transfer_repository_test.go
│   │   └── idempotency_repository_test.go
│   ├── services/
│   │   ├── auth_service_test.go
│   │   ├── transfer_service_test.go
│   │   └── account_service_test.go
│   ├── models/
│   │   ├── account_test.go
│   │   └── transfer_test.go
│   ├── consumer/
│   │   └── transfer_consumer_test.go
│   └── validation/
│       └── validate_test.go
├── tests/
│   ├── integration/
│   │   ├── api_integration_test.go    # API with real Cassandra/Kafka
│   │   ├── consumer_integration_test.go # Consumer with real Kafka
│   │   └── saga_integration_test.go   # Full saga flow
│   └── e2e/
│       └── transfer_workflow_test.go  # Full workflow E2E
├── testdata/
│   ├── accounts.json
│   ├── transfers.json
│   └── seeds/
│       └── seed.go
└── go.mod
```

### Test Placement Rules

- **Unit tests live alongside code** under `internal/` — same package, `*_test.go` suffix.
- **Integration tests live under `tests/integration/`** — require real Cassandra and Kafka.
- **E2E tests live under `tests/e2e/`** — full workflow through the API.
- **Test fixtures live under `testdata/`** — synthetic data, never real PII.
- **Test helpers are not in production code** — test-only utilities live in `tests/` or `testdata/`.

## 2. Frontend Test Structure

```text
packages/frontend/app/
├── src/
│   ├── components/
│   │   ├── auth/
│   │   │   ├── LoginForm.tsx
│   │   │   └── LoginForm.test.tsx
│   │   ├── balance/
│   │   │   ├── BalanceCard.tsx
│   │   │   └── BalanceCard.test.tsx
│   │   ├── transfer/
│   │   │   ├── TransferForm.tsx
│   │   │   └── TransferForm.test.tsx
│   │   └── ui/
│   │       ├── Button.tsx
│   │       └── Button.test.tsx
│   ├── lib/
│   │   ├── api/
│   │   │   └── client.test.ts
│   │   └── env.ts
│   ├── hooks/
│   │   ├── useBalance.ts
│   │   └── useBalance.test.ts
│   ├── pages/
│   │   ├── LoginPage.tsx
│   │   └── LoginPage.test.tsx
│   └── test/
│       └── setup.ts            # Vitest setup file
├── tests/
│   └── e2e/
│       ├── transfer.spec.ts
│       ├── auth.unauth.spec.ts
│       └── negative.spec.ts
├── vitest.config.ts
└── package.json
```

### Test Placement Rules

- **Unit tests live alongside components** under `src/` — `*.test.tsx` next to the component.
- **E2E tests live under `tests/e2e/`** — Playwright specs.
- **Test setup lives under `src/test/`** — Vitest setup, MSW server, mocks.
- **No tests in production builds** — test files are excluded from `vite build`.

## 3. Naming Conventions

| File Type | Pattern | Example |
|-----------|---------|---------|
| Unit test (Go) | `*_test.go` | `auth_service_test.go` |
| Integration test (Go) | `*_integration_test.go` | `api_integration_test.go` |
| E2E test (Go) | `*_e2e_test.go` | `transfer_workflow_e2e_test.go` |
| Unit test (TS) | `*.test.tsx` or `*.test.ts` | `LoginForm.test.tsx` |
| E2E spec (TS) | `*.spec.ts` | `transfer.spec.ts` |
| Test helper | `*_test_helper.go` or `test_*.ts` | `test_setup.ts` |

## 4. Test Categorisation

Tests are tagged for selective execution:

```bash
# Unit tests only (fast, no external deps)
go test -short ./...
pnpm test --run unit

# Integration tests (need Cassandra + Kafka)
go test ./tests/integration/...
pnpm test:e2e --project chromium

# Security tests
go test -run Security ./...
```

Tags:
- `unit` — no external dependencies, runs in seconds.
- `integration` — needs real Cassandra/Kafka, runs in minutes.
- `e2e` — needs full stack, runs in Playwright browser.
- `security` — authorisation, authentication, input validation tests.
- `performance` — load and latency tests.

## Cross-references

- Testing strategy: [`09-testing-strategy.md`](../09-testing-strategy.md#2-test-layers)
- Backend testing patterns: [`docs/08-tech-stack/backend/testing.md`](../08-tech-stack/backend/testing.md)
- Frontend testing patterns: [`docs/08-tech-stack/frontend/testing/vitest.md`](../08-tech-stack/frontend/testing/vitest.md)
- CI pipeline: [`04-quality-gates.md`](./04-quality-gates.md)
- Planned test: `test_organization` (verifies test files are in correct locations)