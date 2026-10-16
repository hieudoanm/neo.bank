# FEATURE-03: Balance Inquiry

## Priority

P0 (FR-020..FR-024)

## Description

Authenticated users can view their current account balance. The system returns the balance in minor units (cents) with currency code, account ID, and last updated timestamp. Balance requests require valid JWT authentication. The Account Service is the authoritative source for balance data.

## User-visible behaviour

- Clear display of current balance with currency symbol (e.g., $1,234.56).
- Display of account ID and last updated timestamp.
- Loading state (skeleton or spinner) while fetching balance from the API.
- Error state if the request fails (with retry option for network errors).
- Redirect to login if not authenticated or tokens are expired.
- Auto-refresh of balance after a successful transfer completes.

## Inputs

| Field | Type | Location | Constraints |
|-------|------|----------|-------------|
| Authorization | string (Bearer JWT) | Header | Required, valid access token |
| accountId | string (UUID v7) | Path | Must belong to authenticated user |

## Outputs

On success (`200 OK`):

| Field | Type | Description |
|-------|------|-------------|
| accountId | string (UUID v7) | Account identifier |
| balanceCents | integer | Balance in minor units (cents) |
| currencyCode | string | ISO 4217 currency code (e.g., "USD") |
| updatedAt | string (ISO 8601 UTC) | Last balance update timestamp |

On error:

| Status | Code | Condition |
|--------|------|-----------|
| `401` | `UNAUTHENTICATED` | Missing or invalid access token |
| `404` | `ACCOUNT_NOT_FOUND` | Account does not exist (generic, no enumeration) |
| `503` | `SERVICE_UNAVAILABLE` | Account Service temporarily unavailable |

## Behaviour

1. **Validate JWT access token**: Verify signature, expiry, issuer, and audience.
2. **Extract user ID**: From the `sub` claim of the JWT.
3. **Retrieve account record**: Query Account Service's private Cassandra store by user ID.
4. **Verify ownership**: Confirm the requested `accountId` belongs to the authenticated user.
5. **Return balance**: Balance in cents, currency code, account ID, and last updated timestamp.
6. **On ownership mismatch**: Return `404` (not `403`) to avoid account enumeration.

## API Contract

```
GET /api/v1/accounts/{accountId}/balance
Authorization: Bearer <accessToken>

Response (200):
{
  "accountId": "550e8400-e29b-41d4-a716-446655440000",
  "balanceCents": 123456,
  "currencyCode": "USD",
  "updatedAt": "2026-10-15T20:00:00Z"
}

Response (401):
{
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "Authentication required",
    "requestId": "req_..."
  }
}

Response (404):
{
  "error": {
    "code": "ACCOUNT_NOT_FOUND",
    "message": "Account not found",
    "requestId": "req_..."
  }
}
```

## States

| State | Description |
|-------|-------------|
| `loading` | Fetching balance from the API |
| `success` | Balance displayed successfully |
| `error` | Request failed or authentication error |
| `unauthenticated` | User is not logged in |

## Edge cases

- **Invalid or expired access token**: Returns `401 UNAUTHENTICATED`; client redirects to login.
- **Account does not exist**: Returns `404 ACCOUNT_NOT_FOUND` (should not happen for an authenticated user with a valid account).
- **Network error during fetch**: Frontend shows error state with retry option.
- **Concurrent balance updates**: The most recently committed balance is returned (Cassandra QUORUM read).
- **Account Service unavailable**: Returns `503 SERVICE_UNAVAILABLE` with `Retry-After` header.

## Acceptance conditions

- Valid access token returns balance in minor units (cents) with currency code.
- Response includes account ID and last updated timestamp (ISO 8601 UTC).
- Invalid or expired token returns `401` error.
- Non-existent account returns `404 ACCOUNT_NOT_FOUND` (distinct from `401 UNAUTHENTICATED`, so unauthenticated and not-found cases are never conflated).
- Cross-account access (user A querying user B's balance) returns `404`, not `403`.
- Balance inquiry completes within 100 ms (p95) per [`NFR-003`](../01-requirements/non-functional-requirements.md#performance).
- All CQL queries are parameterised; account ID is never interpolated into queries.
- Balance amounts are integers in minor units; no floating-point representation.

## Cross-references

- Functional requirements: [`FR-020..FR-024`](../01-requirements/functional-requirements.md#balance-inquiry)
- User story: [`US-004`](../03-user-stories/US-004.md)
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md)
- Account Service: [`docs/04-architecture.md`](../04-architecture.md#44-account-service)
- Data model: [`docs/05-data-model.md`](../05-data-model.md)
- Security: [`FEATURE-06`](./FEATURE-06-Security-and-Access-Control.md) (account ownership enforcement)
- Planned test: `test_balance` (unit + integration with real Cassandra)