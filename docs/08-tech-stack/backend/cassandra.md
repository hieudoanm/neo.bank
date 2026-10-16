# NEO Bank — Cassandra Best Practices

_Backend · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [backend index](./README.md)._

## 1. Schema Organization

All schema lives in versioned migration scripts under each service's own module, `packages/backend/<service>/internal/db/migrations/`. Each service owns and applies only its own migrations; no service reads or writes another service's keyspace.

```
packages/backend/<service>/
└── internal/db/
    ├── cassandra.go            # session, cluster config, health
    ├── migrations.go           # runner + schema_migrations tracking
    └── migrations/
        ├── 001_create_tables.cql
        ├── 002_add_lookup_tables.cql
        └── ...
```

Rules:

- One logical change per migration; migrations run in numeric order.
- Each service's migrations target only its own keyspace (for example `auth_keyspace` for `auth-service`); there is no shared keyspace and no shared migration.
- Files are `.cql` and idempotent (`CREATE TABLE IF NOT EXISTS`). Do **not** use `CREATE INDEX`/`CREATE CUSTOM INDEX`: Amazon Keyspaces does not support secondary indexes, so lookups use dedicated tables.
- The runner records `(version, applied_at, checksum)`; a changed checksum on an applied migration fails the run.
- Destructive operations (`DROP`, column removal) are separate, reviewed migrations never mixed with additive changes.

## 2. Data Modeling (Query-First)

Model from access patterns, not from entities. Decide the query, then design the table.

| Query                                              | Table                  | Key                                      |
| -------------------------------------------------- | ---------------------- | ---------------------------------------- |
| Balance by account                                  | `accounts`             | `account_id`                             |
| Accounts by user                                    | `accounts_by_user`     | `((user_id), account_id)`                |
| User by email (login)                               | `users_by_email`       | `(email)`                                |
| Refresh token lookup                                | `refresh_tokens`       | `(token_hash)`                           |
| Refresh tokens by user                              | `refresh_tokens_by_user` | `((user_id), token_hash)`              |
| Transfer by ID                                      | `transfers`            | `(transfer_id)`                          |
| Transfers by account                                | `transfers_by_account` | `((account_id), transfer_id)`            |
| Transactions by account, newest first               | `transactions`         | `((account_id), created_at, transaction_id)` |
| Idempotency key lookup                              | `idempotency_keys`     | `(idempotency_key)`                      |
| Audit events by time                                | `audit_events`         | `((service, day), timestamp, event_id)`  |

Every non-primary-key access path is a **lookup table** written by the owning service, never a secondary index — both because indexes scale poorly in Cassandra and because Amazon Keyspaces rejects `CREATE INDEX`.

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

| Operation                        | Production (RF=3) | Local / CI (RF=1)    |
| -------------------------------- | ------------------ | -------------------- |
| Reads                            | `gocql.Quorum`     | `gocql.One`          |
| Writes (non-LWT)                 | `gocql.Quorum`     | `gocql.One`          |
| Lightweight transactions (paxos) | `gocql.Serial`     | `gocql.Serial`       |
| Read-after-write on own balance  | `gocql.Quorum`     | `gocql.One`          |

Set consistency explicitly at the query, and keep timeouts bounded so a slow node fails loudly. The level comes from configuration (environment), never from a hard-coded constant in shared code: a single-node cluster running `QUORUM` is meaningless, and asserting RF=3 locally would fail. Keyspaces in AWS honours `QUORUM`/`SERIAL` against its managed RF=3.

## 4. Lightweight Transactions & Atomicity

Each balance change must be atomic and must never allow a negative balance. The debit is a single conditional (paxos) statement; the credit is a separate statement. Because the two accounts may live in different services/keyspaces, they are **not** one atomic unit — the saga reconciles them with a compensating reversal if the credit fails. Do **not** wrap the debit and credit in a `BATCH`: multi-partition batches are not atomic, and Amazon Keyspaces does not support lightweight transactions inside batches.

```go
// internal/repositories/transfers.go
const debit = `
    UPDATE accounts SET balance_cents = balance_cents - ?, updated_at = toTimestamp(now())
        WHERE account_id = ? IF balance_cents >= ?;
`
const credit = `
    UPDATE accounts SET balance_cents = balance_cents + ?, updated_at = toTimestamp(now())
        WHERE account_id = ?;
`

// Debit source (conditional): applied == false means insufficient funds.
func (r *AccountRepository) Debit(ctx context.Context, accountID gocql.UUID, amount int64) error {
    iter := r.session.Query(debit, amount, accountID, amount).
        WithContext(ctx).Consistency(gocql.Serial).Iter()
    applied, err := iter.ScanCAS()
    if err != nil {
        return fmt.Errorf("debit: %w", err)
    }
    if !applied {
        return ErrInsufficientFunds
    }
    return iter.Close()
}

// Credit destination (separate statement, no cross-account transaction).
func (r *AccountRepository) Credit(ctx context.Context, accountID gocql.UUID, amount int64) error {
    return r.session.Query(credit, amount, accountID).
        WithContext(ctx).Consistency(gocql.Quorum).Exec()
}
```

Rules:

- The conditional (`IF balance_cents >= ?`) update is the guard: insufficient funds means `applied == false`, not an error.
- The debit and credit are **separate** statements; never assume both applied. The saga reverses a completed debit when the credit fails.
- Always `Close()` the iterator and check `ScanCAS` before proceeding on a conditional statement.
- One LWT per logical operation; serialize paxos rounds to avoid contention on the same partition.
- Never read-then-write outside a conditional statement for balances — that is a race.

## 5. Query Patterns & Performance

- Prefer single-partition queries; a query should touch one partition key.
- Avoid `ALLOW FILTERING`. If a query needs it, add a dedicated table (query-first) instead.
- There are **no secondary indexes** (`CREATE INDEX`). Serve low-cardinality lookups (`user_id`, `email`) with a dedicated lookup table partitioned by the lookup column.
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
- CI runs migrations against each service's own test keyspace (`<service>_keyspace_test`, `SimpleStrategy`, RF 1) and fails on drift.

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
