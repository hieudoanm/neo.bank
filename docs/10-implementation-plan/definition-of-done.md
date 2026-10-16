# NEO Bank — Implementation Plan — Definition of Done

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Purpose

Defines when a phase is considered complete. A phase is not "done" until every criterion below is satisfied and verified.

## Phase Completion Criteria

A phase is complete when:

| #   | Criterion                   | Detail                                                                                          |
| --- | --------------------------- | ----------------------------------------------------------------------------------------------- |
| 1   | All tasks completed         | Every task in the phase's task list is checked off.                                             |
| 2   | All tests pass              | Unit, integration, and frontend tests for the phase pass in CI.                                 |
| 3   | Code reviewed and approved  | At least one reviewer approves all phase changes.                                               |
| 4   | Documentation updated       | All affected docs are updated to reflect the phase's changes.                                   |
| 5   | CI pipeline passes          | Lint, typecheck, tests, and security scans all pass.                                            |
| 6   | Phase deliverables verified | Each deliverable listed in the phase document is demonstrated.                                  |
| 7   | Dependencies satisfied      | All prior phases are complete and their deliverables are used.                                  |
| 8   | No regression               | Existing tests continue to pass after phase changes.                                            |
| 9   | Coverage thresholds met     | Backend line coverage ≥ 80%; frontend line coverage ≥ 80% for code changed in the phase.        |
| 10  | Decisions recorded          | Significant decisions are recorded as ADRs in `docs/decisions/` (per AGENTS.md delivery rules). |

## Phase-Level Definition of Done Examples

### Phase 0: Foundation DoD

- [ ] Monorepo structure with pnpm workspace exists.
- [ ] Go backend compiles (`go build ./...`).
- [ ] React frontend builds (`pnpm build`).
- [ ] Docker Compose starts Cassandra and Kafka (`docker-compose up -d`).
- [ ] CI pipeline runs on empty repo (lint, typecheck, test skeleton).
- [ ] README has local development setup instructions.
- [ ] Git hooks (Husky + lint-staged) are configured.
- [ ] All acceptance criteria from `phase-0-foundation.md` are met.

### Phase 2: Authentication DoD

- [ ] `POST /register` endpoint works with real Cassandra.
- [ ] `POST /login` endpoint issues JWT access + refresh tokens.
- [ ] `POST /refresh` endpoint issues new access token from refresh token.
- [ ] `POST /logout` endpoint invalidates refresh token.
- [ ] Passwords hashed with Argon2id (not plain text).
- [ ] JWT tokens validated on protected endpoints (401 on missing/invalid).
- [ ] Register and Login frontend screens work.
- [ ] Unit tests pass for auth service.
- [ ] Integration tests pass for auth endpoints.
- [ ] Frontend tests pass for auth screens.
- [ ] All acceptance criteria from `phase-2-authentication.md` are met.

### Phase 5: Kafka Consumer DoD

- [ ] The Account Service consumes `account.debit.requested`/`account.credit.requested`, and the Transfer Service advances the saga on the resulting events.
- [ ] Idempotency check prevents double processing.
- [ ] Lightweight transactions (Paxos) used for atomic per-account debit and credit, with saga compensation.
- [ ] Transfer status updated to `completed` or `failed`.
- [ ] Transaction entries recorded for both accounts.
- [ ] Compensation reverses debit on credit failure.
- [ ] Duplicate Kafka events handled idempotently.
- [ ] Consumer runs in Docker Compose.
- [ ] Integration tests pass with real Kafka and Cassandra.

## Release-Level Definition of Done

A release is ready when:

- All features in the release meet their individual phase DoD.
- All critical test scenarios (FIN/SEC/EVT/EDG) pass.
- E2E test suite passes in the target environment.
- Performance tests meet targets under load.
- Security scan reports zero high-severity findings.
- Coverage thresholds are met (backend ≥ 80%, frontend ≥ 80%).
- Rollback plan is documented and tested.
- Documentation is updated for all user-facing changes.
- Significant decisions are recorded as ADRs in `docs/decisions/`.
- AGENTS.md is consistent with the implementation.
- This document is the single canonical Definition of Done; phase documents and milestones reference it rather than restating criteria.

## Cross-references

- Implementation plan: [`10-implementation-plan.md`](../10-implementation-plan.md#6-definition-of-done)
- Feature completion: [`docs/02-features.md`](../02-features.md#3-feature-completion-definition)
- Testing DoD: [`docs/09-testing-strategy/11-definition-of-done.md`](../09-testing-strategy/11-definition-of-done.md)
- Delivery rules: `AGENTS.md`
- Planned test: `test_dod` (verifies all DoD criteria are met before phase completion)