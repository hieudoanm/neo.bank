# NEO Bank — Implementation Plan — Phase 3: Balance Inquiry

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Implement the balance inquiry endpoint and frontend dashboard. Phase 3 delivers the first read-only banking operation on top of authentication.

**Week:** 3

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| Backend endpoint | `GET /api/v1/accounts/{accountId}/balance` returning balance in cents with currency code |
| Frontend screen | Dashboard with balance display |
| Account repository | CQL queries for account lookup by authenticated user |
| Tests | Unit and integration tests for balance endpoint and dashboard |

## Tasks

- [ ] Implement account repository (CQL queries: find account by user ID).
- [ ] Implement `GET /api/v1/accounts/{accountId}/balance` endpoint (authenticated, returns balance) using the authentication middleware owned by Phase 2. The `accountId` comes from the login response (US-002) / the JWT `accountId` claim — there is no `/accounts/me` alias.
- [ ] Implement Dashboard screen (frontend, balance display, loading/empty states).
- [ ] Implement balance display component (currency formatting, minor units).
- [ ] Implement TanStack Query hook for balance fetching (polling, stale time).
- [ ] Write unit tests for account service (balance lookup, formatting).
- [ ] Write integration tests for balance endpoint with real Cassandra.
- [ ] Write frontend tests for dashboard (Vitest).

## Acceptance Criteria

- Authenticated user can view their balance after login.
- Balance returned in minor units (cents) with currency code (e.g., USD).
- Unauthenticated request returns `401 UNAUTHENTICATED`.
- Authorised access to another account's balance returns `404` (no leakage).
- Balance inquiry completes within 100 ms (p95).
- Tests pass (unit, integration, frontend).

## Dependencies

Phase 2 (JWT validation middleware owned by Phase 2 — see [`phase-2-authentication.md`](./phase-2-authentication.md) — and user lookup).

## Testing

- Unit tests: account service, balance formatting, currency conversion.
- Integration tests: authenticated balance lookup, unauthorised access (404).
- Frontend tests: dashboard rendering, loading/empty/error states.
- Security tests: cross-account access returns 404, not 403.

## Risks & Considerations

- **Information leakage:** unauthorised access to another account returns 404, never 403.
- **Balance staleness:** use short stale time and polling for balance changes.
- **Currency formatting:** always display with the correct currency code and minor units.

## Cross-references

- Feature: [`FEATURE-03`](../02-features/FEATURE-03-Balance-Inquiry.md)
- User story: [`US-004`](../03-user-stories/US-004.md)
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md#2-service-boundaries-and-api-files)
- Backend accounts: [`docs/08-tech-stack/backend/structure.md`](../08-tech-stack/backend/structure.md)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_3_balance_inquiry` (verifies balance inquiry deliverables)