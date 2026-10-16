# FEATURE-08: Error Handling

## Priority

P0

## Description

The system provides clear, actionable error messages without exposing sensitive information. Errors are structured with stable error codes, human-readable messages, and optional field details. All errors include a `requestId` for correlation with server logs. The system logs errors with correlation IDs for debugging but never logs secrets, credentials, or stack traces.

## User-visible behaviour

- Clear error messages explaining what went wrong in plain language.
- Actionable guidance for fixing errors (e.g., "Password must be at least 12 characters").
- Generic error messages for security-sensitive failures (no account enumeration).
- Request IDs displayed to users for support ticket correlation.
- Retry options for transient errors (network, rate limiting, service unavailable).
- No stack traces, SQL, CQL, credentials, or sensitive data shown to users.

## Inputs

Various request inputs that may fail validation:
- Malformed JSON body
- Missing required fields
- Invalid email format
- Weak password
- Invalid UUID format
- Expired or invalid JWT
- Insufficient funds
- Non-existent destination account

## Outputs

All error responses follow the standard error envelope:

| Field | Type | Description |
|-------|------|-------------|
| error.code | string | Stable, machine-readable error code |
| error.message | string | Human-readable error message |
| error.details | object | Optional field-level validation details |
| error.requestId | string | Unique request ID for correlation |

| Status | Code | Condition | Retryable |
|--------|------|-----------|-----------|
| `400` | `VALIDATION_ERROR` | Input validation failed (including malformed pagination cursor) | No |
| `400` | `INVALID_AMOUNT` | Transfer amount not positive | No |
| `400` | `INSUFFICIENT_FUNDS` | Source balance too low | No |
| `400` | `SAME_ACCOUNT` | Source and destination accounts are the same | No |
| `401` | `UNAUTHENTICATED` | Missing or invalid JWT | No (after re-login) |
| `401` | `INVALID_CREDENTIALS` | Wrong email or password | No |
| `401` | `TOKEN_EXPIRED` | Access token expired | Yes (use refresh) |
| `401` | `INVALID_REFRESH_TOKEN` | Refresh token expired/revoked/already used | No |
| `404` | `ACCOUNT_NOT_FOUND` | Account not found or not owned (never `403`) | No |
| `404` | `TRANSFER_NOT_FOUND` | Transfer ID not found | No |
| `409` | `EMAIL_ALREADY_REGISTERED` | Email in use | No |
| `409` | `IDEMPOTENCY_KEY_EXISTS` | Same idempotency key reused with a different payload | No |
| `429` | `RATE_LIMITED` | Rate limit exceeded | Yes (after Retry-After) |
| `500` | `INTERNAL_ERROR` | Unexpected server error | Yes (with caution) |
| `503` | `SERVICE_UNAVAILABLE` | Downstream service unavailable | Yes (with backoff) |

These are the only error codes the system emits; clients can rely on this closed set.

## Behaviour

1. **Validate request inputs**: Check for missing fields, invalid formats, and business rule violations.
2. **Return validation errors**: Include field-specific details for `VALIDATION_ERROR`.
3. **Catch exceptions**: Map internal exceptions to stable error codes (never expose raw exceptions).
4. **Log errors**: Record with correlation ID (`X-Request-ID` or `X-Correlation-ID`).
5. **Return generic messages**: For security-sensitive failures (auth, authorisation), return generic messages without revealing internal state.
6. **Include request ID**: Every response (error or success) includes a `requestId` header or body field.

## Error Envelope Example

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": {
      "password": "Password must be at least 12 characters with mixed case, number, and symbol"
    },
    "requestId": "req_550e8400-e29b-41d4-a716-446655440000"
  }
}
```

```json
{
  "error": {
    "code": "INSUFFICIENT_FUNDS",
    "message": "Insufficient funds for this transfer",
    "details": {
      "availableCents": 1000,
      "requestedCents": 5000
    },
    "requestId": "req_550e8400-e29b-41d4-a716-446655440001"
  }
}
```

## States

| State | Description |
|-------|-------------|
| `validation_error` | Input validation failed; client should correct and retry |
| `authentication_error` | Authentication failed; client should re-authenticate |
| `authorisation_error` | User does not own the resource; surfaced as `404 ACCOUNT_NOT_FOUND`, never `403` |
| `not_found_error` | Resource not found (generic, no enumeration) |
| `conflict_error` | Resource conflict (e.g., duplicate email, idempotency key reused with a different payload) |
| `rate_limit_error` | Rate limit exceeded; client should wait and retry |
| `unexpected_error` | Internal server error; client may retry with backoff |

## Edge cases

- **Same idempotency key reused with a different payload**: `409 IDEMPOTENCY_KEY_EXISTS` with details about the conflict.
- **Database connection failures**: `503 SERVICE_UNAVAILABLE`; client retries with exponential backoff.
- **Kafka connection failures**: `503 SERVICE_UNAVAILABLE`; transfer initiation is rolled back.
- **Unexpected exceptions**: `500 INTERNAL_ERROR`; logged with full stack trace server-side, generic message to client.
- **Malformed JSON body**: `400 VALIDATION_ERROR` with details about the parse failure.
- **Missing required fields**: `400 VALIDATION_ERROR` listing missing fields.

## Acceptance conditions

- Validation errors include field-specific details for programmatic handling.
- Authentication/authorisation errors do not reveal sensitive information (no account enumeration).
- Server errors include `requestId` for correlation with server logs.
- Errors are logged with correlation IDs for distributed tracing.
- No stack traces, SQL, CQL, credentials, or sensitive data in error responses or logs.
- All error codes are stable and documented (clients can rely on them).
- Error messages are actionable and written in plain language.

## Cross-references

- API specification: [`docs/06-api-specification.md`](../06-api-specification.md#5-error-response-format)
- Security: [`FEATURE-06`](FEATURE-06-Security-and-Access-Control.md)
- Architecture: [`docs/04-architecture.md`](../04-architecture.md#8-failure-recovery)
- Failure recovery: [`docs/04-architecture.md`](../04-architecture.md#8-failure-recovery)
- Planned test: `test_error_handling` (every error code has a corresponding test)