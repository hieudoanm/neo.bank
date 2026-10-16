# NEO Bank — Technology Stack & Decisions

## 1. Purpose

This document records **which technologies are chosen and why**. It is the decision record: the trade-offs, the alternatives considered, and the rules that follow from the choices.

It does **not** explain how to write code in each technology. That lives in the companion best-practice documents under [`08-tech-stack/frontend/`](./08-tech-stack/frontend/), [`08-tech-stack/backend/`](./08-tech-stack/backend/), and [`08-tech-stack/infrastructure/`](./08-tech-stack/infrastructure/):

**Frontend** — see [`frontend/README.md`](./08-tech-stack/frontend/README.md) for the index and the shared frontend rules (arrow functions only, named exports, strict TypeScript). Documents are grouped into [`core/`](./08-tech-stack/frontend/core/), [`styles/`](./08-tech-stack/frontend/styles/), and [`testing/`](./08-tech-stack/frontend/testing/).

| Frontend best practices                                                     | Covers                                               |
| --------------------------------------------------------------------------- | ---------------------------------------------------- |
| [`frontend/README.md`](./08-tech-stack/frontend/README.md)                  | Index + shared frontend rules                        |
| [`core/README.md`](./08-tech-stack/frontend/core/README.md)                 | Framework group index + rules                        |
| [`core/react.md`](./08-tech-stack/frontend/core/react.md)                   | Components, client state, UI states, accessibility   |
| [`core/react-router.md`](./08-tech-stack/frontend/core/react-router.md)     | Route structure, protected routes, lazy loading      |
| [`core/tanstack-query.md`](./08-tech-stack/frontend/core/tanstack-query.md) | Server state, caching, mutations, polling            |
| [`core/typescript.md`](./08-tech-stack/frontend/core/typescript.md)         | Strict typing, API client, runtime validation        |
| [`core/vite.md`](./08-tech-stack/frontend/core/vite.md)                     | Build config, env vars, aliases, code splitting      |
| [`styles/README.md`](./08-tech-stack/frontend/styles/README.md)             | Token-layer group index + rules                      |
| [`styles/tailwindcss.md`](./08-tech-stack/frontend/styles/tailwindcss.md)   | Design tokens, utility conventions, responsive       |
| [`styles/daisyui.md`](./08-tech-stack/frontend/styles/daisyui.md)           | Theme mapping, component classes, shared UI wrappers |
| [`tools/eslint.md`](./08-tech-stack/frontend/tools/eslint.md)               | Lint rules for TypeScript and React                  |
| [`tools/prettier.md`](./08-tech-stack/frontend/tools/prettier.md)           | Format rules with Tailwind class ordering            |
| [`testing/README.md`](./08-tech-stack/frontend/testing/README.md)           | Test-layer group index + rules                       |
| [`testing/vitest.md`](./08-tech-stack/frontend/testing/vitest.md)           | Component/hook tests, mocking, coverage              |
| [`testing/playwright.md`](./08-tech-stack/frontend/testing/playwright.md)   | End-to-end flows, auth state, CI                     |

**Backend** — see [`backend/README.md`](./08-tech-stack/backend/README.md) for the index and the shared backend rules (financial integrity, idempotency, parameterised CQL).

| Backend best practices                                   | Covers                                                    |
| -------------------------------------------------------- | --------------------------------------------------------- |
| [`backend/README.md`](./08-tech-stack/backend/README.md) | Index + shared backend rules                              |
| [`go.md`](./08-tech-stack/backend/go.md)                 | Structure, layers, errors, config, testing, observability |
| [`gin.md`](./08-tech-stack/backend/gin.md)               | Gin router, middleware, request binding, error responses  |
| [`jwt.md`](./08-tech-stack/backend/jwt.md)               | JWT access/refresh tokens, key rotation                   |
| [`cassandra.md`](./08-tech-stack/backend/cassandra.md)   | Query-first modelling, LWTs, migrations, sessions         |
| [`kafka.md`](./08-tech-stack/backend/kafka.md)           | Topic design, event contract, idempotent consumers, DLQ   |
| [`zookeeper.md`](./08-tech-stack/backend/zookeeper.md)   | Kafka coordination service: config, quorum, health checks |

