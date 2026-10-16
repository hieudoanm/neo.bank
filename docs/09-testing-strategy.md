# NEO Bank — Testing Strategy

## 1. Purpose

This document defines the testing strategy for NEO Bank, including test layers, fixtures, acceptance criteria, and quality gates.

The testing strategy must ensure:

- Financial data integrity and consistency.
- Security boundaries (authentication, authorisation).
- Idempotency guarantees for all mutating operations.
- Reliability of event-driven processing (Kafka).
- Correctness of balance updates and transaction recording.
- Coverage of critical workflows end-to-end.

## 2. Test Layers

Testing is organised into seven layers, each documented separately:

- [`01-test-layers.md`](./09-testing-strategy/01-test-layers.md) — Unit, Integration, Contract, E2E, Security, Performance, Reliability
- [`02-test-fixtures.md`](./09-testing-strategy/02-test-fixtures.md) — Synthetic data and test seeds
- [`03-test-organization.md`](./09-testing-strategy/03-test-organization.md) — Backend and frontend test structure
- [`04-quality-gates.md`](./09-testing-strategy/04-quality-gates.md) — Pre-commit, CI pipeline, PR requirements
- [`05-test-data-management.md`](./09-testing-strategy/05-test-data-management.md) — Test keyspace, topics, cleanup
- [`06-test-execution.md`](./09-testing-strategy/06-test-execution.md) — Local and CI execution commands
- [`07-test-metrics.md`](./09-testing-strategy/07-test-metrics.md) — Coverage, flakiness, duration targets
- [`08-critical-test-scenarios.md`](./09-testing-strategy/08-critical-test-scenarios.md) — Financial integrity, security, event processing, edge cases
- [`09-test-documentation.md`](./09-testing-strategy/09-test-documentation.md) — Test cases and reports
- [`10-test-maintenance.md`](./09-testing-strategy/10-test-maintenance.md) — Regular review, test debt
- [`11-definition-of-done.md`](./09-testing-strategy/11-definition-of-done.md) — Testing DoD

## 3. Guiding Principles

- Tests assert observable behaviour, not internal implementation.
- Cassandra-dependent tests use a real database, not mocks.
- Kafka-dependent tests use a real Kafka instance, not mocks.
- Account isolation and authorisation are always tested negatively.
- Consumer tests assume duplicate message delivery and idempotency.
- Fixtures are synthetic and deterministic; never commit real financial data.
