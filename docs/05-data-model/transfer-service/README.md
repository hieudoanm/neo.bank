# Transfer Service — Data Model

## Purpose

Saga orchestrator. Owns transfer aggregate state and orchestrates the debit/credit flow via Kafka events. Manages saga state, idempotency for orchestration, and compensation logic. Private Cassandra keyspace; no other service may access these tables directly. The Transfer Service is the source of truth for transfer status and saga state.

## Tables

### transfers

| Column                | Type      | Description                                                        |
| --------------------- | --------- | ------------------------------------------------------------------ |
| transfer_id           | uuid      | Primary key, UUID v7                                               |
| source_account_id     | uuid      | Source account ID                                                  |
| destination_account_id| uuid      | Destination account ID                                             |
| amount_cents          | bigint    | Transfer amount in minor units (cents)                             |
| status                | text      | Transfer status (pending, processing, completed, failed, reversed) |
| idempotency_key       | text      | Idempotency key for duplicate detection                            |
| correlation_id        | text      | Correlation ID for distributed tracing                             |
| created_at            | timestamp | Transfer creation timestamp (ISO 8601 UTC)                         |
| updated_at            | timestamp | Last status update timestamp (ISO 8601 UTC)                        |
| failure_reason        | text      | Failure reason (null if not failed)                                |

**Primary key:** `(transfer_id)`

**Constraints:**
- `amount_cents` must be > 0.
- `status` must be one of: `pending`, `processing`, `completed`, `failed`, `reversed`.
- `source_account_id` != `destination_account_id`.
- `idempotency_key` identifies the request; the authoritative key-to-resource mapping and cached response live in `idempotency_keys` (Cassandra does not support `UNIQUE` constraints, so this is not enforced by the table).

**State transitions:**
- `pending` → `processing` → `completed` (successful debit + credit)
- `pending` → `processing` → `failed` (debit or credit failure, after compensation)
- `completed` → `reversed` (future enhancement, not in MVP)

**Access patterns:**
- `SELECT * FROM transfers WHERE transfer_id = ?` — status lookup by transfer ID (primary key, QUORUM read).
- `INSERT INTO transfers ...` — create transfer (on initiation).
- `UPDATE transfers SET status = ?, updated_at = ? WHERE transfer_id = ?` — status update (saga progression).
- `UPDATE transfers SET failure_reason = ?, status = 'failed', updated_at = ? WHERE transfer_id = ?` — mark failed.

Account-scoped transfer history is served by `transfers_by_account`; `transfers` carries no secondary indexes.

**CQL:**

```cql
CREATE TABLE transfers (
    transfer_id uuid PRIMARY KEY,
    source_account_id uuid,
    destination_account_id uuid,
    amount_cents bigint,
    status text,
    idempotency_key text,
    correlation_id text,
    created_at timestamp,
    updated_at timestamp,
    failure_reason text
);
```

**Example queries:**

```cql
-- Create transfer (initiation)
INSERT INTO transfers (transfer_id, source_account_id, destination_account_id, amount_cents, status, idempotency_key, correlation_id, created_at, updated_at)
VALUES (?, ?, ?, ?, 'pending', ?, ?, toTimestamp(now()), toTimestamp(now()));

-- Update status to processing (saga step 1)
UPDATE transfers SET status = 'processing', updated_at = toTimestamp(now()) WHERE transfer_id = ?;

-- Update status to completed (saga success)
UPDATE transfers SET status = 'completed', updated_at = toTimestamp(now()) WHERE transfer_id = ?;

-- Update status to failed (saga failure)
UPDATE transfers SET status = 'failed', failure_reason = ?, updated_at = toTimestamp(now()) WHERE transfer_id = ?;
```

### transfers_by_account

| Column                  | Type      | Description                                          |
| ----------------------- | --------- | ---------------------------------------------------- |
| account_id              | uuid      | Partition key; source or destination account         |
| transfer_id             | uuid      | Clustering key; UUID v7, DESC (newest first)         |
| direction               | text      | `source` or `destination` relative to `account_id`   |
| counterparty_account_id | uuid      | The other account in the transfer                    |
| amount_cents            | bigint    | Transfer amount in minor units (cents)               |
| status                  | text      | Transfer status                                      |
| correlation_id          | text      | Correlation ID for distributed tracing               |
| created_at              | timestamp | Transfer creation timestamp (ISO 8601 UTC)           |
| updated_at              | timestamp | Last status update timestamp (ISO 8601 UTC)          |
| failure_reason          | text      | Failure reason (null if not failed)                  |

**Primary key:** `((account_id), transfer_id)`

**Clustering key:** `transfer_id DESC` (newest first)

**Purpose:** the design principle is to partition all account-related data by `account_id`. Because `transfers` is partitioned by `transfer_id`, this lookup table provides the account-scoped access path so no secondary index is required. One row is written per transfer per involved account.

**Access patterns:**
- `SELECT * FROM transfers_by_account WHERE account_id = ?` — transfer history for an account (single-partition read, newest first).
- `INSERT INTO transfers_by_account ...` — write a lookup row for the source and destination account.
- `UPDATE transfers_by_account SET status = ?, updated_at = ? WHERE account_id = ? AND transfer_id = ?` — keep lookup status in sync with `transfers`.

**CQL:**

```cql
CREATE TABLE transfers_by_account (
    account_id uuid,
    transfer_id uuid,
    direction text,
    counterparty_account_id uuid,
    amount_cents bigint,
    status text,
    correlation_id text,
    created_at timestamp,
    updated_at timestamp,
    failure_reason text,
    PRIMARY KEY ((account_id), transfer_id)
) WITH CLUSTERING ORDER BY (transfer_id DESC);
```

