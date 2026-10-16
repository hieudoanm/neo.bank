# NEO Bank — Features

## 1. Purpose

This document describes how NEO Bank features should behave from the user's perspective and how their behaviour should be implemented.

It complements:

- `00-overview.md` — product vision and scope.
- `01-requirements.md` — functional and non-functional requirements (FR-001..FR-052).
- `03-user-stories.md` — user goals and acceptance scenarios.
- `04-architecture.md` — system boundaries and technical responsibilities.
- `05-data-model.md` — entities, relationships, and constraints.
- `06-api-specification.md` — REST endpoints and request/response models.
- `08-tech-stack.md` — technology choices and rationale.
- `07-ui-ux-design.md` — interface and interaction design.

Features should be delivered in vertical slices: a user should be able to complete a useful workflow through the frontend, API, database, and relevant background services before the next major workflow is started.

The canonical entities referenced throughout are `accounts`, `users`, `refresh_tokens`, `transfers`, `transactions`, `idempotency_keys`, and `audit_events`.

NEO Bank is a software demonstration. It does not provide financial advice, process real money, or claim compliance with banking regulations.

Each feature section below states its **user-visible behaviour**, **inputs**, **outputs**, **states**, **edge cases**, and **acceptance conditions**. Priority references map to the P0–P1 definitions and to functional requirements FR-001..FR-052.

## 2. Feature Priorities

| Priority | Meaning                             | Product stage |
| -------- | ----------------------------------- | ------------- |
| P0       | Required for a usable, secure MVP   | Core platform |
| P1       | Important for a complete experience | Enhancement   |

## 3. Feature Completion Definition

A feature is complete when: behaviour documented, FE/BE implement contract, authz enforced, validation/failure handled, tests pass, migrations added, logs/ops adequate, docs updated, no critical security issues, E2E exercised where practical, quality gates pass.

- Feature behaves as documented in the specification.
- Frontend and backend implementations match the agreed-upon contract.
- Authorization is properly enforced at all access points.
- Input validation and error handling are implemented correctly.
- All tests pass, including unit, integration, and end-to-end where applicable.
- Database migrations are added and tested.
- Logging and observability are adequate for monitoring and debugging.
- Documentation is updated to reflect the implemented feature.
- No critical security issues are identified during review.
- End-to-end testing is performed where practical.
- Quality gates (linting, testing, security checks) all pass.

## 4. MVP Definition of Done

MVP ready when core workflow works end-to-end: register → login → view balance → initiate transfer → monitor status → view transaction history. Idempotency enforced, security tested, error handling complete, UI states implemented, health/audit/observability in place.

- Users can successfully complete the core banking workflow:
  1. Register a new account
  2. Log in to the account
  3. View account balance
  4. Initiate a fund transfer
  5. Monitor transfer status
  6. View transaction history
- Idempotency is properly enforced for all mutating operations.
- Security measures are implemented and tested.
- Error handling is complete and follows established patterns.
- All UI states (loading, empty, success, error, etc.) are implemented.
- Health checks, audit logging, and observability components are functioning.

## 5. Features

Each feature is documented in a separate file:

- [FEATURE-01: Account Registration](./02-features/FEATURE-01-Account-Registration.md)
- [FEATURE-02: Authentication](./02-features/FEATURE-02-Authentication.md)
- [FEATURE-03: Balance Inquiry](./02-features/FEATURE-03-Balance-Inquiry.md)
- [FEATURE-04: Fund Transfer](./02-features/FEATURE-04-Fund-Transfer.md)
- [FEATURE-05: Transaction History](./02-features/FEATURE-05-Transaction-History.md)
- [FEATURE-06: Security and Access Control](./02-features/FEATURE-06-Security-and-Access-Control.md)
- [FEATURE-07: Idempotency](./02-features/FEATURE-07-Idempotency.md)
- [FEATURE-08: Error Handling](./02-features/FEATURE-08-Error-Handling.md)
- [FEATURE-09: UI States (Global)](./02-features/FEATURE-09-UI-States-Global.md)
