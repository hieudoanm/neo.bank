# FEATURE-05: Transaction History

## Priority

P1 (FR-050..FR-052)

## Description

Authenticated users can view their transaction history. The system returns a paginated list of transactions with cursor-based pagination and deterministic ordering (newest first). Each transaction includes transaction ID, type (debit/credit), amount, counterparty account ID, status, and timestamp. The Transaction Service maintains an append-only read model projected from transfer lifecycle events.

## User-visible behaviour

- Transaction list with type, amount, counterparty, status, and timestamp.
- Cursor-based pagination controls (load more, previous page).
- Loading state (skeleton or spinner) while fetching transactions.
- Empty state if no transactions (with guidance to make a first transfer).
- Error state if request fails (with retry option).
- Transactions sorted newest first by default.

## Inputs

| Field | Type | Location | Constraints |
|-------|------|----------|-------------|
| Authorization | string (Bearer JWT) | Header | Required, valid access token |
| accountId | string (UUID v7) | Path | Must belong to authenticated user |
| cursor | string | Query | Optional; opaque pagination cursor |
| limit | integer | Query | Optional; default 20, max 100 |

## Outputs

On success (`200 OK`):

| Field | Type | Description |
|-------|------|-------------|
| transactions | array | List of transaction objects |
| nextCursor | string | Opaque cursor for the next page (null if no more) |
| hasMore | boolean | Whether additional pages exist |

Each transaction object:

| Field | Type | Description |
|-------|------|-------------|
| transactionId | string (UUID v7) | Transaction identifier |
| type | string | `debit` or `credit` |
| amountCents | integer | Amount in minor units (cents) |
| counterpartyAccountId | string (UUID v7) | Other account involved |
| status | string | `completed` or `failed` |
| createdAt | string (ISO 8601 UTC) | Transaction timestamp |

On error:

| Status | Code | Condition |
|--------|------|-----------|
| `401` | `UNAUTHENTICATED` | Missing or invalid access token |
| `404` | `ACCOUNT_NOT_FOUND` | Account does not exist |
| `400` | `VALIDATION_ERROR` | Malformed pagination cursor (details identify the `cursor` field) |
| `503` | `SERVICE_UNAVAILABLE` | Transaction Service temporarily unavailable |

## Behaviour

1. **Validate JWT access token**: Verify signature, expiry, issuer, audience.
2. **Extract user ID**: From the `sub` claim of the JWT.
3. **Retrieve account ID**: Map user ID to account ID via Account Service.
4. **Verify ownership**: Confirm the requested `accountId` belongs to the authenticated user.
5. **Query transactions**: Transaction Service queries its private read model (Cassandra) for transactions belonging to the account.
6. **Apply pagination**: Cursor-based pagination with deterministic ordering (`createdAt DESC, transactionId DESC`).
7. **Return response**: Transaction list, next cursor, and `hasMore` flag.

## API Contract

```
GET /api/v1/accounts/{accountId}/transactions?cursor=<opaque>&limit=20
Authorization: Bearer <accessToken>

Response (200):
{
  "transactions": [
    {
      "transactionId": "550e8400-e29b-41d4-a716-446655440010",
      "type": "debit",
      "amountCents": 5000,
      "counterpartyAccountId": "550e8400-e29b-41d4-a716-446655440001",
      "status": "completed",
      "createdAt": "2026-10-15T20:30:00Z"
    },
    {
      "transactionId": "550e8400-e29b-41d4-a716-446655440009",
      "type": "credit",
      "amountCents": 10000,
      "counterpartyAccountId": "550e8400-e29b-41d4-a716-446655440002",
      "status": "completed",
      "createdAt": "2026-10-15T20:15:00Z"
    }
  ],
  "nextCursor": "opaque_cursor_...",
  "hasMore": true
}

Response (401):
{
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "Authentication required",
    "requestId": "req_..."
  }
}
```

## States

| State | Description |
|-------|-------------|
| `loading` | Fetching transactions from the API |
| `success` | Transactions displayed successfully |
| `empty` | No transactions to display |
| `error` | Request failed or authentication error |

## Edge cases

- **Invalid or expired access token**: Returns `401 UNAUTHENTICATED`; client redirects to login.
- **Empty transaction history**: Returns empty list with `hasMore: false`; frontend shows empty state.
- **Network error during fetch**: Frontend shows error state with retry option.
- **Invalid pagination cursor**: Returns `400 VALIDATION_ERROR` (details identify the `cursor` field); client resets to first page.
- **Concurrent new transactions**: New transactions appear on subsequent pages (deterministic ordering prevents skipping).
- **Transaction Service unavailable**: Returns `503 SERVICE_UNAVAILABLE` with `Retry-After` header.

## Acceptance conditions

- Valid access token returns paginated transaction list.
- Transactions are ordered newest first (`createdAt DESC, transactionId DESC`).
- Each transaction includes transaction ID, type, amount, counterparty, status, and timestamp.
- Pagination uses cursor-based approach with deterministic ordering.
- Empty history returns empty list with clear empty state.
- Cross-account access returns `404` (not `403`) to prevent enumeration.
- Query latency target: 200 ms (p95) for typical page sizes — planned target; no NFR currently defines a latency target for transaction history.
- All amounts are integers in minor units (cents); no floating-point representation.

## Cross-references

- Functional requirements: [`FR-050..FR-052`](../01-requirements/functional-requirements.md#transaction-history)
- User story: [`US-008`](../03-user-stories/US-008.md)
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md)
- Transaction Service: [`docs/04-architecture.md`](../04-architecture.md#46-transaction-service)
- Data model: [`docs/05-data-model.md`](../05-data-model.md)
- Planned test: `test_history` (unit + integration with real Cassandra)