# NEO Bank — Implementation Plan — Phase 8: Security and Idempotency Tests

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Implement comprehensive security and idempotency tests. Phase 8 verifies that financial integrity and security boundaries hold under attack and failure conditions.

**Week:** 6

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Security tests | Unauthorised access, rate limiting, input validation |
| API Gateway rate limiting | Per-IP limits on the public auth routes and per-user (JWT subject) limits on authenticated routes; `429 RATE_LIMITED` in the error envelope with a `Retry-After` header (limits in [`06-api-specification.md` §8](../06-api-specification.md#8-rate-limiting)) |
| Idempotency tests | Duplicate requests, duplicate Kafka events |
| Performance tests | API response times, consumer throughput |
| Security scans | Trivy filesystem and image scanning in CI |
| Tests | Security, idempotency, and performance test suites |

## Tasks

- [x] Write security tests for unauthorised account access (404, not 403).
- [x] Write security tests for SQL/CQL injection prevention.
- [x] Implement API Gateway rate limiting: per-IP buckets on the public auth routes and per-user (JWT subject) buckets on authenticated routes, returning `429 RATE_LIMITED` with `Retry-After` and login lockout after 5 failed attempts in 15 minutes.
- [x] Write security tests for rate limiting (5 failures per 15 minutes).
- [x] Write idempotency tests for duplicate transfer requests (same key).
- [x] Write idempotency tests for duplicate Kafka events (at-least-once).
- [ ] Write performance tests for API endpoints (latency targets).
- [ ] Write performance tests for Kafka consumer (throughput, lag).
- [x] Configure Trivy security scan in CI (filesystem + image).
- [ ] Update CI pipeline with security and performance test gates. (Security gates landed — formatting, `go vet`, and Trivy filesystem + image scans; performance gates pending tasks 30–31.)

## Acceptance Criteria

- Unauthorised account access returns `404` (no information leakage).
- SQL/CQL injection attempts rejected with `400 VALIDATION_ERROR`.
- Rate limiting enforced at the API Gateway on auth (per IP) and authenticated (per user) endpoints.
- Duplicate transfer requests with same idempotency key handled correctly.
- Duplicate Kafka events processed idempotently (no double debit/credit).
- Performance targets met (API response times, consumer throughput).
- Security scans pass (zero high-severity Trivy findings).
- CI pipeline includes security and performance tests as gates.

## Dependencies

Phase 7 (complete read-write flow to test).

## Testing

- Security tests: FIN-001..008, SEC-001..009, EVT-001..006, EDG-001..007 scenarios.
- Idempotency tests: duplicate HTTP requests, duplicate Kafka messages.
- Performance tests: k6 or Locust load tests against NFR targets.
- Security scans: Trivy on filesystem and container images.

## Risks & Considerations

- **Test realism:** security tests must simulate real attack patterns, not just happy paths.
- **Performance baseline:** measure before optimizing; don't over-optimize prematurely.
- **Flakiness:** timing-dependent performance tests must be deterministic or isolated.

## Cross-references

- Critical test scenarios: [`docs/09-testing-strategy/08-critical-test-scenarios.md`](../09-testing-strategy/08-critical-test-scenarios.md)
- Security rules: `AGENTS.md`
- Financial integrity rules: `AGENTS.md`
- Non-functional requirements: [`docs/01-requirements/non-functional-requirements.md`](../01-requirements/non-functional-requirements.md)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_8_security_idempotency` (verifies security/idempotency deliverables)