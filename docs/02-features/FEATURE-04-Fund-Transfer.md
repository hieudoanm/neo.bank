# FEATURE-04: Fund Transfer

## Priority

P0 (FR-030..FR-043)

## Description

Authenticated users can initiate transfers to other accounts. The system validates sufficient funds, valid destination account, positive amount, and rejects transfers to the same account. Transfers are published to Kafka for asynchronous processing. The system returns a transfer ID and initial status (pending) immediately. Kafka consumers process each balance change atomically using lightweight transactions (debit source, then credit destination), coordinate the two via the saga, update transfer status, and record transaction entries. The debit and credit are separate statements — not one atomic unit — so a failed credit is compensated by reversing the debit. The system enforces idempotency via idempotency keys.

## User-visible behaviour

- Transfer form with destination account ID and amount fields.
- Real-time validation of amount (positive integer) and destination account ID (UUID v7).
- Success message with transfer ID upon submission.
- Transfer status display (pending → processing → completed/failed).
- Loading state while submitting the transfer.
- Error messages for validation failures (insufficient funds, invalid destination, same account, invalid amount).
- Polling for transfer status updates until completion.
- Transaction history updated after transfer completes.

## Inputs

| Field | Type | Location | Required | Constraints |
|-------|------|----------|----------|-------------|
| destinationAccountId | string (UUID v7) | Body | Yes | Must exist, must not equal source account |
| amountCents | integer | Body | Yes | Must be > 0, in minor units (cents) |
| Idempotency-Key | string | Header | Yes | 1–255 chars matching `[a-zA-Z0-9_-]`; required for retry safety |

The source account is derived server-side from the authenticated user and is never a request field.

## Outputs

On initiation (`202 Accepted`):

| Field | Type | Description |
|-------|------|-------------|
| transferId | string (UUID v7) | Unique transfer identifier |
| status | string | Initial status: "pending" |
| amountCents | integer | Transfer amount in minor units |
| sourceAccountId | string (UUID v7) | Sender's account ID |
| destinationAccountId | string (UUID v7) | Recipient's account ID |
| createdAt | string (ISO 8601 UTC) | Transfer creation timestamp |

On error:

| Status | Code | Condition |
|--------|------|-----------|
| `400` | `INSUFFICIENT_FUNDS` | Source account balance is too low |
| `400` | `SAME_ACCOUNT` | Source and destination are the same |
| `400` | `INVALID_AMOUNT` | Amount is not a positive integer |
| `401` | `UNAUTHENTICATED` | Missing or invalid access token |
| `404` | `ACCOUNT_NOT_FOUND` | Destination account does not exist (same code as any unknown or foreign account; no enumeration) |
| `409` | `IDEMPOTENCY_KEY_EXISTS` | Same `Idempotency-Key` replayed with a different payload |
| `429` | `RATE_LIMITED` | Too many transfer requests |
| `503` | `SERVICE_UNAVAILABLE` | Transfer Service or Kafka unavailable |

## Behaviour

### Transfer Initiation (Transfer Service)

1. **Validate JWT access token**: Verify signature, expiry, issuer, audience.
2. **Extract user ID**: From the `sub` claim of the JWT.
3. **Retrieve source account**: Query Account Service to get source account balance.
4. **Validate destination account exists**: Query Account Service (read-only pre-check).
5. **Validate sufficient funds**: Source balance >= amountCents.
6. **Validate amount > 0**: Reject zero or negative amounts.
7. **Validate source != destination**: Reject self-transfers.
8. **Check idempotency key**: If the `Idempotency-Key` was already processed with the same payload, return the cached original response (`202` with the original `transferId`). If the same key arrives with a different payload, return `409 IDEMPOTENCY_KEY_EXISTS`.
9. **Create transfer record**: Status `pending`, stored in Transfer Service's private keyspace.
10. **Publish to Kafka**: Event `transfer.initiated` to the `transfers` topic, keyed by source account ID.
11. **Return response**: `202 Accepted` with transfer ID and pending status.

### Transfer Execution (Kafka Consumer → Account Service)

1. **Consume `transfer.initiated`**: From the `transfers` Kafka topic.
2. **Check idempotency**: If idempotency key already processed, skip (emit `transfer.completed` with existing result).
3. **Debit source account**: Cassandra lightweight transaction (compare-and-set) ensuring balance >= amount.
4. **On debit success**: Emit `account.debited` event to Kafka.
5. **Credit destination account**: Cassandra lightweight transaction adding amount to destination.
6. **On credit success**: Emit `account.credited` event to Kafka.
7. **Update transfer status**: Mark transfer `completed`.
8. **Emit `transfer.completed`**: To Kafka for Transaction Service projection.
9. **On failure at any step**: Emit `account.debit.failed` or `account.credit.failed`; execute compensation (reverse debit if source was already debited); mark transfer `failed`; emit `transfer.failed`.

