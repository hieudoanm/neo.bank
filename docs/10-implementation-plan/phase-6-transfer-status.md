# NEO Bank — Implementation Plan — Phase 6: Transfer Status

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Implement the transfer status endpoint and frontend display. Phase 6 lets users monitor the asynchronous processing of their transfers from `pending` to `completed` or `failed`.

**Week:** 5

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Backend endpoint | `GET /api/v1/transfers/{transferId}` returning transfer status and details |
| Frontend screen | Transfer status screen with polling |
| Status polling | Frontend polls every 2 seconds until terminal state |
| Status badge | Visual indicator for pending/processing/completed/failed/reversed |
| Tests | Unit and integration tests for status endpoint and screen |

## Tasks

- [x] Implement `GET /api/v1/transfers/{transferId}` endpoint (authenticated, returns status).
- [x] Implement Transfer Status screen (frontend, displays status and details).
- [x] Implement status polling (every 2 seconds, stops at terminal state).
- [x] Implement status badge component (color-coded by status).
- [x] Implement failure reason display for failed transfers.
- [x] Verify dashboard balance reflects completed transfers (TanStack Query invalidation) after the Phase 5 consumer completes a transfer.
- [x] Write unit tests for transfer service (status lookup, authorization).
- [ ] Write integration tests for status endpoint with real Cassandra.
- [x] Write frontend tests for status screen (Vitest, polling behavior).

## Acceptance Criteria

- Authenticated user can view transfer status by transfer ID.
- Status transitions from `pending` to `completed` or `failed`.
- Status polls automatically every 2 seconds until terminal state.
- Failed transfers display the failure reason.
- Dashboard balance reflects completed transfers (via TanStack Query invalidation).
- Cross-account access to another user's transfer returns `404` (no leakage).
- Transfer status lookup completes within 100 ms (p95).
- Tests pass (unit, integration, frontend).

## Dependencies

Phase 5 (consumer updates transfer status to terminal state).

## Testing

- Unit tests: transfer service status lookup, authorization checks.
- Integration tests: status endpoint after consumer processing, cross-account 404.
- Frontend tests: status badge rendering, polling interval, terminal state handling.
- Security tests: user cannot view another user's transfer (404, not 403).

## Risks & Considerations

- **Polling overhead:** stop polling at terminal state; use short intervals only while in flight.
- **No leakage:** transfer status for another user's account returns 404, never 403.
- **Stale data:** polling interval balances freshness with API load.

## Cross-references

- Feature: [`FEATURE-04`](../02-features/FEATURE-04-Fund-Transfer.md)
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md#2-service-boundaries-and-api-files)
- Security rules: `AGENTS.md`
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_6_transfer_status` (verifies transfer status deliverables)