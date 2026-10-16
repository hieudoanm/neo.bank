# NEO Bank — Implementation Plan

## 1. Purpose

This document defines the phased implementation plan for NEO Bank, including milestones, dependencies, deliverables, and acceptance criteria. The plan is organised into 12 phases (0–11) delivered over approximately 8 weeks.

The implementation plan must:

- Deliver a complete vertical slice before expanding scope.
- Prioritise financial integrity and security.
- Ensure each milestone is runnable and reviewable.
- Account for dependencies between phases.
- Keep every phase independently testable and demonstrable.

## 2. Implementation Principles

1. **Vertical slices first:** deliver complete workflows (register → login → balance → transfer → history) before expanding features.
2. **Financial integrity first:** ensure balance consistency, transfer consistency (saga compensation), and idempotency before any other features.
3. **Security first:** implement authentication and authorisation before any protected operations.
4. **Test-driven development:** write tests alongside implementation, not after.
5. **Documentation updates:** update docs when behaviour or scope changes; record significant decisions as ADRs in `docs/decisions/` per AGENTS.md delivery rules.
6. **Incremental delivery:** keep every milestone runnable and reviewable.
7. **Fail fast:** healthchecks and readiness probes gate startup and traffic.
8. **Idempotent and reversible:** every deploy has a documented rollback.

## 3. Phases Overview

| Phase    | Document                                                                                        | Focus                                      | Duration  |
| -------- | ----------------------------------------------------------------------------------------------- | ------------------------------------------ | --------- |
| Phase 0  | [`phase-0-foundation.md`](./10-implementation-plan/phase-0-foundation.md)                       | Project structure, tooling, infrastructure | Week 1    |
| Phase 1  | [`phase-1-data-model.md`](./10-implementation-plan/phase-1-data-model.md)                       | Cassandra schema, migrations, fixtures     | Weeks 1-2 |
| Phase 2  | [`phase-2-authentication.md`](./10-implementation-plan/phase-2-authentication.md)               | Register, login, JWT, Argon2id             | Weeks 2-3 |
| Phase 3  | [`phase-3-balance-inquiry.md`](./10-implementation-plan/phase-3-balance-inquiry.md)             | GET /accounts/{accountId}/balance, dashboard | Week 3    |
| Phase 4  | [`phase-4-transfer-initiation.md`](./10-implementation-plan/phase-4-transfer-initiation.md)     | POST /transfers, Kafka producer            | Week 4    |
| Phase 5  | [`phase-5-kafka-consumer.md`](./10-implementation-plan/phase-5-kafka-consumer.md)               | Saga execution, LWT, idempotency           | Weeks 4-5 |
| Phase 6  | [`phase-6-transfer-status.md`](./10-implementation-plan/phase-6-transfer-status.md)             | GET /transfers/{transferId}, polling        | Week 5    |
| Phase 7  | [`phase-7-transaction-history.md`](./10-implementation-plan/phase-7-transaction-history.md)     | GET /accounts/{accountId}/transactions, pagination | Weeks 5-6 |
| Phase 8  | [`phase-8-security-idempotency.md`](./10-implementation-plan/phase-8-security-idempotency.md)   | Security, idempotency, performance tests   | Week 6    |
| Phase 9  | [`phase-9-health-observability.md`](./10-implementation-plan/phase-9-health-observability.md)   | Health checks, logging, metrics            | Weeks 6-7 |
| Phase 10 | [`phase-10-documentation-demo.md`](./10-implementation-plan/phase-10-documentation-demo.md)     | Documentation, demo                        | Week 7    |
| Phase 11 | [`phase-11-cloud-infrastructure.md`](./10-implementation-plan/phase-11-cloud-infrastructure.md) | Terraform AWS/GCP, K8s manifests (optional) | Week 8+  |

## 4. Milestones

See [`milestones.md`](./10-implementation-plan/milestones.md) for milestone details:

- **Milestone 1: Core Authentication** (End of Week 3) — Registration, login, token refresh, logout.
- **Milestone 2: Basic Banking** (End of Week 5) — Balance inquiry, transfer initiation, saga execution (debit/credit consumers), transfer status.
- **Milestone 3: Complete MVP** (End of Week 7) — Transaction history, security/idempotency tests, health/observability, documentation.
- **Milestone 4: Cloud Deployment** (Optional, Week 8+) — Terraform for AWS/GCP, cloud deployment.

## 5. Risk Mitigation

See [`risk-mitigation.md`](./10-implementation-plan/risk-mitigation.md) for the five tracked risks: Cassandra complexity, Kafka consumer idempotency, financial data integrity, time estimation, and security vulnerabilities.

## 6. Definition of Done

See [`definition-of-done.md`](./10-implementation-plan/definition-of-done.md) — the single canonical Definition of Done for this plan. Phase documents and milestones reference it rather than restating criteria.

## 7. Dependency Graph

```
Phase 0 → Phase 1 → Phase 2 → Phase 3 → Phase 4 → Phase 5 → Phase 6 → Phase 7 → Phase 8 → Phase 9 → Phase 10 → Phase 11
```

- Phase 0 has no dependencies.
- Phase 1 depends on Phase 0 (project structure).
- Phase 2 depends on Phase 1 (data model for users/refresh_tokens).
- Phase 3 depends on Phase 2 (authentication for protected endpoints).
- Phase 4 depends on Phase 3 (account validation for transfers).
- Phase 5 depends on Phase 4 (transfer events to consume; the Account Service applies the debit/credit, the Transfer Service advances the saga).
- Phase 6 depends on Phase 5 (transfer status set by the saga).
- Phase 7 depends on Phase 6 (transaction records from transfer completion).
- Phase 8 depends on Phase 7 (security tests on complete flow).
- Phase 9 depends on Phase 8 (observability on complete system).
- Phase 10 depends on Phase 9 (documentation of complete system).
- Phase 11 depends on Phase 10 (cloud deployment of complete system).

## 8. Acceptance Criteria (Plan Level)

- Each phase has clearly defined deliverables and acceptance criteria.
- Each phase has identified dependencies on prior phases.
- Each phase has explicit acceptance criteria that are verifiable.
- Each phase includes test tasks (unit, integration, frontend).
- Each phase includes documentation tasks.
- Significant decisions are recorded as ADRs in `docs/decisions/` per AGENTS.md delivery rules.
- Milestones align with phase boundaries.
- Risk mitigation strategies are documented for top risks.

## Cross-references

- Requirements: [`docs/01-requirements.md`](01-requirements.md)
- Features: [`docs/02-features.md`](02-features.md)
- Architecture: [`docs/04-architecture.md`](04-architecture.md)
- Testing strategy: [`docs/09-testing-strategy.md`](09-testing-strategy.md)
- Deployment: [`docs/11-deployment.md`](11-deployment.md)
- Operations: [`docs/12-operations.md`](12-operations.md)
- Planned test: `test_plan_consistency` (verifies phase dependencies are consistent)
