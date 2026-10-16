# NEO Bank — Quality Gates

## Purpose

Defines the pre-commit hooks, CI pipeline checks, and pull request requirements that act as quality gates for all changes to NEO Bank.

## 1. Pre-commit

Husky + lint-staged runs on every commit:

### Backend (Go)

| Check         | Command                   | Purpose                                   |
| ------------- | ------------------------- | ----------------------------------------- |
| golangci-lint | `golangci-lint run ./...` | Aggregate linting (style, bugs, security) |
| go vet        | `go vet ./...`            | Static analysis                           |
| gofmt         | `gofmt -l .`              | Formatting check                          |

### Frontend (TypeScript)

| Check      | Command             | Purpose                       |
| ---------- | ------------------- | ----------------------------- |
| ESLint     | `pnpm lint`         | Code style and potential bugs |
| Prettier   | `pnpm format:check` | Formatting                    |
| TypeScript | `pnpm typecheck`    | Type safety (`tsc --noEmit`)  |

Rules:

- Pre-commit runs on staged files only (lint-staged).
- A failing check blocks the commit; the developer must fix before retrying.
- `--no-verify` is discouraged and requires justification in the commit message.

## 2. CI Pipeline

GitHub Actions runs on every push and pull request.

### 2.1 Backend

| Step               | Command                           | Gate                             |
| ------------------ | --------------------------------- | -------------------------------- |
| Checkout           | `actions/checkout`                | Get source                       |
| Setup Go           | `actions/setup-go`                | Install Go 1.22+                 |
| Lint               | `golangci-lint run ./...`         | Must pass                        |
| Vet                | `go vet ./...`                    | Must pass                        |
| Test (unit)        | `go test -short ./...`            | Must pass                        |
| Test (integration) | `go test ./tests/integration/...` | Must pass (real Cassandra/Kafka) |
| Coverage           | `bash scripts/coverage/backend.sh <module>` | Must be ≥ 80% per backend service module (excludes `cmd/`, `internal/db/`, `internal/repositories/`) |
| Security           | `trivy fs .`                      | No high-severity findings        |
| Build              | `go build ./...`                  | Must pass                        |

### 2.2 Frontend

| Step        | Command                          | Gate                       |
| ----------- | -------------------------------- | -------------------------- |
| Checkout    | `actions/checkout`               | Get source                 |
| Setup Node  | `actions/setup-node`             | Install Node + pnpm        |
| Install     | `pnpm install --frozen-lockfile` | Must pass                  |
| Lint        | `pnpm lint`                      | Must pass                  |
| Format      | `pnpm format:check`              | Must pass                  |
| Typecheck   | `pnpm typecheck`                 | Must pass                  |
| Test (unit) | `pnpm test`                      | Must pass (≥ 80% coverage) |
| Test (E2E)  | `pnpm test:e2e`                  | Must pass (Playwright)     |
| Build       | `pnpm build`                     | Must pass                  |

### 2.3 Integration

| Step              | Command                                | Purpose                       |
| ----------------- | -------------------------------------- | ----------------------------- |
| Docker Compose    | `docker-compose up -d cassandra kafka` | Start dependencies            |
| Wait for health   | `wait-for-it` scripts                  | Dependencies ready            |
| Migrations        | `go run ./cmd/migrate`                 | Apply CQL migrations          |
| Integration tests | `go test ./tests/integration/...`      | Test with real deps           |
| Cleanup           | `docker-compose down -v`               | Remove containers and volumes |

## 3. Pull Request Requirements

| Requirement                 | Description                                        |
| --------------------------- | -------------------------------------------------- |
| All tests pass              | Every CI check must be green                       |
| Coverage does not decrease  | New code is covered; no uncovered lines added      |
| No security vulnerabilities | Trivy reports zero high-severity findings          |
| Manual review               | At least one reviewer approves non-trivial changes |
| Documentation updated       | Behavior changes are reflected in `docs/`          |
| Changelog entry             | User-facing changes get a changelog entry          |

## 4. Merge Requirements

- All status checks pass.
- At least one approval from a code owner.
- No merge conflicts.
- Base branch is up to date.
- Squash merge preferred (clean history).

## 5. Rollback Criteria

A deploy is rolled back if:

- Error rate > 1% for 5 minutes.
- p99 latency > 2s for 10 minutes.
- Any critical alert fires.
- Rollback uses `kubectl rollout undo` or previous image tag.

## Cross-references

- Testing strategy: [`09-testing-strategy.md`](../09-testing-strategy.md#2-test-layers)
- Backend lint: [`docs/08-tech-stack/backend/code-style.md`](../08-tech-stack/backend/code-style.md)
- Frontend lint: [`docs/08-tech-stack/frontend/core/typescript.md`](../08-tech-stack/frontend/core/typescript.md)
- CI pipeline: [`docs/08-tech-stack/infrastructure/github-actions.md`](../08-tech-stack/infrastructure/github-actions.md)
- Planned test: `test_quality_gates` (verifies CI config is valid)