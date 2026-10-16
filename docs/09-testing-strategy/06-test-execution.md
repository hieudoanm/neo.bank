# NEO Bank — Test Execution

## Purpose

Defines how to run tests locally and in CI, including commands, dependencies, and execution order.

## 1. Local Development

### Backend

```bash
# Unit tests only (fast, no external dependencies)
# Each service is its own module under packages/backend/<service>/ — run per service.
cd packages/backend/<service>
go test -short ./...

# All backend tests including integration
go test ./...

# Integration tests only (requires Cassandra and Kafka)
docker-compose up -d cassandra kafka
go test ./tests/integration/...

# Specific package
go test -v ./internal/services/...

# With coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Frontend

```bash
# Unit tests
cd packages/frontend/app
pnpm test

# Unit tests with coverage
pnpm test

# Watch mode (re-run on file changes)
pnpm test:watch

# E2E tests
pnpm test:e2e

# E2E tests in headed mode (see browser)
pnpm test:e2e --headed

# E2E tests with UI (debug mode)
pnpm test:e2e --ui
```

### Prerequisites

- **Backend:** Go 1.22+, Docker (for integration tests).
- **Frontend:** Node.js, pnpm (version pinned in `package.json`).
- **Integration tests:** Docker Compose with Cassandra and Kafka.

## 2. CI Execution

CI runs on every push and pull request via GitHub Actions.

### 2.1 Backend CI Pipeline

```yaml
# .github/workflows/ci.yml (excerpt)
jobs:
  backend-test:
    runs-on: ubuntu-latest
    services:
      cassandra:
        image: cassandra:4.1
        ports: ["9042:9042"]
      zookeeper:
        image: confluentinc/cp-zookeeper:7.4.0
        env: { ZOOKEEPER_CLIENT_PORT: 2181 }
      kafka:
        image: confluentinc/cp-kafka:7.4.0
        ports: ["9092:9092"]
        env:
          KAFKA_ZOOKEEPER_CONNECT: zookeeper:2181
          KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR: 1
          KAFKA_DEFAULT_REPLICATION_FACTOR: 1
          KAFKA_MIN_INSYNC_REPLICAS: 1
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with: { go-version: '1.22' }
      - run: go test -short ./...
      - run: go test ./tests/integration/...
      - run: go test -coverprofile=coverage.out ./...
      - uses: codecov/codecov-action@v3
        with: { file: coverage.out }
```

Service definitions are identical in [`github-actions.md`](../08-tech-stack/infrastructure/github-actions.md#3-integration-with-real-services) — Kafka runs in ZooKeeper mode (a `zookeeper` service is required) at RF=1 for the single CI broker.

### 2.2 Frontend CI Pipeline

```yaml
# .github/workflows/ci.yml (excerpt)
jobs:
  frontend-test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: pnpm/action-setup@v4
      - run: pnpm install --frozen-lockfile
      - run: pnpm test
      - run: pnpm test:e2e
      - uses: codecov/codecov-action@v3
```

### 2.3 Execution Order

1. **Lint and typecheck** — fast feedback on code quality.
2. **Unit tests** — fast, no external dependencies.
3. **Integration tests** — need real Cassandra/Kafka (service containers).
4. **E2E tests** — need full stack, run last.
5. **Coverage upload** — report to Codecov/Coveralls.
6. **Security scan** — Trivy image/filesystem scan.

## 3. Test Selection by Tag

```bash
# Unit tests only
go test -short -run '^(Test|Benchmark)' ./...

# Integration tests only
go test -run 'Integration' ./tests/integration/...

# Security tests only
go test -run 'Security|Auth|Authz' ./...

# Frontend unit tests
pnpm test --run 'Login|Transfer|Balance'

# Frontend E2E tests
pnpm test:e2e --grep '@critical'
```

## 4. Parallel Execution

- Go tests run in parallel by default (`t.Parallel()` for safe tests).
- Vitest runs test files in parallel.
- Playwright runs spec files in parallel (fullyParallel: true).
- CI runs backend and frontend jobs in parallel.

## 5. Caching

- Go module cache (`go mod download`) cached between runs.
- pnpm store cached between runs.
- Playwright browsers cached between runs.
- Docker images for Cassandra/Kafka cached locally.

## 6. Artifacts

| Artifact               | Purpose                 | Retention |
| ---------------------- | ----------------------- | --------- |
| Coverage report        | Code coverage metrics   | 30 days   |
| Test results XML       | CI test reporting       | 30 days   |
| Playwright HTML report | E2E test details        | 30 days   |
| Test traces            | Failed test diagnosis   | 7 days    |
| Screenshots            | Failed E2E test visuals | 7 days    |

## 7. Troubleshooting

| Problem                  | Solution                                           |
| ------------------------ | -------------------------------------------------- |
| Cassandra not ready      | Wait for healthcheck (`cqlsh` responds)            |
| Kafka consumer lag       | Increase consumer instances or reduce message rate |
| Flaky test               | Fix determinism or disable with documented reason  |
| Coverage below threshold | Add tests for uncovered lines                      |
| E2E timeout              | Increase timeout or check service dependencies     |

## Cross-references

- Testing strategy: [`09-testing-strategy.md`](../09-testing-strategy.md#2-test-layers)
- Backend testing: [`docs/08-tech-stack/backend/testing.md`](../08-tech-stack/backend/testing.md)
- Frontend testing: [`docs/08-tech-stack/frontend/testing/vitest.md`](../08-tech-stack/frontend/testing/vitest.md)
- CI pipeline: [`docs/08-tech-stack/infrastructure/github-actions.md`](../08-tech-stack/infrastructure/github-actions.md)
- Planned test: `test_execution` (verifies all test commands run successfully)