**Infrastructure** — see [`infrastructure/README.md`](./08-tech-stack/infrastructure/README.md) for the index and the shared infrastructure rules (reproducibility, least privilege, no secrets in the repo).

| Infrastructure best practices                                           | Covers                                                     |
| ----------------------------------------------------------------------- | ---------------------------------------------------------- |
| [`infrastructure/README.md`](./08-tech-stack/infrastructure/README.md)  | Index + shared infrastructure rules                        |
| [`docker.md`](./08-tech-stack/infrastructure/docker.md)                 | Dockerfiles, Compose topology, healthchecks, images        |
| [`kubernetes.md`](./08-tech-stack/infrastructure/kubernetes.md)         | Workloads, probes, scaling, secrets, rollouts, hardening   |
| [`terraform.md`](./08-tech-stack/infrastructure/terraform.md)           | Layout, remote state, modules, secrets, plan/apply         |
| [`github-actions.md`](./08-tech-stack/infrastructure/github-actions.md) | CI/CD workflows, service containers, caching, deploy gates |

**Tools** — see [`tools/README.md`](./08-tech-stack/tools/README.md) for cross-cutting tooling (Ruff for repository Python scripts).

The implementation must remain consistent with the normative documents [`01`–`12`](./README.md). This document and its companions refine the technology decisions; where they disagree with a numbered document, the numbered document governs.

## 2. How Decisions Are Made

1. **Prefer the boring, mature option.** A technology is chosen because it fits the load, the team, and the operational model — not because it is new.
2. **One tool per responsibility.** No two libraries may own the same job; the overlap must be resolved or documented.
3. **Types and validation at the boundaries.** Go structs and TypeScript types make contracts explicit; runtime validation guards every entry point.
4. **Reproducibility by default.** Pinned dependencies, versioned migrations, and infrastructure as code.
5. **Justify every dependency.** Each non-trivial library, service, or runtime needs a documented reason and an assigned responsibility.
6. **Record the alternative.** A decision without a rejected alternative is a preference, not a decision.

Two rules are non-negotiable, and every decision below is subordinate to them:

- **Financial operations never run in web request handlers.** The API validates input and publishes events to Kafka; Kafka consumers perform balance changes and record transactions.
- **Dependency choices require justification.** Every non-trivial addition has a documented reason, an owner, and no overlapping alternative already in the stack.

## 3. Stack at a Glance

