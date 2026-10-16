# NEO Bank — Agent and Contributor Guide

This file contains the repository conventions that automated agents and human contributors must follow.

## 1. Project

NEO Bank is a full-stack banking application demonstrating core financial operations: account registration, authentication, balance inquiry, and fund transfers. It is built as event-driven microservices: an API gateway in front of Auth, Account, Transfer (saga orchestrator), and Transaction services, each owning its own Cassandra keyspace. The project showcases end-to-end software engineering across a React and TypeScript frontend, Go microservices with Cassandra and Kafka, containerised with Docker, orchestrated with Kubernetes (EKS on AWS, GKE on GCP), and deployed via Terraform.

Read `docs/` before implementing. The documents are normative:

| Document                         | Governs                                                |
| -------------------------------- | ------------------------------------------------------ |
| `docs/00-overview.md`            | Product vision and scope                               |
| `docs/01-requirements.md`        | Functional/non-functional requirements (`FR-`, `NFR-`) |
| `docs/02-features.md`            | User-visible behaviour                                 |
| `docs/03-user-stories.md`        | User goals and acceptance scenarios                    |
| `docs/04-architecture.md`        | Module boundaries, request lifecycle, tenancy          |
| `docs/05-data-model.md`          | Schema, constraints, state transitions                 |
| `docs/06-api-specification.md`   | REST contract, status codes, error envelope            |
| `docs/07-ui-ux-design.md`        | UI tokens, layout, states, accessibility               |
| `docs/08-tech-stack.md`          | Allowed technologies and dependency policy             |
| `docs/09-testing-strategy.md`    | Test layers, fixtures, acceptance criteria             |
| `docs/10-implementation-plan.md` | Phased delivery plan                                   |
| `docs/11-deployment.md`          | CI, Docker Compose, and cloud deployment                |
| `docs/12-operations.md`          | Operations: backup, monitoring, security, rollback, recovery |

When an implementation and a document disagree, stop and resolve the disagreement explicitly rather than silently choosing one.

## 2. Repository layout

```text
neo.bank/
├── packages/
│   ├── backend/
│   │   ├── api-gateway/      # Go API gateway (routing, auth enforcement, rate limiting)
│   │   ├── auth/             # Go Auth Service (auth_keyspace: users, refresh_tokens, audit_events)
│   │   ├── accounts/         # Go Account Service (account_keyspace: accounts, accounts_by_user, processed_commands, audit_events)
│   │   ├── transfer/         # Go Transfer Service — saga orchestrator (transfer_keyspace: transfers, transfers_by_account, idempotency_keys, audit_events)
│   │   └── transactions/     # Go Transaction Service — read model (transaction_keyspace: transactions, audit_events)
│   └── frontend/
│       └── app/              # React + TypeScript frontend
├── infrastructure/          # Terraform (AWS/GCP) and Kubernetes manifests (infrastructure/kubernetes/<service>/)
├── docker-compose.yaml      # Local topology: gateway + 4 services + frontend + cassandra + kafka + zookeeper
├── Makefile                 # Convenience commands
├── docs/                    # Normative project documentation
└── .github/workflows/       # CI
```

Each service is an independent Go module with its own entry point, migrations, Cassandra keyspace, and Kubernetes manifests. No service reads or writes another service's keyspace; cross-service interaction happens through the API gateway, explicit service APIs, and Kafka.

Do not create empty placeholder modules. Prefer a small, coherent implementation.

## 3. Toolchain

| Area           | Tool                                                                      |
| -------------- | ------------------------------------------------------------------------- |
| Go             | Go 1.22+, Gin, golangci-lint, go vet, go test, testify                    |
| Node           | pnpm, version pinned in package.json and CI                                |
| Backend lint   | golangci-lint, go vet                                                      |
| Backend tests  | Go testing, testify (real Cassandra and Kafka in CI)                        |
| Frontend lint  | ESLint, Prettier                                                           |
| Frontend types | TypeScript (`tsc --noEmit`)                                               |
| Frontend tests | Vitest, @testing-library/react, Playwright (E2E)                          |

Run the relevant checks through the root Makefile or CI pipeline. Never weaken a check to make CI pass.

## 4. Backend conventions

### Structure

Each service lives in `packages/backend/<service>/` and is an independent Go module. It owns its binary, its CQL migrations, and its Cassandra keyspace.

