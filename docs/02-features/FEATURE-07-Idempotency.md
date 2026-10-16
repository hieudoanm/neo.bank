# FEATURE-07: Idempotency

## Priority

P0 (FR-031, FR-041, NFR-022)

## Description

The system enforces idempotency for mutating operations using the `Idempotency-Key` header. Repeating a request with the same key and an identical payload returns the cached original response (the original status code and body — `202 Accepted` for transfers) without re-executing the operation. Reusing a key with a different payload is rejected with `409 IDEMPOTENCY_KEY_EXISTS`. Kafka consumers are idempotent: duplicate events with the same idempotency key do not cause double debit/credit. Idempotency keys expire after a configurable period (default 24 hours) to prevent unbounded storage growth.

## User-visible behaviour

- Client provides the idempotency key in the `Idempotency-Key` request header. The key is never sent in the request body.
- Repeated requests with the same idempotency key and identical payload return the cached original response (including the original status code, `202 Accepted` for transfers).
- The same key with a different payload returns `409 IDEMPOTENCY_KEY_EXISTS`.
- No duplicate financial operations occur (no double debit/credit).
- Retry after network failure is safe: the client can resend with the same key.
- Expired idempotency keys result in the request being processed as new.

## Inputs

| Field | Type | Location | Constraints |
|-------|------|----------|-------------|
| Idempotency-Key | string | Header | Canonical mechanism. Required on `POST /api/v1/transfers`; optional but supported on other mutating endpoints. 1–255 chars matching `[a-zA-Z0-9_-]` |

## Outputs

| Condition | Result |
|-----------|--------|
| New idempotency key | Request processed normally; result stored with key |
| Existing idempotency key, identical payload | Cached original response returned (status + body, e.g., `202`); no operation performed |
| Existing idempotency key, different payload | `409 IDEMPOTENCY_KEY_EXISTS` |
| Expired idempotency key | Request processed as new; new result stored |

## Behaviour

### Server-side Idempotency (Transfer Service)

1. **Check idempotency key**: Look up the key in the Transfer Service's private Cassandra store (`idempotency_keys`).
2. **If key exists, not expired, and the payload matches**: Return the cached original response (status code and body — `202 Accepted` with the original `transferId`) without creating a new transfer.
3. **If key exists with a different payload**: Return `409 IDEMPOTENCY_KEY_EXISTS`.
4. **If key does not exist or has expired**: Process the request normally.
5. **Store result**: After successful processing, store the result (transfer ID, status, payload hash) with the idempotency key and expiry timestamp.
6. **Kafka consumer idempotency**: Before processing a `transfer.initiated` event, check if the idempotency key has already been processed. If so, skip processing and emit the existing result.

### Client-side Retry Safety

- The client generates a UUID v4 idempotency key for each mutating request.
- On network failure (timeout, connection error), the client retries with the same key.
- The server recognises the key and returns the cached result instead of reprocessing.

## API Contract

```
POST /api/v1/transfers
Content-Type: application/json
Authorization: Bearer <accessToken>
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440002

Request:
{
  "destinationAccountId": "550e8400-e29b-41d4-a716-446655440001",
  "amountCents": 5000
}

Response (202 Accepted - first request):
{
  "transferId": "550e8400-e29b-41d4-a716-446655440003",
  "status": "pending"
}

Response (202 Accepted - duplicate request with same key and identical payload; cached original response):
{
  "transferId": "550e8400-e29b-41d4-a716-446655440003",
  "status": "pending"
}

Response (409 - same key with a different payload):
{
  "error": {
    "code": "IDEMPOTENCY_KEY_EXISTS",
    "message": "Idempotency key already used with a different payload",
    "requestId": "req_..."
  }
}
```

## States

| State | Description |
|-------|-------------|
| `new_idempotency_key` | Idempotency key not seen before; request will be processed |
| `existing_idempotency_key` | Key already processed with the same payload; cached original response returned |
| `idempotency_key_conflict` | Key reused with a different payload; `409 IDEMPOTENCY_KEY_EXISTS` |
| `expired_idempotency_key` | Key has expired; request processed as new |

## Edge cases

- **Expired idempotency key**: Key past its TTL (default 24 hours); request processed as new.
- **Same key, different payload**: Returns `409 IDEMPOTENCY_KEY_EXISTS`; client must use a new key or resend the original payload.
- **Concurrent requests with same idempotency key**: Cassandra LWT ensures only one request processes; others wait or receive the result.
- **Different operations with same idempotency key**: Keys are scoped per operation type (e.g., transfer vs. balance update); no cross-operation collision.
- **Idempotency key storage growth**: Keys expire after 24 hours; expired keys are cleaned up by a background TTL compaction.
- **Kafka duplicate delivery**: Consumer checks idempotency before processing; duplicate events are skipped.

## Acceptance conditions

- Duplicate requests with the same idempotency key and an identical payload return the cached original response (for transfers, the original `202 Accepted` with the original `transferId`) without creating duplicate operations.
- The same idempotency key with a different payload returns `409 IDEMPOTENCY_KEY_EXISTS`.
- No duplicate financial operations occur (debited once, credited once per transfer).
- Idempotency keys expire after configurable period (default 24 hours).
- Kafka consumers are idempotent under duplicate message delivery.
- Idempotency key check completes within 10 ms (p95) — planned target; no NFR currently defines a latency target for the idempotency check ([`NFR-003`](../01-requirements/non-functional-requirements.md#performance) covers balance inquiry only).
- `POST /api/v1/transfers` requires an idempotency key per [`NFR-022`](../01-requirements/non-functional-requirements.md#consistency-and-durability); other mutating endpoints support the header optionally.

## Cross-references

- Functional requirements: [`FR-031`](../01-requirements/functional-requirements.md#fund-transfer), [`FR-041`](../01-requirements/functional-requirements.md#fund-transfer)
- Non-functional requirements: [`NFR-022`](../01-requirements/non-functional-requirements.md#consistency-and-durability)
- Fund Transfer: [`FEATURE-04`](FEATURE-04-Fund-Transfer.md)
- Saga pattern: [`docs/04-architecture.md`](../04-architecture.md#32-saga-pattern-for-transfers)
- Planned test: `test_idempotency` (duplicate delivery, concurrent requests, expired keys)