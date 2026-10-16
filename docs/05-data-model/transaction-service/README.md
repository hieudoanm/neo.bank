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
-- created_at is taken from the event's occurredAt so a redelivery is
-- deterministic; transaction_id is a fresh UUID v7 generated once per movement.
INSERT INTO transactions (account_id, created_at, transaction_id, transfer_id, type, amount_cents, counterparty_account_id, status)
VALUES (?, ?, ?, ?, 'debit', ?, ?, 'completed');

-- Insert credit transaction (projected from account.credited)
INSERT INTO transactions (account_id, created_at, transaction_id, transfer_id, type, amount_cents, counterparty_account_id, status)
VALUES (?, ?, ?, ?, 'credit', ?, ?, 'completed');
```

### transfer_movements

Lookup index that makes the projection idempotent and lets transfer lifecycle events
find the rows they must finalise. A transaction is written once per
`(transfer_id, movement)`; a redelivered movement event finds the existing entry and
is ignored instead of inserting a duplicate row.

| Column         | Type      | Description                                                       |
| -------------- | --------- | ----------------------------------------------------------------- |
| transfer_id    | uuid      | Partition key; transfer the movement belongs to                   |
| movement       | text      | Movement identifier within the transfer, clustering key           |
| transaction_id | uuid      | Transaction row this movement produced (UUID v7)                  |
| account_id     | uuid      | Account the movement was applied to (partition key of `transactions`) |
| created_at     | timestamp | Mirrors the transaction's `created_at` (clustering key of `transactions`) |

**Primary key:** `((transfer_id), movement)`

**Constraints:**
- `movement` is derived from the command's idempotency key suffix: `debit`, `credit`, or `compensation` (the compensating credit to the source when a credit fails).
- The row exists solely to make projection idempotent and to locate the `transactions` row for status finalisation; it carries no amount or status.

**CQL:**

```cql
CREATE TABLE transfer_movements (
    transfer_id uuid,
    movement text,
    transaction_id uuid,
    account_id uuid,
    created_at timestamp,
    PRIMARY KEY ((transfer_id), movement)
);
```

**Access patterns:**
- `SELECT transaction_id, account_id, created_at FROM transfer_movements WHERE transfer_id = ? AND movement = ?` — idempotency check before inserting a projected transaction.
- `SELECT movement, transaction_id, account_id, created_at FROM transfer_movements WHERE transfer_id = ?` — enumerate a transfer's movements to finalise their status.
- `INSERT INTO transfer_movements ...` — record a movement after inserting its transaction (QUORUM write).

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
- Tables: `transactions`, `transfer_movements`, `audit_events`

## Projection Logic

The Transaction Service projects the saga lifecycle into the `transactions` table. Account-movement events create rows; transfer lifecycle events finalise their status:

| Kafka Event | Projection Action |
|-------------|-------------------|
| `account.debited` | Insert a `debit` transaction for the source account with `status = 'completed'` |
| `account.credited` | Insert a `credit` transaction for the destination account with `status = 'completed'` |
| `transfer.completed` | Set `status = 'completed'` on the transactions for that `transfer_id` |
| `transfer.failed` | Set `status = 'failed'` on the transactions for that `transfer_id` (including the compensating movement) |

Each transaction record includes `transfer_id` to link debit/credit pairs to the same transfer, and `counterparty_account_id` so the frontend can display who sent/received funds.

**Idempotency.** A movement event is projected at most once: before inserting, the
service looks up `transfer_movements` by `(transfer_id, movement)`; if an entry
exists the event is already applied and is skipped. `created_at` comes from the
event's `occurredAt`, so a replayed insert would be the same row anyway, and status
updates are naturally re-appliable. `transfer.completed`/`transfer.failed` enumerate
the transfer's rows through `transfer_movements` and update each. Unsupported event
versions and malformed payloads are routed to `transfers.dlq` rather than feared.

## Cross-references

- Data model overview: [`docs/05-data-model.md`](../../05-data-model.md#2-service-data-ownership)
- Audit shape: [`shared/README.md`](../shared/README.md)
- Transfer Service transfers table: [`transfer-service/README.md`](../transfer-service/README.md#transfers)
- Transaction history feature: [`docs/02-features/FEATURE-05-Transaction-History.md`](../../02-features/FEATURE-05-Transaction-History.md)
- Saga pattern: [`docs/04-architecture.md`](../../04-architecture.md#32-saga-pattern-for-transfers)
- Kafka event contract: [`docs/08-tech-stack/backend/kafka.md`](../../08-tech-stack/backend/kafka.md)
- Planned test: `test_history` (integration with real Cassandra, verifies projection from events)