- `cmd/api/main.go` — HTTP API server entry point.
- `cmd/consumer/main.go` — Kafka consumer entry point.
- `cmd/migrate/main.go` — Cassandra migration runner.
- `internal/api/` — routes, middleware, handlers. Keep handlers thin.
- `internal/core/` — configuration, logging, errors, security.
- `internal/db/` — Cassandra connection and CQL migrations for this service's own keyspace.
- `internal/repositories/` — explicit CQL queries and row mapping only.
- `internal/services/` — business operations coordinating repositories.
- `internal/models/` — Go structs for domain entities.
- `internal/consumer/` — Kafka consumer logic.
- `internal/producer/` — Kafka producer logic.

Kubernetes manifests for each service live under `infrastructure/kubernetes/<service>/`. Health endpoints are served per service: gateway exposes `GET /healthz` and `GET /readyz` at the root, outside `/api/v1`.

Never import route handlers into business logic. Never expose database rows directly as API responses. Never read or write another service's keyspace; cross-service calls go through the gateway, explicit service APIs, or Kafka topics (`transfers`, `transfers.dlq`).

### Database rules

1. All values are passed as parameterised CQL queries. Never interpolate user input into CQL.
2. Dynamic identifiers come only from trusted allowlists.
3. Use lightweight transactions (paxos) for balance updates to ensure atomicity.
4. Every account-scoped query filters by account_id verified by the application layer, never by a raw client value.
5. Every schema change ships as a reviewed, versioned CQL migration in the owning service's migrations directory.
6. Connection pooling is bounded; do not share one connection across unrelated tasks.
7. Each service accesses only its own keyspace. No shared or cross-service keyspace access.

### API rules

1. Base path is `/api/v1`.
2. Responses use the error envelope from `docs/06-api-specification.md`.
3. UUIDs identify resources; timestamps are ISO 8601 UTC (`...Z`).
4. Unknown request fields are rejected on writes.
5. List endpoints use cursor pagination, deterministic ordering, and page size limits.
6. Never return stack traces, CQL, credentials, or sensitive data in responses/logs.

### Security rules

1. Deny by default. Authenticate before authorising.
2. `404` is preferred over disclosing another account's resource (not `403` to avoid information leakage).
3. Validate and bound every input before processing.
4. Never treat possession of an account ID as authorisation.
5. Keep secrets out of the repository, logs, and frontend bundles.
6. All CQL queries use parameterised queries.
7. Passwords hashed with Argon2id before storage.
8. JWT tokens validated on every protected endpoint.
9. Every authorisation path needs a negative test.

## 5. Frontend conventions

1. Strict TypeScript, no `any`.
2. Design tokens from `docs/07-ui-ux-design.md`; avoid scattered arbitrary values.
3. Every data screen implements loading, empty, validation-error, permission-error, network-error, and unexpected-error states.
4. Never show success before the server confirms it.
5. Access the API through TanStack Query; never call the backend from the browser with raw secrets.
6. Hide unavailable actions for usability, but treat the server as the only authority.

## 6. Financial integrity rules

1. All financial amounts stored as integers in minor units (cents).
2. Financial operations must never run in web request handlers.
3. Balance updates use lightweight transactions (paxos) to ensure atomicity.
4. Transfers are processed asynchronously via the Transfer Service saga orchestrator and Kafka consumers.
5. Idempotency keys tracked for all mutating operations.
6. Every balance change has a corresponding transaction record.
7. Account balances must never go negative (enforced by application logic and lightweight transactions).

## 7. Testing conventions

1. Tests assert observable behaviour, not internal implementation.
2. Cassandra-dependent tests use a real database, not mocks (CI provides Cassandra service).
3. Kafka-dependent tests use a real Kafka instance, not mocks (CI provides Kafka service).
4. Account isolation and authorisation are always tested negatively.
5. Consumer tests assume duplicate message delivery and idempotency.
6. Fixtures are synthetic and deterministic. Never commit real financial data.
7. Mark tests that need external services so they can be skipped deliberately, never silently.

## 8. Delivery rules

1. Keep every commit runnable.
2. Add tests alongside the behaviour they protect.
3. Update documentation when behaviour or scope changes.
4. Record significant decisions in `docs/decisions/` (if created).
5. Do not commit generated data, credentials, or build artifacts.
6. Do not claim a feature that is not implemented and tested.
7. Do not process real money or personally identifiable information.
8. Do not claim compliance with banking regulations or certifications without independently establishing requirements and evidence.
