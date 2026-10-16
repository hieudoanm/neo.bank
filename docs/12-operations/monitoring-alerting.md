# Monitoring and Alerting

## Purpose

Defines the metrics, logging, and alerting strategy for NEO Bank in production. Observability enables rapid detection and response to issues affecting financial operations.

## 1. Metrics

### API metrics

| Metric | Type | Description |
|--------|------|-------------|
| Request latency | Histogram | API endpoint latency by method and path |
| Request rate | Counter | Requests per second by endpoint |
| Error rate | Counter | Errors per second by status code |
| Active connections | Gauge | Concurrent connections to the API |

### Consumer metrics

| Metric | Type | Description |
|--------|------|-------------|
| Kafka consumer lag | Gauge | Messages behind the latest offset |
| Transfer processing rate | Counter | Transfers processed per second |
| Transfer success rate | Counter | Transfers completed successfully |
| Transfer failure rate | Counter | Transfers that failed |

### Infrastructure metrics

| Metric | Type | Description |
|--------|------|-------------|
| CPU usage | Gauge | Pod CPU utilisation |
| Memory usage | Gauge | Pod memory utilisation |
| Disk usage | Gauge | Node disk utilisation |
| Network I/O | Counter | Network bytes in/out |

## 2. SLOs and Error Budgets

### Service level objectives

Targets align with the non-functional requirements (NFR-001–005, NFR-010):

| SLI | SLO | Window |
|-----|-----|--------|
| Availability | 99.9% of successful requests | Monthly |
| Registration latency (p95) | < 500 ms | Monthly |
| Login latency (p95) | < 300 ms | Monthly |
| Balance inquiry latency (p95) | < 100 ms | Monthly |
| Transfer initiation latency (p95) | < 200 ms | Monthly |
| Transfer processing latency (p95) | < 2 s | Monthly |

### Error budget

- 99.9% availability yields an error budget of **0.1%** per month (~43 minutes of allowed unavailability).
- Error budget is consumed by failed requests and by time below the latency SLOs.
- Fast burn: > 2% of the monthly budget consumed within 1 hour pages on-call (critical).
- Slow burn: > 10% of the monthly budget consumed within 24 hours alerts the team (warning).

### Error-budget policy

- When more than 50% of the budget is consumed before mid-month, freeze non-essential feature deploys and prioritise reliability work.
- When the budget is exhausted, halt feature deploys and dedicate the team to restoring reliability until the budget recovers.
- Any budget exhaustion triggers an incident review and recorded follow-up actions.

## 3. Logging

### Structured JSON logs

All services emit structured JSON to stdout/stderr:

```json
{
  "timestamp": "2026-10-15T10:30:00Z",
  "level": "info",
  "requestId": "uuid",
  "operation": "transfer_initiated",
  "userId": "uuid",
  "accountId": "uuid",
  "transferId": "uuid",
  "amountCents": 5000
}
```

### Log fields

| Field | Description |
|-------|-------------|
| `timestamp` | ISO 8601 UTC timestamp |
| `level` | Log level (debug, info, warn, error, fatal) |
| `requestId` | Correlation ID for request tracing |
| `operation` | Business operation being performed |
| `userId` | Authenticated user ID |
| `accountId` | Account ID involved |
| `transferId` | Transfer ID (for transfer operations) |
| `amountCents` | Amount in minor units (financial operations) |

### Log aggregation

- AWS: CloudWatch Logs
- GCP: Cloud Logging
- Logs retained for 30 days (hot), 1 year (cold archive).
- Never log secrets, credentials, or raw financial data.

## 4. Alerting

### Alert rules

| Alert | Condition | Severity |
|-------|-----------|----------|
| High error rate | Error rate > 1% for 5 minutes | Critical |
| High p99 latency | p99 latency > 2s for 5 minutes | Critical |
| Kafka consumer lag | Consumer lag > 1000 for 5 minutes | Warning |
| Disk usage | Disk usage > 80% | Warning |
| High CPU usage | CPU usage > 90% for 5 minutes | Warning |
| High memory usage | Memory usage > 90% for 5 minutes | Warning |
| Negative balance detected | Any account balance < 0 (integer cents) | Critical |
| Reconciliation mismatch | Ledger/transaction imbalance or reconciliation mismatch detected | Critical |
| DLQ growth | `transfers.dlq` message count > 0 and increasing for 5 minutes | Warning |
| DLQ backlog | `transfers.dlq` message count > 100 for 10 minutes | Critical |
| Backup failure | Scheduled backup or snapshot fails | Warning |
| Backup schedule miss | Expected backup/PITR window missed | Critical |
| Error budget fast burn | > 2% of monthly budget consumed in 1 hour | Critical |

### Notification channels

- Email: non-critical alerts and daily summaries.
- Slack (webhook): warning and critical alerts.
- PagerDuty: critical alerts requiring immediate response.

### Alert routing

- Critical alerts page on-call via PagerDuty.
- Warning alerts go to Slack channel.
- Info-level events logged but not alerted.

## 5. Dashboards

- API dashboard: latency, error rate, request rate, active connections.
- Consumer dashboard: processing rate, success/failure rates, lag.
- Infrastructure dashboard: CPU, memory, disk, network per pod/node.
- Business dashboard: transfers completed, revenue volume, active users.

## 6. Health Checks

- `GET /healthz` — liveness probe (process health).
- `GET /readyz` — readiness probe (dependencies: Cassandra, Kafka).
- Probes configured in Kubernetes; failures trigger pod restarts or traffic removal.

## Cross-references

- Observability phase: [`docs/10-implementation-plan/phase-9-health-observability.md`](../10-implementation-plan/phase-9-health-observability.md)
- Kubernetes probes: [`docs/08-tech-stack/infrastructure/kubernetes.md`](../08-tech-stack/infrastructure/kubernetes.md#3-deployments--probes)
- Backend observability: [`docs/08-tech-stack/backend/observability.md`](../08-tech-stack/backend/observability.md)
- Operations index: [`12-operations.md`](../12-operations.md)
- Planned test: `test_operations_monitoring` (verifies monitoring config is complete)