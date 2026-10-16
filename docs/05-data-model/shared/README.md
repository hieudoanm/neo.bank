# Audit Events — Shared Shape

## Purpose

Defines the canonical shape of the append-only `audit_events` table used by every service. Each service owns its **own** `audit_events` table inside its **own** private keyspace:

| Service             | Keyspace               |
| ------------------- | ---------------------- |
| Auth Service        | `auth_keyspace`        |
| Account Service     | `account_keyspace`     |
| Transfer Service    | `transfer_keyspace`    |
| Transaction Service | `transaction_keyspace` |

There is **no `shared` keyspace and no shared audit store**. This document describes the shared *shape* only; the rows live in each service's keyspace and no service reads another service's audit log. Cross-service correlation is done by matching `correlation_id` across the relevant service keyspaces, not by a shared table.

## Table shape

### audit_events

| Column         | Type            | Description                                                |
| -------------- | --------------- | ---------------------------------------------------------- |
| service        | text            | Partition key; owning service identifier                   |
| day            | date            | Partition key; UTC day, keeps partitions bounded           |
| timestamp      | timestamp       | Event timestamp (ISO 8601 UTC), clustering key             |
| event_id       | uuid            | Event ID, UUID v7, clustering key                          |
| actor_user_id  | uuid            | Actor user ID (null for system events)                     |
| action         | text            | Action performed (login, transfer_initiated, etc.)         |
| resource_type  | text            | Resource type (account, transfer, user, etc.)              |
| resource_id    | uuid            | Resource ID                                                |
| correlation_id | text            | Correlation ID for distributed tracing                     |
| metadata       | map<text, text> | Additional metadata (key-value pairs, no sensitive data)   |

**Primary key:** `((service, day), timestamp, event_id)`

**Clustering key:** `timestamp DESC, event_id DESC` (newest first)

**Constraints:**
- Records are append-only (immutable); never updated or deleted.
- No sensitive data (passwords, tokens, PII) in metadata.
- `correlation_id` links related events across service boundaries.
- Retention: 1 year (configurable).
- `event_id` is a UUID v7 and is only a clustering component for uniqueness; it is not used for write ordering (the `timestamp DESC` clustering does that).

**Access patterns:**
- `INSERT INTO audit_events ...` — write an audit event (QUORUM write).
- `SELECT * FROM audit_events WHERE service = ? AND day = ?` — read a service's events for a day, newest first. Filter by `actor_user_id`, `resource_id`, or `correlation_id` in the application.
- A lookup by `correlation_id` alone is **not** a valid query: `correlation_id` is not a partition key. To trace a request, read the relevant service/day partitions (one per service involved) and match `correlation_id` across them, or add a dedicated lookup table inside the owning keyspace if that access pattern becomes hot.

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

-- Read a service's events for a day (then match correlation_id in the application)
SELECT timestamp, action, resource_type, resource_id, correlation_id, metadata
FROM audit_events
WHERE service = ? AND day = ?;
```

## Data Ownership

- Owner: the service that wrote the row. Each service writes only its own `audit_events` table in its own keyspace; there are no multiple writers to a single table and no cross-service direct database access.
- Writers authenticate via their own service credentials; no cross-service direct DB access.
- Table: `audit_events` (one per service keyspace).

## Audit logging behaviour (normative here)

Audit logging is described in this document rather than in the observability guide (which covers logs, metrics, and tracing). The rules are:

- Every security-relevant or financially-relevant action writes one immutable `audit_events` row in the acting service's keyspace.
- Writes are best-effort within the same request/consumer as the action; failures are logged but do not roll back the primary operation.
- `correlation_id` is propagated from the originating request so events for one workflow can be correlated across service keyspaces.

## Cross-references

- Data model overview: [`docs/05-data-model.md`](../../05-data-model.md#2-service-data-ownership)
- Failure recovery: [`docs/04-architecture.md`](../../04-architecture.md#8-failure-recovery)
- Observability (logs, metrics, tracing): [`docs/08-tech-stack/backend/observability.md`](../../08-tech-stack/backend/observability.md)
- Planned test: `test_audit_logging` (verifies audit events are written for security-relevant actions)
