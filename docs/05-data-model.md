# NEO Bank — Data Model

## 1. Purpose

This document defines the data model for NEO Bank, including entities, relationships, constraints, and state transitions. The model is implemented in Apache Cassandra 4.x.

The data model must support:

- Account and user management.
- Secure authentication with hashed passwords and refresh tokens.
- Balance tracking in minor units (cents).
- Transfer lifecycle management (pending, processing, completed, failed).
- Transaction history recording.
- Idempotency guarantees for all mutating operations.
- Audit logging for security-relevant events.

All financial amounts are stored as integers in minor units (cents) to avoid floating-point precision issues.

## 2. Service Data Ownership

Each microservice owns a private Cassandra keyspace. Cross-service direct database access is forbidden; state is shared only via APIs or Kafka events. There is no shared keyspace and no shared audit store: each service records its own `audit_events` table in its own keyspace.

| Service             | Data Model File                                                            | Tables                                                     | Keyspace               |
| ------------------- | -------------------------------------------------------------------------- | ---------------------------------------------------------- | ---------------------- |
| Auth Service        | [`README.md`](./05-data-model/auth-service/README.md)               | `users`, `users_by_email`, `refresh_tokens`, `refresh_tokens_by_user`, `audit_events` | `auth_keyspace`        |
| Account Service     | [`README.md`](./05-data-model/account-service/README.md)         | `accounts`, `audit_events`                                 | `account_keyspace`     |
| Transfer Service    | [`README.md`](./05-data-model/transfer-service/README.md)       | `transfers`, `transfers_by_account`, `idempotency_keys`, `audit_events` | `transfer_keyspace`    |
| Transaction Service | [`README.md`](./05-data-model/transaction-service/README.md) | `transactions`, `audit_events`                             | `transaction_keyspace` |

## 3. Design Principles

1. **Partition by account_id:** All account-related data is partitioned by account_id for data locality and efficient queries. Data that is keyed by transfer but read by account (e.g. transfer history) is served by the `transfers_by_account` lookup table so it too is partitioned by account_id.
2. **Use lightweight transactions:** Balance updates use lightweight transactions (paxos) to ensure atomicity and prevent negative balances. Cassandra LWTs are per-statement; because there is no cross-account transaction, multi-account changes are coordinated by the saga with compensating actions.
3. **Parameterised CQL:** All queries use parameterised CQL to prevent injection.
4. **Idempotency tracking:** All mutating operations track idempotency keys to prevent duplicate operations. The cached response is stored with the key so a repeat request can return the same response.
5. **Audit logging:** All security-relevant events are recorded in an append-only audit log. Each service writes its own `audit_events` table in its own keyspace; there is no shared audit store.
6. **Immutable history:** Transaction records are immutable; corrections create new records.
7. **UUID v7 for identifiers:** All entities use UUID v7 for time-ordered, unique identifiers.

## 4. Relationships

### 4.1 User to Account

- One-to-one relationship: each user has exactly one account.
- `users.user_id` → `accounts.user_id`

### 4.2 Account to Transactions

- One-to-many relationship: each account has many transactions.
- `accounts.account_id` → `transactions.account_id`

### 4.3 Transfer to Transactions

- One-to-many relationship: each transfer creates two transactions (debit from source, credit to destination).
- `transfers.transfer_id` → `transactions.transfer_id`

### 4.4 Account to Transfers

- One-to-many relationship: each account can be source or destination of many transfers.
- `accounts.account_id` → `transfers.source_account_id`
- `accounts.account_id` → `transfers.destination_account_id`
- Because `transfers` is partitioned by `transfer_id`, account-scoped transfer history is read from the `transfers_by_account` lookup table (`account_id` partition key → `transfer_id` clustering key).

### 4.5 User to Refresh Tokens

- One-to-many relationship: each user can have multiple refresh tokens (multiple devices).
- `users.user_id` → `refresh_tokens.user_id`

## 5. Key Workflows

### 5.1 Account Registration

1. Validate email format and password strength.
2. Check email uniqueness via the `users_by_email` lookup (SELECT user_id FROM users_by_email WHERE email = ?).
3. Hash password with Argon2id.
4. Auth Service inserts the user and the email lookup (INSERT INTO users; INSERT INTO users_by_email) in `auth_keyspace`.
5. Generate account_id (UUID v7).
6. Account Service inserts the account with balance_cents = 0 (INSERT INTO accounts) in `account_keyspace`.
7. Generate JWT tokens.
8. Auth Service hashes the refresh token and inserts it (INSERT INTO refresh_tokens) in `auth_keyspace`.
9. Each owning service records its own audit event (INSERT INTO audit_events) in its own keyspace.

### 5.2 Balance Inquiry

1. Validate JWT token.
2. Extract user_id from token.
3. Query account by user_id (SELECT accounts WHERE user_id = ?).
4. Return balance_cents and currency_code.

### 5.3 Transfer Initiation

