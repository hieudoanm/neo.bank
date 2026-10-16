# NEO Bank — Test Metrics

## Purpose

Defines coverage targets, flakiness policy, duration thresholds, and metrics collection for all tests in NEO Bank.

## 1. Coverage Targets

| Layer       | Backend Minimum | Backend Target         | Frontend Minimum | Frontend Target      |
| ----------- | --------------- | ---------------------- | ---------------- | -------------------- |
| Unit tests  | 80%             | 90% (critical paths)   | 80%              | 90% (critical paths) |
| Integration | N/A             | 100% of critical paths | N/A              | N/A                  |
| E2E         | N/A             | Core journeys only     | N/A              | Core journeys only   |

Critical paths include: authentication, balance inquiry, transfer initiation, idempotency, authorisation.

### Coverage Measurement

- Backend: `scripts/coverage/backend.sh <module> [threshold]` (or `--all`) runs `go test -coverprofile` for each service module and fails if its statement coverage is below the threshold (default 80%). Wired into `make test` and the CI `backend` job.
- Frontend: Vitest with V8 provider (`provider: 'v8'`), thresholds in `vitest.config.ts`.
- Reports uploaded to Codecov/Coveralls as CI artifacts.
- Coverage is enforced as a gate — PRs that reduce coverage below the threshold are rejected.

### Exclusions

Test files, generated code, and `main()` entry points are excluded from coverage. The backend gate additionally excludes packages that cannot be exercised without a live Cassandra, from the denominator:

- `cmd/...` — thin `main()` entry points (dependency injection only)
- `internal/db/...` — embedded CQL schema and connection wiring
- `internal/repositories/` — Cassandra CQL queries and row mapping

```go
// main.go is excluded (thin entry point, no logic to test)
func main() { /* dependency injection only */ }
```

## 2. Flakiness Policy

| Policy            | Detail                                                                      |
| ----------------- | --------------------------------------------------------------------------- |
| Zero tolerance    | Flaky tests are not allowed in CI                                           |
| Fix or disable    | Flaky tests must be fixed within 1 sprint or disabled                       |
| Documented reason | Disabled tests need a comment explaining why and a fix ticket               |
| No silent skips   | `testing.Short()` skips are deliberate; no silent `t.Skip()` without reason |
| Quarantine        | Repeatedly flaky tests are moved to a quarantine directory                  |

### Detecting Flakiness

- CI reports test duration and pass/fail history.
- Tests that pass locally but fail in CI are investigated.
- Timing-dependent tests (sleep-based assertions) are converted to event-based assertions.

### Common Flakiness Sources

- Fixed `sleep` instead of polling/waiting on state.
- Race conditions in concurrent transfer tests.
- Kafka consumer offset assumptions.
- Cassandra eventual consistency (read your own writes).
- Browser timing in Playwright (use web-first assertions).

## 3. Duration Thresholds

| Test Type             | Target       | Maximum    | Action if Exceeded               |
| --------------------- | ------------ | ---------- | -------------------------------- |
| Unit tests (backend)  | < 30 seconds | 2 minutes  | Profile and optimise             |
| Unit tests (frontend) | < 30 seconds | 2 minutes  | Profile and optimise             |
| Integration tests     | < 2 minutes  | 5 minutes  | Reduce test data or parallelise  |
| E2E tests             | < 5 minutes  | 10 minutes | Split into smaller specs         |
| Total CI pipeline     | < 10 minutes | 20 minutes | Parallelise or remove slow tests |

## 4. Metrics Collected

| Metric              | Source | Purpose                             |
| ------------------- | ------ | ----------------------------------- |
| Test pass rate      | CI     | Trend over time; detect regressions |
| Test duration       | CI     | Detect slowdowns                    |
| Coverage percentage | CI     | Ensure coverage doesn't decrease    |
| Flaky test count    | CI     | Track and reduce                    |
| E2E failure rate    | CI     | Detect environment issues           |

## 5. Reporting

- Coverage reports published to CI artifacts (HTML + text).
- Test results published in JUnit XML format.
- Flaky test alerts sent to the team channel.
- Performance regression alerts when test duration increases > 20%.

## Cross-references

- Testing strategy: [`09-testing-strategy.md`](../09-testing-strategy.md#2-test-layers)
- Backend testing: [`docs/08-tech-stack/backend/testing.md`](../08-tech-stack/backend/testing.md)
- Frontend testing: [`docs/08-tech-stack/frontend/testing/vitest.md`](../08-tech-stack/frontend/testing/vitest.md)
- Non-functional requirements: [`NFR-050..053`](../01-requirements/non-functional-requirements.md#observability)
- Planned test: `test_metrics` (verifies coverage thresholds are met)