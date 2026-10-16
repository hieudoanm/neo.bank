# NEO Bank — Cassandra Best Practices

_Backend · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [backend index](./README.md)._

## 1. Schema Organization

All schema lives in versioned migration scripts under each service's own module, `packages/<service>/internal/db/migrations/`. Each service owns and applies only its own migrations; no service reads or writes another service's keyspace.

```
packages/<service>/
└── internal/db/
    ├── cassandra.go            # session, cluster config, health
    ├── migrations.go           # runner + schema_migrations tracking
    └── migrations/
        ├── 001_create_tables.cql
        ├── 002_add_indexes.cql
        └── ...
```

Rules:

- One logical change per migration; migrations run in numeric order.
- Each service's migrations target only its own keyspace (for example `auth_keyspace` for `auth-service`); there is no shared keyspace and no shared migration.
- Files are `.cql` and idempotent (`CREATE TABLE IF NOT EXISTS`, `CREATE INDEX IF NOT EXISTS`).
- The runner records `(version, applied_at, checksum)`; a changed checksum on an applied migration fails the run.
- Destructive operations (`DROP`, column removal) are separate, reviewed migrations never mixed with additive changes.

## 2. Data Modeling (Query-First)

Model from access patterns, not from entities. Decide the query, then design the table.

| Query                                              | Table              | Key                                  |
| -------------------------------------------------- | ------------------ | ------------------------------------ |
| Balance by account                                  | `accounts`         | `account_id`                         |
| Account by user                                     | `accounts`         | secondary index on `user_id`         |
| User by email (login)                               | `users`            | secondary index on `email`           |
| Refresh token lookup                                | `refresh_tokens`   | `token_hash`                         |
| Transfer by ID                                      | `transfers`        | `transfer_id`                        |
| Transfers by source / destination                   | `transfers`        | secondary indexes                    |
| Transactions by account, newest first               | `transactions`     | `(account_id, transaction_id DESC)`  |
| Idempotency key lookup                              | `idempotency_keys` | `idempotency_key`                    |
| Audit events by time                                | `audit_events`     | `(event_id, timestamp DESC)`         |

Rules:

- Partition key must be the value you query by. `transactions` partitions by `account_id` so history reads are single-partition.
- Clustering order encodes the dominant sort (`transaction_id DESC` for newest first).
- Denormalize freely: a table per query is normal in Cassandra. Do not join.
- All identifiers are UUID v7 (time-ordered, unique).
- All financial amounts are `bigint` in minor units (cents); never `float`/`double`.
- Each table belongs to the keyspace of the service that owns it; a service never queries another service's tables (see [`05-data-model.md`](../../05-data-model.md)).

## 3. CQL Conventions

### Parameterised queries only

Every user-controlled value is bound, never interpolated:

```go
// internal/repositories/accounts.go
const getAccountByID = `
    SELECT account_id, user_id, balance_cents, currency_code, created_at, updated_at
    FROM accounts
    WHERE account_id = ?
`

func (r *AccountRepository) GetByID(ctx context.Context, accountID gocql.UUID) (*Account, error) {
    var acc Account
    err := r.session.Query(getAccountByID, accountID).
        WithContext(ctx).
        Consistency(gocql.Quorum).
        Scan(&acc.AccountID, &acc.UserID, &acc.BalanceCents, &acc.CurrencyCode, &acc.CreatedAt, &acc.UpdatedAt)
    if err != nil {
        if errors.Is(err, gocql.ErrNotFound) {
            return nil, ErrAccountNotFound
        }
        return nil, fmt.Errorf("get account: %w", err)
    }
    return &acc, nil
}
```

Rules:

- Query constants are package-level `const` strings, named after the operation.
- Never build CQL with `fmt.Sprintf`; dynamic identifiers (`keyspace`, table) come only from trusted allowlists.
- Always pass a `context.Context` via `WithContext(ctx)`.
- Use `gocql.UUID` for UUID values, not `string`.

### Consistency levels

| Operation                        | Consistency             |
| -------------------------------- | ----------------------- |
| Reads                            | `gocql.Quorum`          |
| Writes (non-LWT)                 | `gocql.Quorum`          |
| Lightweight transactions (paxos) | `gocql.Serial`          |
| Read-after-write on own balance  | `gocql.Quorum` (default) |

Set consistency explicitly at the query, and keep timeouts bounded so a slow node fails loudly.

## 4. Lightweight Transactions & Atomicity

Balance changes must be atomic and must never allow a negative balance. Use a `BATCH` with a conditional (paxos) update on the debit.

```go
// internal/repositories/transfers.go
const applyTransfer = `
    BEGIN BATCH
        UPDATE accounts SET balance_cents = balance_cents - ?, updated_at = ?
            WHERE account_id = ? IF balance_cents >= ?;
        UPDATE accounts SET balance_cents = balance_cents + ?, updated_at = ?
            WHERE account_id = ?;
    APPLY BATCH;
