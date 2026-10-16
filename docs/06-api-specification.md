# NEO Bank — API Specification

## 1. Purpose

This document defines the REST API contract for NEO Bank, including endpoints, request/response models, status codes, and error handling.

The API must:

- Use RESTful conventions where appropriate.
- Return consistent error responses.
- Use UUID v7 for all resource identifiers.
- Store and return financial amounts in minor units (cents).
- Require JWT authentication for all protected endpoints.
- Enforce account ownership for all account-scoped operations.
- Support idempotency for all mutating operations.

## 2. Service Boundaries and API Files

NEO Bank is decomposed into independently deployable microservices. Each service owns its OpenAPI specification:

| Service             | OpenAPI File                                                                  | Endpoints                                    | Owns                                                           |
| ------------------- | ----------------------------------------------------------------------------- | -------------------------------------------- | -------------------------------------------------------------- |
| API Gateway         | [`api-gateway-openapi.yaml`](./06-api-specification/api-gateway/api-gateway-openapi.yaml)                 | `/healthz`, `/readyz` (root)                 | Routing, auth validation, rate limiting, CORS, correlation IDs |
| Auth Service        | [`auth-service-openapi.yaml`](./06-api-specification/auth-service/auth-service-openapi.yaml)               | `/register`, `/login`, `/refresh`, `/logout` | Users, credentials, JWT issuance/refresh                       |
| Account Service     | [`account-service-openapi.yaml`](./06-api-specification/account-service/account-service-openapi.yaml)         | `/accounts/{accountId}/balance`              | Accounts, balances                                             |
| Transfer Service    | [`transfer-service-openapi.yaml`](./06-api-specification/transfer-service/transfer-service-openapi.yaml)       | `/transfers` (POST), `/transfers/{transferId}` (GET) | Transfer saga orchestration                                    |
| Transaction Service | [`transaction-service-openapi.yaml`](./06-api-specification/transaction-service/transaction-service-openapi.yaml) | `/accounts/{accountId}/transactions`         | Transaction history read model                                 |

All service OpenAPI files live in [`./06-api-specification/`](./06-api-specification/). Each file is independently versioned and deployable with its service.

## 3. Base URL

- Development: `http://localhost:8080/api/v1`
- Production: `https://api.neo.bank/api/v1`

Gateway health endpoints (`/healthz`, `/readyz`) are served at the root (gateway servers: `http://localhost:8080` and `https://api.neo.bank`). All other endpoints are served under `/api/v1`.

## 4. Authentication

All protected endpoints require a JWT access token in the `Authorization` header:

```
Authorization: Bearer <access_token>
```

Access tokens expire after 15 minutes. Refresh tokens (7 day expiry) can be used to obtain new access tokens via the token refresh endpoint.

## 5. Error Response Format

All error responses follow this structure:

```json
{
  "error": {
    "code": "ERROR_CODE",
    "message": "Human-readable error message",
    "details": {
      "field": "Additional field-specific details"
    },
    "requestId": "uuid-for-correlation"
  }
}
```

### Common Error Codes

| Code                    | HTTP Status | Description                     |
| ----------------------- | ----------- | ------------------------------- |
| `VALIDATION_ERROR`      | 400         | Request validation failed       |
| `INVALID_AMOUNT`        | 400         | Invalid transfer amount         |
| `INSUFFICIENT_FUNDS`    | 400         | Insufficient funds              |
| `SAME_ACCOUNT`          | 400         | Source equals destination       |
| `UNAUTHENTICATED`       | 401         | Missing or invalid auth         |
| `INVALID_CREDENTIALS`   | 401         | Invalid credentials             |
| `TOKEN_EXPIRED`         | 401         | Access token expired            |
| `INVALID_REFRESH_TOKEN` | 401         | Invalid refresh token           |
| `ACCOUNT_NOT_FOUND`     | 404         | Account not found               |
| `TRANSFER_NOT_FOUND`    | 404         | Transfer not found              |
| `EMAIL_ALREADY_REGISTERED` | 409      | Email already registered        |
| `IDEMPOTENCY_KEY_EXISTS`| 409         | Idempotency key reused with different payload |
| `RATE_LIMITED`          | 429         | Rate limit exceeded             |
| `INTERNAL_ERROR`        | 500         | Internal server error           |
| `SERVICE_UNAVAILABLE`   | 503         | Service unavailable             |

## 6. Requirements Traceability

| Requirement | Service     | Endpoints                                        |
| ----------- | ----------- | ------------------------------------------------ |
| FR-001..006 | Auth        | `POST /api/v1/register`                          |
| FR-010..015 | Auth        | `POST /api/v1/login`, `POST /api/v1/refresh`, `POST /api/v1/logout` |
| FR-020..024 | Account     | `GET /api/v1/accounts/{accountId}/balance`        |
| FR-030..043 | Transfer    | `POST /api/v1/transfers`, `GET /api/v1/transfers/{transferId}` |
| FR-050..052 | Transaction | `GET /api/v1/accounts/{accountId}/transactions`   |

## 7. Idempotency

All mutating operations (POST, PUT, DELETE) support idempotency via the `Idempotency-Key` header.

**Idempotency-Key header:**

```
Idempotency-Key: <unique-key>
```

If a request with the same idempotency key is repeated, the API returns the cached response from the original request without re-executing the operation.

Idempotency keys expire after a configurable period (default: 24 hours).

## 8. Rate Limiting

Rate limits are enforced per IP for auth endpoints and per user for authenticated endpoints:

| Endpoint                | Rate Limit                                           |
| ----------------------- | ---------------------------------------------------- |
| POST /register          | 100 req/min per IP; lockout after 5 failed logins/15 min |
| POST /login             | 100 req/min per IP; lockout after 5 failed logins/15 min |
| Authenticated endpoints | 1000 req/min per user                                |

Rate-limited responses include a `Retry-After` header:

```
Retry-After: 60
```

## 9. CORS

CORS is restricted to the known frontend origin:

```
Access-Control-Allow-Origin: https://app.neo.bank
Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS
Access-Control-Allow-Headers: Authorization, Content-Type, Idempotency-Key
Access-Control-Max-Age: 86400
```

## 10. Request IDs

All API responses include a `X-Request-ID` header for correlation:

```
X-Request-ID: 0189a5c7-7b1a-7e3a-8d4f-2c5b6a7c8d9e
```

Clients should include this ID in support requests.

## 11. Validation Rules

### 11.1 Email

- Must be a valid email format
- Maximum length: 255 characters
- Case-insensitive for uniqueness checks

### 11.2 Password

- Minimum length: 12 characters
- Must contain: uppercase letter, lowercase letter, number, symbol
- Maximum length: 128 characters

### 11.3 Account ID

- Must be a valid UUID v7
- Must exist in the accounts table

### 11.4 Amount

- Must be a positive integer
- Must be in minor units (cents)
- Maximum value: 9,999,999,999 cents ($99,999,999.99)

### 11.5 Transfer Status

- Must be one of: pending, processing, completed, failed, reversed

## 12. Security Considerations

- All sensitive operations require JWT authentication
- Access tokens expire after 15 minutes
- Refresh tokens are hashed in the database
- Passwords are hashed with Argon2id
- Account ownership is verified for every operation
- Unauthorised access returns 404 (not 403) to avoid information leakage
- All CQL queries use parameterised queries
- No sensitive data (passwords, tokens) in logs or error responses
- TLS 1.2+ enforced for all HTTP traffic