### Status Polling (Frontend → Transfer Service)

1. **Poll `GET /api/v1/transfers/{transferId}`**: Every 2-3 seconds until status is `completed` or `failed`.
2. **Display status**: pending → processing → completed/failed.
3. **On completed**: Show success message, update balance, add to transaction history.
4. **On failed**: Show error message with reason; offer retry option.

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

Response (202 Accepted):
{
  "transferId": "550e8400-e29b-41d4-a716-446655440003",
  "status": "pending",
  "amountCents": 5000,
  "sourceAccountId": "550e8400-e29b-41d4-a716-446655440000",
  "destinationAccountId": "550e8400-e29b-41d4-a716-446655440001",
  "createdAt": "2026-10-15T20:00:00Z"
}

GET /api/v1/transfers/{transferId}

Response (200):
{
  "transferId": "...",
  "status": "failed",
  "amountCents": 5000,
  "sourceAccountId": "...",
  "destinationAccountId": "...",
  "failureReason": "INSUFFICIENT_FUNDS",
  "createdAt": "...",
  "updatedAt": "..."
}
```

`failureReason` is present only when `status` is `failed` and carries a canonical error code.

## States

| State | Description |
|-------|-------------|
| `pending` | Transfer initiated, waiting for Kafka processing |
| `processing` | Transfer being processed by Kafka consumer |
| `completed` | Transfer completed successfully (debited + credited) |
| `failed` | Transfer failed (insufficient funds, system error, etc.) |
| `reversed` | Transfer reversed (future enhancement, not in MVP) |

## Edge cases

- **Insufficient funds**: Rejects at initiation with `400 INSUFFICIENT_FUNDS`; no Kafka event published.
- **Invalid destination account ID**: Rejects at initiation with `404 ACCOUNT_NOT_FOUND` (same generic code used for any unknown account).
- **Transfer to same account**: Rejects at initiation with `400 SAME_ACCOUNT`.
- **Amount ≤ 0**: Rejects at initiation with `400 INVALID_AMOUNT`.
- **Duplicate transfer request (same idempotency key, same payload)**: Returns the cached original `202` response with the original `transferId` without reprocessing.
- **Same idempotency key, different payload**: Returns `409 IDEMPOTENCY_KEY_EXISTS`.
- **Kafka consumer failure mid-processing**: Compensation reverses any partial debit; transfer marked `failed`.
- **Concurrent transfers from same account**: Cassandra LWT prevents overdraft; one transfer succeeds, others fail with `INSUFFICIENT_FUNDS`.
- **Destination credit fails after source debit**: Compensation reverses the source debit; transfer marked `failed`.

## Acceptance conditions

- Valid transfer request returns transfer ID and `pending` status within 200 ms (p95) per [`NFR-004`](../01-requirements/non-functional-requirements.md#performance).
- Insufficient funds rejects request with clear error; no money moved.
- Unknown destination account rejects the request with `404 ACCOUNT_NOT_FOUND` (same generic code for any unknown or foreign account; no enumeration).
- Transfer to same account rejects request.
- Amount ≤ 0 rejects request.
- Transfer is published to Kafka with all details (keyed by source account ID).
- Kafka consumer applies each balance change atomically using Cassandra lightweight transactions, coordinating the debit and credit with saga compensation.
- Transfer status updates to `completed` or `failed`.
- Transaction entries are recorded for both source and destination accounts.
- Duplicate requests with the same idempotency key and an identical payload return the cached original `202` response (same `transferId`) without reprocessing.
- The same idempotency key with a different payload returns `409 IDEMPOTENCY_KEY_EXISTS`.
- Consumer is idempotent: duplicate events do not cause double debit/credit.
- Transfer processing completes within 2 s (p95) per [`NFR-005`](../01-requirements/non-functional-requirements.md#performance).
- Account balances never go negative per [`NFR-020`](../01-requirements/non-functional-requirements.md#consistency-and-durability).

## Cross-references

- Functional requirements: [`FR-030..FR-043`](../01-requirements/functional-requirements.md#fund-transfer)
- User story: [`US-005`](../03-user-stories/US-005.md)
- API specification: [`docs/06-api-specification.md`](../06-api-specification.md)
- Saga pattern: [`docs/04-architecture.md`](../04-architecture.md#32-saga-pattern-for-transfers)
- Kafka: [`docs/08-tech-stack/backend/kafka.md`](../08-tech-stack/backend/kafka.md)
- Idempotency: [`FEATURE-07`](FEATURE-07-Idempotency.md)
- Planned test: `test_transfer` (unit + integration with real Cassandra and Kafka)