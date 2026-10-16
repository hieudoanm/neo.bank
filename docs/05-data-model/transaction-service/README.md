# Transaction Service — Data Model

## Purpose

Append-only read model for transaction history. Consumes transfer/account lifecycle events from Kafka to build a denormalised view optimised for paginated reads by account. Exposes paginated transaction queries via REST API. Private Cassandra keyspace; no other service may access these tables directly. The Transaction Service is a pure projection — it never participates in saga write paths and never writes another service's keyspace.

## Tables

### transactions

| Column                  | Type      | Description                                          |
| ----------------------- | --------- | ---------------------------------------------------- |
| account_id              | uuid      | Partition key; owning account                        |
| created_at              | timestamp | Transaction timestamp (ISO 8601 UTC), clustering key |
| transaction_id          | uuid      | Transaction ID, UUID v7; clustering key              |
| transfer_id             | uuid      | Foreign key to transfers (null for initial deposits) |
| type                    | text      | Transaction type (debit, credit)                     |
| amount_cents            | bigint    | Transaction amount in minor units (cents)            |
| counterparty_account_id | uuid      | Counterparty account ID                              |
| status                  | text      | Transaction status (completed, failed)               |

**Primary key:** `((account_id), created_at, transaction_id)`

**Clustering key:** `created_at DESC, transaction_id DESC` (deterministic newest-first ordering, matching `FEATURE-05`)

**Constraints:**
- `type` must be one of: `debit`, `credit`.
- `status` must be one of: `completed`, `failed` (required by FR-051).
- `amount_cents` must be > 0.
- Records are immutable (append-only); corrections create new offsetting records.
- `transfer_id` may be null for system-initiated transactions (e.g., initial deposit).

**Access patterns:**
- `SELECT * FROM transactions WHERE account_id = ?` — paginated history for an account (QUORUM read, clustering order gives newest first).
- `INSERT INTO transactions ...` — write transaction (projection from Kafka events, QUORUM write).
- `UPDATE transactions SET status = ? WHERE account_id = ? AND created_at = ? AND transaction_id = ?` — finalise status from a transfer lifecycle event.

**CQL:**

```cql
CREATE TABLE transactions (
    account_id uuid,
    created_at timestamp,
    transaction_id uuid,
    transfer_id uuid,
    type text,
    amount_cents bigint,
    counterparty_account_id uuid,
    status text,
    PRIMARY KEY ((account_id), created_at, transaction_id)
) WITH CLUSTERING ORDER BY (created_at DESC, transaction_id DESC);
```

**Example queries:**

```cql
-- First page of history for an account (newest first)
SELECT transaction_id, transfer_id, type, amount_cents, counterparty_account_id, status, created_at
FROM transactions
WHERE account_id = ?
LIMIT ?;

-- Next page: cursor is the (created_at, transaction_id) of the last item returned
SELECT transaction_id, transfer_id, type, amount_cents, counterparty_account_id, status, created_at
FROM transactions
WHERE account_id = ? AND (created_at, transaction_id) < (?, ?)
LIMIT ?;

-- Insert debit transaction (projected from account.debited)
INSERT INTO transactions (account_id, created_at, transaction_id, transfer_id, type, amount_cents, counterparty_account_id, status)
VALUES (?, toTimestamp(now()), ?, ?, 'debit', ?, ?, 'completed');

-- Insert credit transaction (projected from account.credited)
INSERT INTO transactions (account_id, created_at, transaction_id, transfer_id, type, amount_cents, counterparty_account_id, status)
VALUES (?, toTimestamp(now()), ?, ?, 'credit', ?, ?, 'completed');
```

### audit_events

Each service records its own audit log; the Transaction Service owns this table inside `transaction_keyspace`. The table shape is shared across services (see [`../shared/README.md`](../shared/README.md)).

| Column         | Type            | Description                                                |
| -------------- | --------------- | ---------------------------------------------------------- |
| service        | text            | Owning service (constant `transaction` here)               |
| day            | date            | Partition day (UTC) for bounded partitions                 |
| timestamp      | timestamp       | Event timestamp (ISO 8601 UTC)                             |
| event_id       | uuid            | Event ID, UUID v7                                          |
| actor_user_id  | uuid            | Actor user ID (null for system events)                     |
| action         | text            | Action performed (transaction_projected, status_updated)   |
| resource_type  | text            | Resource type (transaction)                                |
| resource_id    | uuid            | Resource ID                                                |
| correlation_id | text            | Correlation ID for distributed tracing                     |
| metadata       | map<text, text> | Additional metadata (key-value pairs, no sensitive data)   |

**Primary key:** `((service, day), timestamp, event_id)`

**Clustering key:** `timestamp DESC, event_id DESC` (newest first)

**CQL:**

```cql
CREATE TABLE audit_events (
    service text,
    day date,
    timestamp timestamp,
    event_id uuid,
    actor_user_id uuid,
    action text,
    resource_type text,
    resource_id uuid,
    correlation_id text,
    metadata map<text, text>,
    PRIMARY KEY ((service, day), timestamp, event_id)
) WITH CLUSTERING ORDER BY (timestamp DESC, event_id DESC);
```

## Data Ownership

- Service: **Transaction Service**
- Keyspace: `transaction_keyspace`
- Tables: `transactions`, `audit_events`

## Projection Logic

The Transaction Service projects the saga lifecycle into the `transactions` table. Account-movement events create rows; transfer lifecycle events finalise their status:

| Kafka Event | Projection Action |
|-------------|-------------------|
| `account.debited` | Insert a `debit` transaction for the source account with `status = 'completed'` |
| `account.credited` | Insert a `credit` transaction for the destination account with `status = 'completed'` |
| `transfer.completed` | Set `status = 'completed'` on the transactions for that `transfer_id` |
| `transfer.failed` | Set `status = 'failed'` on the transactions for that `transfer_id` (including the compensating movement) |

Each transaction record includes `transfer_id` to link debit/credit pairs to the same transfer, and `counterparty_account_id` so the frontend can display who sent/received funds. Projection is idempotent: duplicate deliveries either re-insert the same deterministic row (keyed by `(account_id, created_at, transaction_id)`) or re-apply the same status update.

## Cross-references

- Data model overview: [`docs/05-data-model.md`](../../05-data-model.md#2-service-data-ownership)
- Audit shape: [`shared/README.md`](../shared/README.md)
- Transfer Service transfers table: [`transfer-service/README.md`](../transfer-service/README.md#transfers)
- Transaction history feature: [`docs/02-features/FEATURE-05-Transaction-History.md`](../../02-features/FEATURE-05-Transaction-History.md)
- Saga pattern: [`docs/04-architecture.md`](../../04-architecture.md#32-saga-pattern-for-transfers)
- Kafka event contract: [`docs/08-tech-stack/backend/kafka.md`](../../08-tech-stack/backend/kafka.md)
- Planned test: `test_history` (integration with real Cassandra, verifies projection from events)
