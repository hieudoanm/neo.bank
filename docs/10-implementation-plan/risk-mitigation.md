# NEO Bank — Implementation Plan — Risk Mitigation

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Risk 1: Cassandra Complexity

**Risk:** Cassandra's data modelling requirements (query-first design, partition keys, lightweight transactions) may slow development and cause data consistency issues.

**Impact:** High — incorrect data modelling can cause data loss, performance degradation, or incorrect balances.

**Probability:** Medium — the team has limited Cassandra experience.

### Mitigation

- Start with simple queries and denormalised tables optimised for read patterns.
- Use lightweight transactions (Paxos) only for balance updates where atomicity is required.
- Document CQL patterns and anti-patterns in [`docs/05-data-model.md`](../05-data-model.md).
- Use testcontainers with real Cassandra in CI to catch modelling issues early.
- Review all CQL queries in pull requests (query-first design review).
- Consider managed Cassandra (Amazon Keyspaces (AWS) / Cassandra on GKE (GCP)) to reduce operational burden.

### Contingency

- If Cassandra proves too complex, evaluate PostgreSQL as a fallback for the MVP (with a migration path to Cassandra later).

---

## Risk 2: Kafka Consumer Idempotency

**Risk:** Kafka provides at-least-once delivery, meaning consumers may receive duplicate messages. Without proper idempotency, duplicate transfer events could cause double debit/credit.

**Impact:** Critical — financial data integrity is violated if money is duplicated or lost.

**Probability:** Medium — at-least-once delivery is Kafka's default guarantee.

### Mitigation

- Implement idempotency checks early (Phase 5), not as an afterthought.
- Track idempotency keys in Cassandra with TTL-based expiry (24 hours).
- Test extensively with duplicate messages (consumer tests assume duplicate delivery).
- Use Kafka transactions exactly-once semantics where feasible (producer side).
- Document the idempotency strategy in [`docs/02-features/FEATURE-07-Idempotency.md`](../02-features/FEATURE-07-Idempotency.md).
- Implement compensation logic for partial failures (reverse debit if credit fails).

### Contingency

- If idempotency proves difficult, implement a dual-write pattern with a database outbox table as a fallback.

---

## Risk 3: Financial Data Integrity

**Risk:** Balance inconsistencies, negative balances, or lost transactions would violate the core guarantee of a banking system.

**Impact:** Critical — financial data integrity is the foundation of the product.

**Probability:** Low — with lightweight transactions and idempotency, the risk is manageable.

### Mitigation

- Use Cassandra lightweight transactions (Paxos) for all balance updates: `UPDATE accounts SET balance_cents = balance_cents - ? WHERE account_id = ? IF balance_cents >= ?`.
- Enforce non-negative balances at the application layer AND the database layer.
- Write comprehensive tests for concurrent transfers (race conditions).
- Audit all financial operations (every balance change has a transaction record).
- Implement compensating transactions for saga failures (reverse debit on credit failure).
- Review all financial code paths in pull requests with a second reviewer.

### Contingency

- If integrity issues are found, pause feature development and implement a reconciliation process to detect and correct inconsistencies.

---

## Risk 4: Time Estimation

**Risk:** The 8-week timeline may be optimistic given the complexity of distributed systems, event-driven processing, and financial integrity requirements.

**Impact:** Medium — delayed delivery, scope reduction, or quality issues.

**Probability:** High — estimates for distributed systems are often optimistic.

### Mitigation

- Buffer time for each phase (20-30% contingency).
- Prioritise the critical path (registration → login → balance → transfer → history).
- Be prepared to defer non-essential features (dark mode, advanced filtering, etc.).
- Deliver vertical slices early so stakeholders can see progress.
- Re-estimate at each milestone boundary based on actual velocity.

### Contingency

- If the timeline slips, defer Phase 11 (cloud infrastructure) and focus on the MVP (Phases 0-10).
- Reduce scope of non-critical features rather than compromising on financial integrity or security.

---

## Risk 5: Security Vulnerabilities

**Risk:** Authentication or authorisation vulnerabilities could expose user accounts or financial data.

**Impact:** Critical — security breaches destroy trust and may have legal consequences.

**Probability:** Low — with standard patterns (JWT, Argon2id, parameterised queries), the risk is manageable.

### Mitigation

- Use established libraries (argon2, jose, gocql) rather than custom crypto.
- All queries are parameterised; no string interpolation into CQL.
- `404` preferred over `403` for unauthorised access (prevents account enumeration).
- Rate limiting on auth endpoints (5 failures per 15 minutes).
- Security tests for every authorisation path (negative tests).
- Trivy security scans in CI for container and dependency vulnerabilities.
- Regular dependency updates for security patches.

### Contingency

- If a vulnerability is found, pause deployment, patch immediately, and rotate affected credentials/tokens.

## Cross-references

- Implementation plan: [`10-implementation-plan.md`](../10-implementation-plan.md#5-risk-mitigation)
- Financial integrity rules: `AGENTS.md`
- Security rules: `AGENTS.md`
- Testing strategy: [`docs/09-testing-strategy.md`](../09-testing-strategy.md)
- Planned test: `test_risk_mitigation` (verifies risk mitigations are implemented)