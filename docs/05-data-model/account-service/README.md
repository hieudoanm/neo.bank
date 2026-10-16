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
- Initial state: `balance_cents = 0` (on account creation).
- Debit: `balance_cents -= amount` (must remain ≥ 0; enforced by LWT `IF balance_cents >= ?`).
- Credit: `balance_cents += amount`.

**Indexes:**
- Secondary index on `user_id`. Direction: this lets the service query accounts **by** `user_id` (user → account). It does not give a mapping from account to user; that is the `user_id` column on the row itself.

**Access patterns:**
- `SELECT balance_cents FROM accounts WHERE account_id = ?` — balance inquiry (QUORUM read).
- `SELECT * FROM accounts WHERE user_id = ?` — account lookup by owning user (secondary index, QUORUM read).
- `UPDATE accounts SET balance_cents = balance_cents - ?, updated_at = ? WHERE account_id = ? IF balance_cents >= ?` — debit (single LWT, SERIAL consistency).
- `UPDATE accounts SET balance_cents = balance_cents + ?, updated_at = ? WHERE account_id = ?` — credit (QUORUM write).
- `INSERT INTO accounts ...` — account creation (on registration).

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

CREATE INDEX idx_accounts_user_id ON accounts (user_id);
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
- Tables: `accounts`, `audit_events`

## Cross-references

- Data model overview: [`docs/05-data-model.md`](../../05-data-model.md#2-service-data-ownership)
- Audit shape: [`shared/README.md`](../shared/README.md)
- Auth Service users table: [`auth-service/README.md`](../auth-service/README.md#users)
- Balance inquiry feature: [`docs/02-features/FEATURE-03-Balance-Inquiry.md`](../../02-features/FEATURE-03-Balance-Inquiry.md)
- Fund transfer feature: [`docs/02-features/FEATURE-04-Fund-Transfer.md`](../../02-features/FEATURE-04-Fund-Transfer.md)
- Lightweight transactions: [`docs/05-data-model.md`](../../05-data-model.md#3-design-principles)
- Planned test: `test_account_balance` (integration with real Cassandra, tests negative balance prevention)