1. Validate JWT token at the API Gateway.
2. Extract user_id from token.
3. Transfer Service resolves the source account by user_id (Account Service).
4. Validate destination account exists (Account Service).
5. Validate amount_cents > 0 and source_account_id != destination_account_id.
6. Check the idempotency key: SELECT response_status, response_body, resource_id FROM idempotency_keys WHERE idempotency_key = ?. If a completed record exists, return the cached response (`response_status`, `response_body`) without side effects. If a pending marker exists, treat the request as in-flight.
7. Generate transfer_id (UUID v7).
8. Insert transfer with status = pending (INSERT INTO transfers) in `transfer_keyspace`, and insert the `transfers_by_account` lookups for both source and destination.
9. Insert a pending idempotency marker (INSERT INTO idempotency_keys) when one does not already exist.
10. Publish `transfer.initiated` to Kafka topic `transfers`.
11. Record an audit event in `transfer_keyspace`.
12. Return `202 Accepted` with transferId and status.

### 5.4 Transfer Processing (saga orchestration)

Ownership rule: each service writes only to its own keyspace. The Transfer Service orchestrates; the Account Service performs debits/credits in `account_keyspace`; the Transaction Service projects transactions in `transaction_keyspace`.

1. Transfer Service consumes `transfer.initiated` from Kafka topic `transfers`.
2. If a completed idempotency record already exists for the key, skip (idempotent).
3. Transfer Service marks its own transfer as processing (UPDATE transfers SET status = 'processing' WHERE transfer_id = ?).
4. Transfer Service issues `account.debit`; Account Service executes the debit in `account_keyspace` as a single lightweight transaction: UPDATE accounts SET balance_cents = balance_cents - ?, updated_at = ? WHERE account_id = ? IF balance_cents >= ?. If the condition fails, Account Service emits `account.debit.failed`.
5. On `account.debited`, Transfer Service issues `account.credit`; Account Service executes the credit in `account_keyspace` as a separate statement. If the credit fails, Account Service emits `account.credit.failed` and Transfer Service triggers compensation by commanding Account Service to reverse the debit (credit the source account back).
6. Transaction Service consumes `account.debited` and `account.credited` and projects transaction rows into `transaction_keyspace` (the read model); it never writes to another service's keyspace.
7. On success, Transfer Service marks its transfer completed (UPDATE transfers SET status = 'completed' WHERE transfer_id = ?) and records the idempotency key with the cached response: INSERT INTO idempotency_keys (idempotency_key, operation_type, resource_id, expires_at, created_at, response_status, response_body) VALUES (?, ?, ?, ?, ?, ?, ?) USING TTL ?.
8. On failure, after compensation, Transfer Service marks its transfer failed (UPDATE transfers SET status = 'failed', failure_reason = ? WHERE transfer_id = ?).
9. Each service records its own `audit_events` in its own keyspace.

**Consistency note:** Cassandra has no `BEGIN`/`ROLLBACK`. Each lightweight transaction (paxos) is a single statement, so the debit and the credit are not atomic with respect to one another. Cross-account consistency is achieved by the saga: a completed debit that cannot be credited is reversed by a compensating credit to the source account.

## 6. Cassandra Configuration

### 6.1 Keyspace

Each service uses its own keyspace. There is no shared keyspace: each service owns its `audit_events` table inside its own keyspace.

```cql
CREATE KEYSPACE auth_keyspace
WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': 3}
AND durable_writes = true;

CREATE KEYSPACE account_keyspace
WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': 3}
AND durable_writes = true;

CREATE KEYSPACE transfer_keyspace
WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': 3}
AND durable_writes = true;

CREATE KEYSPACE transaction_keyspace
WITH replication = {'class': 'NetworkTopologyStrategy', 'datacenter1': 3}
AND durable_writes = true;
```

### 6.2 Consistency Levels

- Reads: QUORUM
- Writes: QUORUM
- Lightweight transactions: SERIAL

### 6.3 Replication

- Replication factor: 3
- Consistency: QUORUM (2 out of 3 replicas must acknowledge)

## 7. Data Migration Strategy

CQL migrations are versioned and applied in order. Each migration is a CQL script that is idempotent.

Migrations live with the owning service under `packages/<service>/internal/db/migrations/` with filenames like `001_create_tables.cql`, `002_add_indexes.cql`, etc. Each service owns and applies only its own migrations.

Migration state is tracked in a `schema_migrations` table per keyspace:

```cql
CREATE TABLE schema_migrations (
    version int PRIMARY KEY,
    applied_at timestamp,
    checksum text
);
```

## 8. Data Retention

- Idempotency keys: expire after 24 hours (configurable).
- Refresh tokens: expire after 7 days (configurable).
- Audit events: retain for 1 year (configurable).
- Transfers and transactions: retain indefinitely (configurable, subject to regulatory requirements).

## 9. Backup and Recovery

- Cassandra snapshots are taken regularly per keyspace.
- Snapshots are stored in durable object storage (S3, GCS).
- Point-in-time recovery is supported using Cassandra's incremental repair.
- Backup strategy is documented in [`12-operations/backup-recovery.md`](./12-operations/backup-recovery.md).
