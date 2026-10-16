# NEO Bank — Critical Test Scenarios

## Purpose

Defines the most critical test scenarios that must always pass. These scenarios cover financial integrity, security, event processing, and edge cases. They are executed in every CI run and before every release.

## 1. Financial Integrity

These scenarios protect the core guarantee: money is never created or lost.

| ID | Scenario | Expected Result |
|----|----------|-----------------|
| FIN-001 | Transfer with sufficient funds completes | Debit source + credit destination, transfer status `completed` |
| FIN-002 | Transfer with insufficient funds is rejected | Transfer status `failed`, no balance change |
| FIN-003 | Concurrent transfers from same account | Only one succeeds (LWT prevents overdraft) |
| FIN-004 | Balance never goes negative | `balance_cents >= 0` after every operation |
| FIN-005 | Transfer consistency | A completed debit is always matched by a credit or a compensating reversal; no funds lost or created |
| FIN-006 | Duplicate transfer request (idempotency) | Same transfer ID returned, no double debit/credit |
| FIN-007 | Transaction recording | Every balance change has a corresponding transaction record |
| FIN-008 | Consumer failure mid-saga | Compensation reverses partial debit, transfer marked `failed` |

### Test Implementation

```go
func TestTransfer_InsufficientFunds_Rejected(t *testing.T) {
    // Seed account with 1000 cents
    // Attempt transfer of 5000 cents
    // Assert: transfer status = failed, balance unchanged at 1000
}

func TestTransfer_ConcurrentSameAccount(t *testing.T) {
    // Seed account with 1000 cents
    // Launch 10 concurrent transfers of 600 cents each
    // Assert: exactly one succeeds, balance = 400, 9 fail with INSUFFICIENT_FUNDS
}
```

## 2. Security

| ID | Scenario | Expected Result |
|----|----------|-----------------|
| SEC-001 | Unauthenticated request to protected endpoint | `401 UNAUTHENTICATED` |
| SEC-002 | Access another account's balance | `404 ACCOUNT_NOT_FOUND` (not `403`) |
| SEC-003 | Invalid login credentials | `401 INVALID_CREDENTIALS` (generic) |
| SEC-004 | Rate limit exceeded (5 failures / 15 min) | `429 RATE_LIMITED` |
| SEC-005 | Expired access token | `401 TOKEN_EXPIRED` |
| SEC-006 | Revoked refresh token | `401 INVALID_REFRESH_TOKEN` |
| SEC-007 | SQL injection in input | Rejected with `400 VALIDATION_ERROR` |
| SEC-008 | CQL injection in input | Rejected with `400 VALIDATION_ERROR` |
| SEC-009 | Password stored as plain text | Test asserts password_hash is not equal to plaintext |

### Test Implementation

```go
func TestAuth_AccessAnotherAccount_Returns404(t *testing.T) {
    // Login as user A
    // Request user B's balance
    // Assert: 404 ACCOUNT_NOT_FOUND (not 403)
}
```

## 3. Event Processing

| ID | Scenario | Expected Result |
|----|----------|-----------------|
| EVT-001 | Consumer idempotent under duplicate delivery | No double debit/credit when an event is delivered more than once |
| EVT-002 | Consumer failure does not cause data loss | Message redelivered, processed successfully |
| EVT-003 | Consumer handles duplicate messages | Idempotency check skips already-processed events |
| EVT-004 | Transfer status transitions correctly | `pending` → `processing` → `completed` or `failed` |
| EVT-005 | Compensation on credit failure | Source debit reversed, transfer marked `failed` |
| EVT-006 | DLQ for poison messages | Malformed events routed to `transfers.dlq` |

### Test Implementation

```go
func TestConsumer_IdempotentUnderDuplicateDelivery(t *testing.T) {
    // Publish same transfer.initiated event twice
    // Assert: account debited exactly once, transfer status = completed
}
```

## 4. Edge Cases

| ID | Scenario | Expected Result |
|----|----------|-----------------|
| EDG-001 | Transfer to non-existent account | `404 ACCOUNT_NOT_FOUND` |
| EDG-002 | Transfer with amount ≤ 0 | `400 INVALID_AMOUNT` |
| EDG-003 | Transfer to same account | `400 SAME_ACCOUNT` |
| EDG-004 | Expired idempotency key | Request processed as new |
| EDG-005 | Network error during transfer initiation | `503 SERVICE_UNAVAILABLE`, retry safe |
| EDG-006 | Empty transaction history | Empty list with `hasMore: false` |
| EDG-007 | Invalid pagination cursor | `400 VALIDATION_ERROR` |

## 5. Execution Requirements

- All critical scenarios run in every CI pipeline run.
- Critical scenarios are tagged `@critical` in Playwright and `//go:build` tags in Go.
- A failure in any critical scenario blocks the merge.
- Critical scenarios are reviewed quarterly for completeness.

## Cross-references

- Testing strategy: [`09-testing-strategy.md`](../09-testing-strategy.md#2-test-layers)
- Financial integrity rules: `AGENTS.md`
- Security rules: `AGENTS.md`
- Saga pattern: [`docs/04-architecture.md`](../04-architecture.md#32-saga-pattern-for-transfers)
- Planned test: `test_critical_scenarios` (runs all FIN/SEC/EVT/EDG scenarios)