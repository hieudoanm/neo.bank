# NEO Bank — Test Fixtures

## Purpose

Defines synthetic data and test seeds used across all test layers. Test data is synthetic, deterministic, and never contains real financial data or personally identifiable information.

## 1. Synthetic Data

All test data is synthetic and deterministic:

- Synthetic accounts with predictable balances.
- Synthetic users with known credentials.
- Synthetic transfers with known states.
- Synthetic transactions with known patterns.

### Account Fixtures

| Field | Value | Description |
|-------|-------|-------------|
| accountId | `acc-test-1` (UUID v7) | Primary test account |
| userId | `user-test-1` (UUID v7) | Primary test user |
| balanceCents | `10000` | $100.00 starting balance |
| currencyCode | `USD` | Currency |
| createdAt | `2026-01-01T00:00:00Z` | Fixed creation timestamp |

### User Fixtures

| Field | Value | Description |
|-------|-------|-------------|
| userId | `user-test-1` (UUID v7) | Primary test user |
| email | `user-test-1@example.com` | Synthetic email |
| passwordHash | `argon2id_hash_...` | Pre-computed hash |
| passwordPlain | `Str0ng!Test123` | Known plaintext (tests only) |

### Transfer Fixtures

| Field | Value | Description |
|-------|-------|-------------|
| transferId | `transfer-test-1` (UUID v7) | Primary test transfer |
| sourceAccountId | `acc-test-1` | Source |
| destAccountId | `acc-test-2` | Destination |
| amountCents | `5000` | $50.00 |
| status | `completed` | Terminal state |
| idempotencyKey | `idem-test-1` | For idempotency tests |

### Transaction Fixtures

| Field | Value | Description |
|-------|-------|-------------|
| transactionId | `txn-test-1` (UUID v7) | Primary test transaction |
| accountId | `acc-test-1` | Account |
| transferId | `transfer-test-1` | Related transfer |
| type | `debit` | Debit from source |
| amountCents | `5000` | $50.00 |
| counterpartyAccountId | `acc-test-2` | Other side |
| createdAt | `2026-01-01T00:00:00Z` | Fixed timestamp |

## 2. Rules

- No real financial data.
- No personally identifiable information (emails are `example.com`, names are synthetic).
- Deterministic and reproducible — same seed, same result every time.
- Clean up after tests (truncate tables, reset offsets).
- Fixtures are version-controlled alongside test code.
- Balance changes in fixtures respect financial integrity rules (non-negative balances, compensated debit/credit pairs).

## 3. Test Seeds

Seed data is loaded before test suites via a seed script:

### Seed Script

```go
// packages/backend/<service>/tests/seeds/seed.go
func SeedTestData(session *gocql.Session) error {
    // Create test accounts
    // Create test users
    // Create test transfers in various states
    // Return error if any insert fails
}
```

### Seed Coverage

- Accounts with initial balances (covering zero, positive, and near-zero balances).
- Users with known credentials (covering valid and invalid login scenarios).
- Existing transfers in various states (pending, completed, failed).
- Idempotency keys (covering used, unused, and expired keys).

### Seed Loading

- Loaded once per test suite (not per test).
- Individual tests truncate and re-insert only the rows they need.
- Seed data is isolated to the test keyspace/topic and never touches production.

## 4. Cleanup

- **Truncate tables after each test** — each test starts with a clean slate.
- **Delete test data after test suite completion** — the seed script provides a cleanup function.
- **Reset Kafka consumer offsets** — consumers start from the beginning of the test topic.
- **Reset state between tests** — Cassandra has no transactions; reset by truncating tables or deleting the specific rows each test created.
- **Never leave partial data** — failed tests still trigger cleanup via `defer` or `AfterEach`.

## Cross-references

- Testing strategy: [`09-testing-strategy.md`](../09-testing-strategy.md#2-test-layers)
- Test data management: [`05-test-data-management.md`](./05-test-data-management.md)
- Backend testing: [`docs/08-tech-stack/backend/testing.md`](../08-tech-stack/backend/testing.md)
- Frontend testing: [`docs/08-tech-stack/frontend/testing/vitest.md`](../08-tech-stack/frontend/testing/vitest.md)
- Planned test: `test_fixtures` (verifies fixtures load cleanly and cleanup works)