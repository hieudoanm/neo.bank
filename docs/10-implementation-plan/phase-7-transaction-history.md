# NEO Bank — Implementation Plan — Phase 7: Transaction History

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Implement the transaction history endpoint and frontend display with cursor-based pagination. Phase 7 completes the read-side of the banking workflow.

**Week:** 5-6

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Backend endpoint | `GET /transactions` with cursor-based pagination |
| Frontend screen | Transactions list with pagination controls |
| Transaction repository | CQL queries with cursor-based pagination |
| Tests | Unit and integration tests for transactions endpoint and screen |

## Tasks

- [ ] Implement transaction repository (CQL queries with cursor-based pagination).
- [ ] Implement `GET /transactions` endpoint (authenticated, paginated list).
- [ ] Implement Transactions screen (frontend, list with pagination).
- [ ] Implement transaction list component (type, amount, counterparty, date).
- [ ] Implement pagination controls (next cursor, load more).
- [ ] Implement TanStack Query infinite query for transaction fetching.
- [ ] Write unit tests for transaction service (pagination, ordering).
- [ ] Write integration tests for transactions endpoint with real Cassandra.
- [ ] Write frontend tests for transactions screen (Vitest).

## Acceptance Criteria

- Authenticated user can view paginated transaction history.
- Transactions ordered newest first.
- Cursor-based pagination works correctly (stable ordering, no duplicates/skips).
- Each transaction displays type (debit/credit), amount, counterparty, date.
- Pagination completes within 100 ms (p95).
- Page size bounded (maximum 100 items per page).
- Tests pass (unit, integration, frontend).

## Dependencies

Phase 6 (transaction records created by the consumer on transfer completion).

## Testing

- Unit tests: pagination logic, cursor handling, transaction formatting.
- Integration tests: paginated transaction list with real data, cursor stability.
- Frontend tests: transaction list rendering, pagination controls, empty state.
- Security tests: user sees only their own transactions.

## Risks & Considerations

- **Cursor stability:** pagination cursor is opaque and server-generated; never construct client-side.
- **Ordering:** transactions ordered newest first; deterministic ordering for stable pagination.
- **Page size limits:** maximum page size prevents unbounded queries.

## Cross-references

- Feature: [`FEATURE-05`](../02-features/FEATURE-05-Transaction-History.md)
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md#2-service-boundaries-and-api-files)
- TanStack Query: [`docs/08-tech-stack/frontend/core/tanstack-query.md`](../08-tech-stack/frontend/core/tanstack-query.md)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_7_transaction_history` (verifies transaction history deliverables)