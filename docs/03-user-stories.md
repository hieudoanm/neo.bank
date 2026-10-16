# NEO Bank — User Stories

## 1. Purpose

This document describes user stories for NEO Bank, expressing user goals in terms of role, motivation, and acceptance criteria.

It complements:

- `00-overview.md` — product vision and scope.
- `01-requirements.md` — functional and non-functional requirements (FR-001..FR-052).
- `02-features.md` — user-visible behaviour and implementation.
- `04-architecture.md` — system boundaries and technical responsibilities.
- `05-data-model.md` — entities, relationships, and constraints.
- `06-api-specification.md` — REST endpoints and request/response models.

User stories guide implementation and testing. Each story states a role, a goal, a benefit, and concrete acceptance criteria.

## 2. User Roles

### 2.1 Personas

| Persona            | Primary goal                                       |
| ------------------ | -------------------------------------------------- |
| Account Holder     | Manage account, view balance, transfer funds       |
| Developer/Operator | Deploy, maintain, and monitor the banking platform |

### 2.2 Account Context

Each authenticated user has access to their own account and can perform operations within that account's scope. The system does not support shared accounts, joint accounts, or beneficiaries in the initial release.

## User Stories

Each user story is documented in a separate file:

- [US-001: Register a new account](./03-user-stories/US-001.md)
- [US-002: Log in to an existing account](./03-user-stories/US-002.md)
- [US-003: Log out and invalidate session](./03-user-stories/US-003.md)
- [US-004: View current balance](./03-user-stories/US-004.md)
- [US-005: Initiate a transfer to another account](./03-user-stories/US-005.md)
- [US-006: Monitor transfer status](./03-user-stories/US-006.md)
- [US-007: Process transfer asynchronously](./03-user-stories/US-007.md)
- [US-008: View transaction history](./03-user-stories/US-008.md)
- [US-009: Prevent unauthorised account access](./03-user-stories/US-009.md)
- [US-010: Prevent duplicate transfers](./03-user-stories/US-010.md)
- [US-011: Handle transfer failures gracefully](./03-user-stories/US-011.md)
- [US-012: Handle concurrent transfers](./03-user-stories/US-012.md)
- [US-013: Monitor system health](./03-user-stories/US-013.md)
- [US-014: Audit financial events](./03-user-stories/US-014.md)
