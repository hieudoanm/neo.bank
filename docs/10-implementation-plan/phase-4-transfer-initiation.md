# NEO Bank — Implementation Plan — Phase 4: Transfer Initiation

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Implement the transfer initiation endpoint and frontend form. Phase 4 delivers the first mutating banking operation, publishing events to Kafka for asynchronous processing.

**Week:** 4

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Backend endpoint | `POST /transfers` creating a pending transfer record |
| Frontend screen | Transfer form with validation |
| Transfer repository | CQL queries for transfer records |
| Kafka producer | Publishes `transfer.initiated` events |
| Idempotency | Idempotency key handling for duplicate requests |
| Tests | Unit and integration tests for transfer endpoint and screen |

## Tasks

- [ ] Implement transfer repository (CQL queries: insert, find by ID, find by user).
- [ ] Implement Kafka producer (segmentio/kafka-go) for `transfer.initiated` events.
- [ ] Implement idempotency key repository (CQL queries: insert, find by key).
- [ ] Implement `POST /transfers` endpoint (validation, create pending transfer).
- [ ] Implement transfer validation (sufficient funds, valid destination, positive amount, source != destination).
- [ ] Implement Transfer screen (frontend, form with validation, error states).
- [ ] Implement transfer form with validation (amount, destination, idempotency key).
- [ ] Implement TanStack Query mutation for transfer submission.
- [ ] Write unit tests for transfer service (validation, idempotency).
- [ ] Write integration tests for transfer endpoint with real Cassandra/Kafka.
- [ ] Write frontend tests for transfer screen (Vitest).

## Acceptance Criteria

- Authenticated user can initiate a transfer to another account.
- Transfer validated: sufficient funds, valid destination, positive amount, source != destination.
- Transfer record created with `pending` status.
- Transfer event published to Kafka (`transfer.initiated`).
- Idempotency key handling: duplicate key returns existing transfer, no double-processing.
- Transfer initiation completes within 200 ms (p95).
- Tests pass (unit, integration, frontend).

## Dependencies

Phase 3 (authentication, account lookup for balance validation).

## Testing

- Unit tests: transfer validation, idempotency key logic, amount formatting.
- Integration tests: transfer initiation → Kafka publish → pending record, real Cassandra/Kafka.
- Frontend tests: transfer form validation, submission, loading states.
- Security tests: cross-account transfer rejected, insufficient funds rejected.

## Risks & Considerations

- **Idempotency:** duplicate requests with the same key must not create duplicate transfers.
- **At-least-once delivery:** Kafka producer may retry; consumer must handle duplicates (Phase 5).
- **Validation order:** validate before creating any record to avoid orphaned pending transfers.

## Cross-references

- Feature: [`FEATURE-04`](../02-features/FEATURE-04-Fund-Transfer.md)
- User story: [`US-005`](../03-user-stories/US-005.md)
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md#2-service-boundaries-and-api-files)
- Kafka backend: [`docs/08-tech-stack/backend/kafka.md`](../08-tech-stack/backend/kafka.md)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_4_transfer_initiation` (verifies transfer initiation deliverables)