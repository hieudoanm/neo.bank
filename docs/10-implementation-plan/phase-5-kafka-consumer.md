# NEO Bank — Implementation Plan — Phase 5: Kafka Consumer

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Implement the Kafka consumer service that processes transfer events asynchronously. Phase 5 is the heart of the banking workflow — atomic debit/credit with idempotency and compensation.

**Week:** 4-5

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Kafka consumer service | Go service consuming `transfer.initiated` events |
| Lightweight transaction | Paxos-based atomic debit/credit on balance updates |
| Idempotency check | Skip already-processed events |
| Transfer status | Update to `completed` or `failed` |
| Transaction recording | Debit and credit entries for both accounts |
| Compensation | Reverse debit on credit failure |
| Tests | Unit and integration tests with real Kafka and Cassandra |

## Tasks

- [ ] Implement Kafka consumer (segmentio/kafka-go) for `transfer.initiated` events.
- [ ] Implement idempotency check (CQL query on idempotency key before processing).
- [ ] Implement lightweight transaction for balance updates (Paxos conditional update).
- [ ] Implement debit operation (balance check and atomic decrement).
- [ ] Implement credit operation (atomic increment on destination account).
- [ ] Implement transaction recording (debit and credit entries).
- [ ] Implement transfer status update (`pending` → `completed` or `failed`).
- [ ] Implement failure handling (insufficient funds, invalid destination, credit failure).
- [ ] Implement compensation (reverse debit if credit fails).
- [ ] Write unit tests for consumer logic (idempotency, LWT, compensation).
- [ ] Write integration tests with real Kafka and Cassandra.
- [ ] Configure consumer in Docker Compose (alongside API and frontend).

## Acceptance Criteria

- Consumer processes `transfer.initiated` events from Kafka.
- Consumer checks idempotency key before processing (skip duplicates).
- Consumer uses lightweight transaction for atomic debit/credit.
- Consumer updates transfer status to `completed` or `failed`.
- Consumer records transaction entries for both debit and credit accounts.
- Consumer handles duplicate messages idempotently (no double debit/credit).
- Compensation reverses partial debit on credit failure.
- Balance never goes negative (enforced by LWT condition).
- Tests pass (unit, integration with real Kafka/Cassandra).

## Dependencies

Phase 4 (transfer initiation, Kafka producer, pending transfer records).

## Testing

- Unit tests: consumer logic, idempotency checks, LWT conditions, compensation.
- Integration tests: end-to-end transfer with real Kafka consumer and Cassandra.
- Consumer tests assume duplicate message delivery (at-least-once semantics).
- Failure tests: credit failure → compensation, insufficient funds → failed status.

## Risks & Considerations

- **Financial integrity:** LWT conditions must prevent overdrafts; compensation must reverse partial state.
- **Idempotency:** duplicate Kafka events must not cause double debit/credit.
- **Consumer scaling:** bounded by Kafka partition count; more replicas just idle.
- **Poison messages:** malformed events routed to dead-letter queue, not blocking the consumer.

## Cross-references

- Feature: [`FEATURE-04`](../02-features/FEATURE-04-Fund-Transfer.md), [`FEATURE-07`](../02-features/FEATURE-07-Idempotency.md)
- Saga pattern: [`docs/04-architecture.md`](../04-architecture.md#32-saga-pattern-for-transfers)
- Kafka backend: [`docs/08-tech-stack/backend/kafka.md`](../08-tech-stack/backend/kafka.md)
- Critical test scenarios: [`docs/09-testing-strategy/08-critical-test-scenarios.md`](../09-testing-strategy/08-critical-test-scenarios.md#3-event-processing)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_5_kafka_consumer` (verifies consumer deliverables are met)