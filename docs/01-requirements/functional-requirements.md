# Functional Requirements

This document specifies the functional requirements for NEO Bank, organized by business capability. Each requirement carries a unique ID (FR-###), a description, and a priority. Requirements are traced to user stories, features, architecture components, and tests in the traceability table at the end. IDs are grouped by capability block and are intentionally sparse; gaps between blocks (for example FR-007..FR-009, FR-016..FR-019) are reserved for future requirements within that capability, so the number of requirements is lower than the highest ID suggests.

## Account Registration

| ID     | Requirement                                                                                                        | Priority |
| ------ | ------------------------------------------------------------------------------------------------------------------ | -------- |
| FR-001 | The system shall allow a user to register a new account with email and password.                                   | Must     |
| FR-002 | The system shall validate email format and enforce password strength (min 12 chars, mixed case, number, symbol).   | Must     |
| FR-003 | The system shall reject registration if the email is already registered.                                           | Must     |
| FR-004 | The system shall create an account with a unique account ID (UUID v7) and initial balance of 0.                    | Must     |
| FR-005 | The system shall hash passwords using Argon2id before storage.                                                     | Must     |
| FR-006 | The system shall return the new user ID, account ID, access token, and refresh token upon successful registration. | Must     |

**Acceptance criteria:**

- Registration endpoint: `POST /api/v1/register` with body `{ email, password }`.
- On success, returns `201 Created` with `{ userId, accountId, accessToken, refreshToken, tokenType, expiresIn }`.
- On duplicate email, returns `409 Conflict` with error envelope `EMAIL_ALREADY_REGISTERED`.
- On invalid email or weak password, returns `400 Bad Request` with error envelope `VALIDATION_ERROR` and field-level validation details.
- Passwords are never stored in plain text; Argon2id hash is stored in the Auth Service's private Cassandra keyspace.

**Rationale:** Registration is the entry point for all banking operations. Strong password requirements and email uniqueness prevent account takeover and duplicate accounts.

**See also:** [`US-001`](../03-user-stories/US-001.md), [`F-001`](../02-features/FEATURE-01-Account-Registration.md), [`Auth Service`](../08-tech-stack/backend/README.md).

---

## Authentication

| ID     | Requirement                                                                                                                                    | Priority |
| ------ | ---------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| FR-010 | The system shall allow a registered user to log in with email and password.                                                                    | Must     |
| FR-011 | The system shall verify the password hash using Argon2id.                                                                                      | Must     |
| FR-012 | The system shall issue a JWT access token (15 min expiry) and an opaque refresh token (7 day expiry) upon successful login.                    | Must     |
| FR-013 | The system shall rate limit auth endpoints to 100 requests/minute per IP and lock the account after 5 failed login attempts within 15 minutes. | Must     |
| FR-014 | The system shall allow token refresh using a valid refresh token.                                                                              | Must     |
| FR-015 | The system shall invalidate refresh tokens on logout or password change.                                                                       | Must     |

**Acceptance criteria:**

- Login endpoint: `POST /api/v1/login` with body `{ email, password }`.
- On success, returns `200 OK` with `{ userId, accountId, accessToken, refreshToken, tokenType, expiresIn }`.
- The access token is a JWT (RS256, 15 min TTL); the refresh token is opaque, stored SHA-256 hashed, with a 7 day TTL.
- On invalid credentials, returns `401 Unauthorized` with error envelope `INVALID_CREDENTIALS`.
- Auth endpoints are rate limited to 100 requests/minute per IP; exceeding the limit returns `429 Too Many Requests` with error envelope `RATE_LIMITED` and a `Retry-After` header.
- After 5 failed login attempts within 15 minutes, the account is locked and the request returns `429 Too Many Requests` with error envelope `RATE_LIMITED` and a `Retry-After` header.
- Refresh endpoint: `POST /api/v1/refresh` with body `{ refreshToken }`.
- Logout endpoint: `POST /api/v1/logout` invalidates the refresh token server-side.

**Rationale:** Short-lived access tokens reduce the window of exposure if a token is stolen. Refresh tokens enable long sessions without re-entering credentials. Rate limiting prevents brute-force attacks.

**See also:** [`US-002`](../03-user-stories/US-002.md), [`US-003`](../03-user-stories/US-003.md), [`F-002`](../02-features/FEATURE-02-Authentication.md), [`JWT`](../08-tech-stack/backend/jwt.md).

---

## Balance Inquiry

| ID     | Requirement                                                                                           | Priority |
| ------ | ----------------------------------------------------------------------------------------------------- | -------- |
| FR-020 | The system shall return the current balance for the authenticated user's account.                     | Must     |
| FR-021 | The balance shall be returned in minor units (cents) with currency code.                              | Must     |
| FR-022 | The system shall include the account ID and last updated timestamp in the response.                   | Must     |
| FR-023 | The system shall reject balance requests without a valid access token.                                | Must     |
| FR-024 | The system shall return 404 if the account does not exist (should not happen for authenticated user). | Must     |

**Acceptance criteria:**

- Balance endpoint: `GET /api/v1/accounts/{accountId}/balance`.
- Response: `{ accountId, balanceCents, currency, updatedAt }` where `updatedAt` is ISO 8601 UTC.
- Balance amounts are integers in minor units (cents); no floating-point representation.
- Unauthenticated requests return `401 Unauthorized` with error envelope `UNAUTHENTICATED` or `TOKEN_EXPIRED`.
- Account must belong to the authenticated user; cross-account access returns `404` (not `403`) to avoid information leakage.

**Rationale:** Balance inquiry is a read-heavy operation. Using integer cents avoids floating-point rounding errors in financial data. Returning `404` for unauthorized account access prevents account enumeration.

**See also:** [`US-004`](../03-user-stories/US-004.md), [`F-003`](../02-features/FEATURE-03-Balance-Inquiry.md), [`F-006`](../02-features/FEATURE-06-Security-and-Access-Control.md), [`Account Service`](../08-tech-stack/backend/README.md).

---

## Fund Transfer

| ID     | Requirement                                                                                                                                                                                                                     | Priority |
| ------ | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- | -------- |
| FR-030 | The system shall allow an authenticated user to initiate a transfer to another account.                                                                                                                                         | Must     |
| FR-031 | The transfer request shall require: destination account ID, amount in minor units (positive), and an `Idempotency-Key` request header. The source account is derived from the authenticated user, never supplied by the client. | Must     |
| FR-032 | The system shall validate that the source account has sufficient funds.                                                                                                                                                         | Must     |
| FR-033 | The system shall validate that the destination account exists.                                                                                                                                                                  | Must     |
| FR-034 | The system shall reject transfers to the same account (source == destination).                                                                                                                                                  | Must     |
| FR-035 | The system shall reject transfers with amount ≤ 0.                                                                                                                                                                              | Must     |
| FR-036 | The system shall publish a transfer event to Kafka with all transaction details.                                                                                                                                                | Must     |
| FR-037 | The system shall return a transfer ID and initial status (pending) immediately.                                                                                                                                                 | Must     |
| FR-038 | The system shall process transfers asynchronously via Kafka consumers.                                                                                                                                                          | Must     |
| FR-039 | The consumer shall apply the source debit and destination credit as separate atomic lightweight transactions (never allowing a negative balance); the saga coordinates them so a failed credit is compensated by reversing the debit. | Must     |
| FR-040 | The consumer shall update the transfer status to completed or failed.                                                                                                                                                           | Must     |
| FR-041 | The consumer shall be idempotent: duplicate events with the same idempotency key shall not cause double debit/credit.                                                                                                           | Must     |
| FR-042 | The system shall provide a transfer status endpoint for polling.                                                                                                                                                                | Must     |
| FR-043 | The system shall record a transaction entry for both source and destination accounts.                                                                                                                                           | Must     |

**Acceptance criteria:**

- Transfer endpoint: `POST /api/v1/transfers` with headers `Authorization: Bearer <accessToken>` and `Idempotency-Key: <key>` (required), and body `{ destinationAccountId, amountCents }`. The source account is derived from the authenticated user; the client cannot supply a source account.
- On initiation, returns `202 Accepted` with `{ transferId, status: "pending" }`.
- Validation errors return `400 Bad Request` with error envelope reason codes (`INSUFFICIENT_FUNDS`, `SAME_ACCOUNT`, `INVALID_AMOUNT`, `ACCOUNT_NOT_FOUND` for a missing destination account).
- Transfer events are published to the `transfers` Kafka topic, keyed by source account ID for ordering.
- The saga orchestrator (Transfer Service) coordinates debit → credit with compensation on failure.
- Status endpoint: `GET /api/v1/transfers/{transferId}` returns `{ transferId, status, amountCents, sourceAccountId, destinationAccountId, createdAt, updatedAt }`.
- Duplicate idempotency keys return the original result without reprocessing.

**Rationale:** Transfers are the core financial operation. Asynchronous processing via Kafka ensures financial operations never block web request handlers. Idempotency keys prevent double-charging on retry. The saga pattern with compensation makes transfers consistent across services (no funds lost or created) without distributed transactions.

**See also:** [`US-005`](../03-user-stories/US-005.md), [`US-006`](../03-user-stories/US-006.md), [`US-007`](../03-user-stories/US-007.md), [`F-004`](../02-features/FEATURE-04-Fund-Transfer.md), [`F-007`](../02-features/FEATURE-07-Idempotency.md), [`Transfer Service saga`](../04-architecture.md#32-saga-pattern-for-transfers).

---

## Transaction History

| ID     | Requirement                                                                                                  | Priority |
| ------ | ------------------------------------------------------------------------------------------------------------ | -------- |
| FR-050 | The system shall return a paginated list of transactions for the authenticated account.                      | Should   |
| FR-051 | Each transaction shall include: ID, type (debit/credit), amount, counterparty account ID, status, timestamp. | Should   |
| FR-052 | The list shall support cursor-based pagination with deterministic ordering (newest first).                   | Should   |

**Acceptance criteria:**

- Transaction history endpoint: `GET /api/v1/accounts/{accountId}/transactions`.
- Query parameters: `cursor` (optional), `limit` (default 20, max 100).
- Response: `{ transactions: [...], nextCursor, hasMore }`.
- Each transaction: `{ id, type, amountCents, counterpartyAccountId, status, createdAt }`.
- Ordering: newest first, deterministic by `(createdAt DESC, id DESC)`.
- Only transactions belonging to the authenticated account are returned.

**Rationale:** Transaction history provides the audit trail for all financial operations. Cursor-based pagination ensures consistent performance as the history grows. Deterministic ordering prevents skipped or duplicated entries across pages.

**See also:** [`US-008`](../03-user-stories/US-008.md), [`F-005`](../02-features/FEATURE-05-Transaction-History.md), [`Transaction Service`](../08-tech-stack/backend/README.md).

---

## Requirements Traceability

| Requirement | User Story             | Feature | Architecture Component            | Test          |
| ----------- | ---------------------- | ------- | --------------------------------- | ------------- |
| FR-001..006 | US-001                 | F-001   | Auth Service, Cassandra           | test_register |
| FR-010..015 | US-002, US-003         | F-002   | Auth Service, JWT                 | test_login    |
| FR-020..024 | US-004                 | F-003   | Account Service                   | test_balance  |
| FR-030..043 | US-005, US-006, US-007 | F-004   | Transfer Service, Kafka, Consumer | test_transfer |
| FR-050..052 | US-008                 | F-005   | Transaction Service               | test_history  |

**Cross-references:**

- Priorities: `Must` = required for MVP; `Should` = required for v1.1.
- IDs are grouped by capability, so the ranges above are sparse by design (for example, FR-007..FR-009 and FR-044..FR-049 are reserved for future requirements in their blocks).
- Additional feature coverage: `FR-023` → [`F-006`](../02-features/FEATURE-06-Security-and-Access-Control.md); `FR-030`, `FR-031`, `FR-041` → [`F-007`](../02-features/FEATURE-07-Idempotency.md).
- Related non-functional requirements: [`NFR-020..024`](../01-requirements/non-functional-requirements.md#consistency-and-durability), [`NFR-030..037`](../01-requirements/non-functional-requirements.md#security).
- Architecture overview: [`docs/04-architecture.md`](../04-architecture.md).
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md).
