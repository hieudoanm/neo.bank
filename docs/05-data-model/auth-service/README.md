# Auth Service — Data Model

## Purpose

Owns identity: users, credentials (Argon2id hashes), refresh tokens, and its own audit log. Private Cassandra keyspace; no other service may access these tables directly. Passwords are never stored in plain text; refresh tokens are hashed (SHA-256) at rest.

## Tables

### users

| Column        | Type      | Description                            |
| ------------- | --------- | -------------------------------------- |
| user_id       | uuid      | Primary key, UUID v7                   |
| email         | text      | Unique email address                   |
| password_hash | text      | Argon2id password hash                 |
| created_at    | timestamp | User creation timestamp (ISO 8601 UTC) |
| updated_at    | timestamp | Last update timestamp (ISO 8601 UTC)   |

**Primary key:** `(user_id)`

**Constraints:**
- `email` is unique (enforced by the `users_by_email` lookup, not by a secondary index).
- `password_hash` is Argon2id hash (memory=64MB, iterations=3, parallelism=4).

**Access patterns:**
- `SELECT * FROM users WHERE user_id = ?` — user lookup by ID (primary key, QUORUM read).
- `INSERT INTO users ...` — user creation (on registration).

**CQL:**

```cql
CREATE TABLE users (
    user_id uuid PRIMARY KEY,
    email text,
    password_hash text,
    created_at timestamp,
    updated_at timestamp
);
```

**Example queries:**

```cql
-- User creation
INSERT INTO users (user_id, email, password_hash, created_at, updated_at)
VALUES (?, ?, ?, toTimestamp(now()), toTimestamp(now()));
```

### users_by_email

| Column     | Type      | Description                            |
| ---------- | --------- | -------------------------------------- |
| email      | text      | Primary key (normalised email address) |
| user_id    | uuid      | Owning user                            |
| created_at | timestamp | Creation timestamp (ISO 8601 UTC)      |

**Primary key:** `(email)`

**Purpose and index direction:** login looks a user up **by email**, so the access path must be partitioned by `email`. A secondary index on `users(email)` is the wrong tool: Cassandra secondary indexes are per-node, non-unique, and do not enforce uniqueness, so a check-then-insert race can create duplicate emails. Partitioning `users_by_email` by `email` makes the lookup a single-partition read and lets registration claim the email atomically with `INSERT ... IF NOT EXISTS` (LWT).

**Access patterns:**
- `SELECT user_id FROM users_by_email WHERE email = ?` — login lookup by email (single-partition read, QUORUM).
- `INSERT INTO users_by_email ... IF NOT EXISTS` — claim the email on registration; a failed conditional write means the email is already taken.
- `DELETE FROM users_by_email WHERE email = ?` — release the email (administrative only; user deletion).

**CQL:**

```cql
CREATE TABLE users_by_email (
    email text PRIMARY KEY,
    user_id uuid,
    created_at timestamp
);
```

**Example queries:**

```cql
-- Login lookup by email
SELECT user_id FROM users_by_email WHERE email = ?;

-- Claim the email (fails if already present)
INSERT INTO users_by_email (email, user_id, created_at)
VALUES (?, ?, toTimestamp(now()))
IF NOT EXISTS;
```

### refresh_tokens

| Column     | Type      | Description                                                     |
| ---------- | --------- | --------------------------------------------------------------- |
| token_hash | text      | SHA-256 hash of refresh token (primary key)                     |
| user_id    | uuid      | Foreign key to users                                            |
| expires_at | timestamp | Token expiration timestamp (ISO 8601 UTC, 7 days from creation) |
| created_at | timestamp | Token creation timestamp (ISO 8601 UTC)                         |
| revoked_at | timestamp | Token revocation timestamp (ISO 8601 UTC), null if not revoked  |

**Primary key:** `(token_hash)`

**Constraints:**
- `token_hash` is unique.
- `expires_at` must be in the future when created.
- `revoked_at` is null for active tokens.

**Expiry / cleanup:** expired tokens are reclaimed automatically by `USING TTL`; there is no range `DELETE` on `expires_at` (Cassandra does not allow range deletes on non-key columns). Explicit revocation sets `revoked_at`.

**State transitions:**
- Created: `revoked_at = null` (active token).
- Revoked: `revoked_at = <timestamp>` (on logout or password change).

**Access patterns:**
- `SELECT * FROM refresh_tokens WHERE token_hash = ?` — token validation (QUORUM read).
- `INSERT INTO refresh_tokens ... USING TTL ?` — token creation (on login/refresh).
- `UPDATE refresh_tokens SET revoked_at = ? WHERE token_hash = ?` — token revocation (on logout).

**CQL:**

```cql
CREATE TABLE refresh_tokens (
    token_hash text PRIMARY KEY,
    user_id uuid,
    expires_at timestamp,
    created_at timestamp,
    revoked_at timestamp
);
```

**Example queries:**

