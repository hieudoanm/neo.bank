# Account Service — Data Model

## Purpose

Manages accounts and balances. Authoritative for account validation and balance reads/writes related to transfers. Private Cassandra keyspace; no other service may access these tables directly. All balance updates use lightweight transactions (Paxos) to ensure atomicity and prevent negative balances.

## Tables

### accounts

| Column        | Type      | Description                                  |
| ------------- | --------- | -------------------------------------------- |
| account_id    | uuid      | Primary key, UUID v7                         |
| user_id       | uuid      | Foreign key to users (Auth Service)          |
| balance_cents | bigint    | Current balance in minor units (cents)       |
| currency_code | text      | Currency code (e.g., USD)                    |
| created_at    | timestamp | Account creation timestamp (ISO 8601 UTC)    |
| updated_at    | timestamp | Last balance update timestamp (ISO 8601 UTC) |

**Primary key:** `(account_id)`

**Constraints:**
- `balance_cents` must be ≥ 0 (enforced by application logic and lightweight transactions).
- `currency_code` must be a valid ISO 4217 code.
- `account_id` is unique.

**State transitions:**
- Initial state: `balance_cents` is set to the configured opening balance on account creation (`INITIAL_BALANCE_CENTS`, `0` by default). The Account Service provisions the account when it consumes the Auth Service's `auth.user.created` event, using an idempotent `INSERT … IF NOT EXISTS` so at-least-once delivery cannot duplicate the account or reset the balance.
- Debit: `balance_cents -= amount` (must remain ≥ 0; enforced by LWT `IF balance_cents >= ?`).
- Credit: `balance_cents += amount`.

**Access paths (query-first):**
- The row itself maps account → user via the `user_id` column.
- The reverse path (user → accounts) is served by the `accounts_by_user` lookup table, **not** a secondary index. Amazon Keyspaces does not support `CREATE INDEX`, and the transfer/auth services already use the same lookup-table pattern.

**Access patterns:**
- `SELECT balance_cents FROM accounts WHERE account_id = ?` — balance inquiry (QUORUM read).
- `SELECT account_id FROM accounts_by_user WHERE user_id = ?` — accounts owned by a user (single-partition read).
- `UPDATE accounts SET balance_cents = balance_cents - ?, updated_at = ? WHERE account_id = ? IF balance_cents >= ?` — debit (single LWT, SERIAL consistency).
- `UPDATE accounts SET balance_cents = balance_cents + ?, updated_at = ? WHERE account_id = ?` — credit (QUORUM write).
- `INSERT INTO accounts ...` / `INSERT INTO accounts_by_user ...` — account creation (on registration).

**CQL:**

```cql
CREATE TABLE accounts (
    account_id uuid PRIMARY KEY,
    user_id uuid,
    balance_cents bigint,
    currency_code text,
    created_at timestamp,
    updated_at timestamp
);
```

**Example queries:**

```cql
-- Balance inquiry (read)
SELECT balance_cents, currency_code, updated_at FROM accounts WHERE account_id = ?;

-- Debit with lightweight transaction (atomic, prevents negative balance).
-- One LWT per statement; there is no cross-account transaction.
UPDATE accounts
SET balance_cents = balance_cents - ?, updated_at = toTimestamp(now())
WHERE account_id = ?
IF balance_cents >= ?;

-- Credit (separate statement, QUORUM)
UPDATE accounts
SET balance_cents = balance_cents + ?, updated_at = toTimestamp(now())
WHERE account_id = ?;
```

### accounts_by_user

Reverse lookup from a user to the account(s) they own. Replaces a secondary index on `accounts(user_id)`, which Amazon Keyspaces does not support. Partitioning by `user_id` makes "list this user's accounts" a single-partition read. Written in the same operation that creates the account.

| Column     | Type      | Description                          |
| ---------- | --------- | ------------------------------------ |
| user_id    | uuid      | Partition key — owning user          |
| account_id | uuid      | Clustering key — the user's account  |
| created_at | timestamp | Account creation timestamp           |

**Primary key:** `((user_id), account_id)`

**Access patterns:**
- `SELECT account_id FROM accounts_by_user WHERE user_id = ?` — list accounts for the authenticated user (for example to resolve the `accountId` returned at login).
- `INSERT INTO accounts_by_user ...` — one row per account on creation.

**CQL:**

```cql
CREATE TABLE accounts_by_user (
    user_id uuid,
    account_id uuid,
    created_at timestamp,
    PRIMARY KEY ((user_id), account_id)
);
```

