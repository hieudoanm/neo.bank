# NEO Bank — Definition of Done (Testing)

## Purpose

Defines when a feature is considered complete from a testing perspective. A feature is not "done" until all criteria below are met.

The single canonical Definition of Done lives in [`docs/10-implementation-plan/definition-of-done.md`](../10-implementation-plan/definition-of-done.md). This document extends it with the testing-specific criteria that feed into it; where the two disagree, the canonical document governs.

## Criteria

A feature is complete when:

| #   | Criterion                              | Detail                                                                           |
| --- | -------------------------------------- | -------------------------------------------------------------------------------- |
| 1   | Unit tests written and passing         | Every new function/service has unit tests covering success and failure paths.    |
| 2   | Integration tests written and passing  | Features touching Cassandra/Kafka have integration tests with real dependencies. |
| 3   | E2E tests written and passing          | Critical user journeys have Playwright E2E tests.                                |
| 4   | Security tests written and passing     | Authorisation (negative), authentication, and input validation tests pass.       |
| 5   | Performance tests meet targets         | Latency and throughput targets from `NFR-001..006` are met.                      |
| 6   | Code coverage meets minimum thresholds | Backend ≥ 80%, Frontend ≥ 80% (no decrease from baseline).                       |
| 7   | No flaky tests in CI                   | All tests pass consistently; no flaky tests in the suite.                        |
| 8   | Test documentation updated             | Critical scenarios documented; traceability to FR/NFR IDs.                       |
| 9   | Migrations tested                      | CQL migrations apply cleanly and are reversible where possible.                  |
| 10  | Observability verified                 | Structured logs include correlation IDs; metrics are emitted.                    |

## Feature-Level DoD Examples

### Account Registration Feature DoD

- [ ] Unit test: password hashing, email validation, duplicate email check.
- [ ] Integration test: register → login flow with real Cassandra.
- [ ] E2E test: register → login → view dashboard.
- [ ] Security test: invalid credentials rejected without revealing email existence.
- [ ] Performance test: registration < 500 ms (p95).
- [ ] Coverage: Auth Service handlers and services ≥ 80%.

### Fund Transfer Feature DoD

- [ ] Unit test: transfer validation (insufficient funds, same account, invalid amount).
- [ ] Integration test: transfer initiation → Kafka → consumer → balance update.
- [ ] E2E test: transfer → monitor status → view transaction history.
- [ ] Security test: unauthorised transfer returns 401; cross-account returns 404.
- [ ] Idempotency test: duplicate idempotency key returns same transfer ID.
- [ ] Performance test: transfer initiation < 200 ms (p95), processing < 2 s (p95).
- [ ] Coverage: Transfer Service and consumer ≥ 80%.

### Balance Inquiry Feature DoD

- [ ] Unit test: balance lookup, currency formatting.
- [ ] Integration test: authenticated balance lookup, cross-account returns 404.
- [ ] E2E test: login → view balance on dashboard.
- [ ] Security test: unauthenticated access returns 401.
- [ ] Performance test: balance inquiry < 100 ms (p95).
- [ ] Coverage: Account service ≥ 80%.

## Release-Level Testing DoD

The authoritative release Definition of Done is maintained in [`docs/10-implementation-plan/definition-of-done.md`](../10-implementation-plan/definition-of-done.md#release-level-definition-of-done). A release is testing-ready when:

- All features in the release meet their individual DoD.
- All critical test scenarios (FIN/SEC/EVT/EDG) pass.
- E2E test suite passes in staging environment.
- Performance tests meet targets under load.
- Security scan reports zero high-severity findings.
- Rollback plan is documented and tested.
- Documentation is updated for all user-facing changes.

## Cross-references

- Testing strategy: [`09-testing-strategy.md`](../09-testing-strategy.md#2-test-layers)
- Quality gates: [`04-quality-gates.md`](./04-quality-gates.md)
- Test maintenance: [`10-test-maintenance.md`](./10-test-maintenance.md)
- Canonical Definition of Done: [`docs/10-implementation-plan/definition-of-done.md`](../10-implementation-plan/definition-of-done.md)
- Feature completion: [`docs/02-features.md`](../02-features.md#3-feature-completion-definition)
- Delivery rules: `AGENTS.md`
- Planned test: `test_dod` (verifies all DoD criteria are met before release)