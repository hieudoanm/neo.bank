# NEO Bank — Kafka Best Practices

_Backend · part of the [NEO Bank tech stack](../../08-tech-stack.md) decision record · see the [backend index](./README.md)._

## 1. Topic Design

```
transfers          # transfer events, primary workflow topic
transfers.dlq      # failed events after retries are exhausted (future)
transfers-test     # integration-test topic
```

Rules:

- One topic per event family; do not multiplex unrelated events.
- Topic names are lowercase, dot-separated, and declared in configuration (`KAFKA_TOPIC_TRANSFERS`).
- `auto.create.topics.enable=false`; topics are created explicitly — by Terraform in the cloud and by a Compose/CI init step (`kafka-topics --create`) locally — never implicitly by a client.
- Durability settings by environment (`acks=all` is the one invariant):
  - **Production:** `replication.factor=3`, `min.insync.replicas=2`, producer `acks=all`
  - **Local / CI (single broker):** `replication.factor=1`, `min.insync.replicas=1`, producer `acks=all` — a one-broker cluster cannot satisfy RF=3, and internal topics with RF=3 never come up. See [`NFR-013`](../../01-requirements/non-functional-requirements.md#availability).
- Partition count is chosen for parallelism; the transfers topic partitions by **source account ID** so all events for one account are ordered.

## 2. Event Contract

Events are append-only facts, versioned, and self-describing.

```go
// internal/models/events.go
type TransferEvent struct {
    EventID              string `json:"eventId"`          // UUID v7
    EventType            string `json:"eventType"`        // "transfer_initiated"
    EventVersion         int    `json:"eventVersion"`     // schema version
    TransferID           string `json:"transferId"`
    IdempotencyKey       string `json:"idempotencyKey"`
    CorrelationID        string `json:"correlationId"`
    SourceAccountID      string `json:"sourceAccountId"`
    DestinationAccountID string `json:"destinationAccountId"`
    AmountCents          int64  `json:"amountCents"`
    OccurredAt           string `json:"occurredAt"`       // ISO 8601 UTC
}
```

Rules:

- Include `eventId`, `eventType`, `eventVersion`, `idempotencyKey`, and `correlationId` on every event.
- Event payload fields are camelCase JSON; amounts are integer minor units (cents); timestamps are ISO 8601 UTC.
- Additive changes only within a version; a breaking change bumps `eventVersion` and consumers support both during migration.
- Never put secrets, tokens, or PII in an event payload.

### 2.1 Account movement command/result events

The transfer saga communicates with the Account Service as commands and results on the same `transfers` topic. Messages are discriminated by `eventType`, keyed by the affected account ID (ordering per account), and each service filters the event types it owns. Consumers run in distinct consumer groups (`transfer-service-saga`, `account-service-saga`) so every service observes every message.

```go
// internal/models/events.go (Transfer and Account services)
type AccountCommand struct {
    EventID                 string `json:"eventId"`        // UUID v7
    EventType               string `json:"eventType"`      // "account.debit.requested" | "account.credit.requested"
    EventVersion            int    `json:"eventVersion"`
    TransferID              string `json:"transferId"`
    IdempotencyKey          string `json:"idempotencyKey"`
    CorrelationID           string `json:"correlationId"`
    AccountID               string `json:"accountId"`
    AmountCents             int64  `json:"amountCents"`
    CounterpartyAccountID   string `json:"counterpartyAccountId"`
    OccurredAt              string `json:"occurredAt"`
}

type AccountEvent struct {
    EventID                 string `json:"eventId"`        // UUID v7
    EventType               string `json:"eventType"`      // "account.debited" | "account.debit.failed" | "account.credited" | "account.credit.failed"
    EventVersion            int    `json:"eventVersion"`
    TransferID              string `json:"transferId"`
    IdempotencyKey          string `json:"idempotencyKey"`
    CorrelationID           string `json:"correlationId"`
    AccountID               string `json:"accountId"`
    AmountCents             int64  `json:"amountCents"`
    CounterpartyAccountID   string `json:"counterpartyAccountId"`
    BalanceCents            int64  `json:"balanceCents,omitempty"`
    FailureReason           string `json:"failureReason,omitempty"`
    OccurredAt              string `json:"occurredAt"`
}
```

Rules:

- The **Transfer Service** issues `account.debit.requested` then, on `account.debited`, `account.credit.requested`. The **Account Service** consumes those commands, moves funds with a Cassandra LWT, and emits the matching result event.
- `counterpartyAccountId` is the *other* account in the transfer: the destination on a source-account movement, the source on a destination-account movement. The saga sets it on the command and the Account Service echoes it on the result, so the Transaction Service can denormalise the counterparty into the read model without a cross-keyspace lookup.
- Command `idempotencyKey` is **deterministic per transfer and movement**: `<transferId>:debit`, `<transferId>:credit`, `<transferId>:compensation`. The Account Service claims it in `processed_commands` (see the [Account Service data model](../../05-data-model/account-service/README.md#processed_commands)), so a redelivered command re-emits the stored result instead of moving funds twice.
- On credit failure the Transfer Service compensates by issuing `account.credit.requested` for the **source** account with the `:compensation` key, then marks the transfer `failed`.
- The Transaction Service consumes `account.debited`/`account.credited` to create read-model rows and `transfer.completed`/`transfer.failed` to finalise their status (see the [Transaction Service data model](../../05-data-model/transaction-service/README.md#projection-logic)).

### 2.2 Account provisioning event

When a user registers, the Auth Service publishes `auth.user.created` on the same `transfers` topic (discriminated by `eventType`, keyed by the new account ID). The Account Service consumes it to provision the account, so registration and account creation are decoupled and each service writes only its own keyspace.

```go
// internal/models/events.go (Auth and Account services)
type UserCreatedEvent struct {
    EventID      string `json:"eventId"`      // UUID v7
    EventType    string `json:"eventType"`    // "auth.user.created"
    EventVersion int    `json:"eventVersion"` // 1
    UserID       string `json:"userId"`
    AccountID    string `json:"accountId"`
    OccurredAt   string `json:"occurredAt"`   // ISO 8601 UTC
}
```

Rules:

- The Account Service provisions with an idempotent `INSERT … IF NOT EXISTS`, so at-least-once delivery never duplicates the account or resets a balance.
- The opening balance and currency come from `INITIAL_BALANCE_CENTS` and `CURRENCY_CODE` (see [configuration](./configuration.md)).
- The Auth Service publishes **after** the user is created and rolls the user back if the publish fails, returning `503`; the user and account therefore stay consistent.
- Other consumers (Transfer saga, Transaction projection) ignore event types they do not own and require `eventVersion == 1`.

## 3. Producing

```go
// internal/services/transfer.go
func (p *TransferProducer) Publish(ctx context.Context, ev TransferEvent) error {
    data, err := json.Marshal(ev)
    if err != nil {
        return fmt.Errorf("marshal transfer event: %w", err)
    }

    return p.writer.WriteMessages(ctx, kafka.Message{
        Topic: p.topic,
        Key:   []byte(ev.SourceAccountID), // ordering per source account
        Value: data,
        Headers: []kafka.Header{
            {Key: "correlation_id", Value: []byte(ev.CorrelationID)},
            {Key: "event_type", Value: []byte(ev.EventType)},
        },
    })
}
```

Rules:

- The **message key is the source account ID**, so ordering is preserved per account.
- The producer is configured `RequiredAcks = kafka.RequireAll` and `AllowAutoTopicCreation = false`.
- Marshal with `encoding/json`; never hand-build the payload string.
- Publish happens **after** the transfer row is written (status `pending`); if publish fails, the transfer is rolled back and the API returns `503`.

## 4. Consuming

```go
// internal/consumer/transfers.go
func (c *TransferConsumer) Run(ctx context.Context) error {
    for {
        msg, err := c.reader.FetchMessage(ctx)
        if err != nil {
            if errors.Is(err, context.Canceled) {
                return nil
            }
            c.logger.Error("fetch message", "err", err)
            continue
        }

        if err := c.handle(ctx, msg); err != nil {
            c.logger.Error("handle transfer", "err", err, "partition", msg.Partition, "offset", msg.Offset)
            continue // retry: do not commit
        }

        if err := c.reader.CommitMessages(ctx, msg); err != nil {
            c.logger.Error("commit message", "err", err)
        }
    }
}
```

Rules:

- Use `FetchMessage` + explicit `CommitMessages`; commit **only after** the side effects succeed.
- Consumers run in a consumer group (`KAFKA_CONSUMER_GROUP`); partition assignment gives each event to one consumer.
- Graceful shutdown: watch `ctx` (SIGTERM), commit the in-flight message, then exit.
- A message that fails handling is left uncommitted so it is redelivered; the consumer never "skips" a financial event.

## 5. Idempotency & Exactly-Once Processing

Kafka delivers **at least once**. Correctness comes from idempotent processing, not from the broker.

```go
// internal/consumer/transfers.go
func (c *TransferConsumer) handle(ctx context.Context, msg kafka.Message) error {
    var ev TransferEvent
    if err := json.Unmarshal(msg.Value, &ev); err != nil {
        c.logger.Error("unmarshal transfer event", "err", err)
        c.sendToDLQ(ctx, msg) // poison pill: do not block the partition forever
        return nil
    }

    processed, err := c.idempotency.Claim(ctx, ev.IdempotencyKey)
    if err != nil {
        return fmt.Errorf("claim idempotency key: %w", err)
    }
    if processed {
        c.logger.Info("duplicate event skipped", "idempotencyKey", ev.IdempotencyKey)
        return nil
    }

    if err := c.transfers.Apply(ctx, ev); err != nil {
        _ = c.idempotency.Release(ctx, ev.IdempotencyKey) // allow retry
        return fmt.Errorf("apply transfer: %w", err)
    }

    return c.idempotency.Complete(ctx, ev.IdempotencyKey)
}
```

Rules:

- Claim the idempotency key **before** applying side effects; skip when already claimed.
- Each balance change is a Cassandra lightweight transaction (`docs/08-tech-stack/backend/cassandra.md`); the debit and credit are separate statements coordinated by the saga with compensation, not a single atomic unit.
- Release (or leave unclaimed) on failure so a redelivery can retry; complete only after the state update and transaction rows are written.
- Duplicate delivery is normal: tests must deliver the same event twice and assert a single balance change.

## 6. Error Handling, Retries & DLQ

- **Transient** errors (Cassandra/Kafka unavailable, timeout): leave uncommitted and retry with backoff.
- **Poison-pill** errors (malformed payload, unknown version): route to `transfers.dlq` with the original headers and a failure reason, then commit so the partition advances.
- Bounded retries with exponential backoff and jitter; after the bound, send to the DLQ and alert.
- The DLQ consumer is a manual/operational tool: inspect, fix, replay. Never silently drop.

## 7. Ordering Guarantees

- Ordering holds **within a partition only**. Partitioning by source account ID guarantees a single account's transfers apply in order.
- Never assume global ordering across accounts; make each event self-contained.
- Do not repartition an existing topic with in-flight events unless the consumer is drained and offsets are managed.

## 8. Observability

Every log line for a transfer event carries the correlation and transfer identifiers.

```go
c.logger.InfoContext(ctx, "transfer processed",
    "transferId", ev.TransferID,
    "sourceAccountId", ev.SourceAccountID,
    "destinationAccountId", ev.DestinationAccountID,
    "amountCents", ev.AmountCents,
    "correlationId", ev.CorrelationID,
    "partition", msg.Partition,
    "offset", msg.Offset,
)
```

Metrics to export:

- `kafka_consumer_lag` (gauge) — the primary health signal; alert when lag > 1000 for 5 minutes.
- `transfers_processed_total` (counter), `transfers_failed_total` (counter).
- Processing latency histogram.
- DLQ depth.

## 9. Testing

- Integration tests use a **real Kafka** (`transfers-test` topic), never a mock — the delivery contract is the point.
- Deliver the same message twice and assert idempotency (one balance change).
- Kill the consumer mid-processing and assert the event is reprocessed without double debit.
- Assert malformed payloads reach the DLQ and the partition keeps advancing.
- Reset consumer offsets and truncate test topics between suites.

## 10. Code Style & Rules

- Financial operations live **only** in consumers, never in HTTP handlers.
- Every event carries `eventId`, `eventType`, `eventVersion`, `idempotencyKey`, `correlationId`.
- Message key = source account ID; `acks=all`; `auto.create.topics.enable=false`.
- Commit only after side effects succeed; release the idempotency key on failure.
- Duplicate delivery and redelivery are the normal case — design and test for them.
- The consumer is stateless; all state lives in Cassandra keyed by idempotency key.
