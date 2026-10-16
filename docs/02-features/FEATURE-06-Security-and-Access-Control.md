# FEATURE-06: Security and Access Control

## Priority

P0 (FR-023, FR-030, NFR-030..NFR-037)

## Description

The system enforces account ownership for every operation. Users can only access their own accounts. All API endpoints (except register, login, health) require valid JWT authentication. The system uses parameterised CQL queries and prevents injection attacks. CORS is restricted to known frontend origins. Rate limiting is enforced on auth endpoints (100 req/min per IP) and authenticated endpoints (1000 req/min per user). Unauthorised access returns `404` (not `403`) to avoid information leakage.

## User-visible behaviour

- Unauthenticated requests to protected endpoints return `401` error.
- Requests to other users' accounts return `404` error (not `403`, to avoid information leakage).
- Rate-limited requests return `429` error with `Retry-After` header.
- CORS errors for unauthorised origins.
- No stack traces, CQL, credentials, or sensitive data in error responses or logs.

## Inputs

| Field | Type | Location | Constraints |
|-------|------|----------|-------------|
| Authorization | string (Bearer JWT) | Header | Required on all protected endpoints |
| accountId | string (UUID v7) | Path/Body | Must match authenticated user's account |

## Outputs

| Status | Code | Condition |
|--------|------|-----------|
| `401` | `UNAUTHENTICATED` | Missing or invalid JWT token |
| `404` | `ACCOUNT_NOT_FOUND` | Account not found or not owned by user |
| `429` | `RATE_LIMITED` | Rate limit exceeded |

CORS rejections for unauthorised origins do not use the error envelope; the API Gateway rejects the request before handling, and the browser blocks the response. The error envelope is reserved for the canonical error codes only.

## Behaviour

1. **Validate JWT access token**: Verify signature (RS256), expiry, issuer (`iss`), and audience (`aud`) on every protected endpoint.
2. **Extract user ID**: From the `sub` claim of the validated JWT.
3. **Retrieve account ID**: Map user ID to account ID via Auth Service or Account Service.
4. **Verify ownership**: Confirm the requested `accountId` matches the authenticated user's account ID.
5. **Return `404` on mismatch**: Never return `403` for unauthorised account access (prevents account enumeration).
6. **Enforce rate limiting**: Per-IP for auth endpoints; per-user for authenticated endpoints.
7. **Validate CORS origin**: Reject requests from origins not in the allowlist.
8. **Parameterise all CQL queries**: User input is never interpolated into CQL strings.

## API Contract

```
GET /api/v1/accounts/{accountId}/balance
Authorization: Bearer <accessToken>

Response (401 - missing/invalid token):
{
  "error": {
    "code": "UNAUTHENTICATED",
    "message": "Authentication required",
    "requestId": "req_..."
  }
}

Response (404 - not owned by user):
{
  "error": {
    "code": "ACCOUNT_NOT_FOUND",
    "message": "Account not found",
    "requestId": "req_..."
  }
}

Response (429 - rate limited):
{
  "error": {
    "code": "RATE_LIMITED",
    "message": "Too many requests. Retry after 900 seconds.",
    "requestId": "req_..."
  }
}
```

## States

| State | Description |
|-------|-------------|
| `authenticated` | User has a valid, unexpired access token |
| `unauthenticated` | User lacks a valid token (missing, expired, or malformed) |
| `unauthorised` | User does not own the requested account |
| `rate_limited` | User or IP exceeded the rate limit |

## Edge cases

- **Invalid or expired access token**: Returns `401 UNAUTHENTICATED`; no account existence information revealed.
- **User attempts to access another account**: Returns `404 ACCOUNT_NOT_FOUND` (same message as non-existent account).
- **Rate limit exceeded**: Returns `429 RATE_LIMITED` with `Retry-After` header indicating the wait time.
- **CORS violation**: The API Gateway rejects requests from origins not in the allowlist before handling; the browser blocks the response. No `403` is returned for account access, and CORS rejections do not use a canonical error-envelope code.
- **Concurrent requests from same user**: Each request independently validated; rate limits tracked per user/IP.
- **Token revoked on logout/password change**: Subsequent requests with the old token return `401`.
- **Database unavailable during auth check**: Returns `503 SERVICE_UNAVAILABLE`; request is not processed.

## Acceptance conditions

- Protected endpoints require valid JWT; missing or invalid token returns `401`.
- Unauthorised account access returns `404` (not `403`) to prevent information leakage.
- Rate limiting enforced: 100 requests/minute per IP for auth endpoints; 1000 requests/minute per user for authenticated endpoints.
- CORS restricted to the known frontend origin.
- All CQL queries use parameterised queries; no string interpolation of user input.
- No stack traces, CQL, credentials, or sensitive data in error responses or logs.
- Every authorisation path has a negative test (unauthorised access must be rejected).

## Cross-references

- Functional requirements: [`FR-023`](../01-requirements/functional-requirements.md#balance-inquiry), [`FR-030`](../01-requirements/functional-requirements.md#fund-transfer)
- Non-functional requirements: [`NFR-030..NFR-037`](../01-requirements/non-functional-requirements.md#security)
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md#5-error-response-format)
- Security boundaries: [`docs/04-architecture.md`](../04-architecture.md#9-security-boundaries)
- JWT: [`docs/08-tech-stack/backend/jwt.md`](../08-tech-stack/backend/jwt.md)
- Planned test: `test_auth` (negative authorisation tests for every protected endpoint)