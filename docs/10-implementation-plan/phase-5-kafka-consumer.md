# NEO Bank — Implementation Plan — Phase 5: Saga Execution and Balance Commands

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Implement the transfer saga: the **Account Service** command consumer that applies debit/credit balance movements, and the **Transfer Service** saga consumer that advances the workflow and applies compensation. Phase 5 is the heart of the banking workflow — conditional per-account balance updates with idempotency and compensation. There is no standalone "consumer service": each consumer lives inside the service that owns the data it touches.

**Week:** 4-5

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Account Service command consumer | Go consumer in `account-service` that consumes `account.debit.requested` and `account.credit.requested` from the `transfers` topic |
| Lightweight transaction | Paxos conditional update per account: conditional debit (never negative) and credit, as **separate** statements |
| Command idempotency (Account Service) | Record the command `idempotency_key` before applying; a redelivered command must not move funds twice |
| Result events (Account Service) | Emit `account.debited` / `account.debit.failed` / `account.credited` / `account.credit.failed` |
| Transfer Service saga consumer | Go consumer in `transfer-service` that consumes `transfer.initiated`, `account.debited`, `account.debit.failed`, `account.credit.failed`, advances the saga, and publishes the next command |
| Compensation | Transfer Service issues `account.credit.requested` (source reversal) when the destination credit fails |
| Transfer status transitions | Transfer Service moves `pending` → `processing` → `completed`/`failed` and emits `transfer.completed` / `transfer.failed` |
| Transaction projection (owned elsewhere) | Written by the **Transaction Service** in Phase 7, which consumes `account.debited`/`account.credited` and `transfer.completed`/`transfer.failed`. Phase 5 only emits those events |
| Tests | Unit and integration tests with real Kafka and Cassandra |

## Tasks

- [x] Implement the Account Service command consumer (segmentio/kafka-go) for `account.debit.requested` and `account.credit.requested`.
- [x] Implement command idempotency: check/record the command idempotency key in `account_keyspace` before applying the movement.
- [x] Implement the debit as a single conditional (paxos) statement: `... IF balance_cents >= ?`.
- [x] Implement the credit as a separate statement — no cross-account `BATCH` and no distributed transaction.
- [x] Emit result events (`account.debited` / `account.debit.failed` / `account.credited` / `account.credit.failed`) carrying `correlationId`, `causationId`, and `idempotencyKey`.
- [x] Implement the Transfer Service saga consumer: on `transfer.initiated` issue `account.debit.requested`; on `account.debited` issue `account.credit.requested` to the destination.
- [x] Handle `account.debit.failed` (for example insufficient funds): mark the transfer `failed` and emit `transfer.failed`.
- [x] Implement compensation on `account.credit.failed`: issue `account.credit.requested` back to the source (reversal), then mark the transfer `failed` and emit `transfer.failed`.
- [x] Implement transfer status transitions and persist them in `transfer_keyspace`.
- [x] Ensure every consumer is idempotent under at-least-once delivery (duplicate commands and events cause no double movement).
- [x] Configure the Account Service and Transfer Service consumers in Docker Compose (alongside API and frontend).
- [x] Write unit tests for LWT conditions, command idempotency, the saga state machine, and compensation.
- [ ] Write integration tests with real Kafka and Cassandra (`docker-compose`).
- [x] Leave transaction projection to the Transaction Service (Phase 7); add a test asserting this phase emits the events the projection consumes.

## Acceptance Criteria

- Account Service consumes debit/credit commands and applies exactly one balance movement per idempotency key.
- Debits are conditional (balance never negative); insufficient funds yields `account.debit.failed` (not an error crash).
- The credit is a separate statement; there is no cross-account transaction or multi-partition batch.
- Transfer Service advances the saga: debit → credit → `completed`, or failure → compensation → `failed`.
- Compensation reverses the source debit when the destination credit fails, leaving the source balance unchanged.
- Transfer status reaches `completed` or `failed` and the matching terminal event is emitted.
- Duplicate commands/events are handled idempotently (no double debit/credit).
- Events carry `correlationId`/`causationId`/`idempotencyKey`.
- Tests pass (unit, integration with real Kafka/Cassandra).

## Dependencies

Phase 4 (transfer initiation, Kafka producer, pending transfer records).

## Testing

- Unit tests: consumer logic, command idempotency, LWT conditions, saga state machine, compensation.
- Integration tests: end-to-end transfer with the real Account Service and Transfer Service consumers, real Kafka, and real Cassandra.
- Consumer tests assume at-least-once delivery and assert idempotency.
- Failure tests: insufficient funds → `failed`; credit failure → compensation; duplicate delivery → single movement.

## Risks & Considerations

- **Ownership:** only the Account Service mutates balances; only the Transfer Service owns saga state. Do not let one consumer write the other's keyspace.
- **Financial integrity:** LWT conditions must prevent overdrafts; compensation must reverse partial state.
- **Idempotency:** duplicate Kafka events must not cause double debit/credit.
- **Consumer scaling:** bounded by Kafka partition count; more replicas just idle.
- **Poison messages:** malformed events routed to `transfers.dlq`, not blocking the consumer.

## Cross-references

- Feature: [`FEATURE-04`](../02-features/FEATURE-04-Fund-Transfer.md), [`FEATURE-07`](../02-features/FEATURE-07-Idempotency.md)
- Saga pattern: [`docs/04-architecture.md`](../04-architecture.md#32-saga-pattern-for-transfers)
- Async command/event contract: [`docs/06-api-specification.md` §13](../06-api-specification.md#13-asynchronous-command--event-contracts) and [`account-service-asyncapi.yaml`](../06-api-specification/account-service/account-service-asyncapi.yaml)
- Kafka backend: [`docs/08-tech-stack/backend/kafka.md`](../08-tech-stack/backend/kafka.md)
- Cassandra backend: [`docs/08-tech-stack/backend/cassandra.md`](../08-tech-stack/backend/cassandra.md)
- Critical test scenarios: [`docs/09-testing-strategy/08-critical-test-scenarios.md`](../09-testing-strategy/08-critical-test-scenarios.md#3-event-processing)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_5_saga_execution` (verifies saga deliverables are met)