`

func (r *TransferRepository) ApplyTransfer(ctx context.Context, ev TransferEvent) error {
    iter := r.session.Query(applyTransfer,
        ev.AmountCents, now, ev.SourceAccountID, ev.AmountCents, // debit, conditional
        ev.AmountCents, now, ev.DestinationAccountID,            // credit, unconditional
    ).WithContext(ctx).Consistency(gocql.Serial).Iter()

    applied, err := iter.ScanCAS()
    if err != nil {
        return fmt.Errorf("apply transfer: %w", err)
    }
    if !applied {
        return ErrInsufficientFunds
    }
    return iter.Close()
}
```

Rules:

- The conditional (`IF`) update is the guard: insufficient funds means `applied == false`, not an error.
- Always `Close()` the iterator and check `ScanCAS` before proceeding.
- One LWT per logical operation; serialize paxos rounds to avoid contention on the same partition.
- Never read-then-write outside a transaction for balances — that is a race.

## 5. Query Patterns & Performance

- Prefer single-partition queries; a query should touch one partition key.
- Avoid `ALLOW FILTERING`. If a query needs it, add a dedicated table (query-first) instead.
- Secondary indexes are for low-cardinality lookups only (`user_id`, `email`); do not index high-cardinality or rapidly changing columns.
- Paginate with `PagingState`, never `LIMIT` guessing:

```go
func (r *TransactionRepository) ListByAccount(ctx context.Context, accountID gocql.UUID, pageSize int, state []byte) ([]Transaction, []byte, error) {
    iter := r.session.Query(listTransactions, accountID).
        WithContext(ctx).
        PageSize(pageSize).
        PageState(state).
        Consistency(gocql.Quorum).
        Iter()
    defer iter.Close()

    txs := make([]Transaction, 0, pageSize)
    for {
        var tx Transaction
        if !iter.Scan(&tx.TransactionID, &tx.Type, &tx.AmountCents, &tx.CounterpartyAccountID, &tx.CreatedAt) {
            break
        }
        txs = append(txs, tx)
    }
    if err := iter.Close(); err != nil {
        return nil, nil, fmt.Errorf("list transactions: %w", err)
    }
    return txs, iter.PageState(), nil
}
```

- Bound page sizes (for example `≤ 100`) and return the opaque paging state as the API cursor.
- Set per-query timeouts so a slow partition does not stall the request.

## 6. Migrations

- Migrations are applied by `cmd/migrate` before the API or consumer starts; CI applies them before integration tests.
- Migrations are additive on release; dropping or renaming a column is a two-step release.
- Never edit an applied migration — add a new one.
- CI runs migrations against the test keyspace (`neo_bank_test`, `SimpleStrategy`, RF 1) and fails on drift.

## 7. Idempotency, Retries & Failure

- Track processed operations in `idempotency_keys` (primary key `idempotency_key`) before applying side effects.
- Retries are safe only because writes are idempotent; never rely on a read-then-write for correctness.
- On timeout, treat the outcome as unknown and re-check the idempotency key rather than assuming success or failure.
- Unavailable Cassandra maps to API `503` with `Retry-After`; the consumer pauses and retries with backoff.

## 8. Connection & Session Management

```go
// internal/db/cassandra.go
func NewSession(ctx context.Context, cfg CassandraConfig) (*gocql.Session, error) {
    cluster := gocql.NewCluster(cfg.ContactPoints...)
    cluster.Keyspace = cfg.Keyspace
    cluster.Consistency = gocql.Quorum
    cluster.Timeout = 5 * time.Second
    cluster.ConnectTimeout = 5 * time.Second
    cluster.NumConns = 4
    cluster.PoolConfig.HostSelectionPolicy = gocql.TokenAwareHostPolicy(gocql.RoundRobinHostPolicy())

    session, err := cluster.CreateSession()
    if err != nil {
        return nil, fmt.Errorf("connect cassandra: %w", err)
    }
    return session, nil
}
```

Rules:

- One `*gocql.Session` per process, created at startup and closed on shutdown.
- Use `TokenAwareHostPolicy` so queries go to a replica.
- Bound pools and timeouts; never create a session per request.
- Fail fast at startup if the cluster is unreachable; expose readiness separately.

## 9. Testing

- Repository tests run against a real Cassandra (Docker service in CI), never a mock — the query is the contract.
- Use a dedicated test keyspace; truncate tables between tests.
- Assert both the happy path and the failure paths: insufficient funds (`ScanCAS` false), nonexistent rows (`ErrAccountNotFound`), and duplicate idempotency keys.
- Test concurrent debits on the same account: exactly one should succeed when funds are only enough for one.

## 10. Code Style & Rules

- Schema change requires a migration, a data-model update (`docs/05-data-model.md`), and a test.
- Every query is parameterised, `WithContext`, and consistency-explicit.
- Amounts are `bigint` cents; identifiers are UUID v7.
- No `ALLOW FILTERING`; no joins; no read-modify-write for balances.
- Repositories return domain errors (`ErrAccountNotFound`, `ErrInsufficientFunds`), never raw `gocql` errors.
- One responsibility per repository file; keep functions small and queries named.
