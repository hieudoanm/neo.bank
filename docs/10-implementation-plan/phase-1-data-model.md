# NEO Bank — Implementation Plan — Phase 1: Data Model and Migrations

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Define and implement the Cassandra data model, migration runner, and test fixtures. Phase 1 establishes the schema that every data-touching phase depends on.

**Week:** 1-2

## Deliverables

| Deliverable | Description |
|-------------|-------------|
| CQL schema | Tables for accounts, users, refresh_tokens, transfers, transactions, idempotency_keys, audit_events |
| Migration runner | Go-based migration runner applying versioned CQL scripts |
| Test fixtures | Synthetic, deterministic test data for all tables |
| Documentation | [`docs/05-data-model.md`](../05-data-model.md) updated with full schema |

## Tasks

- [ ] Define CQL schema for `accounts`, `users`, `refresh_tokens`, `transfers`, `transactions`, `idempotency_keys`, `audit_events`.
- [ ] Choose partition keys and clustering columns optimised for query-first access patterns.
- [ ] Write CQL migration scripts (`001_create_accounts.cql`, `002_create_users.cql`, etc.).
- [ ] Implement migration runner in Go (reads `.cql` files, tracks applied migrations).
- [ ] Create the per-service test keyspaces (`auth_keyspace_test`, `account_keyspace_test`, `transfer_keyspace_test`, `transaction_keyspace_test`) in Cassandra (`SimpleStrategy`, replication factor 1).
- [ ] Write seed script for deterministic synthetic test data.
- [ ] Verify schema with manual testing (CQL shell or Go client).
- [ ] Document table definitions, indexes, and access patterns in `05-data-model.md`.
- [ ] Write unit tests for migration runner (idempotent apply, version tracking).

## Acceptance Criteria

- Migration runner applies all migrations successfully against a fresh Cassandra.
- All tables created with correct schema (partition keys, clustering columns, TTLs where applicable).
- Seed script populates test data deterministically.
- Data model documented with table definitions and CQL examples.
- Migration runner is idempotent (re-applying is a no-op).
- Tests pass (unit, integration with real Cassandra).

## Dependencies

Phase 0 (project structure, Docker Compose with Cassandra).

## Testing

- Unit tests for migration runner logic (version tracking, idempotent apply).
- Integration tests against real Cassandra verifying schema creation and seed data.
- Cassandra-dependent tests use a real database, not mocks.

## Risks & Considerations

- **Query-first data modelling:** schema must be designed around access patterns, not normalisation.
- **Lightweight transactions:** balance updates require Paxos (conditional updates) for atomicity.
- **Schema evolution:** migrations must be additive and idempotent; no destructive changes in production.

## Cross-references

- Data model: [`docs/05-data-model.md`](../05-data-model.md)
- Test data management: [`docs/09-testing-strategy/05-test-data-management.md`](../09-testing-strategy/05-test-data-management.md)
- Cassandra backend: [`docs/08-tech-stack/backend/cassandra.md`](../08-tech-stack/backend/cassandra.md)
- Definition of Done: [`definition-of-done.md`](./definition-of-done.md)
- Planned test: `test_phase_1_data_model` (verifies schema and migrations are correct)