**Example queries:**

```cql
-- Account transfer history (newest first)
SELECT transfer_id, direction, counterparty_account_id, amount_cents, status, created_at
FROM transfers_by_account
WHERE account_id = ?;

-- Insert lookup row
INSERT INTO transfers_by_account (account_id, transfer_id, direction, counterparty_account_id, amount_cents, status, correlation_id, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, 'pending', ?, toTimestamp(now()), toTimestamp(now()));
```

### idempotency_keys

| Column          | Type      | Description                                          |
| --------------- | --------- | ---------------------------------------------------- |
| idempotency_key | text      | Primary key                                          |
| operation_type  | text      | Operation type (transfer, etc.)                      |
| resource_id     | uuid      | Resource ID (e.g., transfer_id)                      |
| request_hash    | text      | Hash of the request payload (to detect key reuse)    |
| response_status | int       | Cached HTTP status for a completed operation         |
| response_body   | text      | Cached response body to replay for the same key      |
| expires_at      | timestamp | Expiration timestamp (ISO 8601 UTC)                  |
| created_at      | timestamp | Creation timestamp (ISO 8601 UTC)                    |

**Primary key:** `(idempotency_key)`

**Idempotency rule:** the key is recorded when the operation is initiated as a *pending marker* (`response_status`/`response_body` null). When the operation completes, the row is upserted with the cached response (`response_status`, `response_body`) so a repeat request with the same key and payload receives the identical response. A repeat request whose `request_hash` differs is rejected as key misuse.

**Expiry:** keys expire automatically via `USING TTL`; there is no `DELETE ... WHERE expires_at < ?` cleanup statement (Cassandra does not allow range deletes on non-key columns).

**Access patterns:**
- `SELECT response_status, response_body, resource_id FROM idempotency_keys WHERE idempotency_key = ?` — replay a cached response.
- `INSERT INTO idempotency_keys ... USING TTL ?` — store the pending marker or the completed response.

**CQL:**

```cql
CREATE TABLE idempotency_keys (
    idempotency_key text PRIMARY KEY,
    operation_type text,
    resource_id uuid,
    request_hash text,
    response_status int,
    response_body text,
    expires_at timestamp,
    created_at timestamp
);
```

**Example queries:**

```cql
-- Check for a completed key and its cached response
SELECT response_status, response_body, resource_id, request_hash
FROM idempotency_keys WHERE idempotency_key = ?;

-- Store pending marker with 24-hour TTL
INSERT INTO idempotency_keys (idempotency_key, operation_type, resource_id, request_hash, response_status, response_body, expires_at, created_at)
VALUES (?, 'transfer', ?, ?, null, null, ?, toTimestamp(now()))
USING TTL 86400;

-- Upsert cached response at completion (same TTL)
INSERT INTO idempotency_keys (idempotency_key, operation_type, resource_id, request_hash, response_status, response_body, expires_at, created_at)
VALUES (?, 'transfer', ?, ?, ?, ?, ?, toTimestamp(now()))
USING TTL 86400;
```

### audit_events

Each service records its own audit log; the Transfer Service owns this table inside `transfer_keyspace`. The table shape is shared across services (see [`../shared/README.md`](../shared/README.md)).

| Column         | Type            | Description                                                      |
| -------------- | --------------- | ---------------------------------------------------------------- |
| service        | text            | Owning service (constant `transfer` here)                        |
| day            | date            | Partition day (UTC) for bounded partitions                       |
| timestamp      | timestamp       | Event timestamp (ISO 8601 UTC)                                   |
| event_id       | uuid            | Event ID, UUID v7                                                |
| actor_user_id  | uuid            | Actor user ID (null for system events)                           |
| action         | text            | Action performed (transfer_initiated, completed, failed)         |
| resource_type  | text            | Resource type (transfer)                                         |
| resource_id    | uuid            | Resource ID                                                      |
| correlation_id | text            | Correlation ID for distributed tracing                           |
| metadata       | map<text, text> | Additional metadata (key-value pairs, no sensitive data)         |

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

- Service: **Transfer Service**
- Keyspace: `transfer_keyspace`
- Tables: `transfers`, `transfers_by_account`, `idempotency_keys`, `audit_events`

## Saga State Machine

The Transfer Service maintains the following state machine per transfer:

```
pending
  │
  ▼ (Kafka consumer picks up transfer.initiated)
processing
  │
  ├─► completed (account.debited + account.credited both succeed)
  │
  └─► failed (account.debited or account.credited fails; compensation applied)
```

On failure after a successful debit, compensation reverses the debit by crediting the source account back, then marks the transfer as `failed`.

## Cross-references

- Data model overview: [`docs/05-data-model.md`](../../05-data-model.md#2-service-data-ownership)
- Audit shape: [`shared/README.md`](../shared/README.md)
- Transaction Service transactions table: [`transaction-service/README.md`](../transaction-service/README.md#transactions)
- Account Service accounts table: [`account-service/README.md`](../account-service/README.md#accounts)
- Fund transfer feature: [`docs/02-features/FEATURE-04-Fund-Transfer.md`](../../02-features/FEATURE-04-Fund-Transfer.md)
- Idempotency feature: [`docs/02-features/FEATURE-07-Idempotency.md`](../../02-features/FEATURE-07-Idempotency.md)
- Saga pattern: [`docs/04-architecture.md`](../../04-architecture.md#32-saga-pattern-for-transfers)
- Kafka event contract: [`docs/08-tech-stack/backend/kafka.md`](../../08-tech-stack/backend/kafka.md)
- Planned test: `test_transfer` (integration with real Cassandra and Kafka, tests saga state transitions and compensation)