```cql
-- Validate refresh token
SELECT user_id, expires_at, revoked_at FROM refresh_tokens WHERE token_hash = ?;

-- Store new refresh token (TTL matches expires_at)
INSERT INTO refresh_tokens (token_hash, user_id, expires_at, created_at, revoked_at)
VALUES (?, ?, ?, toTimestamp(now()), null)
USING TTL 604800;

-- Revoke refresh token
UPDATE refresh_tokens SET revoked_at = toTimestamp(now()) WHERE token_hash = ?;
```

### refresh_tokens_by_user

| Column     | Type      | Description                                                    |
| ---------- | --------- | -------------------------------------------------------------- |
| user_id    | uuid      | Partition key; owning user                                     |
| token_hash | text      | Clustering key; SHA-256 hash of refresh token                  |
| expires_at | timestamp | Token expiration timestamp (ISO 8601 UTC)                      |
| created_at | timestamp | Token creation timestamp (ISO 8601 UTC)                        |
| revoked_at | timestamp | Token revocation timestamp (ISO 8601 UTC), null if not revoked |

**Primary key:** `((user_id), token_hash)`

**Purpose:** `refresh_tokens` is partitioned by `token_hash` (needed to validate a presented token), so it cannot list or revoke all tokens for a user. This lookup table is partitioned by `user_id`, making "revoke all sessions on password change" expressible: read the user's tokens, then revoke each.

**Access patterns:**
- `SELECT token_hash FROM refresh_tokens_by_user WHERE user_id = ?` — list a user's refresh tokens.
- `UPDATE refresh_tokens_by_user SET revoked_at = ? WHERE user_id = ? AND token_hash = ?` — revoke one (kept in sync with `refresh_tokens`).

**CQL:**

```cql
CREATE TABLE refresh_tokens_by_user (
    user_id uuid,
    token_hash text,
    expires_at timestamp,
    created_at timestamp,
    revoked_at timestamp,
    PRIMARY KEY ((user_id), token_hash)
);
```

**Example queries:**

```cql
-- List a user's tokens (for revoke-all on password change)
SELECT token_hash FROM refresh_tokens_by_user WHERE user_id = ?;

-- Revoke one token
UPDATE refresh_tokens_by_user SET revoked_at = toTimestamp(now())
WHERE user_id = ? AND token_hash = ?;
```

### audit_events

Each service records its own audit log; the Auth Service owns this table inside `auth_keyspace`. The table shape is shared across services (see [`../shared/README.md`](../shared/README.md)).

| Column         | Type            | Description                                                |
| -------------- | --------------- | ---------------------------------------------------------- |
| service        | text            | Owning service (constant `auth` here)                      |
| day            | date            | Partition day (UTC) for bounded partitions                 |
| timestamp      | timestamp       | Event timestamp (ISO 8601 UTC)                             |
| event_id       | uuid            | Event ID, UUID v7                                          |
| actor_user_id  | uuid            | Actor user ID (null for system events)                     |
| action         | text            | Action performed (login, logout, register, token_revoked)  |
| resource_type  | text            | Resource type (user, refresh_token)                        |
| resource_id    | uuid            | Resource ID                                                |
| correlation_id | text            | Correlation ID for distributed tracing                     |
| metadata       | map<text, text> | Additional metadata (key-value pairs, no sensitive data)   |

**Primary key:** `((service, day), timestamp, event_id)`

**Clustering key:** `timestamp DESC, event_id DESC` (newest first)

**Access patterns:**
- `INSERT INTO audit_events ...` — write audit event (QUORUM write).
- `SELECT * FROM audit_events WHERE service = ? AND day = ?` — recent events for a service/day partition; filter by `actor_user_id`, `resource_id`, or `correlation_id` in the application.

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

**Example queries:**

```cql
-- Record an audit event
INSERT INTO audit_events (service, day, timestamp, event_id, actor_user_id, action, resource_type, resource_id, correlation_id, metadata)
VALUES (?, toDate(now()), toTimestamp(now()), ?, ?, ?, ?, ?, ?, ?);

-- Recent auth events for a day, then filter by correlation_id in the application
SELECT timestamp, action, resource_type, resource_id, correlation_id
FROM audit_events
WHERE service = 'auth' AND day = ?;
```

## Data Ownership

- Service: **Auth Service**
- Keyspace: `auth_keyspace`
- Tables: `users`, `users_by_email`, `refresh_tokens`, `refresh_tokens_by_user`, `audit_events`

## Cross-references

- Data model overview: [`docs/05-data-model.md`](../../05-data-model.md#2-service-data-ownership)
- Audit shape: [`shared/README.md`](../shared/README.md)
- Account Service accounts table: [`account-service/README.md`](../account-service/README.md#accounts)
- Authentication feature: [`docs/02-features/FEATURE-02-Authentication.md`](../../02-features/FEATURE-02-Authentication.md)
- Account registration feature: [`docs/02-features/FEATURE-01-Account-Registration.md`](../../02-features/FEATURE-01-Account-Registration.md)
- JWT implementation: [`docs/08-tech-stack/backend/jwt.md`](../../08-tech-stack/backend/jwt.md)
- Planned test: `test_user_lookup`, `test_refresh_token_validation` (integration with real Cassandra)