| Layer               | Chosen technology                                  | Decision record                         |
| ------------------- | -------------------------------------------------- | --------------------------------------- |
| Frontend framework  | React 18 + TypeScript + Vite                       | [§4.1](#41-react--typescript--vite)     |
| Styling             | Tailwind CSS + DaisyUI                             | [§4.2](#42-tailwind-css--daisyui)       |
| Frontend data layer | TanStack Query                                     | [§4.3](#43-tanstack-query)              |
| Frontend routing    | React Router                                       | [§4.4](#44-react-router)                |
| Frontend icons      | react-icons (Phosphor Icons, `pi`)                 | [§4.6](#46-icons-react-icons--phosphor) |
| Frontend testing    | Vitest + Playwright                                | [§4.5](#45-vitest--playwright)          |
| Package manager     | pnpm (workspaces)                                  | [§13.1](#13-package-manager--git-hooks) |
| Git hooks           | Husky                                              | [§13.2](#13-package-manager--git-hooks) |
| API                 | Go 1.22+, Gin router                               | [§5.1](#51-go--gin)                     |
| Database driver     | gocql                                              | [§5.3](#53-gocql)                       |
| Database            | Apache Cassandra 4.x                               | [§6](#6-cassandra)                      |
| Message queue       | Apache Kafka 3.x + segmentio/kafka-go              | [§7](#7-kafka)                          |
| Authentication      | JWT (RS256) + Argon2id                             | [§8](#8-authentication)                 |
| Containerisation    | Docker + Docker Compose                            | [§9](#9-local-topology--ci)             |
| CI/CD               | GitHub Actions                                     | [§9](#9-local-topology--ci)             |
| Orchestration       | Kubernetes (EKS / GKE)                             | [§10](#10-infrastructure)               |
| Infrastructure      | Terraform (EKS/MSK/Keyspaces, GKE/Cassandra/Kafka) | [§10](#10-infrastructure)               |
| Observability       | Structured logs, request IDs, Prometheus metrics   | [§11](#11-observability)                |

## 4. Frontend Decisions

### 4.1 React + TypeScript + Vite

- **Chosen:** React 18 with TypeScript, bundled by Vite.
- **Considered:** Next.js (with SSR), plain Vite + React.
- **Why:** NEO Bank is an authenticated single-page application. It needs no SEO, no server rendering, and no edge routes, so SSR would add a runtime and a data-fetching model we do not need. Vite gives a fast dev server and a small client build; React 18's concurrent features and ecosystem (Router, TanStack Query, Testing Library) are mature. TypeScript is non-negotiable for modelling transfer states and API contracts.
- **Consequences:** Client-side routing and code splitting are our responsibility; every data screen must implement its own loading/empty/error states.
- **How:** [`react.md`](./08-tech-stack/frontend/core/react.md), [`typescript.md`](./08-tech-stack/frontend/core/typescript.md), [`vite.md`](./08-tech-stack/frontend/core/vite.md).

### 4.2 Tailwind CSS + DaisyUI

- **Chosen:** Tailwind CSS with the DaisyUI component plugin.
- **Considered:** CSS Modules, styled-components, plain Tailwind.
- **Why:** Tailwind makes the design tokens in `07-ui-ux-design.md` the single source of styling and removes dead CSS at build time. DaisyUI layers semantic component classes (`btn`, `card`, `badge`, `alert`) on top, which keeps financial UI states consistent and accessible without long utility chains.
- **Consequences:** Tokens live in one theme; dynamic class names are forbidden; components are the unit of reuse.
- **How:** [`tailwindcss.md`](./08-tech-stack/frontend/styles/tailwindcss.md), [`daisyui.md`](./08-tech-stack/frontend/styles/daisyui.md).

### 4.3 TanStack Query

- **Chosen:** TanStack Query for server state.
- **Considered:** SWR, hand-rolled `fetch` + `useEffect`.
- **Why:** Balance, transfers, and transactions are server state with caching, retries, revalidation, and polling needs. TanStack Query models all of that explicitly and has first-class TypeScript support and a stable mutation/invalidation API. Hand-rolled fetching reimplements caching badly.
- **Consequences:** Access to the API goes through the query layer; components never call `fetch` directly.
- **How:** [`tanstack-query.md`](./08-tech-stack/frontend/core/tanstack-query.md), [`react.md`](./08-tech-stack/frontend/core/react.md).

### 4.4 React Router

- **Chosen:** React Router for client-side routing.
- **Considered:** TanStack Router, hand-rolled view switching.
- **Why:** React Router is the established SPA router with nested/layout routes, protected-route patterns, and lazy loading — everything the authenticated shell needs. Hand-rolled routing reimplements history, params, and redirects.
- **Consequences:** Routes are declared centrally; protected routes redirect unauthenticated users; pages are lazy-loaded.
- **How:** [`react-router.md`](./08-tech-stack/frontend/core/react-router.md).

### 4.5 Vitest + Playwright

- **Chosen:** Vitest with React Testing Library for unit/component tests; Playwright for end-to-end flows.
- **Considered:** Jest, Cypress.
- **Why:** Vitest shares the Vite transform pipeline (one config, fast, native ESM/TS), and Playwright runs real browsers against the full stack with reliable auto-waiting and trace capture. Together they match the frontend layers in [`09-testing-strategy.md`](./09-testing-strategy.md).
- **Consequences:** Component tests assert behaviour at the network boundary (MSW); critical journeys are covered end to end and gated in CI.
- **How:** [`vitest.md`](./08-tech-stack/frontend/testing/vitest.md), [`playwright.md`](./08-tech-stack/frontend/testing/playwright.md).

### 4.6 Icons (react-icons / Phosphor)

- **Chosen:** `react-icons`, using the Phosphor Icons set through the `react-icons/pi` entry point.
- **Considered:** hand-authored inline SVG; a Phosphor-only package (`@phosphor-icons/react`); another icon set (Heroicons, Lucide) through `react-icons`.
- **Why:** Icons are standard UI, not bespoke artwork, so they should come from one maintained set rather than being hand-drawn. `react-icons` exposes tree-shakeable React components (`PiUser`, `PiLockKey`, …) that render `currentColor` and inherit size from `font-size`, mapping directly onto the `icon-xs`…`icon-xl` size tokens and the colour tokens in [`07-ui-ux-design.md`](./07-ui-ux-design.md). Phosphor gives a large, consistent, outline-oriented set with strong coverage of navigation, finance, and status; one library keeps both the bundle and the visual language consistent.
- **Consequences:** Icons are imported from `react-icons/pi` — never hand-authored SVG and never a second icon library. Icon style, sizes, and meanings are fixed by [`iconography.md`](./07-ui-ux-design/system/iconography.md).
- **How:** [`iconography.md`](./07-ui-ux-design/system/iconography.md), [`react.md`](./08-tech-stack/frontend/core/react.md).

## 5. Backend Decisions

### 5.1 Go + Gin

- **Chosen:** Go 1.22+ with the Gin web framework.
- **Considered:** Chi, Echo, Fiber, pure `net/http`.
- **Why:** Gin is the most widely adopted Go web framework, with the largest ecosystem, the most third-party middleware, and the most documentation and community tutorials. For a banking demonstration, that means faster onboarding, more battle-tested security middleware (CORS, recovery, rate limiting, JWT), and easier long-term maintenance. Gin's structured context (`*gin.Context`) makes request binding, validation, and JSON responses explicit, which suits an OpenAPI-driven contract where every endpoint's request/response model is pre-defined. Its performance is production-grade, and it runs cleanly on Kubernetes behind a load balancer.
- **Consequences:** Gin's context and middleware are framework-specific (not `net/http`-native), so handlers couple to `*gin.Context`. Business logic still lives in services; the framework only owns routing, request binding, and response writing.
- **How:** [`go.md`](./08-tech-stack/backend/go.md).

### 5.2 API Framework Comparison

| Framework  | Style                  | Performance | stdlib compat | Middleware         | OpenAPI/codegen | Community / docs |
| ---------- | ---------------------- | ----------- | ------------- | ------------------ | --------------- | ---------------- |
| Gin        | Martini-like           | Very fast   | Partial       | Rich built-in set  | `gin-swagger`   | Largest          |
| Chi        | stdlib-compatible      | Good        | Full          | Built-in/context   | Manual          | Moderate         |
| Echo       | High-performance       | Excellent   | Partial       | Extensive registry | `echo-swagger`  | Large            |
| Fiber      | Express-like, fasthttp | Extreme     | Limited       | fasthttp-specific  | `fiber-swagger` | Growing          |
| `net/http` | Minimal stdlib         | Baseline    | Full          | Manual (stdlib)    | None            | Standard library |

**Why Gin was chosen over Chi (previous choice):** Gin offers the largest community, most middleware, best documentation, and structured request binding/validation that maps cleanly to Neo Bank's OpenAPI-defined per-service contracts. Chi's stdlib purity was a nice fit for a minimal demo, but Gin's ecosystem reduces the amount of custom middleware code needed for CORS, recovery, logging, rate limiting, and JWT auth — all required for a banking surface. Gin's performance is production-grade for the expected load, and its context-based handler model makes validation and JSON responses explicit, matching the requirement that request/response models are pre-defined.

### 5.3 gocql

- **Chosen:** gocql for Cassandra access.
- **Considered:** an ORM/query builder.
- **Why:** gocql exposes CQL directly, which is what this project requires: parameterised queries, explicit consistency levels, and lightweight transactions. An ORM would hide the query shapes that the data model must be designed around.
- **How:** [`go.md`](./08-tech-stack/backend/go.md), [`cassandra.md`](./08-tech-stack/backend/cassandra.md).

## 6. Cassandra

- **Chosen:** Apache Cassandra 4.x for the authoritative financial store.
- **Considered:** PostgreSQL, MySQL, MongoDB.
- **Why:** The workload is write-heavy (every transfer writes a transfer row and two transaction rows) and must scale horizontally with tunable consistency and no single point of failure. Cassandra's partition model fits account-scoped access, and lightweight transactions (paxos) provide the conditional balance update that prevents overdrafts. A relational database would centralise a bottleneck we do not want.
- **Consequences:** Query-first modelling, no joins, denormalised tables, and every balance change goes through a lightweight transaction.
- **How:** [`cassandra.md`](./08-tech-stack/backend/cassandra.md), schema in [`05-data-model.md`](./05-data-model.md).

## 7. Kafka

- **Chosen:** Apache Kafka 3.x with segmentio/kafka-go.
- **Considered:** RabbitMQ, AWS SQS/SNS, Postgres-backed queues.
- **Why:** Transfers must be processed asynchronously, durably, and in order per account. Kafka provides partitioned ordering, replication, replay, and consumer groups out of the box, and maps directly onto the "no financial work in request handlers" rule. segmentio/kafka-go is a pure-Go, lightweight client that matches the Go backend.
- **Consequences:** At-least-once delivery is assumed; correctness comes from idempotent consumers, not the broker. Every event carries an idempotency key.
- **How:** [`kafka.md`](./08-tech-stack/backend/kafka.md), [`zookeeper.md`](./08-tech-stack/backend/zookeeper.md), flow in [`04-architecture.md`](./04-architecture.md).

## 8. Authentication

- **Chosen:** JWT access tokens (RS256, 15 min) with hashed refresh tokens (7 days) and Argon2id password hashing.
- **Considered:** server-side sessions, opaque tokens only, bcrypt.
- **Why:** JWTs keep the API stateless and horizontally scalable; RS256 lets verification happen with a public key only. Refresh tokens are stored hashed so a database leak does not yield usable tokens. Argon2id is the current memory-hard password hash recommended for new systems.
- **Consequences:** Short-lived access tokens, rotating signing keys, and `404` (not `403`) for resources outside the caller's account.
- **How:** [`jwt.md`](./08-tech-stack/backend/jwt.md).

## 9. Local Topology & CI

- **Chosen:** Docker Compose for local development; GitHub Actions for CI.
- **Considered:** a cloud-only dev loop, other CI providers.
- **Why:** A banking stack with Cassandra and Kafka must be reproducible on any machine without cloud credentials. Docker Compose brings up the full event-driven topology with one command: the `api-gateway`, the four Go services (`auth-service`, `account-service`, `transfer-service`, `transaction-service`), the `frontend`, plus `cassandra`, `kafka`, and `zookeeper`. Each service owns its own keyspace; the services interact through the gateway and Kafka topics (`transfers`, `transfers.dlq`). GitHub Actions runs the same checks with real Cassandra and Kafka service containers.
- **Consequences:** Same images across environments; configuration via `.env` with no secrets in the repository. See [`11-deployment.md`](./11-deployment.md) for the topology and [`docker.md`](./08-tech-stack/infrastructure/docker.md) for the Compose service list.
- **How:** [`docker.md`](./08-tech-stack/infrastructure/docker.md), [`github-actions.md`](./08-tech-stack/infrastructure/github-actions.md).

## 10. Infrastructure

- **Chosen:** Kubernetes (EKS on AWS, GKE on GCP) to orchestrate the workloads, with Terraform provisioning the cluster, the data services (MSK and Keyspaces on AWS; Cassandra on GKE and Kafka on GKE on GCP), networking, and IAM.
- **Considered:** serverless containers only (ECS Fargate, Cloud Run) without an orchestrator, manual provisioning, AWS CDK, cloud-specific IaC.
- **Why:** Kubernetes is the deployment target for the API, consumer, and frontend, giving the same declarative, portable workload model on both clouds and the scheduling, probes, autoscaling, and rollout controls the banking workloads need. Terraform keeps the surrounding infrastructure declarative, reviewable, and provider-neutral. Local development continues to use Docker Compose and never depends on the cluster.
- **Consequences:** Workloads are stateless (state stays in Cassandra/Kafka); manifests under `infrastructure/kubernetes/` are the source of truth; state is remote (S3/GCS) and secrets come from the cloud secret manager. See [`11-deployment.md`](./11-deployment.md).
- **How:** [`kubernetes.md`](./08-tech-stack/infrastructure/kubernetes.md), [`terraform.md`](./08-tech-stack/infrastructure/terraform.md), [`docker.md`](./08-tech-stack/infrastructure/docker.md).

## 11. Observability

- **Chosen:** structured JSON logs with request/correlation IDs, plus Prometheus-style metrics.
- **Considered:** unstructured logging, tracing-only.
- **Why:** Financial flows span the API, Kafka, and the consumer, so a correlation ID must travel end to end. Structured logs are queryable; metrics (error rate, latency, consumer lag) drive the alerts in [`11-deployment.md`](./11-deployment.md).
- **How:** [`go.md`](./08-tech-stack/backend/go.md), [`kafka.md`](./08-tech-stack/backend/kafka.md).

## 12. Decision Records

Each row is a decision with its rejected alternative and the reason it was rejected.

| Decision              | Considered                          | Chosen                                            | Rationale                                                                                      |
| --------------------- | ----------------------------------- | ------------------------------------------------- | ---------------------------------------------------------------------------------------------- |
| Database              | PostgreSQL vs Cassandra             | Cassandra                                         | Write-heavy, horizontally scalable, partition by account, tunable consistency.                 |
| Queue                 | RabbitMQ vs Kafka                   | Kafka                                             | Durable, partitioned ordering, replay, cloud integration (MSK).                                |
| Backend language      | Python vs Node.js vs Go             | Go                                                | Static types, concurrency, fast startup, strong Cassandra/Kafka clients.                       |
| Frontend framework    | Next.js vs Vite + React             | Vite + React                                      | Authenticated SPA; no SSR/SEO need; lighter and faster to develop.                             |
| Styling               | CSS Modules vs Tailwind             | Tailwind + DaisyUI                                | Token-driven, purged CSS, semantic component classes for consistent states.                    |
| Data fetching         | SWR vs TanStack Query               | TanStack Query                                    | Mature caching, mutations, TypeScript support, polling.                                        |
| Routing               | TanStack Router vs React Router     | React Router                                      | Nested/layout routes, protected routes, lazy loading out of the box.                           |
| Icons                 | hand-authored SVG vs icon library   | react-icons (Phosphor)                            | One maintained set; tree-shakeable React components mapped to token sizes/colours.             |
| Package manager       | npm / yarn vs pnpm                  | pnpm (root workspaces)                            | Content-addressable store, strict resolution, fast installs, lockfile by default.              |
| Git hooks             | Lefthook vs Husky                   | Husky                                             | Hooks versioned in repo, lifecycle-driven install, integrates with lint-staged.                |
| Frontend testing      | Jest/Cypress vs Vitest/Playwright   | Vitest + Playwright                               | Shares the Vite pipeline; real-browser E2E with auto-waiting and traces.                       |
| Local dev / images    | host installs vs Docker Compose     | Docker + Docker Compose                           | Reproducible topology (Cassandra/Kafka); same images across environments.                      |
| CI                    | other CI vs GitHub Actions          | GitHub Actions                                    | Native to the repo; real Cassandra/Kafka via service containers.                               |
| IaC                   | manual vs CDK vs Terraform          | Terraform                                         | Declarative, reviewable, provider-neutral across AWS and GCP.                                  |
| Orchestration         | ECS Fargate/Cloud Run vs Kubernetes | Kubernetes (EKS/GKE)                              | One portable workload model across both clouds; probes, autoscaling, rollouts.                 |
| Cloud data services   | self-managed vs managed             | MSK/Keyspaces (AWS); Cassandra/Kafka on GKE (GCP) | Managed durability/backups on AWS; proven operators on GKE; no hand-rolled stateful workloads. |
| Idempotency store     | Redis vs Cassandra                  | Cassandra (`idempotency_keys`)                    | Durable across restarts, no extra datastore, survives consumer failure.                        |
| Consistency           | STRONG vs eventual                  | QUORUM (strong-ish)                               | Financial data needs strong consistency; QUORUM balances it with availability.                 |
| Balance updates       | read-modify-write vs LWT            | Lightweight transactions                          | Atomic conditional debit prevents overdrafts; Cassandra has no full ACID.                      |
| Delivered in handlers | financial work in API vs consumer   | Kafka consumer                                    | Keeps requests fast and stateless; processing scales and replays separately.                   |

New significant decisions (for example splitting the consumer into per-event services) must be recorded as an Architecture Decision Record before implementation, per [`04-architecture.md`](./04-architecture.md).

## 13. Package Manager & Git Hooks

### 13.1 pnpm

- **Chosen:** pnpm as the root-level package manager, including workspaces for the frontend app.
- **Considered:** npm, yarn.
- **Why:** pnpm uses a content-addressable store with hard links/symlinks instead of copying packages into every `node_modules`; the project has one frontend app and potentially multiple future packages, so disk usage and install speed stay low and the store is shared across versions. pnpm's strict dependency resolution surfaces missing/hoisted dependencies as errors rather than silently allowing them, matching the "no undeclared dependencies" rule in [`08-tech-stack/frontend/README.md`](./08-tech-stack/frontend/README.md). It ships a lockfile (`pnpm-lock.yaml`) without extra configuration and works with the GitHub Actions `setup-node` + `pnpm/action-setup` flow.
- **Consequences:** All install, run, and test commands use `pnpm`; the lockfile is committed; the root toolchain is referenced from `AGENTS.md`. Husky lifecycle scripts rely on `pnpm install` running `prepare`.
- **How:** [`08-tech-stack/frontend/tools/README.md`](./08-tech-stack/frontend/tools/README.md), `packages/frontend/app/package.json`.

### 13.2 Husky

- **Chosen:** Husky for Git hooks, configured at the repository root.
- **Considered:** Lefthook, simple-link, manual hooks in `.git/hooks`.
- **Why:** Husky installs hooks through the dependency lifecycle (`prepare`), keeps them versioned in the repository rather than in `.git`, and integrates with lint-staged for fast, staged-file-only checks on commit and push. It is the lowest-friction cross-platform option for a TypeScript + Go monorepo where each member installs `pnpm` dependencies.
- **Consequences:** Hooks run formatting, linting, and type checks locally ahead of CI; a bypass requires `git commit --no-verify` (allowed only in emergencies, consistent with the deny-by-default rule). No hook may weaken or skip a quality gate.
- **How:** Root `package.json` `prepare` script and `.husky/` hooks.

## 14. Configuration & Secrets


- Environment variables locally via `.env`; managed secrets in production.
- Required: `APP_ENV`, `JWT_PRIVATE_KEY_PEM`, `JWT_PUBLIC_KEY_PEM`, `CASSANDRA_HOST`, `CASSANDRA_KEYSPACE`, `KAFKA_BROKERS`, `KAFKA_TOPIC_TRANSFERS`, `LOG_LEVEL`, `CORS_ORIGINS`.
- Validate at startup and fail clearly on missing required configuration.
- `.env.example` holds placeholders only. Never commit secrets.

## 15. Definition of Done (Tech Stack)

1. Frontend and backend build, test, typecheck, and lint independently in CI.
2. Cassandra migrations are applied and tested; every query is parameterised and account-scoped.
3. No financial operation runs in a web request handler; consumers are idempotent.
4. All amounts are integers in minor units; balance updates use lightweight transactions.
5. Authentication uses JWT (RS256) with Argon2id password hashing and hashed refresh tokens.
6. Every mutating operation enforces an idempotency key.
7. Local development is reproducible through Docker Compose; no secrets in the repository.
8. Terraform is formatted, validated, and planned in CI; state is remote, encrypted, and locked.
9. Kubernetes manifests are validated in CI; workloads define probes and resource limits and run as non-root with images pinned by digest.
10. Each selected technology has a decision record, an assigned responsibility, and a best-practice document under `frontend/`, `backend/`, or `infrastructure/` that it points to.
