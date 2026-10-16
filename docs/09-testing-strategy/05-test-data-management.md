# NEO Bank — Test Data Management

## Purpose

Defines how test data is stored, isolated, and cleaned up across all test layers. Test data is never production data — it is synthetic, isolated, and cleaned up after every test run.

## 1. Test Keyspaces

Tests preserve the strict **one-keyspace-per-service** boundary: each service's tests create and use only that service's own test keyspace. There is no shared `neo_bank_test` keyspace, because a single shared keyspace would let tests read another service's tables — exactly what production forbids.

```sql
CREATE KEYSPACE auth_keyspace_test
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
AND durable_writes = true;

CREATE KEYSPACE account_keyspace_test
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
AND durable_writes = true;

CREATE KEYSPACE transfer_keyspace_test
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
AND durable_writes = true;

CREATE KEYSPACE transaction_keyspace_test
WITH replication = {'class': 'SimpleStrategy', 'replication_factor': 1}
AND durable_writes = true;
```

| Property | Value | Reason |
|----------|-------|--------|
| Keyspace names | `<service>_keyspace_test` | One per service; distinct from production keyspaces |
| Replication strategy | `SimpleStrategy` | Single-node test environment |
| Replication factor | 1 | Test environment has one node |
| Durable writes | `true` | Test data survives restarts within the session |

Cross-service integration tests drive the services through their public APIs (or the gateway); they never open a second service's keyspace.

### Tables per Test Keyspace

- `auth_keyspace_test` — `users`, `users_by_email`, `refresh_tokens`, `refresh_tokens_by_user`, `audit_events`, `schema_migrations`
- `account_keyspace_test` — `accounts`, `accounts_by_user`, `processed_commands`, `audit_events`, `schema_migrations`
- `transfer_keyspace_test` — `transfers`, `transfers_by_account`, `idempotency_keys`, `audit_events`, `schema_migrations`
- `transaction_keyspace_test` — `transactions`, `audit_events`, `schema_migrations`

## 2. Test Kafka Topic

A separate Kafka topic is used for all tests:

```bash
# Topic name
transfers-test

# Configuration
partitions: 3
replication.factor: 1
retention.ms: 86400000  # 24 hours
```

### Consumer Group

- Consumer group ID: `neo-bank-test-consumer`
- Offset reset: `earliest` (read from beginning for each test)
- Auto-commit: disabled (manual offset control for deterministic tests)

## 3. Test Fixtures vs Production

| Aspect | Test | Production |
|--------|------|------------|
| Data | Synthetic, deterministic | Real user data |
| Keyspace | `<service>_keyspace_test` (one per service) | `auth_keyspace`, `account_keyspace`, etc. |
| Kafka topic | `transfers-test` | `transfers` |
| Retention | 24 hours | 7 days (configurable) |
| Replication | 1 | 3 |
| Cleanup | After every test | Never (immutable financial records) |

## 4. Cleanup Strategy

### Per-Test Cleanup

Each test is responsible for cleaning up its own data:

```go
func TestTransfer_Initiate(t *testing.T) {
    // Arrange: insert test account
    accountID := insertTestAccount(t, 10000)

    // Act + Assert
    // ...

    // Cleanup: truncate or delete test rows
    truncateTestTables(t)
}
```

Methods:

- **Truncate tables after each test** — `TRUNCATE accounts` (one table per statement; Cassandra has no multi-table `TRUNCATE`), scoped to the service's own test keyspace — each test starts clean.
- **Delete specific rows** — `DELETE FROM accounts WHERE account_id = ?` — for targeted cleanup.
- **No database transactions** — Cassandra has no transactions or transactional rollback; reset state by truncating tables or deleting the rows a test created.
- **`defer` guarantees cleanup** — even failed tests trigger cleanup via `defer truncateTestTables(t)`.

### Per-Suite Cleanup

After the full test suite completes:

- Drop and recreate each service's test keyspace (clean slate) — never touch a production keyspace.
- Delete the test Kafka topic and recreate it.
- Reset all consumer offsets.

### Kafka Cleanup

- **Reset consumer offsets** — consumers start from the beginning of `transfers-test` for each test.
- **Purge test topic** — delete all messages after the suite completes.
- **No cross-test contamination** — each test publishes with unique correlation IDs.

## 5. Data Isolation Rules

| Rule | Enforcement |
|------|-------------|
| Each service's test keyspace is separate from production and from other services | Different keyspace name per service; no cross-keyspace queries |
| Test Kafka topic is separate from production | Different topic name; separate consumer group |
| No test data leaks to production | Test environment has no network path to production |
| Cleanup is automatic | `defer` or `AfterEach` in every test |
| Deterministic fixtures | Same seed, same result — no random data |
| No real PII | Emails use `example.com`, names are synthetic |

## Cross-references

- Testing strategy: [`09-testing-strategy.md`](../09-testing-strategy.md#2-test-layers)
- Test fixtures: [`02-test-fixtures.md`](./02-test-fixtures.md)
- Backend testing: [`docs/08-tech-stack/backend/testing.md`](../08-tech-stack/backend/testing.md)
- Cassandra config: [`docs/05-data-model.md`](../05-data-model.md#6-cassandra-configuration)
- Kafka config: [`docs/08-tech-stack/backend/kafka.md`](../08-tech-stack/backend/kafka.md)
- Planned test: `test_data_cleanup` (verifies tables are empty after test suite)