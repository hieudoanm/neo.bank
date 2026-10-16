# Non-Functional Requirements

This document specifies the non-functional requirements (NFRs) for NEO Bank, organized by quality attribute. Each requirement carries a unique ID (NFR-###), a description, and a target value. NFRs are traced to architecture components and tests in the traceability table at the end.

## Performance

| ID      | Requirement                                                                | Target |
| ------- | -------------------------------------------------------------------------- | ------ |
| NFR-001 | Account registration shall complete within 500 ms (p95).                   | 500 ms |
| NFR-002 | User login shall complete within 300 ms (p95).                             | 300 ms |
| NFR-003 | Balance inquiry shall complete within 100 ms (p95).                        | 100 ms |
| NFR-004 | Transfer initiation shall return within 200 ms (p95).                      | 200 ms |
| NFR-005 | Transfer processing (Kafka to completion) shall complete within 2 s (p95). | 2 s    |
| NFR-006 | The system shall support 1,000 concurrent users with < 1% error rate.      | 1,000  |

**Rationale:** Balance inquiry is the most frequent operation and must be sub-100ms. Transfer initiation returns quickly (202 Accepted) because the actual money movement is asynchronous. Registration and login involve password hashing (Argon2id) which is intentionally slower but bounded at 500ms/300ms p95.

**See also:** [`docs/04-architecture.md#10-scalability-considerations`](../04-architecture.md#10-scalability-considerations), [`docs/06-api-specification.md`](../06-api-specification.md).

---

## Availability

| ID      | Requirement                                                                    | Target |
| ------- | ------------------------------------------------------------------------------ | ------ |
| NFR-010 | The API shall be available 99.9% of the time (excluding planned maintenance).  | 99.9%  |
| NFR-011 | In production, the system shall tolerate single AZ failure without data loss.  | Must   |
| NFR-012 | In production, Cassandra shall be configured with RF=3 across 3 AZs and QUORUM consistency. | Must   |
| NFR-013 | In production, Kafka shall be configured with replication factor 3 and min.insync.replicas=2. | Must   |

**Rationale:** 99.9% availability allows ~8.77 hours of downtime per year. Cassandra RF=3 across AZs ensures no data loss on single AZ failure. QUORUM consistency (2 of 3) balances availability with consistency. Kafka RF=3 with min.insync.replicas=2 ensures producers can write as long as 2 of 3 replicas are in sync.

**Environment scoping:** NFR-011..013 describe the **production** profile. Local development and CI run a single Cassandra node and a single Kafka broker: `SimpleStrategy` with `replication_factor: 1`, queries at `ONE`/`LOCAL_ONE`, and Kafka `replication.factor=1` with `min.insync.replicas=1` (a `QUORUM` read on a single-node cluster equals `ONE`). The production profile is the one under test for availability claims; the single-node profile only exercises correctness. See [`docker.md`](../08-tech-stack/infrastructure/docker.md#3-compose-topology) and [`local-development.md`](../11-deployment/local-development.md).

**See also:** [`docs/11-deployment.md`](../11-deployment.md), [`docs/12-operations.md`](../12-operations.md), [`Kubernetes`](../08-tech-stack/infrastructure/kubernetes.md).

---

## Consistency and Durability

| ID      | Requirement                                                                         | Target |
| ------- | ----------------------------------------------------------------------------------- | ------ |
| NFR-020 | Account balances shall never go negative (enforced by lightweight transactions).    | Must   |
| NFR-021 | Transfer consistency: every balance change is atomic (LWT, never negative), and the saga with compensation guarantees no funds are lost or created across accounts. | Must   |
| NFR-022 | No transfer shall be processed more than once (idempotency guarantee).              | Must   |
| NFR-023 | All financial events shall be durably persisted in Cassandra before acknowledgment. | Must   |
| NFR-024 | Kafka events shall be persisted with acks=all.                                      | Must   |

**Rationale:** Financial integrity is non-negotiable. Cassandra lightweight transactions (Paxos) provide compare-and-set semantics for each balance update, preventing negative balances under concurrent debits. Each debit and credit is a separate atomic statement and they are **not** atomic with respect to one another; the saga pattern with compensation makes cross-account transfers consistent (no funds lost or created) without distributed transactions. Idempotency keys tracked per transfer prevent double-processing on consumer redelivery. `acks=all` ensures Kafka events are not lost on broker failure.

**See also:** [`docs/04-architecture.md#32-saga-pattern-for-transfers`](../04-architecture.md#32-saga-pattern-for-transfers), [`docs/05-data-model.md`](../05-data-model.md), [`Cassandra`](../08-tech-stack/backend/cassandra.md), [`Kafka`](../08-tech-stack/backend/kafka.md).

---

## Security

| ID      | Requirement                                                                                                                                                                          | Target |
| ------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------ | ------ |
| NFR-030 | All API endpoints (except register, login, health) shall require valid JWT.                                                                                                          | Must   |
| NFR-031 | Passwords shall be hashed with Argon2id (memory=64MB, iterations=3, parallelism=4).                                                                                                  | Must   |
| NFR-032 | JWT shall be signed with RS256; private key rotated every 90 days.                                                                                                                   | Must   |
| NFR-033 | Refresh tokens shall be hashed in the database (SHA-256).                                                                                                                            | Must   |
| NFR-034 | All HTTP traffic shall use TLS 1.2+ (enforced by load balancer).                                                                                                                     | Must   |
| NFR-035 | CORS shall be restricted to the known frontend origin.                                                                                                                               | Must   |
| NFR-036 | Rate limiting: 100 requests/minute per IP for auth endpoints plus account lockout after 5 failed login attempts within 15 minutes; 1000 requests/minute for authenticated endpoints. | Must   |
| NFR-037 | Audit logs shall record all authentication and financial events.                                                                                                                     | Must   |

**Rationale:** JWT (RS256) enables stateless authentication across microservices. Short access tokens (15 min) limit exposure; refresh tokens (7 day) enable long sessions. Argon2id is memory-hard, resisting GPU-based cracking. Refresh tokens are hashed at rest so a database leak doesn't compromise them. Rate limiting mitigates brute-force and abuse. Audit logs provide traceability for security incidents.

**See also:** [`docs/04-architecture.md#9-security-boundaries`](../04-architecture.md#9-security-boundaries), [`JWT`](../08-tech-stack/backend/jwt.md), [`docs/06-api-specification.md#5-error-response-format`](../06-api-specification.md#5-error-response-format).

---

## Scalability

| ID      | Requirement                                                                             | Target |
| ------- | --------------------------------------------------------------------------------------- | ------ |
| NFR-040 | The system shall scale horizontally by adding backend instances behind a load balancer. | Must   |
| NFR-041 | Cassandra shall scale by adding nodes; partition strategy by account_id.                | Must   |
| NFR-042 | Kafka shall scale by adding partitions to the transfers topic.                          | Must   |
| NFR-043 | Consumer workers shall scale independently based on lag metrics.                        | Must   |

**Rationale:** All services are stateless (state in private DB/Kafka), enabling horizontal scaling behind a load balancer. Cassandra partitions by `account_id` keep per-account data co-located for efficient reads. Kafka partitions by source account ID preserve per-account transfer ordering while allowing consumer parallelism. HPA scales consumer workers based on Kafka lag metrics.

**See also:** [`docs/04-architecture.md#10-scalability-considerations`](../04-architecture.md#10-scalability-considerations), [`docs/11-deployment.md`](../11-deployment.md).

---

## Observability

| ID      | Requirement                                                                                     | Target |
| ------- | ----------------------------------------------------------------------------------------------- | ------ |
| NFR-050 | Structured JSON logs for all requests with correlation IDs.                                     | Must   |
| NFR-051 | Metrics: request latency (histogram), error rate, active users, transfer throughput, Kafka lag. | Must   |
| NFR-052 | Distributed tracing for transfer flow (initiation → Kafka → consumer → DB).                     | Should |
| NFR-053 | Alerts: error rate > 1%, p99 latency > 2s, Kafka consumer lag > 1000, disk usage > 80%.         | Must   |

**Rationale:** Correlation IDs (`X-Request-ID`, `X-Correlation-ID`) trace requests end-to-end across service boundaries. Structured JSON logs are queryable by log aggregation systems. Prometheus metrics feed dashboards and alerts. OpenTelemetry distributed tracing visualizes the full transfer saga flow. Alerts trigger on error rate, latency, consumer lag, and disk usage thresholds.

**See also:** [`docs/04-architecture.md#10-scalability-considerations`](../04-architecture.md#10-scalability-considerations), [`Observability`](../08-tech-stack/backend/observability.md), [`docs/12-operations/monitoring-alerting.md`](../12-operations/monitoring-alerting.md).

---

## Compliance and Operational

| ID      | Requirement                                                                                                                   | Target |
| ------- | ----------------------------------------------------------------------------------------------------------------------------- | ------ |
| NFR-060 | Database schema changes shall be versioned migrations applied in order.                                                       | Must   |
| NFR-061 | Infrastructure shall be defined as Terraform code for AWS (EKS, MSK, Cassandra on Keyspaces) and GCP (GKE, Cassandra on GKE). | Must   |
| NFR-062 | Docker images shall be built for linux/amd64 and linux/arm64.                                                                 | Should |
| NFR-063 | CI pipeline shall run lint, typecheck, unit tests, integration tests, and security scans.                                     | Must   |
| NFR-064 | Deployment shall support blue-green or rolling strategy with health checks.                                                   | Must   |

**Rationale:** Versioned CQL migrations ensure schema changes are reviewed, tested, and reproducible. Terraform provides infrastructure as code for consistent environments: AWS uses EKS + MSK + Cassandra on Keyspaces, and GCP uses GKE with Cassandra on GKE. Multi-arch Docker images support both x86 and ARM deployments. CI quality gates prevent regressions. Rolling/blue-green deployments with health checks ensure zero-downtime releases.

**See also:** [`docs/08-tech-stack/infrastructure/terraform.md`](../08-tech-stack/infrastructure/terraform.md), [`docs/08-tech-stack/infrastructure/docker.md`](../08-tech-stack/infrastructure/docker.md), [`docs/08-tech-stack/infrastructure/github-actions.md`](../08-tech-stack/infrastructure/github-actions.md), [`docs/11-deployment/ci-cd-pipeline.md`](../11-deployment/ci-cd-pipeline.md).

---

## Requirements Traceability

| Requirement  | User Story     | Feature      | Architecture Component       | Test             |
| ------------ | -------------- | ------------ | ---------------------------- | ---------------- |
| NFR-001..006 | US-001..US-008 | F-001..F-005 | All                          | load_test        |
| NFR-010..013 | US-013         | —            | All                          | —                |
| NFR-020..024 | US-012         | F-007        | Transfer Consumer, Cassandra | test_idempotency |
| NFR-030..036 | US-009         | F-006        | Auth Middleware, Gateway     | test_auth        |
| NFR-037      | US-014         | F-006        | All Services, Cassandra      | test_audit       |
| NFR-040..043 | —              | —            | Gateway, Cassandra, Kafka    | —                |
| NFR-050..053 | US-013         | —            | All                          | —                |
| NFR-060..064 | —              | —            | CI/CD, Terraform             | —                |

**Cross-references:**

- Priority labels: `Must` = required for MVP; `Should` = required for v1.1 (shown in the Target column where the requirement is a compliance level rather than a numeric target).
- Related functional requirements: [`FR-001..FR-052`](functional-requirements.md) (IDs are grouped by capability, so the gaps between blocks are intentional).
- Architecture overview: [`docs/04-architecture.md`](../04-architecture.md).
- Deployment: [`docs/11-deployment.md`](../11-deployment.md), [`docs/12-operations.md`](../12-operations.md).