# Disaster Recovery

## Purpose

Defines the disaster recovery (DR) strategy for NEO Bank: RPO/RTO targets, recovery procedure, and testing cadence. DR ensures the system can be restored after a regional outage or catastrophic failure.

## 1. RPO/RTO

| Objective | Target | Basis |
|-----------|--------|-------|
| Recovery Point Objective (RPO) | 1 hour | Backup frequency; maximum acceptable data loss |
| Recovery Time Objective (RTO) | 4 hours | Time from disaster declaration to service restoration |

## 2. DR Procedure

1. **Detect disaster** — Alerting or monitoring detects regional outage or data loss.
2. **Declare disaster** — Incident commander declares a DR event (decision point).
3. **Restore from backup** — Restore Cassandra from the most recent snapshot (PITR for managed services).
4. **Deploy application** — Deploy the application to the DR region using Terraform and Kubernetes manifests.
5. **Update DNS** — Point DNS to the DR region (ALB / Cloud Load Balancing).
6. **Verify application health** — Run health checks (`/healthz`, `/readyz`) and smoke tests.
7. **Communicate with stakeholders** — Notify users and internal teams of status and resolution.

## 3. DR Testing

- Test DR procedure quarterly.
- Document lessons learned from each test.
- Update DR procedure based on test results.
- Simulate regional outage by shutting down the primary region and failing over.

## 4. Data Protection

- Cassandra snapshots taken regularly (hourly snapshots plus continuous PITR), bounding data loss to ≤ 1 hour (the RPO).
- AWS: Amazon Keyspaces with point-in-time recovery (PITR) enabled, 35-day retention.
- GCP: Cassandra on GKE with hourly `nodetool snapshot` jobs exported to Cloud Storage for long-term retention.
- Kafka is the event backbone (`transfers`, DLQ `transfers.dlq`); topics are ephemeral and transfer state is reconstructed from Cassandra on recovery.
- Backup encryption at rest; access restricted to authorised personnel.
- Financial integrity checks (no negative balances, no ledger imbalance) run as part of restore verification.

## 5. Runbook Locations

- Backup procedure: [`backup-recovery.md`](./backup-recovery.md)
- Rollback procedure: [`rollback-strategy.md`](./rollback-strategy.md)
- Monitoring alerts: [`monitoring-alerting.md`](./monitoring-alerting.md)
- Deployment: [`docs/11-deployment.md`](../11-deployment.md)

## Cross-references

- Backup and recovery: [`backup-recovery.md`](./backup-recovery.md)
- Rollback strategy: [`rollback-strategy.md`](./rollback-strategy.md)
- Deployment: [`docs/11-deployment.md`](../11-deployment.md)
- Operations index: [`12-operations.md`](../12-operations.md)
- Planned test: `test_operations_dr` (verifies DR procedure is documented and tested)