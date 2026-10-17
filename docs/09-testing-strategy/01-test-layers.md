# NEO Bank — Test Layers

## Purpose

Defines the seven test layers used across NEO Bank, their scope, coverage targets, and examples.

## 1. Unit Tests

**Purpose:** Test individual functions, services, and repositories in isolation.

**Scope:**

- Backend: Go unit tests for services, repositories, utilities.
- Frontend: Vitest unit tests for components, hooks, utilities.

**Coverage targets:**

- Backend: 80% code coverage minimum.
- Frontend: 80% code coverage minimum.

**Examples:**

- Password hashing with Argon2id.
- JWT token generation and validation.
- Balance calculation logic.
- Idempotency key validation.
- Component rendering and state management.

## 2. Integration Tests

**Purpose:** Test interactions between components with real dependencies.

**Scope:**

- Backend: API handlers with real Cassandra and Kafka.
- Frontend: API client with mocked backend.

**Dependencies:**

- Real Cassandra instance (test keyspace).
- Real Kafka instance (test topic).
- Mocked external services (identity provider, if any).

**Examples:**

- Account registration end-to-end.
- Login and token refresh.
- Transfer initiation and Kafka publishing.
- Kafka consumer processing and balance updates.
- Transaction history retrieval.

## 3. Contract Tests

**Purpose:** Verify API contract between frontend and backend.

**Scope:**

- OpenAPI specification validation.
- Request/response model validation.
- Error response format validation.

**Tools:**

- OpenAPI validator.
- Schema validation tools.

## 4. End-to-End Tests

**Purpose:** Test complete user workflows through the UI.

**Scope:**

- Critical user journeys:
  - Register → Login → View balance → Transfer → View transaction history.
  - Login → Transfer → Monitor status → View completed transfer.
  - Login → View transaction history → Pagination.

**Tools:**

- Playwright for browser automation.
- Test data fixtures.

## 5. Security Tests

**Purpose:** Verify security boundaries and authorisation.

**Scope:**

- Authentication enforcement (unauthorised requests rejected).
- Authorisation enforcement (users cannot access other accounts).
- Idempotency enforcement (duplicate requests handled correctly).
- Rate limiting enforcement.
- Input validation (SQL injection, CQL injection prevention).
- Token security (expired tokens rejected, refresh token invalidation).

**Examples:**

- Attempt to access another account's balance (should return 404).
- Attempt to initiate transfer from another account (should return 404).
- Duplicate transfer request with same idempotency key (should return existing transfer).
- SQL injection attempts (should be rejected).
- CQL injection attempts (should be rejected).

## 6. Performance Tests

**Purpose:** Verify performance requirements under load.

**Scope:**

- API response times (registration, login, balance, transfer initiation).
- Kafka consumer throughput and latency.
- Cassandra query performance.
- Concurrent user handling.

**Targets:**

- Account registration: < 500 ms (p95).
- User login: < 300 ms (p95).
- Balance inquiry: < 100 ms (p95).
- Transfer initiation: < 200 ms (p95).
- Transfer processing: < 2 s (p95).
- 1,000 concurrent users with < 1% error rate.

**Tools:**

- k6 or Locust for load testing.
- Prometheus metrics for monitoring.

## 7. Reliability Tests

**Purpose:** Verify system behavior under failure conditions.

**Scope:**

- Cassandra failure (API returns 503, consumer pauses).
- Kafka failure (API returns 503, partial debit compensated via application-level compensation).
- Consumer failure (message redelivery, idempotency).
- Network partitions.
- Concurrent transfer processing (race conditions).

**Examples:**

- Transfer processing with Cassandra temporarily unavailable.
- Transfer processing with Kafka consumer crash mid-processing.
- Concurrent transfers from same account (lightweight transaction ensures correctness).
