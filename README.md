# NEO Bank

A full-stack banking application demonstrating core financial operations —
account registration, authentication, balance inquiry, and fund transfers —
built as an event-driven microservices system.

## Architecture

```
React SPA ──> API Gateway ──> Auth Service      (auth_keyspace)
                          ├─> Account Service   (account_keyspace)
                          ├─> Transfer Service  (transfer_keyspace, saga orchestrator)
                          └─> Transaction Service (transaction_keyspace, read model)

Services communicate over Kafka (`transfers`, `transfers.dlq`) for the
transfer saga; each service owns its own Cassandra keyspace.
```

- **Backend**: Go 1.22+ microservices (Gin, Cassandra via gocql, Kafka).
- **Frontend**: React 19 + TypeScript, TanStack Query, Tailwind/daisyUI.
- **Infra**: Docker Compose for local development; Terraform + Kubernetes
  manifests under `infrastructure/` for cloud deployment.

## Repository layout

```
packages/
  backend/
    api-gateway/      # Routing, JWT auth enforcement, rate limiting, CORS
    auth/             # Registration, login, JWT issuance, refresh tokens
    accounts/         # Account balances (LWT/paxos updates)
    transfer/         # Transfer saga orchestrator
    transactions/     # Transaction read model (Kafka projection)
  frontend/app/       # React + TypeScript SPA
infrastructure/       # Terraform (AWS/GCP) and Kubernetes manifests
docs/                 # Normative project documentation (start at docs/README.md)
scripts/              # Coverage enforcement and other helpers
```

## Quickstart

```sh
cp .env.example .env   # fill in JWT keys (see comments in the file)
docker compose up --build
```

The frontend is served at <http://localhost:3000>; the API gateway listens on
<http://localhost:8080> (`/healthz`, `/readyz` at the root, API under `/api/v1`).

## Common commands

```sh
make build       # build all backend modules and the frontend
make test        # backend coverage-gated tests + frontend vitest
make typecheck   # go vet + tsc --noEmit
make lint        # gofmt + go vet + frontend ESLint
```

Backend tests require no external services for unit coverage; Cassandra/Kafka
integration paths are exercised in CI (`.github/workflows/ci.yml`).

## Documentation

`docs/` contains the normative specifications: requirements, features, data
model, API specification, UI/UX design, tech stack, testing strategy, and
deployment/operations guides. Read `AGENTS.md` for contributor conventions.
