# NEO Bank — Implementation Plan — Phase 0: Foundation

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Set up the project structure, tooling, local infrastructure, and CI pipeline that every later phase builds on. Phase 0 has no dependencies — it is the foundation for everything else.

**Week:** 1

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Monorepo structure | pnpm workspace with `packages/frontend/` and Go service modules under `packages/` (`api-gateway`, `auth-service`, `account-service`, `transfer-service`, `transaction-service`) |
| Go service skeletons | `go.mod` per service, `cmd/<service>/main.go`, and `internal/` layout per `AGENTS.md` section 4 |
| React frontend skeleton | Vite + TypeScript + Tailwind + DaisyUI scaffold |
| Docker Compose topology | Cassandra, Kafka, Zookeeper, and healthchecks (service containers are added in later phases) |
| CI pipeline | GitHub Actions with lint, typecheck, unit test skeleton |
| README | Local development setup instructions |
| Git hooks | Husky + lint-staged pre-commit checks |

## Tasks

- [ ] Initialise monorepo with pnpm workspace (`pnpm init`, workspace config).
- [ ] Set up Go service skeletons: `go.mod` per service, `cmd/<service>/main.go`, and `internal/` structure per conventions.
- [ ] Set up React frontend: Vite, TypeScript strict, Tailwind CSS, DaisyUI.
- [ ] Add backend and frontend test skeletons (one passing test per package, one passing Vitest test).
- [ ] Configure Docker Compose: Cassandra 4.x, Kafka, Zookeeper with healthchecks.
- [ ] Set up GitHub Actions CI: lint, typecheck, test skeleton on push/PR. The test skeletons from the previous step must exist first so CI can run them.
- [ ] Configure git hooks: Husky + lint-staged (golangci-lint, gofmt, ESLint, Prettier).
- [ ] Write README with local development setup (clone, install, run).
- [ ] Verify `docker-compose up -d` starts Cassandra and Kafka and passes healthchecks.
- [ ] Verify frontend dev server (`pnpm dev`) and Go service builds (`go build ./...`).

## Acceptance Criteria

- Developer can run `docker-compose up -d` to start Cassandra and Kafka.
- Healthchecks confirm Cassandra (CQL) and Kafka (bootstrap) are ready.
- Developer can run frontend dev server on `localhost:3000`.
- Developer can build and run the Go services (`go build ./...`, `go run ./cmd/<service>/main.go`).
- CI pipeline runs successfully on an empty repo (lint, typecheck, test skeleton pass).
- Pre-commit hooks run on staged files and block non-conforming commits.
- AGENTS.md conventions are followed for directory structure and imports.
- Phase completion is governed by the canonical [Definition of Done](./definition-of-done.md).

## Testing

- Backend unit test skeleton: one passing test per package.
- Frontend unit test skeleton: one passing Vitest test.
- CI validates that the skeleton suite passes on every push.

## Risks & Considerations

- **Cassandra version alignment:** local Compose, CI service containers, and cloud managed services must use compatible versions.
- **Port conflicts:** Cassandra (9042) and Kafka (9092) must not collide with other local services.
- **pnpm version pinning:** lock the pnpm version in `package.json` and CI to avoid drift.

## Dependencies

None.

## Cross-references

- Project structure: `AGENTS.md`
- Backend conventions: `AGENTS.md`
- Frontend conventions: `AGENTS.md`
- CI pipeline: [`docs/08-tech-stack/infrastructure/github-actions.md`](../08-tech-stack/infrastructure/github-actions.md)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_0_foundation` (verifies foundation deliverables are met)