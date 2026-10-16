# FEATURE-01: Account Registration

## Priority

P0 (FR-001..FR-006)

## Description

Users can register a new bank account with email and password. The system validates the email format, enforces password strength, checks for duplicate emails, hashes passwords with Argon2id, and creates an account with a unique account ID (UUID v7) and initial balance of 0. Upon successful registration, the user is automatically logged in with a JWT access token and refresh token.

## User-visible behaviour

- Clear registration form with email and password fields.
- Real-time validation feedback for email format and password strength.
- Success message with account ID upon successful registration.
- Automatic login with session token after registration.
- Distinct errors for invalid input vs duplicate email: `400 VALIDATION_ERROR` and `409 EMAIL_ALREADY_REGISTERED`, each with a generic message that leaks no internal details.
- Loading state while the registration request is in flight.
- Redirect to the dashboard after successful registration.

## Inputs

| Field | Type | Required | Constraints |
|-------|------|----------|-------------|
| email | string | Yes | Valid email format (RFC 5322), max 254 chars |
| password | string | Yes | Min 12 chars, at least one uppercase, one lowercase, one number, one symbol |

## Outputs

On success (`201 Created`):

| Field | Type | Description |
|-------|------|-------------|
| userId | string (UUID v7) | Unique user identifier |
| accountId | string (UUID v7) | Unique account identifier |
| accessToken | string | JWT access token (RS256, 15 min expiry) |
| refreshToken | string | Opaque refresh token (7 day expiry) |
| tokenType | string | Always `Bearer` |
| expiresIn | integer | Access token lifetime in seconds (900) |

On error:

| Status | Code | Condition |
|--------|------|-----------|
| `400` | `VALIDATION_ERROR` | Invalid email format or weak password |
| `409` | `EMAIL_ALREADY_REGISTERED` | Email already in use (generic message, no enumeration) |
| `429` | `RATE_LIMITED` | Too many registration attempts from this IP |
| `500` | `INTERNAL_ERROR` | Server error (no internal details exposed) |
| `503` | `SERVICE_UNAVAILABLE` | Database or Auth Service temporarily unavailable (see edge cases) |

## Behaviour

1. **Validate inputs**: Check email format and password strength per policy.
2. **Check for duplicate email**: Query the Auth Service's private Cassandra store. If the email exists, return `409 EMAIL_ALREADY_REGISTERED` with a generic message that reveals no further detail.
3. **Hash password**: Argon2id (memory=64MB, iterations=3, parallelism=4).
4. **Create account record**: Generate UUID v7 account ID, set initial balance to 0, store in Account Service's private keyspace.
5. **Create user record**: Store email and password hash in Auth Service's private keyspace.
6. **Generate tokens**: JWT access token (15 min expiry, RS256 signed) and refresh token (7 day expiry).
7. **Store refresh token**: Hash with SHA-256 and store in Auth Service's private keyspace.
8. **Return response**: Account ID, access token, and refresh token to the client.

## API Contract

```
POST /api/v1/register
Content-Type: application/json

Request:
{
  "email": "user@example.com",
  "password": "Str0ng!Passw0rd"
}

Response (200):
{
  "accountId": "550e8400-e29b-41d4-a716-446655440000",
  "accessToken": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
  "refreshToken": "refresh_opaque_token_..."
}

Response (409):
{
  "error": {
    "code": "EMAIL_ALREADY_REGISTERED",
    "message": "Registration failed",
    "requestId": "req_..."
  }
}
```

## States

| State | Description |
|-------|-------------|
| `valid_input` | Email and password pass all validation checks |
| `invalid_email` | Email format does not match RFC 5322 |
| `weak_password` | Password fails strength requirements |
| `email_exists` | Email is already registered (generic error) |
| `success` | Account created, tokens issued, session established |

## Edge cases

- **Email already registered**: Returns `409 EMAIL_ALREADY_REGISTERED` with a generic message that leaks no internal details.
- **Weak password**: Returns `400` with specific password requirements.
- **Invalid email format**: Returns `400` with field-level validation error.
- **Concurrent registration with same email**: First request succeeds; second receives `409`. Cassandra LWT ensures atomicity.
- **Database connection failure during account creation**: Returns `503` with `Retry-After` header; no partial account created.
- **Argon2id hashing failure**: Returns `500` with generic error; no account created.

## Acceptance conditions

- Valid email and strong password create an account with unique ID (UUID v7) and balance 0.
- Duplicate email is rejected with `409 EMAIL_ALREADY_REGISTERED` (generic message; no internal details leaked).
- Weak password is rejected with clear password requirements.
- Password is hashed with Argon2id before storage (never plain text).
- Session tokens (access + refresh) are issued upon successful registration.
- Registration completes within 500 ms (p95) per [`NFR-001`](../01-requirements/non-functional-requirements.md#performance).
- All CQL queries are parameterised; no SQL injection possible.

## Cross-references

- Functional requirements: [`FR-001..FR-006`](../01-requirements/functional-requirements.md#account-registration)
- User story: [`US-001`](../03-user-stories/US-001.md)
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md)
- JWT implementation: [`docs/08-tech-stack/backend/jwt.md`](../08-tech-stack/backend/jwt.md)
- Architecture: [`docs/04-architecture.md`](../04-architecture.md#43-auth-service)
- Planned test: `test_register` (unit + integration with real Cassandra)