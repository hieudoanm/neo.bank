# NEO Bank — Product Specification

Product strategy · Functional requirements · Architecture · Data model · Delivery roadmap

> This document is the product-strategy source. The normative implementation set is [`00`–`12`](#normative-documents). Where this document and the numbered documents disagree on implementation detail, the numbered documents govern.

## Normative documents

| #   | Document                                                 | Governs                                                      |
| --- | -------------------------------------------------------- | ------------------------------------------------------------ |
| 00  | [`00-overview.md`](00-overview.md)                       | Product vision and scope                                     |
| 01  | [`01-requirements.md`](01-requirements.md)               | Functional/non-functional requirements (`FR-`, `NFR-`)       |
| 02  | [`02-features.md`](02-features.md)                       | User-visible behaviour                                       |
| 03  | [`03-user-stories.md`](03-user-stories.md)               | User goals and acceptance scenarios                          |
| 04  | [`04-architecture.md`](04-architecture.md)               | Module boundaries, request lifecycle, tenancy                |
| 05  | [`05-data-model.md`](05-data-model.md)                   | Schema, constraints, state transitions                       |
| 06  | [`06-api-specification.md`](06-api-specification.md)     | REST contract, status codes, error envelope                  |
| 07  | [`07-ui-ux-design.md`](07-ui-ux-design.md)               | UI tokens, layout, states, accessibility                     |
| 08  | [`08-tech-stack.md`](08-tech-stack.md)                   | Allowed technologies and dependency policy                   |
| 09  | [`09-testing-strategy.md`](09-testing-strategy.md)       | Test layers, fixtures, acceptance criteria                   |
| 10  | [`10-implementation-plan.md`](10-implementation-plan.md) | Phased delivery plan                                         |
| 11  | [`11-deployment.md`](11-deployment.md)                   | CI, Docker Compose, and cloud deployment                     |
| 12  | [`12-operations.md`](12-operations.md)                   | Operations: backup, monitoring, security, rollback, recovery |

---

## 1. Product vision

NEO Bank is a cloud-based banking platform demonstration for developers and financial software enthusiasts.

It enables users to register accounts, authenticate securely, view balances, transfer funds between accounts, and track transaction history — without maintaining their own banking infrastructure.

The product demonstrates end-to-end software engineering across a React and TypeScript frontend, Go microservices with Cassandra and Kafka, containerised with Docker, orchestrated with Kubernetes (EKS on AWS, GKE on GCP), and deployed via Terraform on AWS and GCP.

## Product principles

**Financial integrity first** — Preserve audit trails and distinguish display formatting from stored precision.

**Secure by default** — Deny access unless authentication and authorisation checks succeed.

**Event-driven processing** — Transfers are processed asynchronously via Kafka with idempotent consumers and exactly-once semantics.

**Reproducibility by default** — Record the inputs, parameters, and versions needed to understand a transaction.

### The product promise

> Go from account registration to a successful fund transfer with full audit trails in the cloud.

NEO Bank is a software demonstration, not a financial product. It does not provide financial advice, process real money, or claim compliance with banking regulations.

## 2. Target users

| Persona            | Primary need                       | Desired outcome                            |
| ------------------ | ---------------------------------- | ------------------------------------------ |
| Account Holder     | Manage account, transfer funds     | Complete banking operations securely       |
| Developer/Operator | Deploy, maintain, monitor platform | Reproducible infrastructure and operations |

Primary persona: a developer who wants to understand how to build a secure, event-driven banking platform with distributed databases and message queues.

## 3. Scope and priorities

- **P0** — Core platform: necessary for the first usable release.
- **P1** — Enhancement: important for a complete experience.

The full platform is the product vision. The first release should deliver one complete banking workflow rather than an incomplete version of every feature.

## 4. Functional requirements

See [`01-requirements.md`](01-requirements.md) for the complete FR-001 through FR-052 specification.

### P0 — Core platform

Registration, login, token refresh, balance inquiry, transfer initiation, idempotency.

### P1 — Enhancement

Transaction history with cursor-based pagination.

## 5. User journeys

### Journey A — Register and get started

Goal: create an account and view balance.

1. Visit the application.
2. Register with email and password.
3. Receive account ID and session token.
4. Log in automatically.
5. View current balance.

Success metric: percentage of new users who successfully register and view balance.

### Journey B — Transfer funds

Goal: transfer money to another account.

1. Log in with email and password.
2. View current balance.
3. Initiate transfer to another account.
4. See transfer status as pending.
5. Wait for transfer to complete.
6. View updated balance.
7. Review transaction in history.

Success metric: successful transfer rate, time to completion.

### Journey C — Monitor transfers

Goal: track transfer status and transaction history.

1. Log in.
2. View transfer status (polling).
3. See status update to completed.
4. View transaction history.
5. Paginate through transactions.

Success metric: status update accuracy, history query performance.

## 6. Application structure

| Page            | Purpose                                            |
| --------------- | -------------------------------------------------- |
| Login           | Email and password authentication                  |
| Register        | Account creation with input validation             |
| Home            | Landing page                                       |
| Dashboard       | Balance, recent transactions, quick actions        |
| Balance         | Balance display, quick actions (part of Dashboard) |
| Transfer        | Transfer form, validation, status monitoring       |
| Transfer Status | Transfer status polling, outcome and errors        |
| Transactions    | Transaction history, pagination, filtering         |
| Settings        | Profile, security (future enhancement)             |

## 7. Technical architecture

### 7.1 System design

The platform is an event-driven microservices system (see [`04-architecture.md`](04-architecture.md)). An API gateway fronts four services — Auth, Account, Transfer, and Transaction — each owning its own Cassandra keyspace and writing its own audit events. The Transfer Service is the saga orchestrator: it accepts transfer initiation, records idempotency keys, and publishes to the `transfers` topic. Kafka consumers process the financial operations (balance updates, transaction recording) asynchronously and idempotently, with failures routed to `transfers.dlq`. The frontend requests data through the gateway and polls for transfer status updates.

### 7.2 Technology stack

| Layer            | Technology                                         | Responsibility                                                          |
| ---------------- | -------------------------------------------------- | ----------------------------------------------------------------------- |
| Frontend         | React 18, TypeScript, Vite, TanStack Query, Router | Interactive banking interface                                           |
| Styling          | Tailwind CSS + DaisyUI                             | Responsive styling, design tokens                                       |
| API              | Go 1.22+, Gin router, gocql, kafka-go              | REST API, auth/authorisation, validation                                |
| Database         | Apache Cassandra 4.x                               | Distributed financial state                                             |
| Message Queue    | Apache Kafka 3.x                                   | Transfer event processing                                               |
| Containerisation | Docker, Docker Compose                             | Local topology reproducibility                                          |
| Orchestration    | Kubernetes (EKS on AWS, GKE on GCP)                | Scheduling, probes, scaling, rollouts                                   |
| Infrastructure   | Terraform                                          | AWS (EKS, MSK, Keyspaces) and GCP (GKE, Cassandra on GKE, Kafka on GKE) |
| CI/CD            | GitHub Actions                                     | Quality gate, build, deployment                                         |
| Testing          | Go testing, testify, Vitest, Playwright            | Unit, integration, E2E tests                                            |

## 8. Current implementation status

This snapshot reflects the repository state on 2026-10-17. It is intentionally separate from the product vision above.

### Implemented

- **Backend services** (`packages/backend/`): `api-gateway`, `auth`, `account`, `transfer`, and `transaction` — each an independent Go module with its own Cassandra keyspace, CQL migrations, and `/healthz`/`/readyz` endpoints.
  - Auth: `POST /api/v1/auth/register`, `/login`, `/refresh`, `/logout` (Argon2id passwords, RS256 JWT).
  - Account: `GET /api/v1/accounts/{accountId}/balance`.
  - Transfer: `POST /api/v1/transfers` (idempotency key), `GET /api/v1/transfers/{transferId}`.
  - Transaction: `GET /api/v1/accounts/{accountId}/transactions` (cursor pagination, max page size 100).
- **Event-driven saga** on the `transfers` topic: the Transfer Service orchestrates debit/credit commands, the Account Service applies them with Cassandra LWTs and claims idempotency keys in `processed_commands`, and a Transaction Service projection consumer builds the read model. Consumers run in distinct groups (`transfer-service-saga`, `account-service-saga`, `transaction-service-projection`).
- **Frontend** (`packages/frontend/app/`): React + TypeScript + TanStack Query with Login, Register, Dashboard (balance), Transfer form, Transfer Status (polling), and Transactions (cursor-paginated) screens. Protected calls send `Authorization: Bearer`.
- **Local topology**: `docker-compose.yaml` runs gateway + 4 services + 3 consumers + Cassandra + Kafka + ZooKeeper + frontend.
- **API Gateway rate limiting** (`packages/backend/api-gateway/`): per-IP sliding-window limit on the public auth routes, per-user (JWT subject) limit on authenticated routes, and a per-IP login lockout after repeated failures; rejects with `429 RATE_LIMITED` and a `Retry-After` header (limits in [`06-api-specification.md` §8](06-api-specification.md#8-rate-limiting)).
- **Tests**: Go unit tests with fakes covering API Gateway rate limiting, security (cross-account access returns `404`, injection payloads rejected before reaching Cassandra), and idempotency (duplicate transfer keys and duplicate Kafka events), plus Vitest frontend tests for the authentication, transfer, transfer-status, and transactions screens, UI primitives, hooks, routing, and the API client, with an enforced **80% coverage** threshold (`pnpm test` runs `vitest --coverage`).
- **CI** (`.github/workflows/ci.yml`): per-service Go formatting (`gofmt`), `go vet`, build and tests; frontend typecheck, tests, and build; and Trivy security scanning (filesystem dependencies plus each container image) with SARIF upload to GitHub code scanning, failing on fixable Critical/High findings.

### Not yet implemented

- **Integration/E2E tests** against real Cassandra and Kafka (unit tests use fakes; CI has no Cassandra/Kafka services).
- **Performance tests** and CI performance gates (latency and consumer throughput targets).
- **Observability** beyond structured logs and health endpoints (metrics, tracing, dashboards, alerting).
- **Cloud infrastructure**: Terraform (AWS/GCP) and Kubernetes manifests.

### Phase plan

Phases 0–7 are complete (Phase 7 transaction history landed with backend and frontend); Phase 8 is in progress (API Gateway rate limiting, the security and idempotency test suites, and CI security scanning landed; performance tests remain); Phases 9–11 remain. The phase numbering below is canonical — see [`10-implementation-plan.md`](10-implementation-plan.md) for detailed phases.

Phase 0 Foundation → Phase 1 Data model → Phase 2 Authentication → Phase 3 Balance inquiry → Phase 4 Transfer initiation → Phase 5 Kafka consumer → Phase 6 Transfer status → Phase 7 Transaction history → Phase 8 Security and idempotency tests → Phase 9 Health and observability → Phase 10 Documentation and demo → Phase 11 Cloud infrastructure (optional).

See [`10-implementation-plan.md`](10-implementation-plan.md) for detailed phases.

## 9. Non-functional requirements

See [`01-requirements.md`](01-requirements.md) for complete NFR specification.

### Key targets

- Performance: registration < 500ms, login < 300ms, balance < 100ms, transfer < 200ms (p95)
- Availability: 99.9%, production RF=3 Cassandra and Kafka RF=3 min.insync.replicas=2 (local/CI run single-node RF=1)
- Consistency: balances never negative, per-account atomic updates with saga compensation, idempotency guarantee
- Security: JWT RS256, Argon2id passwords, hashed refresh tokens, TLS 1.2+, rate limiting
- Observability: structured JSON logs, Prometheus metrics, correlation IDs

## 10. Design and UX

See [`07-ui-ux-design.md`](07-ui-ux-design.md) for complete design specifications.

### Design principles

- Clean and professional (banking context)
- Accessible (WCAG 2.1 AA compliant)
- Responsive (mobile, tablet, desktop)
- Consistent across all screens
- Clear feedback (loading, error, success states)

## 11. Testing strategy

See [`09-testing-strategy.md`](09-testing-strategy.md) for complete testing strategy.

### Test layers

Unit (80% backend / 80% frontend coverage), Integration (real Cassandra/Kafka), Contract, E2E (Playwright), Security (authz/idempotency/input), Performance, Reliability

## 12. Deployment

See [`11-deployment.md`](11-deployment.md) for the complete deployment strategy (CI, Docker Compose, and cloud deployment) and [`12-operations.md`](12-operations.md) for ongoing operations.

### Options

- Local: Docker Compose with gateway, four services, frontend, Cassandra, Kafka, and ZooKeeper
- Orchestration: Kubernetes (EKS on AWS, GKE on GCP)
- CI/CD: GitHub Actions with quality gates and security scans (Trivy)
- Operations: backup, monitoring, security, rollback, and recovery (see [`12-operations.md`](12-operations.md))

## 13. Contributing

See `AGENTS.md` for contributor guidelines.

### Repository conventions

- Go: golangci-lint, go vet, go test
- TypeScript: ESLint, Prettier, tsc
- Git hooks: Husky + lint-staged
- Tests: Vitest, Playwright, Go testing, testify
- Documentation: Keep docs in sync with implementation

## 14. License

This is a portfolio project for demonstration purposes. Not for production use or financial transactions.
