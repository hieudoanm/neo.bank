# FEATURE-02: Authentication

## Priority

P0 (FR-010..FR-015)

## Description

Users can log in with email and password. The system verifies the password hash using Argon2id, issues JWT access tokens (15 min expiry) and refresh tokens (7 day expiry), enforces rate limiting on failed login attempts, and allows token refresh using valid refresh tokens. Refresh tokens are opaque random strings (not JWTs) and are stored only as SHA-256 hashes in the Auth Service's private keyspace. Logout and password change invalidate existing refresh tokens.

## User-visible behaviour

- Clear login form with email and password fields.
- Success message upon successful login.
- Automatic token refresh when the access token expires (handled by the API client).
- Error messages that do not reveal whether an email exists (generic `INVALID_CREDENTIALS`).
- Rate limiting message after too many failed attempts (with `Retry-After` header).
- Loading state while the authentication request is in flight.
- Redirect to the dashboard after successful login.
- Redirect to login when tokens are expired or invalid.

## Inputs

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| email | string | Yes | Valid email format |
| password | string | Yes | Must match stored Argon2id hash |
| refreshToken | string | Yes (for refresh) | Valid, unexpired, unrevoked refresh token |

## Outputs

On success (`200 OK`):

| Field | Type | Description |
|-------|------|-------------|
| accessToken | string | JWT access token (RS256, 15 min expiry) |
| refreshToken | string | Opaque refresh token (7 day expiry) |
| accountId | string (UUID v7) | Authenticated user's account ID |

On error:

| Status | Code | Condition |
|--------|------|-----------|
| `400` | `VALIDATION_ERROR` | Invalid email format or missing fields |
| `401` | `INVALID_CREDENTIALS` | Invalid email or password (generic, no enumeration) |
| `401` | `TOKEN_EXPIRED` | Access token expired; use refresh endpoint |
| `401` | `INVALID_REFRESH_TOKEN` | Refresh token expired, revoked, or already used |
| `429` | `RATE_LIMITED` | Too many failed attempts (5 per 15 min window) |
| `500` | `INTERNAL_ERROR` | Server error (no internal details exposed) |
| `503` | `SERVICE_UNAVAILABLE` | Database or Auth Service temporarily unavailable |

## Behaviour

### Login

1. **Validate inputs**: Check email format and password presence.
2. **Rate limiting**: Check failed attempt count for this IP (max 5 per 15 min window).
3. **Retrieve user record**: Query Auth Service's private Cassandra store by email.
4. **Resolve the user's account**: Look up the account id for this user in the Account Service (by user id — see [`accounts_by_user`](../05-data-model/account-service/README.md)) so the `accountId` claim can be minted.
5. **Verify password**: Compare provided password against stored Argon2id hash.
6. **On mismatch**: Increment failed attempt counter, return `401 INVALID_CREDENTIALS` (generic).
7. **On match**: Reset failed attempt counter, generate tokens.
8. **Generate tokens**: JWT access token (15 min expiry, RS256 signed, claims: `sub`, `accountId`, `email`, `iat`, `exp`, `jti`) and refresh token (7 day expiry, opaque random string).
9. **Store refresh token**: SHA-256 hash stored in Auth Service's private keyspace.
10. **Return response**: Access token, refresh token, account ID.

### Token Refresh

1. **Validate refresh token**: Look up the SHA-256 hash in the database.
2. **Check expiry and revocation**: Reject if expired or already used.
3. **Issue new access token**: JWT with 15 min expiry.
4. **Optional rotation**: Issue new refresh token, invalidate old one.
5. **Return response**: New access token (and optionally new refresh token).

### Logout

1. **Validate refresh token**: Look up the SHA-256 hash.
2. **Revoke**: Mark the refresh token as revoked in the database.
3. **Return response**: `204 No Content`.

## API Contract

```
POST /api/v1/login
Content-Type: application/json

Request:
{
  "email": "user@example.com",
  "password": "Str0ng!Passw0rd"
}

Response (200):
{
  "accessToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "refresh_opaque_token_...",
  "accountId": "550e8400-e29b-41d4-a716-446655440000"
}

POST /api/v1/refresh
Request:
{
  "refreshToken": "refresh_opaque_token_..."
}

Response (200):
{
  "accessToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "new_refresh_opaque_token_..."
}

POST /api/v1/logout
Request:
{
  "refreshToken": "refresh_opaque_token_..."
}

Response (204 No Content)
```

## States

| State | Description |
|-------|-------------|
| `authenticated` | User is logged in with valid access and refresh tokens |
| `unauthenticated` | User is not logged in or tokens are invalid |
| `rate_limited` | Too many failed login attempts (5 per 15 min window) |
| `token_expired` | Access token has expired; refresh token may still be valid |
| `refresh_token_invalid` | Refresh token is expired, revoked, or already used |

## Edge cases

- **Invalid email or password**: Returns `401 INVALID_CREDENTIALS` without revealing whether the email exists.
- **Rate limit exceeded**: Returns `429` with `Retry-After` header indicating when to retry.
- **Expired access token**: Client automatically calls refresh endpoint using the refresh token.
- **Expired or already-used refresh token**: Returns `401 INVALID_REFRESH_TOKEN`; client must re-authenticate.
- **Concurrent login attempts**: Each attempt is independent; rate limiting applies per IP.
- **Password change**: Invalidates all existing refresh tokens for the account; user must re-login.
- **Database unavailable during login**: Returns `503` with `Retry-After` header.

## Acceptance conditions

- Valid credentials issue JWT access token (15 min) and refresh token (7 days); the refresh token is opaque and stored only as a SHA-256 hash (never in plain text).
- Invalid credentials are rejected without revealing email existence.
- Rate limiting enforces max 5 failed attempts within 15 minutes per IP.
- Access tokens expire after 15 minutes; client must use refresh endpoint.
- Refresh tokens expire after 7 days and can be used to obtain new access tokens.
- Logout or password change invalidates existing refresh tokens.
- Authentication completes within 300 ms (p95) per [`NFR-002`](../01-requirements/non-functional-requirements.md#performance).
- JWTs are signed with RS256; private key rotated every 90 days per [`NFR-032`](../01-requirements/non-functional-requirements.md#security).

## Cross-references

- Functional requirements: [`FR-010..FR-015`](../01-requirements/functional-requirements.md#authentication)
- User story: [`US-002`](../03-user-stories/US-002.md)
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md)
- JWT implementation: [`docs/08-tech-stack/backend/jwt.md`](../08-tech-stack/backend/jwt.md)
- Auth Service: [`docs/04-architecture.md`](../04-architecture.md#43-auth-service)
- Planned test: `test_login` (unit + integration with real Cassandra)