### processed_commands

Command-side idempotency for balance movements. The saga delivers commands at-least-once, so the Account Service records each command it applies; a redelivered `account.debit.requested`/`account.credit.requested` with the same command `idempotency_key` re-emits the stored result instead of moving funds again. The transfer's request-level idempotency key lives in `transfer_keyspace` and is **not** readable here — this table is the Account Service's own record.

| Column           | Type      | Description                                                                 |
| ---------------- | --------- | --------------------------------------------------------------------------- |
| idempotency_key  | text      | Primary key — command idempotency key (scoped to this transfer + movement)  |
| transfer_id      | uuid      | Originating transfer                                                        |
| account_id       | uuid      | Account debited/credited                                                    |
| event_type       | text      | Command consumed: `account.debit.requested` or `account.credit.requested`   |
| result_event_type| text      | Result emitted: `account.debited`/`account.debit.failed`/`account.credited`/`account.credit.failed` |
| balance_cents    | bigint    | Resulting balance (null until applied)                                      |
| failure_reason   | text      | Reason when the command failed                                              |
| created_at       | timestamp | When the command was first seen (ISO 8601 UTC)                              |
| applied_at       | timestamp | When the movement was applied (null while still pending)                    |

**Primary key:** `(idempotency_key)`

**Idempotency rule:** the consumer claims the command with `INSERT ... IF NOT EXISTS` (a *pending* row), applies the balance movement, then upserts the result onto the same row. A redelivery whose claim already exists does not move funds again: if a result is recorded it re-emits that result; if the row is still pending the previous attempt did not finish, so the consumer re-drives the movement for that same command. The claim and the `accounts` LWT are **separate** statements — Cassandra has no cross-table transaction — so this is not a single atomic unit; correctness relies on the command key being immutable and the movement being re-drivable (see [atomicity note](../../05-data-model.md#3-design-principles)).

**Access patterns:**
- `INSERT INTO processed_commands (...) IF NOT EXISTS` — claim a command.
- `SELECT result_event_type, balance_cents, failure_reason FROM processed_commands WHERE idempotency_key = ?` — read a recorded result to re-emit.
- `UPDATE processed_commands SET result_event_type = ?, balance_cents = ?, failure_reason = ?, applied_at = ? WHERE idempotency_key = ?` — record the outcome.

**CQL:**

```cql
CREATE TABLE processed_commands (
    idempotency_key text PRIMARY KEY,
    transfer_id uuid,
    account_id uuid,
    event_type text,
    result_event_type text,
    balance_cents bigint,
    failure_reason text,
    created_at timestamp,
    applied_at timestamp
);
```

### audit_events

Each service records its own audit log; the Account Service owns this table inside `account_keyspace`. The table shape is shared across services (see [`../shared/README.md`](../shared/README.md)).

| Column         | Type            | Description                                                |
| -------------- | --------------- | ---------------------------------------------------------- |
| service        | text            | Owning service (constant `account` here)                   |
| day            | date            | Partition day (UTC) for bounded partitions                 |
| timestamp      | timestamp       | Event timestamp (ISO 8601 UTC)                             |
| event_id       | uuid            | Event ID, UUID v7                                          |
| actor_user_id  | uuid            | Actor user ID (null for system events)                     |
| action         | text            | Action performed (account_created, debited, credited)      |
| resource_type  | text            | Resource type (account)                                    |
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

- Service: **Account Service**
- Keyspace: `account_keyspace`
- Tables: `accounts`, `accounts_by_user`, `processed_commands`, `audit_events`

## Cross-references

- Data model overview: [`docs/05-data-model.md`](../../05-data-model.md#2-service-data-ownership)
- Audit shape: [`shared/README.md`](../shared/README.md)
- Auth Service users table: [`auth-service/README.md`](../auth-service/README.md#users)
- Balance inquiry feature: [`docs/02-features/FEATURE-03-Balance-Inquiry.md`](../../02-features/FEATURE-03-Balance-Inquiry.md)
- Fund transfer feature: [`docs/02-features/FEATURE-04-Fund-Transfer.md`](../../02-features/FEATURE-04-Fund-Transfer.md)
- Lightweight transactions: [`docs/05-data-model.md`](../../05-data-model.md#3-design-principles)
- Planned test: `test_account_balance` (integration with real Cassandra, tests negative balance prevention)
