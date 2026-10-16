# NEO Bank — Implementation Plan — Milestones

_Part of [10-implementation-plan.md](../10-implementation-plan.md)._

## Milestone 1: Core Authentication (End of Week 3)

**Goal:** Users can register, authenticate, and manage sessions.

### Deliverables

- Registration endpoint (`POST /register`) with email validation and password strength enforcement.
- Login endpoint (`POST /login`) issuing JWT access tokens (15 min) and refresh tokens (7 days).
- Token refresh endpoint (`POST /refresh`) for obtaining new access tokens.
- Logout endpoint (`POST /logout`) invalidating refresh tokens.
- Argon2id password hashing with secure parameters.
- Register and Login frontend screens with validation.
- Token refresh logic in the frontend API client.
- Unit and integration tests for all auth endpoints.

### Acceptance Criteria

- User can register with a valid email and strong password.
- Duplicate email registration is rejected without revealing existence.
- User can log in with valid credentials and receive tokens.
- Invalid credentials are rejected with a generic error (no account enumeration).
- Access tokens expire after 15 minutes; refresh tokens after 7 days.
- Refresh tokens can be used to obtain new access tokens.
- Logout invalidates refresh tokens.
- Authentication completes within 300 ms (p95) per [`NFR-002`](../01-requirements/non-functional-requirements.md#performance).
- All auth tests pass (unit, integration, frontend).

### Related Documents

- [`FEATURE-01`](../02-features/FEATURE-01-Account-Registration.md)
- [`FEATURE-02`](../02-features/FEATURE-02-Authentication.md)
- [`US-001`](../03-user-stories/US-001.md), [`US-002`](../03-user-stories/US-002.md)

---

## Milestone 2: Basic Banking (End of Week 5)

**Goal:** Users can complete the core banking workflow: register → login → view balance → initiate transfer → monitor status.

### Deliverables

- Balance inquiry endpoint (`GET /api/v1/accounts/{accountId}/balance`) returning balance in cents with currency.
- Dashboard frontend screen with balance display and quick actions.
- Transfer initiation endpoint (`POST /api/v1/transfers`) with validation.
- Transfer frontend screen with form and validation.
- Kafka producer publishing transfer events.
- Account and Transfer consumers processing transfers with lightweight transactions.
- Transfer status endpoint (`GET /api/v1/transfers/{transferId}`) with polling.
- Transfer status frontend screen showing pending → completed/failed.
- Idempotency key handling for transfer requests.
- Unit and integration tests for all banking endpoints.

### Acceptance Criteria

- User can view their balance after login.
- Balance is returned in minor units (cents) with currency code.
- User can initiate a transfer to another account.
- Transfer validation rejects insufficient funds, invalid destination, same account, and non-positive amounts.
- Transfer is published to Kafka and processed asynchronously.
- Transfer status updates from `pending` to `completed` or `failed`.
- User can poll for transfer status and see completion.
- Transaction entries are recorded for both debit and credit.
- Duplicate idempotency keys return existing transfer without double-processing.
- Transfer initiation < 200 ms (p95); processing < 2 s (p95).
- All banking tests pass (unit, integration, frontend).

### Related Documents

- [`FEATURE-03`](../02-features/FEATURE-03-Balance-Inquiry.md)
- [`FEATURE-04`](../02-features/FEATURE-04-Fund-Transfer.md)
- [`US-004`](../03-user-stories/US-004.md), [`US-005`](../03-user-stories/US-005.md)

---

## Milestone 3: Complete MVP (End of Week 7)

**Goal:** A complete, demonstrable MVP with transaction history, security, idempotency, health, and observability.

### Deliverables

- Transaction history endpoint (`GET /api/v1/accounts/{accountId}/transactions`) with cursor-based pagination.
- Transactions frontend screen with pagination.
- Security tests (unauthorised access, rate limiting, input validation).
- Idempotency tests (duplicate requests, duplicate events).
- Performance tests (API response times, consumer throughput).
- Health check endpoints (`GET /healthz`, `GET /readyz`).
- Structured JSON logging with correlation IDs.
- Prometheus metrics (latency, error rate, throughput, Kafka lag).
- Complete documentation (all 12 docs).
- Demo script or video showing the complete workflow.

### Acceptance Criteria

- User can view paginated transaction history ordered newest first.
- Unauthorised account access returns `404` (not `403`).
- Rate limiting enforced on auth and authenticated endpoints.
- SQL/CQL injection attempts are rejected.
- Duplicate requests and events are handled idempotently.
- API response times meet NFR targets.
- Health endpoints return correct status.
- Logs are structured JSON with correlation IDs.
- Metrics are collected and exportable.
- All documentation is complete and consistent.
- Demo recorded and verified.
- MVP ready for review.

### Related Documents

- [`FEATURE-05`](../02-features/FEATURE-05-Transaction-History.md)
- [`FEATURE-06`](../02-features/FEATURE-06-Security-and-Access-Control.md)
- [`FEATURE-07`](../02-features/FEATURE-07-Idempotency.md)
- [`FEATURE-08`](../02-features/FEATURE-08-Error-Handling.md)
- [`FEATURE-09`](../02-features/FEATURE-09-UI-States-Global.md)

---

## Milestone 4: Cloud Deployment (Optional, Week 8+)

**Goal:** Deploy NEO Bank to cloud infrastructure (AWS EKS or GCP GKE).

### Deliverables

- Terraform configuration for AWS (EKS, MSK, Keyspaces, ALB).
- Terraform configuration for GCP (GKE, Cassandra on GKE, Kafka on GKE, Load Balancing).
- Kubernetes manifests for `api`, `consumer`, and `frontend` workloads.
- Deployment documentation in [`11-deployment.md`](../11-deployment.md).
- Staging environment deployment and validation.

### Acceptance Criteria

- Terraform applies successfully for AWS.
- Terraform applies successfully for GCP.
- Application deploys and runs on cloud infrastructure.
- Health checks pass in the cloud environment.
- Deployment documentation complete.

### Related Documents

- [`docs/08-tech-stack/infrastructure/terraform.md`](../08-tech-stack/infrastructure/terraform.md)
- [`docs/08-tech-stack/infrastructure/kubernetes.md`](../08-tech-stack/infrastructure/kubernetes.md)
- [`docs/11-deployment.md`](../11-deployment.md)
- [`docs/12-operations.md`](../12-operations.md)

---

## Milestone Dependencies

```
Milestone 1 → Milestone 2 → Milestone 3 → Milestone 4
```

- Milestone 1 has no dependencies (builds on Phase 0-2).
- Milestone 2 depends on Milestone 1 (authentication for protected endpoints).
- Milestone 3 depends on Milestone 2 (complete banking flow).
- Milestone 4 depends on Milestone 3 (complete MVP ready for cloud deployment).

## Cross-references

- Implementation plan: [`10-implementation-plan.md`](../10-implementation-plan.md#4-milestones)
- Phase documents: [`phase-*.md`](./) for detailed tasks and acceptance criteria.
- Planned test: `test_milestones` (verifies milestone deliverables are met)