# Rollback Strategy

## Purpose

Defines how NEO Bank rolls back failed deployments across API, database, and Kafka consumer layers. Rollback restores the last known-good state quickly and safely.

## 1. API Rollback

- `kubectl rollout undo deployment/api` (or re-apply the previous image digest).
- Readiness probes gate the rollout; a failing probe stops the progression before traffic shifts.
- Health checks ensure the new version is healthy before it receives traffic.
- Blue-green via a Service selector switch when a change needs an atomic cutover.
- Rollback triggered by: error rate > 1% or p99 latency > 2s sustained for 5 minutes (the critical alert thresholds in [`monitoring-alerting.md`](./monitoring-alerting.md)), or any critical alert. The decision to roll back and the rollback itself must be completed within 10 minutes of the trigger.

## 2. Database Rollback

- Use Cassandra snapshots for point-in-time recovery.
- Migrations are additive and idempotent; roll forward rather than roll back schema changes.
- Test rollback procedure regularly.
- Document rollback steps for each migration.

### When to roll back database changes

- Schema migration fails or causes data corruption.
- New queries cause performance regression.
- Application is incompatible with the new schema.

### Database rollback procedure

1. Stop the API deployment (scale to 0).
2. Restore Cassandra from the pre-migration snapshot.
3. Re-apply the previous application version.
4. Scale the API back up and verify health.

## 3. Kafka Rollback

- Consumers are idempotent; can reprocess messages after rollback.
- Use Kafka consumer group reset if needed to reprocess from a known offset.
- Transfer state is in Cassandra; reprocessing is safe due to idempotency checks.

### Kafka rollback procedure

1. Pause the consumer deployment.
2. Reset consumer group offsets to the last known-good position.
3. Re-deploy the previous consumer version.
4. Resume consumption; verify idempotent processing.

## 4. Rollback Triggers

Alerts fire at the 5-minute mark (see [`monitoring-alerting.md`](./monitoring-alerting.md)); the on-call engineer has up to 10 minutes from the trigger to decide and complete the rollback.

| Condition | Action |
|-----------|--------|
| Error rate > 1% for 5 minutes | Rollback API |
| p99 latency > 2s for 5 minutes | Rollback API |
| Any critical alert fires | Rollback affected component |
| Failed health check | Rollback to previous revision |

## 5. Rollback Verification

- Health checks pass after rollback (`/healthz`, `/readyz`).
- Error rate returns to baseline (< 0.1%).
- Latency returns to baseline (< 200 ms p95 for transfer initiation).
- Consumer lag decreases to normal levels.
- Smoke tests pass (register, login, balance, transfer).

## Cross-references

- Disaster recovery: [`disaster-recovery.md`](./disaster-recovery.md)
- Backup and recovery: [`backup-recovery.md`](./backup-recovery.md)
- Kubernetes rollouts: [`docs/08-tech-stack/infrastructure/kubernetes.md`](../08-tech-stack/infrastructure/kubernetes.md#8-rollouts--rollback)
- Deployment: [`docs/11-deployment.md`](../11-deployment.md)
- Operations index: [`12-operations.md`](../12-operations.md)
- Planned test: `test_operations_rollback` (verifies rollback procedure is documented)